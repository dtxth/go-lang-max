package main

import (
	"chat-service/internal/app"
	"chat-service/internal/config"
	"chat-service/internal/infrastructure/grpc"
	"chat-service/internal/infrastructure/http"
	"chat-service/internal/infrastructure/max"
	"chat-service/internal/infrastructure/repository"
	"chat-service/internal/usecase"
	"database/sql"

	_ "github.com/lib/pq"

	// swagger docs
	_ "chat-service/internal/infrastructure/http/docs"
)

// @title           Chat Service API
// @version         1.0
// @description     Сервис управления групповыми чатами для мини-приложения
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
	chatRepo := repository.NewChatPostgres(db)
	administratorRepo := repository.NewAdministratorPostgres(db)
	universityRepo := repository.NewUniversityPostgres(db)

	// Инициализируем MAX клиент
	maxClient := max.NewMaxClient()

	// Инициализируем usecase
	chatService := usecase.NewChatService(chatRepo, administratorRepo, universityRepo, maxClient)

	// Инициализируем HTTP handler
	handler := http.NewHandler(chatService)

	// HTTP server
	httpServer := &app.Server{
		Handler: handler.Router(),
		Port:    cfg.Port,
	}

	// gRPC server
	grpcHandler := grpc.NewChatHandler(chatService)
	grpcServer := grpc.NewServer(grpcHandler, cfg.GRPCPort)

	// Запускаем оба сервера
	go func() {
		if err := grpcServer.Run(); err != nil {
			panic(err)
		}
	}()

	httpServer.Run()
}

