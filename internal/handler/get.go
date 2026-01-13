package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/mailru/easyjson"
	"go.uber.org/zap"

	"github.com/htrandev/metrics/internal/model"
	"github.com/htrandev/metrics/internal/repository"
)

// Get обрабатывает HTTP GET /value/{metricType}/{metricName} для получения одной метрики.
// Возвращает значение в plain text формате. 
func (h *MetricHandler) Get(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	scope := zap.String("scope", "handler/Get")

	metricType := r.PathValue("metricType")
	metricName := r.PathValue("metricName")

	mt := model.ParseMetricType(metricType)
	if mt == model.TypeUnknown {
		h.logger.Error("got unknown metric type", zap.String("type", metricType), scope)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	metric, err := h.service.Get(ctx, metricName)
	if errors.Is(err, repository.ErrNotFound) {
		h.logger.Error("metric not found", scope)
		rw.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		h.logger.Error("get from storage", zap.Error(err), scope)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(metric.Value.String()))
}

// GetAll обрабатывает HTTP GET / для получения всех метрик.
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

// GetJSON обрабатывает HTTP POST /value/ для получения одной метрики в JSON.
// Парсит тело запроса, возвращает структурированный JSON-ответ.
func (h *MetricHandler) GetJSON(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	scope := zap.String("scope", "handler/GetJSON")

	req, err := buildGetRequest(r)
	if err != nil {
		h.logger.Error("store error", zap.Error(err), scope)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		h.logger.Error("got empty metric name", scope)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	m, err := h.service.Get(ctx, req.Name)
	if errors.Is(err, repository.ErrNotFound) {
		h.logger.Error("metric not found", zap.Error(err), scope)
		rw.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		h.logger.Error("get from storage", zap.Error(err), scope)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := buildResponse(m)

	body, err := easyjson.Marshal(response)
	if err != nil {
		h.logger.Error("marshal response", zap.Error(err), scope)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	rw.Write(body)
}
