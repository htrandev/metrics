package middleware

import "net/http"

type responseWriter struct {
	http.ResponseWriter
	body       int
	statusCode int
}

func (w *responseWriter) Write(p []byte) (int, error) {
	w.body = len(p)
	return w.ResponseWriter.Write(p)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
