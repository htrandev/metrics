package router

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/htrandev/metrics/internal/handler"
	"github.com/htrandev/metrics/internal/handler/middleware"
	"github.com/htrandev/metrics/pkg/logger"
)

func New(handler *handler.MetricHandler) (*chi.Mux, error) {
	r := chi.NewRouter()

	zl, err := logger.NewZapLogger("error")
	if err != nil {
		return nil, fmt.Errorf("init logger: %w", err)
	}

	r.With(
		middleware.MethodChecker(http.MethodGet),
		middleware.Logger(zl),
	).Get("/", handler.GetAll)

	r.With(
		middleware.MethodChecker(http.MethodGet),
		middleware.Logger(zl),
	).Get("/value/{metricType}/{metricName}", handler.Get)

	r.With(
		middleware.MethodChecker(http.MethodPost),
		middleware.Logger(zl),
	).Post("/update/{metricType}/{metricName}/{metricValue}", handler.Update)

	r.With(
		middleware.MethodChecker(http.MethodPost),
		middleware.Logger(zl),
		middleware.ContentType(),
	).Post("/update/", handler.UpdateViaBody)

	r.With(
		middleware.MethodChecker(http.MethodPost),
		middleware.Logger(zl),
		middleware.ContentType(),
	).Post("/value/", handler.GetViaBody)

	return r, nil
}
