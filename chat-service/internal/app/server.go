package app

import (
	"context"
	"log"
	"net/http"
	"time"
)

type Server struct {
	Handler                 http.Handler
	Port                    string
	ParticipantsIntegration *ParticipantsIntegration
	httpServer              *http.Server
}

func (s *Server) Start() error {
	// Запускаем participants integration если она доступна
	if s.ParticipantsIntegration != nil {
		log.Println("Starting participants integration")
		s.ParticipantsIntegration.Start()
	}

	// Создаем HTTP сервер
	s.httpServer = &http.Server{
		Addr:    ":" + s.Port,
		Handler: s.Handler,
	}

	log.Println("Starting chat-service HTTP server on port", s.Port)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	// Останавливаем participants integration если она запущена
	if s.ParticipantsIntegration != nil {
		log.Println("Stopping participants integration")
		s.ParticipantsIntegration.Stop()
	}

	// Останавливаем HTTP сервер с graceful shutdown
	if s.httpServer != nil {
		log.Println("Stopping HTTP server")
		shutdownCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()
		
		return s.httpServer.Shutdown(shutdownCtx)
	}

	return nil
}

// Run - deprecated, используйте Start() и Stop() для лучшего контроля lifecycle
func (s *Server) Run() {
	log.Println("Starting chat-service server on port", s.Port)
	log.Fatal(http.ListenAndServe(":"+s.Port, s.Handler))
}

