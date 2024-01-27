package service

import (
	"context"

	"github.com/vook88/go-url-shortener/internal/id"
	"github.com/vook88/go-url-shortener/internal/models"
	"github.com/vook88/go-url-shortener/internal/storage"
)

type Shortener struct {
	storage storage.URLStorage
	baseURL string
}

func NewShortener(storage storage.URLStorage, baseURL string) *Shortener {
	return &Shortener{storage: storage, baseURL: baseURL}
}

func (s Shortener) GenerateShortURL(ctx context.Context, URL string) (string, error) {
	shortID, err := id.New()
	if err != nil {
		return "", err
	}

	err = s.storage.AddURL(ctx, shortID, URL)
	if err != nil {
		return "", err
	}

	return s.baseURL + "/" + shortID, nil
}

func (s Shortener) BatchGenerateShortURL(ctx context.Context, URLs []models.BatchLongURL) ([]models.BatchShortURL, error) {
	var urls []models.BatchShortURL

	for _, URL := range URLs {
		shortURL, err := s.GenerateShortURL(ctx, URL.OriginalURL)
		if err != nil {

			return nil, err
		}
		urls = append(urls, models.BatchShortURL{
			CorrelationID: URL.CorrelationID,
			ShortURL:      shortURL,
		})
	}
	return urls, nil
}
