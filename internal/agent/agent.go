package agent

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/htrandev/metrics/internal/model"
)

type Agent struct {
	client *resty.Client
	addr   string
}

func New(addr string) *Agent {
	client := resty.New().
		SetTimeout(30 * time.Second)

	return &Agent{
		client: client,
		addr:   addr,
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
