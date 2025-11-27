package main

import (
	"log"

	"maxbot-service/internal/config"
	grpcServer "maxbot-service/internal/infrastructure/grpc"
	"maxbot-service/internal/infrastructure/maxapi"
	"maxbot-service/internal/usecase"
)

func main() {
	cfg := config.Load()

	apiClient := maxapi.NewClient(cfg.MaxAPIURL, cfg.MaxAPIToken, cfg.RequestTimeout)
	service := usecase.NewMaxBotService(apiClient)

	handler := grpcServer.NewMaxBotHandler(service)
	server := grpcServer.NewServer(handler, cfg.GRPCPort)

	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
