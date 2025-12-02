package app

import (
	"database/sql"
	"fmt"
	"log"
	"migration-service/internal/config"
	"migration-service/internal/infrastructure/chat"
	httpHandler "migration-service/internal/infrastructure/http"
	"migration-service/internal/infrastructure/repository"
	"migration-service/internal/infrastructure/structure"
	"migration-service/internal/usecase"
	"net/http"

	_ "github.com/lib/pq"
)

// Server represents the migration service server
type Server struct {
	config *config.Config
	db     *sql.DB
	server *http.Server
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

	// Initialize external service clients
	chatService := chat.NewHTTPClient(s.config.Services.ChatServiceURL)
	structureService := structure.NewHTTPClient(s.config.Services.StructureServiceURL)

	// Initialize use cases
	databaseUseCase := usecase.NewMigrateFromDatabaseUseCase(
		nil, // sourceDB - would be configured separately
		jobRepo,
		errorRepo,
		universityRepo,
		chatService,
		nil, // logger
	)

	googleSheetsUseCase := usecase.NewMigrateFromGoogleSheetsUseCase(
		jobRepo,
		errorRepo,
		universityRepo,
		chatService,
		s.config.Google.CredentialsPath,
		nil, // logger
	)

	excelUseCase := usecase.NewMigrateFromExcelUseCase(
		jobRepo,
		errorRepo,
		structureService,
		chatService,
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
	if s.db != nil {
		s.db.Close()
	}
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}
