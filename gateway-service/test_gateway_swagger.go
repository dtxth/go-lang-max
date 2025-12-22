package main

import (
	"log"
	"net/http"

	"gateway-service/internal/config"
	httpHandler "gateway-service/internal/infrastructure/http"
	grpcClient "gateway-service/internal/infrastructure/grpc"
)

func main() {
	// Load configuration
	cfg := config.Load()
	
	// Create a dummy client manager (won't be used for swagger)
	clientManager := grpcClient.NewClientManager(cfg)
	
	// Create HTTP router with all service endpoints
	router := httpHandler.NewRouter(cfg, clientManager)
	
	// Create HTTP server
	server := &http.Server{
		Addr:    ":8085",
		Handler: router,
	}
	
	log.Printf("Test Gateway Service starting on port 8085")
	log.Printf("Swagger UI: http://localhost:8085/swagger/")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}