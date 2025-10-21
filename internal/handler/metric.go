package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/mailru/easyjson"
	"go.uber.org/zap"

	"github.com/htrandev/metrics/internal/model"
)

type Service interface {
	Get(ctx context.Context, name string) (model.Metric, error)
	GetAll(ctx context.Context) ([]model.Metric, error)
	Store(ctx context.Context, metric *model.Metric) error
	Ping(ctx context.Context) error
}

type MetricHandler struct {
	service Service
	logger  *zap.Logger
}

func NewMetricsHandler(
	l *zap.Logger,
	s Service,
) *MetricHandler {
	return &MetricHandler{
		logger:  l,
		service: s,
	}
}

func (h *MetricHandler) Get(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	metricType := r.PathValue("metricType")
	metricName := r.PathValue("metricName")

	mt := model.ParseMetricType(metricType)
	if mt == model.TypeUnknown {
		h.logger.Error("got unknown metric type", zap.String("type", metricType), zap.String("scope", "handler/Get"))
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	metric, err := h.service.Get(ctx, metricName)
	if err != nil {
		h.logger.Error("get from storage", zap.Error(err), zap.String("scope", "handler/Get"))
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(metric.Value.String()))
}

func (h *MetricHandler) GetAll(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	metrics, err := h.service.GetAll(ctx)
	if err != nil {
		h.logger.Error("get all metrics", zap.Error(err), zap.String("scope", "handler/GetAll"))
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	var builder strings.Builder
	for _, metric := range metrics {
		builder.Grow(len(metric.Name) + len(metric.Value.String()) + 2)
		builder.WriteString(metric.Name)
		builder.WriteString(": ")
		builder.WriteString(metric.Value.String())
		builder.WriteString("\r")
	}

	rw.Header().Set("Content-Type", "text/html")
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(builder.String()))
}

func (h *MetricHandler) Update(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	metricType := r.PathValue("metricType")
	metricName := r.PathValue("metricName")
	metricValue := r.PathValue("metricValue")

	if metricName == "" {
		h.logger.Error("got empty metric name", zap.String("scope", "handler/Update"))
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	mt := model.ParseMetricType(metricType)
	if mt == model.TypeUnknown {
		h.logger.Error("got unknown metric type", zap.String("type", metricType), zap.String("scope", "handler/Update"))
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	metric := &model.Metric{
		Name: metricName,
		Value: model.MetricValue{
			Type: mt,
		},
	}
	if err := metric.SetValue(metricValue); err != nil {
		h.logger.Error("set value",
			zap.String("value", metricValue),
			zap.String("name", metric.Name),
			zap.Error(err),
			zap.String("scope", "handler/Update"))
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.service.Store(ctx, metric); err != nil {
		h.logger.Error("store error", zap.Error(err), zap.String("scope", "handler/Update"))
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	h.logger.Info("successfully store", zap.Any("metric", metric), zap.String("scope", "handler/Update"))

	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusOK)
}

func (h *MetricHandler) UpdateJSON(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	m, err := buildUpdateRequest(r)
	if err != nil {
		h.logger.Error("build request error", zap.Error(err), zap.String("scope", "handler/UpdateJSON"))
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	if m == nil {
		h.logger.Error("request is nil", zap.String("scope", "handler/UpdateJSON"))
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	if m.Name == "" {
		h.logger.Error("got empty metric name", zap.String("scope", "handler/UpdateJSON"))
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	if err := h.service.Store(ctx, m); err != nil {
		h.logger.Error("store error", zap.Error(err), zap.String("scope", "handler/UpdateJSON"))
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	h.logger.Info("successfully stored")

	response := buildResponse(*m)

	body, err := easyjson.Marshal(response)
	if err != nil {
		h.logger.Error("get from storage", zap.Error(err), zap.String("scope", "handler/UpdateJSON"))
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	rw.Write(body)

}

func (h *MetricHandler) GetJSON(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := buildGetRequest(r)
	if err != nil {
		h.logger.Error("store error", zap.Error(err), zap.String("scope", "handler/GetJSON"))
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		h.logger.Error("got empty metric name", zap.String("scope", "handler/GetJSON"))
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	m, err := h.service.Get(ctx, req.Name)
	if err != nil {
		h.logger.Error("get from storage", zap.Error(err), zap.String("scope", "handler/GetJSON"))
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	response := buildResponse(m)

	body, err := easyjson.Marshal(response)
	if err != nil {
		h.logger.Error("get from storage", zap.Error(err), zap.String("scope", "handler/GetJSON"))
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	rw.Write(body)
}

func (h *MetricHandler) Ping(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := h.service.Ping(ctx); err != nil {
		h.logger.Error("ping", zap.Error(err), zap.String("scope", "handler/Ping"))
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func buildUpdateRequest(r *http.Request) (*model.Metric, error) {
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
		if req.Value != nil {
			m.Value.Gauge = *req.Value
		} else {
			return nil, fmt.Errorf("value for metric is nil")
		}
	case model.TypeCounter.String():
		m.Value.Type = model.TypeCounter
		if req.Delta != nil {
			m.Value.Counter = *req.Delta
		} else {
			return nil, fmt.Errorf("value for metric is nil")
		}
	default:
		return nil, fmt.Errorf("unknown metric type: %s", req.MType)
	}

	return m, nil
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
