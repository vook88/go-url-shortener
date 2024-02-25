package main

import (
	"context"

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
	ctx := context.Background()
	newStorage, err := storage.New(ctx, cfg)
	if err != nil {
		return err
	}
	h := server.NewHandler(ctx, cfg.BaseURL, newStorage)
	s := server.New(cfg.ServerAddress, h)
	return s.Run()
}
