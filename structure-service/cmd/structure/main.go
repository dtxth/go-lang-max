package main

import (
	"database/sql"
	"structure-service/internal/app"
	"structure-service/internal/config"
	"structure-service/internal/infrastructure/grpc"
	"structure-service/internal/infrastructure/http"
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

	db, err := sql.Open("postgres", cfg.DBUrl)
	if err != nil {
		panic(err)
	}

	repo := repository.NewStructurePostgres(db)
	
	// Инициализируем gRPC клиент для chat-service
	chatClient, err := grpc.NewChatClient(cfg.ChatService)
	if err != nil {
		panic(err)
	}
	defer chatClient.Close()

	structureUC := usecase.NewStructureService(repo)
	handler := http.NewHandler(structureUC)

	// HTTP server
	httpServer := &app.Server{
		Handler: handler.Router(),
		Port:    cfg.Port,
	}

	// gRPC server
	grpcHandler := grpc.NewStructureHandler(structureUC)
	grpcServer := grpc.NewServer(grpcHandler, cfg.GRPCPort)

	// Запускаем оба сервера
	go func() {
		if err := grpcServer.Run(); err != nil {
			panic(err)
		}
	}()

	httpServer.Run()
}

