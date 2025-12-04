package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"migration-service/internal/config"
	"migration-service/internal/domain"
	"migration-service/internal/infrastructure/chat"
	"migration-service/internal/infrastructure/grpc"
	httpHandler "migration-service/internal/infrastructure/http"
	"migration-service/internal/infrastructure/repository"
	"migration-service/internal/usecase"
	"net/http"

	_ "github.com/lib/pq"
)

// Server represents the migration service server
type Server struct {
	config          *config.Config
	db              *sql.DB
	server          *http.Server
	structureClient *grpc.StructureClient
}

// NewServer creates a new server instance
func NewServer(cfg *config.Config) (*Server, error) {
	// Connect to database
	db, err := sql.Open("postgres", cfg.Database.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Connected to database")

	return &Server{
		config: cfg,
		db:     db,
	}, nil
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Initialize repositories
	jobRepo := repository.NewMigrationJobPostgresRepository(s.db)
	errorRepo := repository.NewMigrationErrorPostgresRepository(s.db)
	universityRepo := repository.NewUniversityHTTPRepository(s.config.Services.StructureServiceURL)

	// Initialize HTTP client for Chat Service (для поддержки CreateOrGetUniversity)
	chatHTTPClient := chat.NewHTTPClient(s.config.Services.ChatServiceURL)

	// Initialize gRPC clients
	structureClient, err := grpc.NewStructureClient(s.config.Services.StructureServiceGRPC)
	if err != nil {
		return fmt.Errorf("failed to create structure client: %w", err)
	}
	s.structureClient = structureClient

	// Initialize gRPC client for Chat Service (для добавления администраторов без валидации)
	chatGRPCClient, err := grpc.NewChatClient(s.config.Services.ChatServiceGRPC)
	if err != nil {
		log.Printf("Warning: failed to create chat gRPC client: %v. Will use HTTP client.", err)
		chatGRPCClient = nil
	} else {
		log.Printf("Chat gRPC client initialized successfully at %s", s.config.Services.ChatServiceGRPC)
	}

	// Initialize use cases
	databaseUseCase := usecase.NewMigrateFromDatabaseUseCase(
		nil, // sourceDB - would be configured separately
		jobRepo,
		errorRepo,
		universityRepo,
		chatHTTPClient,
		nil, // logger
	)

	googleSheetsUseCase := usecase.NewMigrateFromGoogleSheetsUseCase(
		jobRepo,
		errorRepo,
		universityRepo,
		chatHTTPClient,
		s.config.Google.CredentialsPath,
		nil, // logger
	)

	// Используем gRPC клиент для администраторов, если доступен, иначе HTTP
	var chatClientForAdmins interface {
		CreateOrGetUniversity(ctx context.Context, university *domain.UniversityData) (int, error)
		CreateChat(ctx context.Context, chat *domain.ChatData) (int, error)
		AddAdministrator(ctx context.Context, admin *domain.AdministratorData) error
	}
	if chatGRPCClient != nil {
		// Создаем композитный клиент: gRPC для администраторов, HTTP для остального
		log.Println("Using composite client: gRPC for administrators, HTTP for other operations")
		chatClientForAdmins = &chat.CompositeClient{
			HTTPClient: chatHTTPClient,
			GRPCClient: chatGRPCClient,
		}
	} else {
		log.Println("Using HTTP client for all chat operations (gRPC not available)")
		chatClientForAdmins = chatHTTPClient
	}

	excelUseCase := usecase.NewMigrateFromExcelUseCase(
		jobRepo,
		errorRepo,
		structureClient,
		chatClientForAdmins,
		nil, // logger
	)

	// Initialize HTTP handler
	handler := httpHandler.NewHandler(
		databaseUseCase,
		googleSheetsUseCase,
		excelUseCase,
		jobRepo,
		errorRepo,
	)

	// Setup routes
	mux := httpHandler.SetupRoutes(handler)

	// Create HTTP server
	s.server = &http.Server{
		Addr:    ":" + s.config.Server.Port,
		Handler: mux,
	}

	log.Printf("Starting migration service on port %s", s.config.Server.Port)
	return s.server.ListenAndServe()
}

// Stop stops the server
func (s *Server) Stop() error {
	if s.structureClient != nil {
		s.structureClient.Close()
	}
	if s.db != nil {
		s.db.Close()
	}
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}
