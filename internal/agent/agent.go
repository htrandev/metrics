package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/mailru/easyjson"

	"github.com/htrandev/metrics/internal/model"
)

type Agent struct {
	client *resty.Client
	url    string
}

func New(addr string) *Agent {
	client := resty.New().
		SetTimeout(30 * time.Second)
	url := buildURL(addr)

	return &Agent{
		client: client,
		url:    url,
	}
}

func (a *Agent) SendMetric(ctx context.Context, metric model.Metric) error {
	req := buildRequest(metric)
	body, err := buildBody(req)
	if err != nil {
		return fmt.Errorf("build body: %w", err)
	}

	_, err = a.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetBody(body).
		SetContext(ctx).
		Post(a.url)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("post: %w", err)
	}
	return nil
}

func buildRequest(metric model.Metric) model.Metrics {
	m := model.Metrics{
		ID:    metric.Name,
		MType: metric.Value.Type.String(),
	}

	switch metric.Value.Type {
	case model.TypeGauge:
		m.Value = &metric.Value.Gauge
	case model.TypeCounter:
		m.Delta = &metric.Value.Counter
	}

	return m
}

func buildBody(m model.Metrics) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	p, err := easyjson.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("can't marshal metrics: %w", err)
	}
	_, err = gz.Write(p)
	if err != nil {
		return nil, fmt.Errorf("can't write: %w", err)
	}
	gz.Close()
	return buf.Bytes(), nil
}

func buildURL(addr string) string {
	u := url.URL{
		Scheme: "http",
		Host:   addr,
		Path:   "/update/",
	}

	return u.String()
}
