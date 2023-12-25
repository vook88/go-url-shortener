package main

import "errors"

type URLStorage interface {
	AddURL(id string, url string) error
	GetURL(id string) (string, bool)
}

type MemoryURLStorage struct {
	urls map[string]string
}

func NewMemoryURLStorage() *MemoryURLStorage {
	return &MemoryURLStorage{urls: make(map[string]string)}
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
