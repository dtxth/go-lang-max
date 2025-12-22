package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gateway-service/internal/config"
	grpcClient "gateway-service/internal/infrastructure/grpc"
	httpHandler "gateway-service/internal/infrastructure/http"
)

func main() {
	// Load configuration
	cfg := config.Load()
	
	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Initialize service registry
	serviceRegistry := grpcClient.NewServiceRegistry(cfg)
	serviceRegistry.RegisterServices()
	
	// Initialize gRPC client manager
	clientManager := grpcClient.NewClientManager(cfg)
	
	// Start gRPC clients (continue even if some fail)
	if err := clientManager.Start(ctx); err != nil {
		log.Printf("Warning: Failed to start some gRPC clients: %v", err)
		log.Printf("Gateway will continue running with limited functionality")
	}
	
	// Start health monitoring in background
	go serviceRegistry.StartHealthMonitoring(ctx, clientManager)
	
	// Create HTTP router with all service endpoints
	router := httpHandler.NewRouter(cfg, clientManager)
	
	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}
	
	// Start HTTP server in a goroutine
	go func() {
		log.Printf("Gateway Service starting on port %s", cfg.Server.Port)
		log.Printf("Swagger UI available at: http://localhost:%s/swagger/", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()
	
	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	
	log.Println("Shutting down Gateway Service...")
	
	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	
	// Stop HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}
	
	// Stop gRPC clients
	if err := clientManager.Stop(); err != nil {
		log.Printf("gRPC client manager shutdown error: %v", err)
	}
	
	log.Println("Gateway Service stopped")
}