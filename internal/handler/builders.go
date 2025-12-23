package handler

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mailru/easyjson"

	"github.com/htrandev/metrics/internal/audit"
	"github.com/htrandev/metrics/internal/model"
)

func buildSingleUpdateRequest(r *http.Request) (*model.Metric, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("can't read body: %w", err)
	}
	defer r.Body.Close()

	var req model.Metrics
	if err := easyjson.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("can't unmarshal request: %w", err)
	}

	m, err := buildInternalMetric(req)
	if err != nil {
		return nil, fmt.Errorf("build internal metric: %w", err)
	}
	return &m, nil
}

func buildGetRequest(r *http.Request) (*model.Metric, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("can't read body: %w", err)
	}
	defer r.Body.Close()

	var req model.Metrics
	if err := easyjson.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("can't unmarshal request: %w", err)
	}

	m := &model.Metric{
		Name: req.ID,
	}

	switch req.MType {
	case model.TypeGauge.String():
		m.Value.Type = model.TypeGauge
	case model.TypeCounter.String():
		m.Value.Type = model.TypeCounter
	default:
		return nil, fmt.Errorf("unknown metric type: %s", req.MType)
	}

	return m, nil
}

func buildResponse(metric model.Metric) model.Metrics {
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

func buildManyUpdateRequest(r *http.Request) ([]model.Metric, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("can't read body: %w", err)
	}
	defer r.Body.Close()

	var req model.MetricsSlice
	if err := easyjson.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("can't unmarshal request: %w", err)
	}

	metrics := make([]model.Metric, 0, len(req))
	for _, metric := range req {
		m, err := buildInternalMetric(metric)
		if err != nil {
			return nil, fmt.Errorf("build internal metric: %w", err)
		}
		metrics = append(metrics, m)
	}

	return metrics, nil
}

func buildInternalMetric(metric model.Metrics) (model.Metric, error) {
	var m model.Metric

	switch metric.MType {
	case model.TypeGauge.String():
		if metric.Value != nil {
			m = model.Gauge(metric.ID, *metric.Value)
		} else {
			return m, fmt.Errorf("value for metric is nil")
		}
	case model.TypeCounter.String():
		if metric.Delta != nil {
			m = model.Counter(metric.ID, *metric.Delta)
		} else {
			return m, fmt.Errorf("value for metric is nil")
		}
	default:
		return m, fmt.Errorf("unknown metric type: %s", metric.MType)
	}

	return m, nil
}

func buildAuditInfoMessage(metrics []model.Metric, ip string) audit.AuditInfo {
	names := make([]string, 0, len(metrics))
	for _, metric := range metrics {
		names = append(names, metric.Name)
	}

	return audit.AuditInfo{
		Timestamp: time.Now().Unix(),
		Metrics:   names,
		IP:        ip,
	}
}
