package server

import (
	"ms-notification/internal/config"
	"net"

	"google.golang.org/grpc"
)

type Server struct {
	grpcServer   *grpc.Server
	grpcListener net.Listener
}

func NewServer(cfg *config.HTTPConfig) (*Server, error) {
	listener, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		return &Server{}, err
	}

	grpcSrv := grpc.NewServer()

	return &Server{
		grpcServer:   grpcSrv,
		grpcListener: listener,
	}, nil
}

func (s *Server) Run() error {
	return s.grpcServer.Serve(s.grpcListener)
}

func (s *Server) Stop() {
	s.grpcServer.GracefulStop()
}

func (s *Server) GRPCServer() *grpc.Server {
	return s.grpcServer
}
