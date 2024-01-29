package storage

import (
	"context"
	"encoding/json"
	"errors"
	"os"

	"github.com/google/uuid"

	"github.com/vook88/go-url-shortener/internal/config"
	"github.com/vook88/go-url-shortener/internal/database"
	errors2 "github.com/vook88/go-url-shortener/internal/errors"
)

type Event struct {
	UUID        uuid.UUID `json:"uuid"`
	ShortURL    string    `json:"short_url"`
	OriginalURL string    `json:"original_url"`
}

type URLStorage interface {
	AddURL(ctx context.Context, id string, url string) error
	BatchAddURL(ctx context.Context, insertURLs []database.InsertURL) error
	GetURL(ctx context.Context, id string) (string, bool)
	Ping(ctx context.Context) error
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
	db *database.DB
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
func (s *MemoryURLStorage) HasValue(value string) (bool, string) {
	for k, v := range s.urls {
		if v == value {
			return true, k
		}
	}
	return false, ""
}
func (s *MemoryURLStorage) AddURL(_ context.Context, id string, url string) error {
	if id == "" {
		return errors.New("short URL can't be empty")
	}
	yes, key := s.HasValue(url)
	if yes {
		return errors2.NewDuplicateURLError(key)
	}
	s.urls[id] = url
	return nil
}

func (s *MemoryURLStorage) BatchAddURL(_ context.Context, urls []database.InsertURL) error {
	for _, url := range urls {
		s.urls[url.ShortURL] = url.OriginalURL
	}
	return nil
}

func (s *MemoryURLStorage) GetURL(_ context.Context, id string) (string, bool) {
	url, ok := s.urls[id]
	return url, ok
}

func (s *MemoryURLStorage) Ping(_ context.Context) error {
	return errors.New("MemoryURLStorage doesn't support ping")
}

func (s *MemoryURLStorage) DeleteURL(_ context.Context, id string) {
	delete(s.urls, id)
}

func (s *DBURLStorage) AddURL(ctx context.Context, id string, url string) error {
	err := s.db.AddURL(ctx, id, url)
	if err != nil {
		return err
	}
	return nil
}

func (s *DBURLStorage) BatchAddURL(ctx context.Context, urls []database.InsertURL) error {
	err := s.db.BatchAddURL(ctx, urls)
	if err != nil {
		return err
	}
	return nil
}

func (s *DBURLStorage) GetURL(ctx context.Context, id string) (string, bool) {
	url, b, err := s.db.GetURL(ctx, id)
	if err != nil {
		return "", false
	}
	return url, b
}

func (s *DBURLStorage) Ping(ctx context.Context) error {
	return s.db.Ping(ctx)
}
