package storage

import (
	"context"
	"encoding/json"
	"os"

	"github.com/google/uuid"

	"github.com/vook88/go-url-shortener/internal/config"
	"github.com/vook88/go-url-shortener/internal/database"
	"github.com/vook88/go-url-shortener/internal/models"
)

type Event struct {
	UUID        uuid.UUID `json:"uuid"`
	UserID      int       `json:"user_id"`
	ShortURL    string    `json:"short_url"`
	OriginalURL string    `json:"original_url"`
}

type URLStorage interface {
	AddURL(ctx context.Context, userID int, id string, url string) error
	BatchAddURL(ctx context.Context, userID int, insertURLs []database.InsertURL) error
	GetURL(ctx context.Context, id string) (string, bool, error)
	GetUserURLs(ctx context.Context, userID int) (models.BatchUserURLs, error)
	Ping(ctx context.Context) error
	GenerateUserID(ctx context.Context) (int, error)
	BatchDeleteURLs(ctx context.Context, urls []string) error
}

func New(ctx context.Context, config *config.Config) (URLStorage, error) {
	if config.DatabaseDSN != "" {
		db, _ := database.New(config.DatabaseDSN)
		err := db.Ping(ctx)
		err2 := db.RunMigrations()
		if err == nil && err2 == nil {
			return &DBURLStorage{db: db}, nil
		}
	}
	urls := make(map[int]map[string]string)

	if config.FileStoragePath == "" {

		return &MemoryURLStorage{urls: urls, lastGeneratedUserID: 0}, nil
	}

	file, err := os.OpenFile(config.FileStoragePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	defer file.Close()
	lastGeneratedUserID := 0
	for {
		event := &Event{}
		err2 := json.NewDecoder(file).Decode(event)
		if err2 != nil {
			if err2.Error() == "EOF" {
				break
			}

			return nil, err2
		}
		if event.UserID > lastGeneratedUserID {
			lastGeneratedUserID = event.UserID
		}
		urls[event.UserID][event.ShortURL] = event.OriginalURL
	}

	return &FileURLStorage{
		filepath:         config.FileStoragePath,
		MemoryURLStorage: &MemoryURLStorage{urls: urls, lastGeneratedUserID: lastGeneratedUserID},
	}, nil
}
