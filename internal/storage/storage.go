package storage

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/google/uuid"
)

type Event struct {
	UUID        uuid.UUID `json:"uuid"`
	ShortURL    string    `json:"short_url"`
	OriginalURL string    `json:"original_url"`
}

type URLStorage interface {
	AddURL(id string, url string) error
	GetURL(id string) (string, bool)
}

var _ URLStorage = (*MemoryURLStorage)(nil)

type MemoryURLStorage struct {
	urls map[string]string
}

type FileURLStorage struct {
	*MemoryURLStorage
	filepath string
}

func New(filepath string) (URLStorage, error) {
	urls := make(map[string]string)

	if filepath == "" {
		return &MemoryURLStorage{urls: urls}, nil
	}

	file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
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
		filepath:         filepath,
		MemoryURLStorage: &MemoryURLStorage{urls: urls},
	}, nil
}

func (f *FileURLStorage) AddURL(id string, url string) error {
	err := f.MemoryURLStorage.AddURL(id, url)
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
		f.MemoryURLStorage.DeleteURL(id)
		return err2
	}

	return nil
}

func (s *MemoryURLStorage) AddURL(id string, url string) error {
	if id == "" {
		return errors.New("short URL can't be empty")
	}
	s.urls[id] = url
	return nil
}

func (s *MemoryURLStorage) GetURL(id string) (string, bool) {
	url, ok := s.urls[id]
	return url, ok
}

func (s *MemoryURLStorage) DeleteURL(id string) {
	delete(s.urls, id)
}
