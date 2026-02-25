package client

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net/url"

	"github.com/mailru/easyjson"

	"github.com/htrandev/metrics/internal/model"
	"github.com/htrandev/metrics/pkg/crypto"
)

func (c *HTTPClient) buildManyBody(metrics model.MetricsSlice) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	p, err := easyjson.Marshal(metrics)
	if err != nil {
		return nil, fmt.Errorf("buildManyBody: can't marshal metrics: %w", err)
	}

	if c.opts.key != nil {
		encrypted, err := crypto.Encrypt(c.opts.key, p)
		if err != nil {
			return nil, fmt.Errorf("buildManyBody: encrypt body: %w", err)
		}
		p = encrypted
	}

	_, err = gz.Write(p)
	if err != nil {
		return nil, fmt.Errorf("buildManyBody: can't write: %w", err)
	}

	gz.Close()
	return buf.Bytes(), nil
}

func buildSingleRequest(metric model.MetricDto) model.Metrics {
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

func buildManyRequest(metrics []model.MetricDto) model.MetricsSlice {
	m := make([]model.Metrics, 0, len(metrics))
	for _, metric := range metrics {
		m = append(m, buildSingleRequest(metric))
	}

	return m
}

func buildSingleBody(m model.Metrics) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	p, err := easyjson.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("buildSingleBody: can't marshal metrics: %w", err)
	}
	_, err = gz.Write(p)
	if err != nil {
		return nil, fmt.Errorf("buildSingleBody: can't write: %w", err)
	}
	gz.Close()
	return buf.Bytes(), nil
}

func buildSingleURL(addr string) string {
	u := url.URL{
		Scheme: "http",
		Host:   addr,
		Path:   "/update/",
	}

	return u.String()
}

func buildManyURL(addr string) string {
	u := url.URL{
		Scheme: "http",
		Host:   addr,
		Path:   "/updates/",
	}

	return u.String()
}
