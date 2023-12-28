package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"github.com/vook88/go-url-shortener/internal/server"
	storage2 "github.com/vook88/go-url-shortener/internal/storage"
)

func setupServer() (*server.Server, *chi.Mux) {
	mockStorage := storage2.New()
	return server.New(":8080", "https://example.com", mockStorage)
}

func trimDomainAndSlash(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	// Удаление первого слэша из пути
	trimmedPath := parsedURL.Path
	if len(trimmedPath) > 0 && trimmedPath[0] == '/' {
		trimmedPath = trimmedPath[1:]
	}

	return trimmedPath, nil
}

func TestGenerateShortUrl(t *testing.T) {
	testCases := []struct {
		method       string
		expectedCode int
	}{
		{method: http.MethodGet, expectedCode: http.StatusBadRequest},
		{method: http.MethodPut, expectedCode: http.StatusBadRequest},
		{method: http.MethodDelete, expectedCode: http.StatusBadRequest},
		{method: http.MethodPost, expectedCode: http.StatusCreated},
	}
	_, router := setupServer()
	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {

			body := bytes.NewBufferString(`{"url": "https://longurl.com"}`)
			request, _ := http.NewRequest(tc.method, "/", body)
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)
			assert.Equal(t, tc.expectedCode, response.Code, "Код ответа не совпадает с ожидаемым")
		})
	}
}

func TestGetShortURL(t *testing.T) {
	testCases := []struct {
		method       string
		expectedCode int
	}{
		{method: http.MethodGet, expectedCode: http.StatusTemporaryRedirect},
		{method: http.MethodPut, expectedCode: http.StatusBadRequest},
		{method: http.MethodDelete, expectedCode: http.StatusBadRequest},
		{method: http.MethodPost, expectedCode: http.StatusBadRequest},
	}
	_, router := setupServer()

	testURL := "https://longurl.com"
	body := bytes.NewBufferString(testURL)
	request, _ := http.NewRequest(http.MethodPost, "/", body)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	id, err := trimDomainAndSlash(response.Body.String())
	if err != nil {
		fmt.Println("Ошибка при разборе URL:", err)
		return
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			request1, _ := http.NewRequest(tc.method, `/`+id, nil)
			response1 := httptest.NewRecorder()
			router.ServeHTTP(response1, request1)

			assert.Equal(t, tc.expectedCode, response1.Code, "Код ответа не совпадает с ожидаемым")

			if tc.method == http.MethodGet {
				assert.Equal(t, response1.Header().Get("Location"), testURL)
			}

		})
	}
}
