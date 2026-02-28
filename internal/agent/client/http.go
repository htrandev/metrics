package client

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/htrandev/metrics/internal/agent"
	"github.com/htrandev/metrics/internal/handler/middleware"
	"github.com/htrandev/metrics/internal/model"
	"github.com/htrandev/metrics/pkg/sign"
	"go.uber.org/zap"
)

var _ agent.Client = (*HTTPClient)(nil)

type HTTPClient struct {
	client *resty.Client
	opts   CommonOptions
}

func NewHTTP(client *resty.Client, opts ...Option) *HTTPClient {
	c := &HTTPClient{
		client: client,
	}
	for _, opt := range opts {
		opt(&c.opts)
	}
	return c
}

func (c *HTTPClient) Send(ctx context.Context, metrics []model.MetricDto) error {
	return c.SendManyWithRetry(ctx, metrics)
}

// SendSingleMetric отправляет за раз одну метрику на сервер.
func (c *HTTPClient) SendSingleMetric(ctx context.Context, metric model.MetricDto) error {
	req := buildSingleRequest(metric)
	body, err := buildSingleBody(req)
	if err != nil {
		return fmt.Errorf("build body: %w", err)
	}
	url := buildSingleURL(c.opts.addr)

	_, err = c.client.R().
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

// SendManyMetrics отправляет за раз несколько метрик на сервер.
func (c *HTTPClient) SendManyMetrics(ctx context.Context, metrics []model.MetricDto) error {
	if len(metrics) == 0 {
		return nil
	}

	req := buildManyRequest(metrics)
	body, err := c.buildManyBody(req)
	if err != nil {
		return fmt.Errorf("build body: %w", err)
	}

	url := buildManyURL(c.opts.addr)

	r := c.client.R().
		SetHeader(middleware.IPHeader, c.opts.ip).
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetBody(body).
		SetContext(ctx)

	if c.opts.signature != "" {
		s := sign.Signature(c.opts.signature)
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

// SendManyWithRetry отправляет за раз несколько метрик на сервер,
// если произошла ошибка, пытается повторно отправить указанное в MaxRetry количество раз.
func (c *HTTPClient) SendManyWithRetry(ctx context.Context, metrics []model.MetricDto) error {
	err := c.SendManyMetrics(ctx, metrics)

	if err != nil {
		c.opts.logger.Error("send many metrics", zap.Error(err), zap.String("scope", "agent/sendWithRetry"))
		c.opts.logger.Info("try to resend metrics", zap.String("scope", "agent/sendWithRetry"))

		for i := 0; i < c.opts.maxRetry; i++ {
			c.opts.logger.Debug("", zap.Int("retry", i+1))
			retryDelay := i*2 + 1
			time.Sleep(time.Second * time.Duration(retryDelay))
			err := c.SendManyMetrics(ctx, metrics)
			if err != nil {
				c.opts.logger.Error("send many metrics", zap.Int("retry", i+1), zap.Error(err), zap.String("scope", "agent/sendWithRetry"))
				continue
			}
			return nil
		}
		return fmt.Errorf("send many with retries: %w", err)
	}
	return nil
}
