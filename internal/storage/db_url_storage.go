package storage

import (
	"context"

	"github.com/vook88/go-url-shortener/internal/database"
	"github.com/vook88/go-url-shortener/internal/models"
)

var _ URLStorage = (*DBURLStorage)(nil)

type DBURLStorage struct {
	db *database.DB
}

func (s *DBURLStorage) AddURL(ctx context.Context, userID int, id string, url string) error {
	err := s.db.AddURL(ctx, userID, id, url)
	if err != nil {
		return err
	}
	return nil
}

func (s *DBURLStorage) BatchAddURL(ctx context.Context, userID int, urls []database.InsertURL) error {
	err := s.db.BatchAddURL(ctx, userID, urls)
	if err != nil {
		return err
	}
	return nil
}

func (s *DBURLStorage) GetURL(ctx context.Context, id string) (string, bool, error) {
	url, b, err := s.db.GetURL(ctx, id)
	if err != nil {
		return "", false, err
	}
	return url, b, nil
}

func (s *DBURLStorage) GetUserURLs(ctx context.Context, userID int) (models.BatchUserURLs, error) {
	return s.db.GetUserURLs(ctx, userID)
}

func (s *DBURLStorage) Ping(ctx context.Context) error {
	return s.db.Ping(ctx)
}

func (s *DBURLStorage) GenerateUserID(ctx context.Context) (int, error) {
	return s.db.AddUser(ctx)
}
