package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/vook88/go-url-shortener/internal/database"
	"github.com/vook88/go-url-shortener/internal/logger"
	"github.com/vook88/go-url-shortener/internal/models"
	"github.com/vook88/go-url-shortener/internal/service"
	"github.com/vook88/go-url-shortener/internal/storage"
)

type Handler struct {
	baseURL string
	storage storage.URLStorage
	mux     *chi.Mux
}

func NewHandler(baseURL string, storage storage.URLStorage, databaseDSN string) *Handler {
	r := chi.NewRouter()
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Bad Request", http.StatusBadRequest)
	})

	r.Use(logger.LoggerMiddleware)
	r.Use(gzipMiddleware)

	h := Handler{
		baseURL: baseURL,
		storage: storage,
		mux:     r,
	}

	r.Post("/", h.generateShortURL)
	r.Post("/api/shorten", h.shortenURL)
	r.Get("/{id}", h.getShortURL)
	r.Get("/ping", h.pingDB)
	r.Post("/api/shorten/batch", h.batchShortenURLs)

	return &h
}

func (h *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	h.mux.ServeHTTP(writer, request)
}

func (h *Handler) generateShortURL(res http.ResponseWriter, req *http.Request) {
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

	shortener := service.NewShortener(h.storage, h.baseURL)

	shortURL, err := shortener.GenerateShortURL(req.Context(), string(url))
	if err != nil {
		var dupErr *database.DuplicateURLError
		if errors.As(err, &dupErr) {
			http.Error(res, h.baseURL+"/"+err.Error(), http.StatusConflict)
			return
		}
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	res.WriteHeader(http.StatusCreated)
	_, _ = fmt.Fprintf(res, "%s", shortURL)
}

func (h *Handler) getShortURL(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Only GET requests are allowed!", http.StatusBadRequest)
		return
	}
	prefix := chi.URLParam(req, "id")
	url, ok := h.storage.GetURL(req.Context(), prefix)
	if !ok {
		http.Error(res, "", http.StatusBadRequest)
		return
	}
	http.Redirect(res, req, url, http.StatusTemporaryRedirect)
}

func (h *Handler) shortenURL(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Only GET requests are allowed!", http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	log := logger.GetLogger()
	log.Debug().Msg("decoding request")
	var r models.RequestShortURL
	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&r); err != nil {
		log.Debug().Msg("cannot decode request JSON body")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	shortener := service.NewShortener(h.storage, h.baseURL)
	shortURL, err := shortener.GenerateShortURL(req.Context(), r.URL)
	responseStatus := http.StatusCreated
	if err != nil {
		var dupErr *database.DuplicateURLError
		if !errors.As(err, &dupErr) {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return

		}
		shortURL = h.baseURL + "/" + err.Error()
		responseStatus = http.StatusConflict
	}

	resp := models.ResponseShortURL{
		ShortURL: shortURL,
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(responseStatus)

	// сериализуем ответ сервера
	enc := json.NewEncoder(res)
	if err = enc.Encode(resp); err != nil {
		log.Debug().Msg(`error encoding response" + log.Err(err)`)
		return
	}
	log.Debug().Msg("sending HTTP 200 response")
}

func (h *Handler) batchShortenURLs(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Only GET requests are allowed!", http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	log := logger.GetLogger()
	log.Debug().Msg("decoding request")

	var request models.RequestBatchLongURLs
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	s := service.NewShortener(h.storage, h.baseURL)
	shortURLs, err := s.BatchGenerateShortURL(req.Context(), request)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)

	// сериализуем ответ сервера
	enc := json.NewEncoder(res)
	if err = enc.Encode(shortURLs); err != nil {
		log.Debug().Msg(`error encoding response" + log.Err(err)`)
		return
	}
	log.Debug().Msg("sending HTTP 200 response")
}

func (h *Handler) pingDB(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Only GET requests are allowed!", http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	log := logger.GetLogger()
	log.Debug().Msg("ping DB")

	if err := h.storage.Ping(req.Context()); err != nil {
		log.Debug().Msg("cannot ping to database")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
}
