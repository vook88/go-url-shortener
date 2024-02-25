package service

import (
	"context"

	"github.com/vook88/go-url-shortener/internal/database"
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

func (s Shortener) GenerateShortURL(ctx context.Context, userID int, URL string) (string, error) {
	shortID, err := id.New()
	if err != nil {
		return "", err
	}

	err = s.storage.AddURL(ctx, userID, shortID, URL)
	if err != nil {
		return "", err
	}

	return s.baseURL + "/" + shortID, nil
}

func (s Shortener) BatchGenerateShortURL(ctx context.Context, userID int, URLs []models.BatchLongURL) ([]models.BatchShortURL, error) {
	var urls []models.BatchShortURL
	var insertURLs []database.InsertURL

	for _, URL := range URLs {
		shortID, err := id.New()
		if err != nil {
			return nil, err
		}
		urls = append(urls, models.BatchShortURL{
			CorrelationID: URL.CorrelationID,
			ShortURL:      s.baseURL + "/" + shortID,
		})
		insertURLs = append(insertURLs, database.InsertURL{
			ShortURL:    shortID,
			OriginalURL: URL.OriginalURL,
		})
	}
	err := s.storage.BatchAddURL(ctx, userID, insertURLs)
	if err != nil {
		return nil, err
	}
	return urls, nil
}
