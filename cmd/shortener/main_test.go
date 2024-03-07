package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vook88/go-url-shortener/internal/authn"
	"github.com/vook88/go-url-shortener/internal/config"
	"github.com/vook88/go-url-shortener/internal/logger"
	"github.com/vook88/go-url-shortener/internal/server"
	storage2 "github.com/vook88/go-url-shortener/internal/storage"
)

func setupHandler() *server.Handler {
	ctx := context.Background()
	c := config.Config{}
	mockStorage, _ := storage2.New(ctx, &c)
	log := logger.New(0)
	return server.NewHandler(ctx, "https://example.com", mockStorage, log)
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

	h := setupHandler()
	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {

			body := bytes.NewBufferString("https://longurl.com")
			request, _ := http.NewRequest(tc.method, "/", body)
			response := httptest.NewRecorder()

			h.ServeHTTP(response, request)
			assert.Equal(t, tc.expectedCode, response.Code, "Код ответа не совпадает с ожидаемым")
		})
	}
}

func TestShortenURL(t *testing.T) {
	testCases := []struct {
		method       string
		expectedCode int
	}{
		{method: http.MethodGet, expectedCode: http.StatusBadRequest},
		{method: http.MethodPut, expectedCode: http.StatusBadRequest},
		{method: http.MethodDelete, expectedCode: http.StatusBadRequest},
		{method: http.MethodPost, expectedCode: http.StatusCreated},
	}

	h := setupHandler()
	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			body := bytes.NewBufferString(`{"url": "https://longurl.com"}`)

			request, _ := http.NewRequest(tc.method, "/api/shorten", body)
			response := httptest.NewRecorder()

			h.ServeHTTP(response, request)
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
	h := setupHandler()

	testURL := "https://longurl.com"
	body := bytes.NewBufferString(testURL)
	request, _ := http.NewRequest(http.MethodPost, "/", body)
	response := httptest.NewRecorder()
	h.ServeHTTP(response, request)

	u, _ := url.Parse(response.Body.String())
	if u == nil {
		t.Fatal("Не удалось распарсить URL")
	}
	id := u.Path[1:]

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			request1, _ := http.NewRequest(tc.method, `/`+id, nil)
			response1 := httptest.NewRecorder()
			h.ServeHTTP(response1, request1)

			assert.Equal(t, tc.expectedCode, response1.Code, "Код ответа не совпадает с ожидаемым")

			if tc.method == http.MethodGet {
				assert.Equal(t, response1.Header().Get("Location"), testURL)
			}

		})
	}
}

func TestGetUserURLs(t *testing.T) {

	h := setupHandler()

	t.Run("Without Cookie", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, `/api/user/urls`, nil)
		response := httptest.NewRecorder()
		h.ServeHTTP(response, request)

		assert.Equal(t, http.StatusUnauthorized, response.Code, "Код ответа не совпадает с ожидаемым")
	})

	t.Run("With Cookie / Empty response", func(t *testing.T) {
		encodedValue, err2 := authn.BuildJWTString(2)
		if err2 != nil {
			t.Errorf("Expected no error, got %v", err2)
			return
		}

		request, _ := http.NewRequest(http.MethodGet, `/api/user/urls`, nil)
		request.AddCookie(&http.Cookie{
			Name:     server.CookieAuthName,
			Value:    encodedValue,
			Path:     "/",
			HttpOnly: true,
		})
		response := httptest.NewRecorder()
		h.ServeHTTP(response, request)

		assert.Equal(t, http.StatusNoContent, response.Code, "Код ответа не совпадает с ожидаемым")
	})

	t.Run("With Cookie / Not empty response", func(t *testing.T) {
		body := bytes.NewBufferString(`{"url": "https://longurl.com"}`)

		request, _ := http.NewRequest(http.MethodPost, "/api/shorten", body)
		response := httptest.NewRecorder()
		h.ServeHTTP(response, request)

		res := response.Result()
		defer res.Body.Close()

		cookies := res.Cookies()

		request1, _ := http.NewRequest(http.MethodGet, `/api/user/urls`, nil)
		for _, c := range cookies {
			request1.AddCookie(c)
		}

		response1 := httptest.NewRecorder()
		h.ServeHTTP(response1, request1)

		assert.Equal(t, http.StatusOK, response1.Code, "Код ответа не совпадает с ожидаемым")
	})
}
