package handler

import (
	"net"
	"net/http"

	"github.com/mailru/easyjson"
	"go.uber.org/zap"

	"github.com/htrandev/metrics/internal/model"
)

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

	info := buildAuditInfoMessage(m, getIP(r))
	h.Publisher.Update(ctx, info)

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
}

func getIP(r *http.Request) string {
	ip := r.Header.Get("X-Real-Ip")
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
	}

	if ip == "" {
		ip = r.RemoteAddr
	}

	host, _, err := net.SplitHostPort(ip)
	if err != nil {
		return ip
	}
	return host
}
