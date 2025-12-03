package http

import (
	"chat-service/internal/infrastructure/middleware"
	"net/http"
	"strings"

	httpSwagger "github.com/swaggo/http-swagger"
)

func (h *Handler) Router() http.Handler {
	mux := http.NewServeMux()

	// Чаты (с аутентификацией)
	mux.HandleFunc("/chats", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.authMiddleware.Authenticate(h.SearchChats)(w, r)
		case http.MethodPost:
			h.CreateChat(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/chats/all", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.authMiddleware.Authenticate(h.GetAllChats)(w, r)
	})

	// Обработка /chats/{id}
	mux.HandleFunc("/chats/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/chats/")

		// Если путь пустой после /chats/, это значит запрос к /chats/
		if path == "" {
			switch r.Method {
			case http.MethodGet:
				h.SearchChats(w, r)
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		// Проверяем, является ли это запросом к /chats/{id}/administrators
		parts := strings.Split(path, "/")
		if len(parts) == 2 && parts[1] == "administrators" {
			switch r.Method {
			case http.MethodPost:
				h.AddAdministrator(w, r)
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		// Проверяем, что это числовой ID
		switch r.Method {
		case http.MethodGet:
			h.GetChatByID(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Обработка /administrators/{id}
	mux.HandleFunc("/administrators/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/administrators/")
		if path == "" {
			http.Error(w, "administrator id is required", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodDelete:
			h.RemoveAdministrator(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Swagger UI
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with request ID middleware
	return middleware.RequestIDMiddleware(h.logger)(mux)
}

