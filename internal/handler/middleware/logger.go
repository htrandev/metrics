package middleware

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type Log struct {
	log *zap.Logger
}

func NewLogger(level string) (*Log, error) {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, fmt.Errorf("parse log level: %w", err)
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("build config: %w", err)
	}

	return &Log{log: zl}, nil
}

func (l *Log) Logger() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rw := &responseWriter{
				ResponseWriter: w,
			}
			start := time.Now()

			next.ServeHTTP(rw, r)

			elapsed := time.Since(start)
			l.log.Info("got incoming HTTP request",
				zap.String("uri", r.URL.RequestURI()),
				zap.String("method", r.Method),
				zap.Duration("elapsed", elapsed),
			)

			l.log.Info("sending HTTP response",
				zap.Int("status code", rw.statusCode),
				zap.Int("size", rw.bodySize),
			)
		})
	}
}
