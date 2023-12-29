package server

import (
	"net/http"
)

type Server struct {
	httpServer *http.Server
}

func (s *Server) Run() error {
	return s.httpServer.ListenAndServe()
}

func New(serverAddress string, h *Handler) *Server {
	s := &Server{
		httpServer: &http.Server{
			Addr:    serverAddress,
			Handler: h,
		},
	}

	return s
}
