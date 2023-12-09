package main

type URLStorage interface {
	AddURL(id string, url string)
	GetURL(id string) (string, bool)
}

type MemoryURLStorage struct {
	urls map[string]string
}

func NewMemoryURLStorage() *MemoryURLStorage {
	return &MemoryURLStorage{urls: make(map[string]string)}
}

func (s *MemoryURLStorage) AddURL(id string, url string) {
	s.urls[id] = url
}

func (s *MemoryURLStorage) GetURL(id string) (string, bool) {
	url, ok := s.urls[id]
	return url, ok
}
