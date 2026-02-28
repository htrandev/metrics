package handler

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	"github.com/htrandev/metrics/internal/audit"
	"github.com/htrandev/metrics/internal/contracts"
)

// Publisher предоставляет интерфейс публикации событий.
type Publisher interface {
	Update(ctx context.Context, info audit.AuditInfo)
}

// MetricHandler определяет обработчика запросов для работы с метриками.
type MetricHandler struct {
	service   contracts.Service
	logger    *zap.Logger
	Publisher Publisher
}

// NewMetricsHandler возвращает новый экземпляр MetricsHandler.
func NewMetricsHandler(
	l *zap.Logger,
	s contracts.Service,
	p Publisher,
) *MetricHandler {
	return &MetricHandler{
		logger:    l,
		service:   s,
		Publisher: p,
	}
}

// Ping обрабатывает HTTP-запрос /ping для проверки доступности сервиса.
func (h *MetricHandler) Ping(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := h.service.Ping(ctx); err != nil {
		h.logger.Error("ping", zap.Error(err), zap.String("scope", "handler/Ping"))
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
}
