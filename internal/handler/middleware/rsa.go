package middleware

import (
	"bytes"
	"crypto/rsa"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/htrandev/metrics/pkg/crypto"
)

func RSA(key *rsa.PrivateKey, logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if key == nil {
				next.ServeHTTP(w, r)
				return
			}
			rw := &responseWriter{
				ResponseWriter: w,
			}

			var buf bytes.Buffer
			if _, err := buf.ReadFrom(r.Body); err != nil {
				logger.Error("read body",
					zap.Error(err),
					zap.String("scope", "middleware"),
					zap.String("method", "rsa"),
				)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			body, err := crypto.Decrypt(key, buf.Bytes())
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
			}

			r.Body = io.NopCloser(bytes.NewReader(body))

			next.ServeHTTP(rw, r)
		})
	}
}
