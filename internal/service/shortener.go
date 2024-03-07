package service

import (
	"context"
	"time"

	"github.com/rs/zerolog"

	"github.com/vook88/go-url-shortener/internal/database"
	"github.com/vook88/go-url-shortener/internal/id"
	"github.com/vook88/go-url-shortener/internal/models"
	"github.com/vook88/go-url-shortener/internal/storage"
)

var urlsToBeDeletedChan = make(chan string)

func BatchDeleteURLs(ctx context.Context, storage storage.URLStorage, log zerolog.Logger, batchSize int) {
	var urls []string
	for {
		select {
		case url, ok := <-urlsToBeDeletedChan:
			if !ok {
				log.Info().Msg("Channel urlsToBeDeletedChan closed.")

				if len(urls) > 0 {
					if err := storage.BatchDeleteURLs(ctx, urls); err != nil {
						log.Error().Msgf("Cannot batch delete URLs: %s", err.Error())
					}
				}
				return
			}
			urls = append(urls, url)

			if len(urls) >= batchSize {
				if err := storage.BatchDeleteURLs(ctx, urls); err != nil {
					log.Error().Msgf("Cannot batch delete URLs: %s", err.Error())
				}
				urls = urls[:0]
			}
		case <-time.After(time.Millisecond * 1000):
			if len(urls) > 0 {
				if err := storage.BatchDeleteURLs(ctx, urls); err != nil {
					log.Error().Msgf("Cannot batch delete URLs: %s", err.Error())
				}
				urls = urls[:0]
			}
		case <-ctx.Done():
			log.Error().Msg("Context cancelled, stopping the batch delete operation.")
			return // Выход из функции при отмене контекста
		}

		if ctx.Err() != nil {
			log.Error().Msg("Context cancelled, stopping the batch delete operation.")
			return
		}
	}
}

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
	var shortURLs = make([]models.BatchShortURL, 0, len(URLs))
	var insertURLs = make([]database.InsertURL, 0, len(URLs))

	for _, URL := range URLs {
		shortID, err := id.New()
		if err != nil {
			return nil, err
		}
		shortURLs = append(shortURLs, models.BatchShortURL{
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
	return shortURLs, nil
}

func (s Shortener) BatchDeleteShortURL(_ context.Context, shortURLs []string) {
	for _, url := range shortURLs {
		urlsToBeDeletedChan <- url
	}
}
