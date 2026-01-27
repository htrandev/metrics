package middleware

import "net/http"

// MethodChecker возвращает HTTP middleware для проверки соответсвия ожидаемого и полученного метода запроса.
func MethodChecker(method string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if method != r.Method {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
