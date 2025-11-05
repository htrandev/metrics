package agent

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"

	"github.com/htrandev/metrics/internal/model"
	"github.com/htrandev/metrics/pkg/sign"
)

var defaultOpts = &AgentOptions{
	Addr:     "localhost:8080",
	Key:      "",
	MaxRetry: 3,
	Logger:   zap.NewNop(),
}

type AgentOptions struct {
	Addr     string
	Key      string
	MaxRetry int

	Client *resty.Client
	Logger *zap.Logger
}

type Agent struct {
	opts *AgentOptions
}

func New(opts *AgentOptions) *Agent {
	if opts == nil {
		opts = defaultOpts
	}

	if opts.MaxRetry <= 0 {
		opts.MaxRetry = 3
	}
	if opts.Logger == nil {
		opts.Logger = zap.NewNop()
	}

	return &Agent{opts: opts}
}

func (a *Agent) SendSingleMetric(ctx context.Context, metric model.Metric) error {
	req := buildSingleRequest(metric)
	body, err := buildSingleBody(req)
	if err != nil {
		return fmt.Errorf("build body: %w", err)
	}
	url := buildSingleURL(a.opts.Addr)

	_, err = a.opts.Client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetBody(body).
		SetContext(ctx).
		Post(url)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("post: %w", err)
	}
	return nil
}

func (a *Agent) SendManyMetrics(ctx context.Context, metrics []model.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	req := buildManyRequest(metrics)
	body, err := buildManyBody(req)
	if err != nil {
		return fmt.Errorf("build body: %w", err)
	}

	url := buildManyURL(a.opts.Addr)

	r := a.opts.Client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetBody(body).
		SetContext(ctx)

	if a.opts.Key != "" {
		s := sign.Signature(a.opts.Key)
		signature := s.Sign(body)
		hash := base64.RawURLEncoding.EncodeToString(signature)
		r.SetHeader("HashSHA256", hash)
	}

	_, err = r.Post(url)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("post: %w", err)
	}
	return nil
}

func (a *Agent) SendManyWithRetry(ctx context.Context, metrics []model.Metric) error {
	err := a.SendManyMetrics(ctx, metrics)

	if err != nil {
		a.opts.Logger.Error("send many metrics", zap.Error(err), zap.String("scope", "agent/sendWithRetry"))
		a.opts.Logger.Info("try to resend metrics", zap.String("scope", "agent/sendWithRetry"))

		for i := 0; i < a.opts.MaxRetry; i++ {
			a.opts.Logger.Debug("", zap.Int("retry", i+1))
			retryDelay := i*2 + 1
			time.Sleep(time.Second * time.Duration(retryDelay))
			err := a.SendManyMetrics(ctx, metrics)
			if err != nil {
				a.opts.Logger.Error("send many metrics", zap.Int("retry", i+1), zap.Error(err), zap.String("scope", "agent/sendWithRetry"))
				continue
			}
			return nil
		}
		return fmt.Errorf("send many with retries: %w", err)
	}
	return nil
}
