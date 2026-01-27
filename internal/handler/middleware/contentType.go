package middleware

import "net/http"

// ContentType возвращает HTTP middleware для проверки Content-Type запроса.
func ContentType() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rw := &responseWriter{
				ResponseWriter: w,
			}

			contentType := r.Header.Get("Content-Type")
			if contentType != "application/json" {
				rw.WriteHeader(http.StatusNotAcceptable)
				return
			}

			next.ServeHTTP(rw, r)
		})
	}
}
