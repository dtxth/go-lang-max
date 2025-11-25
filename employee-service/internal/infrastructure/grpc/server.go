package grpc

import (
	"employee-service/api/proto"
	"log"
	"net"

	"google.golang.org/grpc"
)

type Server struct {
	handler *EmployeeHandler
	port    string
}

func NewServer(handler *EmployeeHandler, port string) *Server {
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
	proto.RegisterEmployeeServiceServer(grpcServer, s.handler)

	log.Println("Starting gRPC server on port", s.port)
	return grpcServer.Serve(lis)
}

