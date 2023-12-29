package id

import (
	"crypto/rand"
	"encoding/base64"
)

func New() (string, error) {
	b := make([]byte, 6) // генерирует 8 символов после base64 кодирования
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
