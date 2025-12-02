package main

import (
	"auth-service/internal/app"
	"auth-service/internal/config"
	"auth-service/internal/infrastructure/grpc"
	"auth-service/internal/infrastructure/hash"
	"auth-service/internal/infrastructure/http"
	"auth-service/internal/infrastructure/jwt"
	"auth-service/internal/infrastructure/repository"
	"auth-service/internal/usecase"
	"database/sql"
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
	cfg := config.Load()

	db, err := sql.Open("postgres", cfg.DBUrl)
	if err != nil {
		panic(err)
	}

	repo := repository.NewUserPostgres(db)
	refreshRepo := repository.NewRefreshPostgres(db)
	hasher := hash.NewBcryptHasher()
	jwtManager := jwt.NewManager(cfg.AccessSecret, cfg.RefreshSecret, 15*time.Minute, 7*24*time.Hour)

	authUC := usecase.NewAuthService(repo, refreshRepo, hasher, jwtManager)
	handler := http.NewHandler(authUC)

	// HTTP server
	httpServer := &app.Server{
		Handler: handler.Router(),
		Port:    cfg.Port,
	}

	// gRPC server
	grpcHandler := grpc.NewAuthHandler(authUC)
	grpcServer := grpc.NewServer(grpcHandler, cfg.GRPCPort)

	// Запускаем оба сервера
	go func() {
		if err := grpcServer.Run(); err != nil {
			panic(err)
		}
	}()

	httpServer.Run()
}
