package service

import (
	"context"

	"github.com/vook88/go-url-shortener/internal/id"
	"github.com/vook88/go-url-shortener/internal/storage"
)

func GenerateShortURL(ctx context.Context, URL string, storage storage.URLStorage, baseURL string) (string, error) {
	shortID, err := id.New()
	if err != nil {
		return "", err
	}

	err = storage.AddURL(ctx, shortID, string(URL))
	if err != nil {
		return "", err
	}

	return baseURL + "/" + shortID, nil
}
