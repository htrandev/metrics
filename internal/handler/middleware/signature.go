package middleware

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
)

type Signer interface {
	Sign([]byte) []byte
}

func Sign(s Signer) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rw := &responseWriter{
				ResponseWriter: w,
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

			hash := s.Sign(buf.Bytes())
			gotHash := base64.RawURLEncoding.EncodeToString(hash)

			fmt.Println("received", receivedHash)
			fmt.Println()
			fmt.Println("got", gotHash)
			fmt.Println(receivedHash != gotHash)
			if receivedHash != gotHash {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			next.ServeHTTP(rw, r)
		})
	}
}
