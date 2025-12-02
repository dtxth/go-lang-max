package main

import (
	"chat-service/internal/app"
	"chat-service/internal/config"
	"chat-service/internal/infrastructure/auth"
	"chat-service/internal/infrastructure/grpc"
	"chat-service/internal/infrastructure/http"
	"chat-service/internal/infrastructure/logger"
	"chat-service/internal/infrastructure/max"
	"chat-service/internal/infrastructure/repository"
	"chat-service/internal/usecase"
	"database/sql"
	"log"
	"os"

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

	// Инициализируем logger
	appLogger := logger.New(os.Stdout, logger.INFO)
	log.Println("Starting chat-service server on port", cfg.Port)
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
	chatRepo := repository.NewChatPostgres(db)
	administratorRepo := repository.NewAdministratorPostgres(db)
	universityRepo := repository.NewUniversityPostgres(db)

	// Инициализируем MAX gRPC клиент
	maxClient, err := max.NewMaxClient(cfg.MaxBotAddress, cfg.MaxBotTimeout)
	if err != nil {
		panic(err)
	}
	defer maxClient.Close()

	// Инициализируем Auth gRPC клиент
	authClient, err := auth.NewAuthClient(cfg.AuthAddress, cfg.AuthTimeout)
	if err != nil {
		panic(err)
	}
	defer authClient.Close()

	// Инициализируем usecase
	chatService := usecase.NewChatService(chatRepo, administratorRepo, universityRepo, maxClient)

	// Инициализируем middleware
	authMiddleware := http.NewAuthMiddleware(authClient)

	// Инициализируем HTTP handler с logger
	handler := http.NewHandler(chatService, authMiddleware, appLogger)

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
