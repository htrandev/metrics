package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

func Logger(log *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rw := &responseWriter{
				ResponseWriter: w,
			}
			start := time.Now()

			next.ServeHTTP(rw, r)

			elapsed := time.Since(start)
			log.Info("got incoming HTTP request",
				zap.String("uri", r.URL.RequestURI()),
				zap.String("method", r.Method),
				zap.Duration("elapsed", elapsed),
			)

			log.Info("sending HTTP response",
				zap.Int("status code", rw.statusCode),
				zap.Int("size", rw.bodySize),
			)
		})
	}
}
