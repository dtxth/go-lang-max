package main

import (
	"log"

	"maxbot-service/internal/config"
	grpcServer "maxbot-service/internal/infrastructure/grpc"
	"maxbot-service/internal/infrastructure/maxapi"
	"maxbot-service/internal/usecase"
)

func main() {
	log.Println("Starting MaxBot Service...")

	// Load configuration
	cfg := config.Load()
	log.Printf("Configuration loaded - GRPC Port: %s, Max API URL: %s, Request Timeout: %s",
		cfg.GRPCPort, cfg.MaxAPIURL, cfg.RequestTimeout)

	// Validate required configuration
	if cfg.MaxAPIToken == "" {
		log.Fatal("MAX_API_TOKEN environment variable is required but not set. Please configure the bot token.")
	}
	log.Printf("MAX_API_TOKEN validated (length: %d characters)", len(cfg.MaxAPIToken))

	// Initialize Max API client
	log.Println("Initializing Max API client...")
	apiClient, err := maxapi.NewClient(cfg.MaxAPIURL, cfg.MaxAPIToken, cfg.RequestTimeout)
	if err != nil {
		log.Fatalf("Failed to initialize Max API client: %v. Please check your MAX_API_TOKEN and MAX_API_URL configuration.", err)
	}
	log.Println("Max API client initialized successfully")

	// Initialize service layer
	log.Println("Initializing MaxBot service...")
	service := usecase.NewMaxBotService(apiClient)

	// Initialize gRPC handler and server
	log.Println("Initializing gRPC server...")
	handler := grpcServer.NewMaxBotHandler(service)
	server := grpcServer.NewServer(handler, cfg.GRPCPort)

	// Start gRPC server
	log.Printf("Starting gRPC server on port %s", cfg.GRPCPort)
	if err := server.Run(); err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}
}
