package main

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/vook88/go-url-shortener/cmd/config"
)

func generateID() (string, error) {
	b := make([]byte, 6) // генерирует 8 символов после base64 кодирования
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func main() {
	cfg := config.NewConfig()
	if err := run(cfg); err != nil {
		panic(err)
	}
}

func run(cfg *config.Config) error {
	storage := NewMemoryURLStorage()
	err := Init(cfg, storage)
	return err
}
