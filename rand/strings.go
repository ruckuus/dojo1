package rand

import (
	"crypto/rand"
	"encoding/base64"
)

const RememberTokenBytes = 32

// Bytes create byte slice with the length of n
// fills it with random byte from rand.Read()
func Bytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func NBytes(base64string string) (int, error) {
	b, err := base64.URLEncoding.DecodeString(base64string)
	if err != nil {
		return -1, nil
	}
	return len(b), nil
}

// String will return base64 string of a byte slice
// with the length of nBytes. It returns string and error.
func String(nBytes int) (string, error) {
	b, err := Bytes(nBytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// RememberToken is a helper function designed to generate
// remember tokens of a predetermined byte size.
func RememberToken() (string, error) {
	return String(RememberTokenBytes)
}
