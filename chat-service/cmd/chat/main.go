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
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	chatService := usecase.NewChatService(chatRepo, administratorRepo, maxClient)

	// Инициализируем participants integration если Redis доступен
	var participantsIntegration *app.ParticipantsIntegration
	if app.IsParticipantsIntegrationEnabled() {
		participantsIntegration, err = app.NewParticipantsIntegration(chatRepo, maxClient, appLogger)
		if err != nil {
			appLogger.Error(context.Background(), "Failed to initialize participants integration", map[string]interface{}{
				"error": err.Error(),
			})
			log.Printf("Warning: Participants integration disabled due to error: %v", err)
		} else {
			appLogger.Info(context.Background(), "Participants integration initialized successfully", nil)
		}
	} else {
		appLogger.Info(context.Background(), "Participants integration disabled (Redis not available or explicitly disabled)", nil)
	}

	// Инициализируем middleware
	authMiddleware := http.NewAuthMiddleware(authClient)

	// Инициализируем HTTP handler с logger
	handler := http.NewHandler(chatService, authMiddleware, appLogger)

	// HTTP server
	httpServer := &app.Server{
		Handler:                 handler.Router(),
		Port:                    cfg.Port,
		ParticipantsIntegration: participantsIntegration,
	}

	// gRPC server
	grpcHandler := grpc.NewChatHandler(chatService)
	grpcServer := grpc.NewServer(grpcHandler, cfg.GRPCPort)

	// Настраиваем graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Канал для получения сигналов ОС
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем gRPC сервер в горутине
	go func() {
		if err := grpcServer.Run(); err != nil {
			appLogger.Error(ctx, "gRPC server error", map[string]interface{}{
				"error": err.Error(),
			})
			cancel()
		}
	}()

	// Запускаем HTTP сервер в горутине
	go func() {
		if err := httpServer.Start(); err != nil {
			appLogger.Error(ctx, "HTTP server error", map[string]interface{}{
				"error": err.Error(),
			})
			cancel()
		}
	}()

	// Ждем сигнал завершения или ошибку
	select {
	case sig := <-sigChan:
		appLogger.Info(ctx, "Received shutdown signal", map[string]interface{}{
			"signal": sig.String(),
		})
	case <-ctx.Done():
		appLogger.Info(ctx, "Context cancelled, shutting down", nil)
	}

	// Graceful shutdown
	appLogger.Info(ctx, "Starting graceful shutdown", nil)
	
	// Создаем контекст с таймаутом для shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Останавливаем сервер
	if err := httpServer.Stop(shutdownCtx); err != nil {
		appLogger.Error(shutdownCtx, "Error during server shutdown", map[string]interface{}{
			"error": err.Error(),
		})
	}

	appLogger.Info(shutdownCtx, "Server shutdown completed", nil)
}
