package service

import (
	"context"
	"time"

	"github.com/vook88/go-url-shortener/internal/database"
	"github.com/vook88/go-url-shortener/internal/id"
	"github.com/vook88/go-url-shortener/internal/logger"
	"github.com/vook88/go-url-shortener/internal/models"
	"github.com/vook88/go-url-shortener/internal/storage"
)

var urlsToBeDeletedChan = make(chan string)

func BatchDeleteURLs(ctx context.Context, storage storage.URLStorage, batchSize int) {
	log := logger.GetLogger()

	for {
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
			case <-time.After(time.Millisecond * 1000): // Таймаут ожидания новых URL
				if len(urls) > 0 {
					if err := storage.BatchDeleteURLs(ctx, urls); err != nil {
						log.Error().Msgf("Cannot batch delete URLs: %s", err.Error())
					}
					urls = urls[:0] // Очищаем список URL после удаления
				}
			case <-ctx.Done():
				log.Error().Msg("Context cancelled, stopping the batch delete operation.")
				return // Выход из функции при отмене контекста
			}

			// Проверяем, не был ли контекст отменен после каждой операции
			if ctx.Err() != nil {
				log.Error().Msg("Context cancelled, stopping the batch delete operation.")
				return
			}
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

func (s Shortener) BatchDeleteShortURL(ctx context.Context, shortURLs []string) {
	for _, url := range shortURLs {
		urlsToBeDeletedChan <- url
	}
}
