package main

import (
	"database/sql"
	"employee-service/internal/app"
	"employee-service/internal/config"
	"employee-service/internal/infrastructure/auth"
	"employee-service/internal/infrastructure/grpc"
	"employee-service/internal/infrastructure/http"
	"employee-service/internal/infrastructure/logger"
	"employee-service/internal/infrastructure/max"
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

	db, err := sql.Open("postgres", cfg.DBUrl)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Проверяем подключение к БД
	if err := db.Ping(); err != nil {
		panic(err)
	}

	// Инициализируем репозитории
	employeeRepo := repository.NewEmployeePostgres(db)
	universityRepo := repository.NewUniversityPostgres(db)
	batchUpdateJobRepo := repository.NewBatchUpdateJobPostgres(db)

	// Инициализируем MAX gRPC клиент
	maxClient, err := max.NewMaxClient(cfg.MaxBotAddress, cfg.MaxBotTimeout)
	if err != nil {
		panic(err)
	}
	defer maxClient.Close()

	// Инициализируем Auth gRPC клиент
	authClient, err := auth.NewAuthClient(cfg.AuthServiceAddress)
	if err != nil {
		log.Printf("Warning: Failed to connect to auth service: %v", err)
		authClient = nil
	}
	if authClient != nil {
		defer authClient.Close()
	}

	// Инициализируем usecase
	employeeService := usecase.NewEmployeeService(employeeRepo, universityRepo, maxClient)
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
