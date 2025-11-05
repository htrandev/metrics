package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/mailru/easyjson"
	"go.uber.org/zap"

	"github.com/htrandev/metrics/internal/model"
	"github.com/htrandev/metrics/internal/repository"
)

type Service interface {
	Get(ctx context.Context, name string) (model.Metric, error)
	GetAll(ctx context.Context) ([]model.Metric, error)

	Store(ctx context.Context, metric *model.Metric) error
	StoreMany(ctx context.Context, metric []model.Metric) error
	StoreManyWithRetry(ctx context.Context, metric []model.Metric) error

	Ping(ctx context.Context) error
}

type Signer interface {
	Sign([]byte) []byte
}

type Cipher interface {
	Encrypt([]byte) []byte
}

type MetricHandler struct {
	service Service
	logger  *zap.Logger
	cipher  Cipher
	signer  Signer
}

func NewMetricsHandler(
	l *zap.Logger,
	s Service,
	key string,
) *MetricHandler {
	return &MetricHandler{
		logger:  l,
		service: s,
	}
}

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

	scope := zap.String("scope", "handler/Update")

	metricType := r.PathValue("metricType")
	metricName := r.PathValue("metricName")
	metricValue := r.PathValue("metricValue")

	if metricName == "" {
		h.logger.Error("got empty metric name", scope)
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	mt := model.ParseMetricType(metricType)
	if mt == model.TypeUnknown {
		h.logger.Error("got unknown metric type", zap.String("type", metricType), scope)
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
			scope)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.service.Store(ctx, metric); err != nil {
		h.logger.Error("store error", zap.Error(err), scope)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	h.logger.Info("successfully store", zap.Any("metric", metric), scope)

	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusOK)
}

func (h *MetricHandler) UpdateJSON(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	scope := zap.String("scope", "handler/UpdateJSON")

	m, err := buildSingleUpdateRequest(r)
	if err != nil {
		h.logger.Error("build single update request", zap.Error(err), scope)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	if m == nil {
		h.logger.Error("request is nil", scope)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	if m.Name == "" {
		h.logger.Error("got empty metric name", scope)
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	if err := h.service.Store(ctx, m); err != nil {
		h.logger.Error("store error", zap.Error(err), scope)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	h.logger.Info("successfully stored")

	response := buildResponse(*m)

	body, err := easyjson.Marshal(response)
	if err != nil {
		h.logger.Error("get from storage", zap.Error(err), scope)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	rw.Write(body)

}

func (h *MetricHandler) UpdateManyJSON(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	scope := zap.String("scope", "handler/UpdateManyJSON")

	m, err := buildManyUpdateRequest(r)
	if err != nil {
		h.logger.Error("build many update request", zap.Error(err), scope)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(m) == 0 {
		h.logger.Debug("receive empty metrics batch", scope)
		rw.WriteHeader(http.StatusOK)
		return
	}

	if err := h.service.StoreManyWithRetry(ctx, m); err != nil {
		h.logger.Error("store many with retry", zap.Error(err), scope)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
}

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

func (h *MetricHandler) Ping(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := h.service.Ping(ctx); err != nil {
		h.logger.Error("ping", zap.Error(err), zap.String("scope", "handler/Ping"))
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
}
