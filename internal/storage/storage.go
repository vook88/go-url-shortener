package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"

	"github.com/google/uuid"

	"github.com/vook88/go-url-shortener/internal/config"
	"github.com/vook88/go-url-shortener/internal/database"
)

type Event struct {
	UUID        uuid.UUID `json:"uuid"`
	ShortURL    string    `json:"short_url"`
	OriginalURL string    `json:"original_url"`
}

type URLStorage interface {
	AddURL(ctx context.Context, id string, url string) error
	GetURL(ctx context.Context, id string) (string, bool)
}

var _ URLStorage = (*MemoryURLStorage)(nil)

type MemoryURLStorage struct {
	urls map[string]string
}

type FileURLStorage struct {
	*MemoryURLStorage
	filepath string
}

type DBURLStorage struct {
	db *sql.DB
}

func New(ctx context.Context, config *config.Config) (URLStorage, error) {
	if config.DatabaseDSN != "" {
		db, _ := database.New(config.DatabaseDSN)
		err := database.Ping(ctx, db)
		err2 := database.RunMigrations(db)
		if err == nil && err2 == nil {
			return &DBURLStorage{db: db}, nil
		}
	}
	urls := make(map[string]string)

	if config.FileStoragePath == "" {
		return &MemoryURLStorage{urls: urls}, nil
	}

	file, err := os.OpenFile(config.FileStoragePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	for {
		event := &Event{}
		err2 := json.NewDecoder(file).Decode(event)
		if err2 != nil {
			if err2.Error() == "EOF" {
				break
			}

			return nil, err2
		}

		urls[event.ShortURL] = event.OriginalURL
	}

	return &FileURLStorage{
		filepath:         config.FileStoragePath,
		MemoryURLStorage: &MemoryURLStorage{urls: urls},
	}, nil
}

func (f *FileURLStorage) AddURL(ctx context.Context, id string, url string) error {
	err := f.MemoryURLStorage.AddURL(ctx, id, url)
	if err != nil {
		return err
	}

	newUUID, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}

	event := Event{
		UUID:        newUUID,
		ShortURL:    id,
		OriginalURL: url,
	}

	file, err := os.OpenFile(f.filepath, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	defer file.Close()

	if err2 := json.NewEncoder(file).Encode(&event); err2 != nil {
		f.MemoryURLStorage.DeleteURL(ctx, id)
		return err2
	}

	return nil
}

func (s *MemoryURLStorage) AddURL(_ context.Context, id string, url string) error {
	if id == "" {
		return errors.New("short URL can't be empty")
	}
	s.urls[id] = url
	return nil
}

func (s *MemoryURLStorage) GetURL(_ context.Context, id string) (string, bool) {
	url, ok := s.urls[id]
	return url, ok
}

func (s *MemoryURLStorage) DeleteURL(_ context.Context, id string) {
	delete(s.urls, id)
}

func (s *DBURLStorage) AddURL(ctx context.Context, id string, url string) error {
	err := database.AddURL(ctx, s.db, id, url)
	if err != nil {
		return err
	}
	return nil
}

func (s *DBURLStorage) GetURL(ctx context.Context, id string) (string, bool) {
	url, b, err := database.GetURL(ctx, s.db, id)
	if err != nil {
		return "", false
	}
	return url, b
}
