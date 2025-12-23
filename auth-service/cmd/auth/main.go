package main

import (
	"auth-service/internal/app"
	"auth-service/internal/config"
	"auth-service/internal/domain"
	"auth-service/internal/infrastructure/cleanup"
	"auth-service/internal/infrastructure/database"
	"auth-service/internal/infrastructure/employee"
	"auth-service/internal/infrastructure/grpc"
	"auth-service/internal/infrastructure/hash"
	"auth-service/internal/infrastructure/http"
	"auth-service/internal/infrastructure/jwt"
	"auth-service/internal/infrastructure/logger"
	"auth-service/internal/infrastructure/max"
	"auth-service/internal/infrastructure/maxbot"
	"auth-service/internal/infrastructure/metrics"
	"auth-service/internal/infrastructure/migration"
	"auth-service/internal/infrastructure/notification"
	"auth-service/internal/infrastructure/repository"
	"auth-service/internal/usecase"
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"

	// swagger docs
	_ "auth-service/internal/infrastructure/http/docs"
)

// @title           Auth Service API
// @version         1.0
// @description     Authentication service with JWT access & refresh tokens
// @BasePath        /
func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

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

	repo := repository.NewUserPostgres(db)
	refreshRepo := repository.NewRefreshPostgres(db)
	userRoleRepo := repository.NewUserRolePostgres(db)
	passwordResetRepo := repository.NewPasswordResetPostgres(db)
	hasher := hash.NewBcryptHasher()
	jwtManager := jwt.NewManager(cfg.AccessSecret, cfg.RefreshSecret, 1*time.Hour, 7*24*time.Hour)
	
	// Initialize MAX auth validator
	maxAuthValidator := max.NewAuthValidator()
	
	// Initialize logger
	appLogger := logger.NewDefault()
	
	// Initialize metrics
	metricsCollector := metrics.NewMetrics()
	log.Printf("Initialized metrics collector")
	
	// Initialize notification service based on configuration
	var notificationSvc domain.NotificationService
	if cfg.NotificationServiceType == "max" {
		maxService, err := notification.NewMaxNotificationService(cfg.MaxBotServiceAddr, appLogger)
		if err != nil {
			log.Fatalf("Failed to initialize MAX notification service: %v", err)
		}
		// Wrap with metrics
		notificationSvc = notification.NewMetricsWrapper(maxService, metricsCollector)
		log.Printf("Initialized MAX notification service (MaxBot: %s)", cfg.MaxBotServiceAddr)
	} else {
		mockService := notification.NewMockNotificationService(appLogger)
		// Wrap with metrics
		notificationSvc = notification.NewMetricsWrapper(mockService, metricsCollector)
		log.Printf("Initialized MOCK notification service")
	}

	authUC := usecase.NewAuthService(repo, refreshRepo, hasher, jwtManager, userRoleRepo)
	
	// Set MAX authentication configuration
	authUC.SetMaxAuthValidator(maxAuthValidator)
	authUC.SetMaxBotToken(cfg.MaxBotToken)
	
	// Set password configuration
	authUC.SetPasswordConfig(cfg.MinPasswordLength, time.Duration(cfg.ResetTokenExpiration)*time.Minute)
	
	// Set optional dependencies
	authUC.SetPasswordResetRepository(passwordResetRepo)
	authUC.SetNotificationService(notificationSvc)
	authUC.SetLogger(appLogger)
	authUC.SetMetrics(metricsCollector)
	
	// Initialize MaxBot client if configured
	if cfg.MaxBotServiceAddr != "" {
		maxBotClient, err := maxbot.NewClient(cfg.MaxBotServiceAddr)
		if err != nil {
			log.Printf("Warning: Failed to initialize MaxBot client: %v", err)
			// Use mock client as fallback
			authUC.SetMaxBotClient(maxbot.NewMockClient())
			log.Printf("Using mock MaxBot client as fallback")
		} else {
			authUC.SetMaxBotClient(maxBotClient)
			log.Printf("Initialized MaxBot client (addr: %s)", cfg.MaxBotServiceAddr)
		}
	} else {
		// Use mock client when no address is configured
		authUC.SetMaxBotClient(maxbot.NewMockClient())
		log.Printf("Using mock MaxBot client (no address configured)")
	}
	
	// Initialize Employee client if configured
	if cfg.EmployeeServiceAddr != "" {
		employeeClient := employee.NewClient(cfg.EmployeeServiceAddr)
		authUC.SetEmployeeClient(employeeClient)
		log.Printf("Initialized Employee client (addr: %s)", cfg.EmployeeServiceAddr)
	} else {
		log.Printf("Employee service not configured - employee data updates disabled")
	}
	
	handler := http.NewHandler(authUC)

	// HTTP server
	httpServer := &app.Server{
		Handler: handler.Router(),
		Port:    cfg.Port,
	}

	// gRPC server
	grpcHandler := grpc.NewAuthHandler(authUC)
	grpcServer := grpc.NewServer(grpcHandler, cfg.GRPCPort)

	// Token cleanup job
	cleanupInterval := time.Duration(cfg.TokenCleanupInterval) * time.Minute
	cleanupLogger := log.New(os.Stdout, "[CLEANUP] ", log.LstdFlags)
	cleanupJob := cleanup.NewTokenCleanupJob(passwordResetRepo, cleanupInterval, cleanupLogger)
	ctx := context.Background()
	
	// Start cleanup job in background
	go cleanupJob.Start(ctx)

	// Запускаем оба сервера
	go func() {
		if err := grpcServer.Run(); err != nil {
			panic(err)
		}
	}()

	httpServer.Run()
}
