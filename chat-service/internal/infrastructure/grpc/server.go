package grpc

import (
	"chat-service/api/proto"
	"log"
	"net"

	"google.golang.org/grpc"
)

type Server struct {
	handler *ChatHandler
	port    string
}

func NewServer(handler *ChatHandler, port string) *Server {
	return &Server{
		handler: handler,
		port:    port,
	}
}

func (s *Server) Run() error {
	lis, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	proto.RegisterChatServiceServer(grpcServer, s.handler)

	log.Println("Starting gRPC server on port", s.port)
	return grpcServer.Serve(lis)
}

