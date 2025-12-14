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

		// Проверяем специальные пути
		parts := strings.Split(path, "/")
		if len(parts) == 2 {
			switch parts[1] {
			case "administrators":
				// /chats/{id}/administrators
				switch r.Method {
				case http.MethodPost:
					h.AddAdministrator(w, r)
				default:
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				}
				return
			case "refresh-participants":
				// /chats/{id}/refresh-participants
				switch r.Method {
				case http.MethodPost:
					h.authMiddleware.Authenticate(h.RefreshParticipantsCount)(w, r)
				default:
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				}
				return
			}
		}

		// Проверяем, что это числовой ID
		switch r.Method {
		case http.MethodGet:
			h.GetChatByID(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Обработка /administrators (список всех)
	mux.HandleFunc("/administrators", func(w http.ResponseWriter, r *http.Request) {
		// Точное совпадение пути
		if r.URL.Path != "/administrators" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.GetAllAdministrators(w, r)
	})

	// Обработка /administrators/{id}
	mux.HandleFunc("/administrators/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/administrators/")
		
		// Если путь пустой после /administrators/, это тоже список всех
		if path == "" {
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			h.GetAllAdministrators(w, r)
			return
		}

		// Иначе это запрос к /administrators/{id}
		switch r.Method {
		case http.MethodGet:
			h.GetAdministratorByID(w, r)
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

