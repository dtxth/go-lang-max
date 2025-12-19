package main

import (
	"database/sql"
	"employee-service/internal/app"
	"employee-service/internal/config"
	"employee-service/internal/infrastructure/auth"
	"employee-service/internal/infrastructure/database"
	"employee-service/internal/infrastructure/grpc"
	"employee-service/internal/infrastructure/http"
	"employee-service/internal/infrastructure/logger"
	"employee-service/internal/infrastructure/max"
	"employee-service/internal/infrastructure/migration"
	"employee-service/internal/infrastructure/notification"
	"employee-service/internal/infrastructure/password"
	"employee-service/internal/infrastructure/profile"
	"employee-service/internal/infrastructure/repository"
	"employee-service/internal/usecase"
	"log"
	"os"

	_ "github.com/lib/pq"

	// swagger docs
	_ "employee-service/internal/infrastructure/http/docs"
)

// @title           Employee Service API
// @version         1.0
// @description     Сервис управления сотрудниками вузов для мини-приложения
// @BasePath        /
// @schemes         http https
func main() {
	cfg := config.Load()

	// Инициализируем logger
	appLogger := logger.New(os.Stdout, logger.INFO)
	log.Println("Starting employee-service server on port", cfg.Port)
	log.Println("Starting gRPC server on port", cfg.GRPCPort)

	// Initialize database connection with automatic reconnection
	dbLogger := log.New(os.Stdout, "[DB] ", log.LstdFlags)
	db := database.NewDB(cfg.DBUrl, dbLogger)
	
	if err := db.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize and run migrations with separate connection
	migrationDB, err := sql.Open("postgres", cfg.DBUrl)
	if err != nil {
		panic(err)
	}
	
	migrator := migration.NewMigrator(migrationDB, log.New(os.Stdout, "[MIGRATION] ", log.LstdFlags))
	
	// Wait for database to be ready
	if err := migrator.WaitForDatabase(); err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	
	// Run migrations
	if err := migrator.RunMigrations(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	
	// Close migration connection
	migrationDB.Close()

	// Инициализируем репозитории
	employeeRepo := repository.NewEmployeePostgresWithDSN(db, cfg.DBUrl)
	universityRepo := repository.NewUniversityPostgresWithDSN(db, cfg.DBUrl)
	batchUpdateJobRepo := repository.NewBatchUpdateJobPostgresWithDSN(db, cfg.DBUrl)

	// Инициализируем MAX gRPC клиент
	maxClient, err := max.NewMaxClient(cfg.MaxBotAddress, cfg.MaxBotTimeout)
	if err != nil {
		panic(err)
	}
	defer maxClient.Close()

	// Инициализируем Profile Cache gRPC клиент (используем тот же адрес что и MaxBot)
	profileCacheClient := profile.NewProfileCacheClient(maxClient.GetConnection())

	// Инициализируем Auth gRPC клиент
	log.Printf("Connecting to Auth Service at %s", cfg.AuthServiceAddress)
	authClient, err := auth.NewAuthClient(cfg.AuthServiceAddress)
	if err != nil {
		log.Printf("ERROR: Failed to connect to auth service: %v", err)
		log.Fatal("Auth service is required for employee service to function properly")
	}
	log.Println("Successfully connected to Auth Service")
	defer authClient.Close()

	// Инициализируем password generator
	passwordGenerator := password.NewSecurePasswordGenerator(12)

	// Инициализируем notification service
	notificationService, err := notification.NewMaxNotificationService(cfg.MaxBotAddress, log.Default())
	if err != nil {
		log.Printf("WARNING: Failed to initialize notification service: %v", err)
		log.Println("Password notifications will not be sent")
	}
	if notificationService != nil {
		defer notificationService.Close()
	}

	// Инициализируем usecase
	employeeService := usecase.NewEmployeeService(employeeRepo, universityRepo, maxClient, authClient, passwordGenerator, notificationService, profileCacheClient)
	batchUpdateMaxIdUseCase := usecase.NewBatchUpdateMaxIdUseCase(employeeRepo, batchUpdateJobRepo, maxClient)
	
	// Инициализируем use case для поиска с ролевой фильтрацией
	var searchEmployeesWithRoleFilterUC *usecase.SearchEmployeesWithRoleFilterUseCase
	if authClient != nil {
		searchEmployeesWithRoleFilterUC = usecase.NewSearchEmployeesWithRoleFilterUseCase(employeeRepo, authClient)
	}

	// Инициализируем HTTP handler с logger
	handler := http.NewHandler(employeeService, batchUpdateMaxIdUseCase, searchEmployeesWithRoleFilterUC, authClient, appLogger)

	// HTTP server
	httpServer := &app.Server{
		Handler: handler.Router(),
		Port:    cfg.Port,
	}

	// gRPC server
	grpcHandler := grpc.NewEmployeeHandler(employeeService)
	grpcServer := grpc.NewServer(grpcHandler, cfg.GRPCPort)

	// Запускаем оба сервера
	go func() {
		if err := grpcServer.Run(); err != nil {
			panic(err)
		}
	}()

	httpServer.Run()
}
