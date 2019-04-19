package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"hash"
)

// HMAC is a wrapper of crypto/hmac
type HMAC struct {
	hmac hash.Hash
}

// NewHMAC creates and returns new HMAC object
func NewHMAC(key string) HMAC {
	h := hmac.New(sha256.New, []byte(key))

	return HMAC{
		hmac: h,
	}
}

// Hash takes input string (and the secret key provided
// when HMAC object was created) and returns hmac for that
// input
func (h HMAC) Hash(input string) string {
	h.hmac.Reset()
	h.hmac.Write([]byte(input))
	b := h.hmac.Sum(nil)
	return base64.URLEncoding.EncodeToString(b)
}
