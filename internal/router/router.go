package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/htrandev/metrics/internal/handler"
	"github.com/htrandev/metrics/internal/handler/middleware"
)

func New(logger *zap.Logger, handler *handler.MetricHandler) (*chi.Mux, error) {
	r := chi.NewRouter()

	r.With(
		middleware.MethodChecker(http.MethodGet),
		middleware.Logger(logger),
	).Get("/", handler.GetAll)

	r.With(
		middleware.MethodChecker(http.MethodGet),
		middleware.Logger(logger),
	).Get("/value/{metricType}/{metricName}", handler.Get)

	r.With(
		middleware.MethodChecker(http.MethodPost),
		middleware.Logger(logger),
	).Post("/update/{metricType}/{metricName}/{metricValue}", handler.Update)

	r.With(
		middleware.MethodChecker(http.MethodPost),
		middleware.Logger(logger),
		middleware.ContentType(),
	).Post("/update/", handler.UpdateViaBody)

	r.With(
		middleware.MethodChecker(http.MethodPost),
		middleware.Logger(logger),
		middleware.ContentType(),
	).Post("/value/", handler.GetViaBody)

	return r, nil
}
