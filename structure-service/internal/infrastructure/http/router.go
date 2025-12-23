package http

import (
	"net/http"
	"strings"
	"structure-service/internal/infrastructure/middleware"

	httpSwagger "github.com/swaggo/http-swagger"
)

func (h *Handler) Router() http.Handler {
	mux := http.NewServeMux()

	// Auth middleware
	authMiddleware := middleware.AuthMiddleware()

	// Universities (с авторизацией)
	mux.Handle("/universities", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.GetAllUniversities(w, r)
		case http.MethodPost:
			h.CreateUniversity(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/universities/", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/structure") {
			if r.Method == http.MethodGet {
				h.GetStructure(w, r)
			} else {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		} else if strings.HasSuffix(path, "/name") {
			if r.Method == http.MethodPut {
				h.UpdateUniversityName(w, r)
			} else {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			if r.Method == http.MethodGet {
				h.GetUniversity(w, r)
			} else {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		}
	})))

	// Structure (с авторизацией)
	mux.Handle("/structure", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.CreateStructure(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Import (с авторизацией)
	mux.Handle("/import/excel", authMiddleware(http.HandlerFunc(h.ImportExcel)))

	// Branches (с авторизацией)
	mux.Handle("/branches/", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/name") && r.Method == http.MethodPut {
			h.UpdateBranchName(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Faculties (с авторизацией)
	mux.Handle("/faculties/", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/name") && r.Method == http.MethodPut {
			h.UpdateFacultyName(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Groups (с авторизацией)
	mux.Handle("/groups/", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/chat") && r.Method == http.MethodPut {
			h.LinkGroupToChat(w, r)
		} else if strings.HasSuffix(r.URL.Path, "/name") && r.Method == http.MethodPut {
			h.UpdateGroupName(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Department Managers (с авторизацией)
	mux.Handle("/departments/managers", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.GetAllDepartmentManagers(w, r)
		case http.MethodPost:
			h.AssignOperator(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/departments/managers/", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			h.RemoveOperator(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Swagger UI (без авторизации)
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	// Health check (без авторизации)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with CORS middleware (отключен) и request ID middleware
	return middleware.RequestIDMiddleware(h.logger)(middleware.CORSMiddleware(mux))
}

