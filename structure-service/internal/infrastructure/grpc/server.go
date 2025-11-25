package grpc

import (
	"log"
	"net"
	structurepb "structure-service/api/proto"

	"google.golang.org/grpc"
)

type Server struct {
	handler *StructureHandler
	port    string
}

func NewServer(handler *StructureHandler, port string) *Server {
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
	structurepb.RegisterStructureServiceServer(grpcServer, s.handler)

	log.Println("Starting gRPC server on port", s.port)
	return grpcServer.Serve(lis)
}
