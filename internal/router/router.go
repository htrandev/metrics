package router

import (
	"crypto/rsa"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/htrandev/metrics/internal/handler"
	"github.com/htrandev/metrics/internal/handler/middleware"
)

type RouterOptions struct {
	Signature string
	Subnet    *net.IPNet
	Key       *rsa.PrivateKey
	Logger    *zap.Logger
	Handler   *handler.MetricHandler
}

// New возвращает новый экземляр роутера.
//
// Эндпоинты:
//   - GET    / - получить все метрики
//   - GET    /value/{{metricType}/{metricName} - получить значение метрики
//   - POST   /update/{metricType}/{metricName}/{metricValue} - обновить метрику
//   - POST   /update/ - обновить метрику в формате JSON
//   - GET    /value/ - получить значение метрики в формате JSON
//   - GET    /ping - проверка доступности БД
//   - POST   /updates/ - обновить несколько метрик в формате JSON
func New(opts RouterOptions) *chi.Mux {
	r := chi.NewRouter()

	var (
		getMethodChecker  = middleware.MethodChecker(http.MethodGet)
		postMethodChecker = middleware.MethodChecker(http.MethodPost)

		l = middleware.Logger(opts.Logger)

		ct = middleware.ContentType()

		signer = middleware.Sign(opts.Signature, opts.Logger)

		compressor = middleware.Compress(opts.Logger)

		rsa = middleware.RSA(opts.Key, opts.Logger)
	)

	r.With(getMethodChecker, l, signer, compressor).
		Get("/", opts.Handler.GetAll)

	r.With(getMethodChecker, l, signer).
		Get("/value/{metricType}/{metricName}", opts.Handler.Get)

	r.With(postMethodChecker, l, signer).
		Post("/update/{metricType}/{metricName}/{metricValue}", opts.Handler.Update)

	r.With(postMethodChecker, l, ct, signer, compressor).
		Post("/update/", opts.Handler.UpdateJSON)

	r.With(postMethodChecker, l, ct, signer, compressor).
		Post("/value/", opts.Handler.GetJSON)

	r.With(getMethodChecker).
		Get("/ping", opts.Handler.Ping)

	middlewares := make([]func(http.Handler) http.Handler, 0, 7)
	middlewares = append(middlewares, postMethodChecker, l, ct, rsa, signer, compressor)
	if opts.Subnet != nil {
		middlewares = append(middlewares, middleware.Subnet(opts.Subnet))
	}
	r.With(middlewares...).
		Post("/updates/", opts.Handler.UpdateManyJSON)

	return r
}
