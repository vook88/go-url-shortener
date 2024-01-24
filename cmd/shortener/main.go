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
	newStorage, err := storage.New(cfg.FileStoragePath)
	if err != nil {
		return err
	}
	h := server.NewHandler(cfg.BaseURL, newStorage, cfg.DatabaseDSN)
	s := server.New(cfg.ServerAddress, h)
	return s.Run()
}
