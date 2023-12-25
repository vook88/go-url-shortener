package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/vook88/go-url-shortener/cmd/config"
	"github.com/vook88/go-url-shortener/internal/logger"
)

func Init(cfg *config.Config, storage URLStorage) error {
	r := chi.NewRouter()
	r.Use(LoggerMiddleware)

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		generateShortURL(cfg, storage, w, r)
	})
	r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		getShortURL(storage, w, r)
	})

	err := http.ListenAndServe(cfg.ServerAddress, r)
	if err != nil {
		return err
	}
	return nil
}

func generateShortURL(cfg *config.Config, storage URLStorage, res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	url, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := generateID()
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	err = storage.AddURL(id, string(url))
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	res.WriteHeader(http.StatusCreated)
	_, _ = fmt.Fprintf(res, "%s/%s", cfg.BaseURL, id)

}

func getShortURL(storage URLStorage, res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Only GET requests are allowed!", http.StatusBadRequest)
		return
	}
	id := strings.TrimPrefix(req.URL.Path, "/")
	url, ok := storage.GetURL(id)
	if !ok {
		http.Error(res, "", http.StatusBadRequest)
		return
	}
	http.Redirect(res, req, url, http.StatusTemporaryRedirect)
}

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logger.GetLogger()
		start := time.Now()

		// эндпоинт /ping
		uri := r.RequestURI
		// метод запроса
		method := r.Method

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			responseData:   responseData,
		}
		next.ServeHTTP(&lw, r)

		duration := time.Since(start)

		log.Info().
			Str("Type", "Request").
			Str("URI", uri).
			Str("Method", method).
			Str("Duration", fmt.Sprint(duration)).
			Msg("")

		log.Info().
			Str("Type", "Response").
			Str("StatusCode", fmt.Sprint(lw.responseData.status)).
			Str("Size", fmt.Sprint(lw.responseData.size)).
			Msg("")
	})
}

type (
	// берём структуру для хранения сведений об ответе
	responseData struct {
		status int
		size   int
	}

	// добавляем реализацию http.ResponseWriter
	loggingResponseWriter struct {
		http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
		responseData        *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // захватываем размер
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}
