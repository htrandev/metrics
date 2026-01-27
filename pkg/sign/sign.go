package sign

import (
	"crypto/hmac"
	"crypto/sha256"
)

// Signature алиас для слайса байт.
type Signature []byte

// Sign вычисляет HMAC-SHA256 подпись для предоставленных данных.
// Использует Signature как секретный ключ.
func (s Signature) Sign(data []byte) []byte {
	h := hmac.New(sha256.New, s)
	h.Write(data)
	return h.Sum(nil)
}
