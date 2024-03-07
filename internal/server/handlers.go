package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/vook88/go-url-shortener/internal/contextkeys"
	errors2 "github.com/vook88/go-url-shortener/internal/errors"
	"github.com/vook88/go-url-shortener/internal/logger"
	"github.com/vook88/go-url-shortener/internal/models"
	"github.com/vook88/go-url-shortener/internal/service"
	"github.com/vook88/go-url-shortener/internal/storage"
)

type Handler struct {
	baseURL string
	storage storage.URLStorage
	log     zerolog.Logger
	mux     *chi.Mux
}

func NewHandler(ctx context.Context, baseURL string, storage storage.URLStorage, log zerolog.Logger) *Handler {
	go service.BatchDeleteURLs(ctx, storage, log, 10)

	r := chi.NewRouter()
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Bad Request", http.StatusBadRequest)
	})

	r.Use(logger.LoggerMiddleware(log))
	r.Use(gzipMiddleware)

	h := Handler{
		baseURL: baseURL,
		storage: storage,
		log:     log,
		mux:     r,
	}

	r.With(AuthMiddlewareCheckAndCreate(storage, log)).Post("/", h.generateShortURL)
	r.With(AuthMiddlewareCheckAndCreate(storage, log)).Post("/api/shorten", h.shortenURL)
	r.Get("/{id}", h.getShortURL)
	r.Get("/ping", h.pingDB)
	r.With(AuthMiddlewareCheckAndCreate(storage, log)).Post("/api/shorten/batch", h.batchShortenURLs)
	r.With(AuthMiddlewareCheckOnly(log)).Get("/api/user/urls", h.getUserURLs)
	r.Delete("/api/user/urls", h.deleteUserURLs)

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

	userID, ok := req.Context().Value(contextkeys.UserIDKey).(int)
	if !ok {
		http.Error(res, "user id not found in context", http.StatusInternalServerError)
		return
	}

	shortener := service.NewShortener(h.storage, h.baseURL)

	shortURL, err := shortener.GenerateShortURL(req.Context(), userID, string(url))
	if err != nil {
		var dupErr *errors2.DuplicateURLError
		if errors.As(err, &dupErr) {
			res.WriteHeader(http.StatusConflict)
			_, _ = fmt.Fprintf(res, "%s", h.baseURL+"/"+err.Error())
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
	url, ok, err := h.storage.GetURL(req.Context(), prefix)
	h.log.Debug().Msgf("URL: %s, ShortURL: %s", url, prefix)

	if err != nil {
		h.log.Error().Msg(err.Error())
		if errors.Is(err, errors2.ErrURLDeleted) {
			http.Error(res, "URL not found", http.StatusGone)
			return
		}
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	if !ok {
		h.log.Error().Msg("URL not found")
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

	h.log.Debug().Msg("decoding request")
	var r models.RequestShortURL
	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&r); err != nil {
		h.log.Debug().Msg("cannot decode request JSON body")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	shortener := service.NewShortener(h.storage, h.baseURL)

	userID, ok := req.Context().Value(contextkeys.UserIDKey).(int)
	if !ok {
		http.Error(res, "user id not found in context", http.StatusInternalServerError)
		return
	}
	shortURL, err := shortener.GenerateShortURL(req.Context(), userID, r.URL)
	responseStatus := http.StatusCreated
	if err != nil {
		var dupErr *errors2.DuplicateURLError
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
		h.log.Debug().Msgf("error encoding response: %s", err.Error())
		return
	}
	h.log.Debug().Msg("sending HTTP 200 response")
}

func (h *Handler) batchShortenURLs(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Only GET requests are allowed!", http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	h.log.Debug().Msg("decoding request")

	var request models.RequestBatchLongURLs
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	s := service.NewShortener(h.storage, h.baseURL)

	userID, ok := req.Context().Value(contextkeys.UserIDKey).(int)
	if !ok {
		http.Error(res, "user id not found in context", http.StatusInternalServerError)
		return
	}
	shortURLs, err := s.BatchGenerateShortURL(req.Context(), userID, request)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)

	// сериализуем ответ сервера
	enc := json.NewEncoder(res)
	if err = enc.Encode(shortURLs); err != nil {
		h.log.Debug().Msgf("error encoding response: %s", err.Error())
		return
	}
	h.log.Debug().Msg("sending HTTP 200 response")
}

func (h *Handler) getUserURLs(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Only GET requests are allowed!", http.StatusBadRequest)
		return
	}

	log.Debug().Msg("getting user URLs")

	userID, ok := req.Context().Value(contextkeys.UserIDKey).(int)
	if !ok {
		http.Error(res, "user id not found in context", http.StatusInternalServerError)
		return
	}

	urls, err := h.storage.GetUserURLs(req.Context(), userID)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	for i := range urls {
		urls[i].ShortURL = h.baseURL + "/" + urls[i].ShortURL
	}

	res.Header().Set("Content-Type", "application/json")
	if len(urls) > 0 {
		res.WriteHeader(http.StatusOK)
	} else {
		res.WriteHeader(http.StatusNoContent)
	}

	// сериализуем ответ сервера
	enc := json.NewEncoder(res)
	if err = enc.Encode(urls); err != nil {
		h.log.Debug().Msgf("error encoding response: %s", err.Error())
		return
	}
	h.log.Debug().Msg("sending HTTP 200 response")

}

func (h *Handler) deleteUserURLs(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodDelete {
		http.Error(res, "Only DELETE requests are allowed!", http.StatusBadRequest)
		return
	}

	h.log.Debug().Msg("deleting user URLs")

	var urls models.RequestDeleteShortURL
	err := json.NewDecoder(req.Body).Decode(&urls)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	s := service.NewShortener(h.storage, h.baseURL)
	s.BatchDeleteShortURL(req.Context(), urls)

	res.WriteHeader(http.StatusAccepted)
	h.log.Debug().Msg("sending HTTP 202 response")
}

func (h *Handler) pingDB(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Only GET requests are allowed!", http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	h.log.Debug().Msg("ping DB")

	if err := h.storage.Ping(req.Context()); err != nil {
		h.log.Debug().Msg("cannot ping to database")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
}
