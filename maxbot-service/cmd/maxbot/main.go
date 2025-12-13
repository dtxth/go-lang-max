package main

import (
	"log"

	"maxbot-service/internal/config"
	"maxbot-service/internal/domain"
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

	// Initialize Max API client (real or mock)
	var apiClient domain.MaxAPIClient

	if cfg.MockMode {
		log.Println("Running in MOCK MODE - using mock Max API client")
		apiClient = maxapi.NewMockClient()
	} else {
		// Validate required configuration for real API
		if cfg.MaxAPIToken == "" {
			log.Fatal("MAX_API_TOKEN environment variable is required but not set. Please configure the bot token or enable MOCK_MODE=true.")
		}
		log.Printf("MAX_API_TOKEN validated (length: %d characters)", len(cfg.MaxAPIToken))

		// Initialize real Max API client
		log.Println("Initializing Max API client...")
		realClient, clientErr := maxapi.NewClient(cfg.MaxAPIURL, cfg.MaxAPIToken, cfg.RequestTimeout)
		if clientErr != nil {
			log.Fatalf("Failed to initialize Max API client: %v. Please check your MAX_API_TOKEN and MAX_API_URL configuration.", clientErr)
		}
		apiClient = realClient
		log.Println("Max API client initialized successfully")
	}

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
