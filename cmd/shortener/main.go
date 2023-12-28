package main

import (
	"github.com/vook88/go-url-shortener/internal/config"
	"github.com/vook88/go-url-shortener/internal/server"
	"github.com/vook88/go-url-shortener/internal/storage"
)

func main() {
	cfg := config.NewConfig()
	if err := run(cfg); err != nil {
		panic(err)
	}
}

func run(cfg *config.Config) error {
	newStorage := storage.New()

	s, _ := server.New(cfg.ServerAddress, cfg.BaseURL, newStorage)
	return s.Run()
}
