package main

import (
	"log"
	"migration-service/internal/app"
	"migration-service/internal/config"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Create server
	server, err := app.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down server...")
		if err := server.Stop(); err != nil {
			log.Printf("Error stopping server: %v", err)
		}
		os.Exit(0)
	}()

	// Start server
	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
