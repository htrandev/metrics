package agent

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"

	"github.com/htrandev/metrics/internal/model"
)

type Agent struct {
	client   *resty.Client
	addr     string
	maxRetry int

	logger *zap.Logger
}

func New(addr string, maxRetry int, l *zap.Logger) *Agent {
	client := resty.New().
		SetTimeout(30 * time.Second)

	if maxRetry <= 0 {
		maxRetry = 3
	}
	if l == nil {
		l = zap.NewNop()
	}

	return &Agent{
		client:   client,
		addr:     addr,
		maxRetry: maxRetry,
		logger:   l,
	}
}

func (a *Agent) SendSingleMetric(ctx context.Context, metric model.Metric) error {
	req := buildSingleRequest(metric)
	body, err := buildSingleBody(req)
	if err != nil {
		return fmt.Errorf("build body: %w", err)
	}
	url := buildSingleURL(a.addr)

	_, err = a.client.R().
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

	url := buildManyURL(a.addr)

	_, err = a.client.R().
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

func (a *Agent) SendManyWithRetry(ctx context.Context, metrics []model.Metric) error {
	err := a.SendManyMetrics(ctx, metrics)

	if err != nil {
		a.logger.Error("send many metrics", zap.Error(err), zap.String("scope", "agent/sendWithRetry"))
		a.logger.Info("try to resend metrics", zap.String("scope", "agent/sendWithRetry"))

		for i := 0; i < a.maxRetry; i++ {
			a.logger.Debug("", zap.Int("retry", i+1))
			retryDelay := i*2 + 1
			time.Sleep(time.Second * time.Duration(retryDelay))
			err := a.SendManyMetrics(ctx, metrics)
			if err != nil {
				a.logger.Error("send many metrics", zap.Int("retry", i+1), zap.Error(err), zap.String("scope", "agent/sendWithRetry"))
				continue
			}
			return nil
		}
		return fmt.Errorf("send many with retries: %w", err)
	}
	return nil
}
