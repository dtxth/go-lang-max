package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"bufio"
	"strings"

	_ "maxbot-service/docs" // Import swagger docs
	"maxbot-service/internal/config"
	"maxbot-service/internal/domain"
	"maxbot-service/internal/infrastructure/cache"
	grpcServer "maxbot-service/internal/infrastructure/grpc"
	httpServer "maxbot-service/internal/infrastructure/http"
	"maxbot-service/internal/infrastructure/maxapi"
	"maxbot-service/internal/infrastructure/monitoring"
	"maxbot-service/internal/usecase"
)

func main() {
	log.Println("Starting MaxBot Service...")

	// Load .env file if it exists
	loadEnvFile()

	// Load configuration
	cfg := config.Load()
	log.Printf("Configuration loaded - GRPC Port: %s, HTTP Port: %s, Max API URL: %s, Request Timeout: %s",
		cfg.GRPCPort, cfg.HTTPPort, cfg.MaxAPIURL, cfg.RequestTimeout)

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

	// Initialize profile cache service
	log.Println("Initializing profile cache service...")
	var profileCache domain.ProfileCacheService
	var redisClient *cache.RedisClient
	
	if cfg.MockMode {
		log.Println("Using mock profile cache service")
		profileCache = cache.NewMockProfileCache()
	} else {
		log.Println("Initializing Redis profile cache...")
		redisCache, client, err := cache.NewProfileCacheServiceWithClient(cfg)
		if err != nil {
			log.Printf("Failed to initialize Redis cache, using mock cache: %v", err)
			profileCache = cache.NewMockProfileCache()
		} else {
			profileCache = redisCache
			redisClient = client
			log.Println("Redis profile cache initialized successfully")
		}
	}

	// Initialize monitoring service
	log.Println("Initializing monitoring service...")
	var monitoringService domain.MonitoringService
	
	if cfg.MockMode || redisClient == nil {
		log.Println("Using mock monitoring service")
		monitoringService = monitoring.NewMockMonitoringService()
	} else {
		log.Println("Initializing Redis monitoring service...")
		monitoringService = monitoring.NewRedisMonitoringService(redisClient.Client, profileCache)
		log.Println("Redis monitoring service initialized successfully")
	}

	// Initialize service layer
	log.Println("Initializing MaxBot service...")
	service := usecase.NewMaxBotService(apiClient)
	
	// Initialize webhook handler
	log.Println("Initializing webhook handler...")
	webhookHandler := usecase.NewWebhookHandlerService(profileCache, monitoringService)
	
	// Initialize profile management service
	log.Println("Initializing profile management service...")
	profileManagement := usecase.NewProfileManagementService(profileCache, apiClient)

	// Initialize servers
	log.Println("Initializing servers...")
	
	// gRPC server
	grpcHandler := grpcServer.NewMaxBotHandler(service)
	grpcSrv := grpcServer.NewServer(grpcHandler, cfg.GRPCPort)
	
	// HTTP server
	httpHandler := httpServer.NewMaxBotHTTPHandler(service, webhookHandler, profileManagement, monitoringService)
	httpSrv := httpServer.NewServer(httpHandler, cfg.HTTPPort)

	// Start servers concurrently
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	// Start gRPC server
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("Starting gRPC server on port %s", cfg.GRPCPort)
		if err := grpcSrv.Run(); err != nil {
			log.Printf("gRPC server error: %v", err)
			cancel()
		}
	}()

	// Start HTTP server
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("Starting HTTP server on port %s", cfg.HTTPPort)
		log.Printf("Swagger documentation available at: http://localhost:%s/swagger/index.html", cfg.HTTPPort)
		if err := httpSrv.Run(); err != nil {
			log.Printf("HTTP server error: %v", err)
			cancel()
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		log.Printf("Received signal %v, shutting down...", sig)
	case <-ctx.Done():
		log.Println("Context cancelled, shutting down...")
	}

	// Graceful shutdown
	log.Println("Shutting down servers...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server
	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	// Note: gRPC server doesn't have graceful shutdown in our current implementation
	// In production, you might want to add graceful shutdown for gRPC as well

	log.Println("MaxBot Service stopped")
}

// loadEnvFile loads environment variables from .env file if it exists
func loadEnvFile() {
	// Try to load .env from current directory or parent directory
	envPaths := []string{".env", "../.env"}
	
	for _, envPath := range envPaths {
		if file, err := os.Open(envPath); err == nil {
			defer file.Close()
			log.Printf("Loading environment variables from %s", envPath)
			
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				
				// Skip empty lines and comments
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				
				// Parse KEY=VALUE format
				if parts := strings.SplitN(line, "=", 2); len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					
					// Only set if not already set in environment
					if os.Getenv(key) == "" {
						os.Setenv(key, value)
						// Don't log sensitive values
						if strings.Contains(strings.ToLower(key), "token") || strings.Contains(strings.ToLower(key), "secret") {
							log.Printf("Set %s=***", key)
						} else {
							log.Printf("Set %s=%s", key, value)
						}
					}
				}
			}
			
			if err := scanner.Err(); err != nil {
				log.Printf("Error reading %s: %v", envPath, err)
			}
			return // Stop after first successful load
		}
	}
	
	log.Println("No .env file found, using system environment variables only")
}