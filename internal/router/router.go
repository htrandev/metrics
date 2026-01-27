package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/htrandev/metrics/internal/handler"
	"github.com/htrandev/metrics/internal/handler/middleware"
)

// New возвращает новый экземляр
// 
// Эндпоинты:
//   - GET    / - получить все метрики
//   - GET    /value/{{metricType}/{metricName} - получить значение метрики
//   - POST   /update/{metricType}/{metricName}/{metricValue} - обновить метрику
//   - POST   /update/ - обновить метрику в формате JSON
//   - GET    /value/ - получить значение метрики в формате JSON
//   - GET    /ping - проверка доступности БД
//   - POST   /updates/ - обновить несколько метрик в формате JSON
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
