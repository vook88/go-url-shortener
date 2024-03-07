package storage

import (
	"context"
	"errors"

	"github.com/vook88/go-url-shortener/internal/database"
	errors2 "github.com/vook88/go-url-shortener/internal/errors"
	"github.com/vook88/go-url-shortener/internal/models"
)

type MemoryURLStorage struct {
	urls                map[int]map[string]string
	lastGeneratedUserID int
}

var _ URLStorage = (*MemoryURLStorage)(nil)

func (s *MemoryURLStorage) GenerateUserID(_ context.Context) (int, error) {
	s.lastGeneratedUserID++
	return s.lastGeneratedUserID, nil
}

func (s *MemoryURLStorage) HasValue(_ context.Context, userID int, value string) (bool, string, error) {
	for k, v := range s.urls[userID] {
		if v == value {
			return true, k, nil
		}
	}
	return false, "", nil
}

func (s *MemoryURLStorage) AddURL(ctx context.Context, userID int, id string, url string) error {
	if id == "" {
		return errors.New("short URL can't be empty")
	}
	yes, key, err := s.HasValue(ctx, userID, url)
	if err != nil {
		return err
	}
	if yes {
		return errors2.NewDuplicateURLError(key)
	}
	if s.urls[userID] == nil {
		s.urls[userID] = make(map[string]string)
	}
	s.urls[userID][id] = url
	return nil
}

func (s *MemoryURLStorage) BatchAddURL(_ context.Context, userID int, urls []database.InsertURL) error {
	for _, url := range urls {
		s.urls[userID][url.ShortURL] = url.OriginalURL
	}
	return nil
}

func (s *MemoryURLStorage) GetURL(_ context.Context, id string) (string, bool, error) {
	for _, v := range s.urls {
		for k, v := range v {
			if k == id {
				return v, true, nil
			}
		}

	}
	return "", false, nil
}

func (s *MemoryURLStorage) GetUserURLs(_ context.Context, userID int) (models.BatchUserURLs, error) {
	var urls models.BatchUserURLs
	for k, v := range s.urls[userID] {
		urls = append(urls, models.UserURL{
			ShortURL:    k,
			OriginalURL: v,
		})
	}
	return urls, nil

}

func (s *MemoryURLStorage) Ping(_ context.Context) error {
	return errors.New("MemoryURLStorage doesn't support ping")
}

func (s *MemoryURLStorage) DeleteURL(_ context.Context, userID int, id string) error {
	delete(s.urls[userID], id)
	return nil
}

func (s *MemoryURLStorage) BatchDeleteURLs(_ context.Context, urls []string) error {
	for _, url := range urls {
		for _, v := range s.urls {
			delete(v, url)
		}
	}
	return nil
}
