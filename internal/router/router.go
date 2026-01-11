package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/htrandev/metrics/internal/handler"
	"github.com/htrandev/metrics/internal/handler/middleware"
)

func New(key string, logger *zap.Logger, handler *handler.MetricHandler) *chi.Mux {
	r := chi.NewRouter()

	var (
		getMethodChecker  = middleware.MethodChecker(http.MethodGet)
		postMethodChecker = middleware.MethodChecker(http.MethodPost)

		l = middleware.Logger(logger)

		ct = middleware.ContentType()

		signer = middleware.Sign(key)

		compressor = middleware.Compress()
	)

	r.With(getMethodChecker, l, signer, compressor).
		Get("/", handler.GetAll)

	r.With(getMethodChecker, l, signer).
		Get("/value/{metricType}/{metricName}", handler.Get)

	r.With(postMethodChecker, l, signer).
		Post("/update/{metricType}/{metricName}/{metricValue}", handler.Update)

	r.With(postMethodChecker, l, ct, signer, compressor).
		Post("/update/", handler.UpdateJSON)

	r.With(postMethodChecker, l, ct, signer, compressor).
		Post("/value/", handler.GetJSON)

	r.With(getMethodChecker).
		Get("/ping", handler.Ping)

	r.With(postMethodChecker, l, ct, signer, compressor).
		Post("/updates/", handler.UpdateManyJSON)

	return r
}
