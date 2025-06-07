package server

import (
	"context"
	"net/http"
	"weather_forecast_sub/internal/config"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(cfg *config.Config, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:              ":" + cfg.HTTP.Port,
			Handler:           handler,
			ReadTimeout:       cfg.HTTP.ReadTimeout,
			ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
			WriteTimeout:      cfg.HTTP.WriteTimeout,
		},
	}
}

func (s *Server) Run() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
