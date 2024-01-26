package storage

import (
	"context"
	"os"
	"testing"

	"github.com/vook88/go-url-shortener/internal/config"
)

func TestMemoryURLStorage(t *testing.T) {
	ctx := context.Background()
	// Создаем инстанс MemoryURLStorage
	storage := &MemoryURLStorage{urls: make(map[string]string)}

	// Тестируем добавление URL
	err := storage.AddURL(ctx, "test1", "http://example.com/test1")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Тестируем получение URL
	url, ok := storage.GetURL(ctx, "test1")
	if !ok || url != "http://example.com/test1" {
		t.Errorf("Expected URL 'http://example.com/test1', got '%s'", url)
	}

	// Тестируем удаление URL
	storage.DeleteURL(ctx, "test1")
	_, ok = storage.GetURL(ctx, "test1")
	if ok {
		t.Errorf("Expected URL 'test1' to be deleted, but it's still present")
	}
}

func TestFileURLStorage(t *testing.T) {
	ctx := context.Background()
	// Создаем временный файл для тестирования
	tmpfile := "test_urls.txt"
	c := config.Config{FileStoragePath: tmpfile}

	defer os.Remove(tmpfile)

	// Создаем инстанс FileURLStorage для тестирования
	storage, err := New(ctx, &c)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Тестируем добавление URL
	err = storage.AddURL(ctx, "test2", "http://example.com/test2")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Тестируем получение URL
	url, ok := storage.GetURL(ctx, "test2")
	if !ok || url != "http://example.com/test2" {
		t.Errorf("Expected URL 'http://example.com/test2', got '%s'", url)
	}
}
