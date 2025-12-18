package main

import (
	"database/sql"
	"log"
	"os"
	"structure-service/internal/app"
	"structure-service/internal/config"
	"structure-service/internal/infrastructure/employee"
	"structure-service/internal/infrastructure/grpc"
	"structure-service/internal/infrastructure/http"
	"structure-service/internal/infrastructure/logger"
	"structure-service/internal/infrastructure/migration"
	"structure-service/internal/infrastructure/repository"
	"structure-service/internal/usecase"

	_ "github.com/lib/pq"

	// swagger docs
	_ "structure-service/internal/infrastructure/http/docs"
)

// @title           Structure Service API
// @version         1.0
// @description     Service for managing university structure hierarchy
// @BasePath        /
func main() {
	cfg := config.Load()

	// Инициализируем logger
	appLogger := logger.New(os.Stdout, logger.INFO)
	log.Println("Starting structure-service server on port", cfg.Port)
	log.Println("Starting gRPC server on port", cfg.GRPCPort)

	db, err := sql.Open("postgres", cfg.DBUrl)
	if err != nil {
		panic(err)
	}

	// Initialize and run migrations
	migrator := migration.NewMigrator(db, log.New(os.Stdout, "[MIGRATION] ", log.LstdFlags))
	
	// Wait for database to be ready
	if err := migrator.WaitForDatabase(); err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	
	// Run migrations
	if err := migrator.RunMigrations(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	repo := repository.NewStructurePostgresWithDSN(db, cfg.DBUrl)
	dmRepo := repository.NewDepartmentManagerPostgres(db)
	
	// Инициализируем gRPC клиент для chat-service
	chatClient, err := grpc.NewChatClient(cfg.ChatService)
	if err != nil {
		panic(err)
	}
	defer chatClient.Close()

	// Инициализируем gRPC клиент для employee-service
	employeeClient, err := employee.NewEmployeeClient(cfg.EmployeeService)
	if err != nil {
		panic(err)
	}
	defer employeeClient.Close()

	// Create chat service adapter
	chatServiceAdapter := grpc.NewChatServiceAdapter(chatClient)

	structureUC := usecase.NewStructureService(repo)
	getUniversityStructureUC := usecase.NewGetUniversityStructureUseCase(repo, chatServiceAdapter)
	assignOperatorUC := usecase.NewAssignOperatorToDepartmentUseCase(dmRepo, employeeClient)
	importStructureUC := usecase.NewImportStructureFromExcelUseCase(repo, db)
	createStructureUC := usecase.NewCreateStructureFromRowUseCase(repo)
	handler := http.NewHandler(structureUC, getUniversityStructureUC, assignOperatorUC, importStructureUC, createStructureUC, dmRepo, appLogger)

	// HTTP server
	httpServer := &app.Server{
		Handler: handler.Router(),
		Port:    cfg.Port,
	}

	// gRPC server
	grpcHandler := grpc.NewStructureHandler(structureUC, createStructureUC)
	grpcServer := grpc.NewServer(grpcHandler, cfg.GRPCPort)

	// Запускаем оба сервера
	go func() {
		if err := grpcServer.Run(); err != nil {
			panic(err)
		}
	}()

	httpServer.Run()
}

