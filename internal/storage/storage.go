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

type Producer struct {
	file    *os.File
	encoder *json.Encoder
}

func (p *Producer) WriteEvent(event *Event) error {
	return p.encoder.Encode(event)
}

func (p *Producer) Close() error {
	return p.file.Close()
}

func NewProducer(filename string) (*Producer, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &Producer{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

type Consumer struct {
	file    *os.File
	decoder *json.Decoder
}

func NewConsumer(filename string) (*Consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (c *Consumer) ReadEvent() (*Event, error) {
	event := &Event{}
	err := c.decoder.Decode(event)
	if err != nil {
		return nil, err
	}

	return event, nil
}

func (c *Consumer) Close() error {
	return c.file.Close()
}

type FileURLStorage struct {
	*MemoryURLStorage
	producer *Producer
}

func New(filepath string) (URLStorage, error) {
	urls := make(map[string]string)

	if filepath != "" {
		producer, err := NewProducer(filepath)
		if err != nil {
			return nil, err
		}

		consumer, err := NewConsumer(filepath)
		if err != nil {
			return nil, err
		}
		defer consumer.Close()

		for {
			event, err := consumer.ReadEvent()
			if err != nil {
				if err.Error() == "EOF" {
					break
				}

				return nil, err
			}

			urls[event.ShortURL] = event.OriginalURL
		}

		return &FileURLStorage{
			producer:         producer,
			MemoryURLStorage: &MemoryURLStorage{urls: urls},
		}, nil
	}
	return &MemoryURLStorage{urls: urls}, nil
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

	if err := f.producer.WriteEvent(&event); err != nil {
		f.MemoryURLStorage.DeleteURL(id)
		return err
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
