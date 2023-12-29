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
	h := server.NewHandler(cfg.BaseURL, newStorage)
	s := server.New(cfg.ServerAddress, h)
	return s.Run()
}
