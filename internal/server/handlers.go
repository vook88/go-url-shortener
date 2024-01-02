package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/vook88/go-url-shortener/internal/id"
	"github.com/vook88/go-url-shortener/internal/logger"
	"github.com/vook88/go-url-shortener/internal/models"
	"github.com/vook88/go-url-shortener/internal/storage"
)

type Handler struct {
	baseURL string
	storage storage.URLStorage
	mux     *chi.Mux
}

func NewHandler(baseURL string, storage storage.URLStorage) *Handler {
	r := chi.NewRouter()
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Bad Request", http.StatusBadRequest)
	})
	r.Use(logger.LoggerMiddleware)
	h := Handler{
		baseURL: baseURL,
		storage: storage,
		mux:     r,
	}

	r.Post("/", h.generateShortURL)
	r.Get("/{id}", h.getShortURL)
	r.Post("/api/shorten", h.shortenURL)

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

	shortID, err := id.New()
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.storage.AddURL(shortID, string(url))
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	res.WriteHeader(http.StatusCreated)
	_, _ = fmt.Fprintf(res, "%s/%s", h.baseURL, shortID)
}

func (h *Handler) getShortURL(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Only GET requests are allowed!", http.StatusBadRequest)
		return
	}
	prefix := chi.URLParam(req, "id")
	url, ok := h.storage.GetURL(prefix)
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

	shortID, err := id.New()
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.storage.AddURL(shortID, string(r.URL))
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	resp := models.ResponseShortURL{
		ShortURL: h.baseURL + "/" + shortID,
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)

	// сериализуем ответ сервера
	enc := json.NewEncoder(res)
	if err = enc.Encode(resp); err != nil {
		log.Debug().Msg(`error encoding response" + log.Err(err)`)
		return
	}
	log.Debug().Msg("sending HTTP 200 response")

}
