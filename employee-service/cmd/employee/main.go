package main

import (
	"database/sql"
	"employee-service/internal/app"
	"employee-service/internal/config"
	"employee-service/internal/infrastructure/grpc"
	"employee-service/internal/infrastructure/http"
	"employee-service/internal/infrastructure/max"
	"employee-service/internal/infrastructure/repository"
	"employee-service/internal/usecase"

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

	// Инициализируем MAX клиент
	maxClient := max.NewMaxClient()

	// Инициализируем usecase
	employeeService := usecase.NewEmployeeService(employeeRepo, universityRepo, maxClient)

	// Инициализируем HTTP handler
	handler := http.NewHandler(employeeService)

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

