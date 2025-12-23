package grpc

import (
	"log"
	"net"

	maxbotproto "maxbot-service/api/proto/maxbotproto"

	"google.golang.org/grpc"
)

type Server struct {
	handler *MaxBotHandler
	port    string
}

func NewServer(handler *MaxBotHandler, port string) *Server {
	return &Server{handler: handler, port: port}
}

func (s *Server) Run() error {
	lis, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	maxbotproto.RegisterMaxBotServiceServer(grpcServer, s.handler)

	log.Printf("Starting maxbot gRPC server on port %s", s.port)
	return grpcServer.Serve(lis)
}
