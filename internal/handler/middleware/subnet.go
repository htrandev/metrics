package middleware

import (
	"net"
	"net/http"
)

const (
	IpHeader = "X-Real-IP"
)

func Subnet(subnet *net.IPNet) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rw := &responseWriter{
				ResponseWriter: w,
			}

			ip := net.ParseIP(r.Header.Get(IpHeader))
			if len(ip) != 0 && !subnet.Contains(ip) {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			next.ServeHTTP(rw, r)
		})
	}
}
