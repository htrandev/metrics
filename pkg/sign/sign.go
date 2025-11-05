package sign

import (
	"crypto/hmac"
	"crypto/sha256"
)

type Signature []byte

func (s Signature) Sign(data []byte) []byte {
	h := hmac.New(sha256.New, s)
	h.Write(data)
	return h.Sum(nil)
}
