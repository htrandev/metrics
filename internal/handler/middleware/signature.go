package middleware

import (
	"bytes"
	"encoding/base64"
	"io"
	"net/http"

	"github.com/htrandev/metrics/pkg/sign"
)

func Sign(key string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rw := &responseWriter{
				ResponseWriter: w,
			}

			if len(key) == 0 {
				next.ServeHTTP(rw, r)
				return
			}

			receivedHash := r.Header.Get("HashSHA256")
			if receivedHash == "" {
				next.ServeHTTP(rw, r)
				return
			}

			var buf bytes.Buffer
			if _, err := buf.ReadFrom(r.Body); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(buf.Bytes()))

			s := sign.Signature(key)
			signature := s.Sign(buf.Bytes())
			gotHash := base64.RawURLEncoding.EncodeToString(signature)

			if receivedHash != gotHash {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			next.ServeHTTP(rw, r)
		})
	}
}
