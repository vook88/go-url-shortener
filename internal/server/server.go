package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/vook88/go-url-shortener/internal/logger"
	"github.com/vook88/go-url-shortener/internal/storage"
)

type Server struct {
	httpServer *http.Server

	baseURL string
	storage storage.URLStorage
}

func (s *Server) Run() error {
	return s.httpServer.ListenAndServe()
}

func New(serverAddress string, baseURL string, storage storage.URLStorage) (*Server, *chi.Mux) {
	r := chi.NewRouter()
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Bad Request", http.StatusBadRequest)
	})
	r.Use(logger.LoggerMiddleware)

	s := &Server{
		httpServer: &http.Server{
			Addr:    serverAddress,
			Handler: r,
		},
		baseURL: baseURL,
		storage: storage,
	}

	r.Post("/", s.generateShortURL)
	r.Get("/{id}", s.getShortURL)

	return s, r
}
