package server

import (
	"context"
	"ms-notification/internal/config"
	"net/http"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(cfgHTTP *config.HTTPConfig, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:              ":" + cfgHTTP.Port,
			Handler:           handler,
			ReadTimeout:       cfgHTTP.ReadTimeout,
			ReadHeaderTimeout: cfgHTTP.ReadHeaderTimeout,
			WriteTimeout:      cfgHTTP.WriteTimeout,
		},
	}
}

func (s *Server) Run() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
