package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	_ "maxbot-service/docs" // Import swagger docs
	"maxbot-service/internal/config"
	"maxbot-service/internal/domain"
	"maxbot-service/internal/infrastructure/maxapi"
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

	// Initialize service layer
	log.Println("Initializing MaxBot service...")
	service := usecase.NewMaxBotService(apiClient)

	// Create working HTTP server with proper routing
	log.Println("Creating HTTP server with proper routing...")
	
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request: %s %s", r.Method, r.URL.Path)
		
		// Health check
		if r.URL.Path == "/health" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"status":"ok","service":"maxbot-service"}`)
			return
		}
		
		// Bot info
		if r.URL.Path == "/api/v1/me" && r.Method == "GET" {
			botInfo, err := service.GetMe(r.Context())
			if err != nil {
				log.Printf("Error getting bot info: %v", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, `{"error":"internal_error"}`)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"name":"%s","add_link":"%s"}`, botInfo.Name, botInfo.AddLink)
			return
		}
		
		// Chat endpoint
		if strings.HasPrefix(r.URL.Path, "/api/v1/chats/") && r.Method == "GET" {
			log.Printf("Chat endpoint called: %s", r.URL.Path)
			path := strings.TrimPrefix(r.URL.Path, "/api/v1/chats/")
			if path == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprint(w, `{"error":"chat_id required"}`)
				return
			}
			
			var chatID int64
			if _, err := fmt.Sscanf(path, "%d", &chatID); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprint(w, `{"error":"invalid chat_id"}`)
				return
			}
			
			log.Printf("Getting chat info for chatID: %d", chatID)
			chatInfo, err := service.GetChatInfo(r.Context(), chatID)
			if err != nil {
				log.Printf("Error getting chat info: %v", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprint(w, `{"error":"chat not found","message":"`+err.Error()+`"}`)
				return
			}
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"chat_id":%d,"title":"%s","type":"%s","participants_count":%d}`, 
				chatInfo.ChatID, chatInfo.Title, chatInfo.Type, chatInfo.ParticipantsCount)
			return
		}
		
		// Unknown endpoint
		log.Printf("Unknown endpoint: %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "404 page not found")
	})
	
	httpSrv := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: mux,
	}
	
	log.Printf("HTTP server created, starting on port %s", cfg.HTTPPort)
	log.Fatal(httpSrv.ListenAndServe())
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