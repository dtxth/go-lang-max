package http

import (
	"net/http"
	"strings"
	"structure-service/internal/infrastructure/middleware"

	httpSwagger "github.com/swaggo/http-swagger"
)

func (h *Handler) Router() http.Handler {
	mux := http.NewServeMux()

	// Universities
	mux.HandleFunc("/universities", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.GetAllUniversities(w, r)
		case http.MethodPost:
			h.CreateUniversity(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/universities/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/structure") {
			if r.Method == http.MethodGet {
				h.GetStructure(w, r)
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
	})

	// Import
	mux.HandleFunc("/import/excel", h.ImportExcel)

	// Department Managers
	mux.HandleFunc("/departments/managers", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.GetAllDepartmentManagers(w, r)
		case http.MethodPost:
			h.AssignOperator(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/departments/managers/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			h.RemoveOperator(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Swagger UI
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	// Wrap with request ID middleware
	return middleware.RequestIDMiddleware(h.logger)(mux)
}

