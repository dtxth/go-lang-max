package http

import (
	"employee-service/internal/infrastructure/middleware"
	"net/http"
	"strings"

	httpSwagger "github.com/swaggo/http-swagger"
)

func (h *Handler) Router() http.Handler {
	mux := http.NewServeMux()

	// Сотрудники
	mux.HandleFunc("/employees", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.SearchEmployees(w, r)
		case http.MethodPost:
			h.AddEmployee(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/employees/all", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.GetAllEmployees(w, r)
	})

	// Batch operations
	mux.HandleFunc("/employees/batch-update-maxid", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.BatchUpdateMaxID(w, r)
	})

	mux.HandleFunc("/employees/batch-status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.GetAllBatchJobs(w, r)
	})

	mux.HandleFunc("/employees/batch-status/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.GetBatchStatus(w, r)
	})

	// Обработка /employees/{id}
	mux.HandleFunc("/employees/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/employees/")
		
		// Если путь пустой после /employees/, это значит запрос к /employees/
		if path == "" {
			switch r.Method {
			case http.MethodGet:
				h.SearchEmployees(w, r)
			case http.MethodPost:
				h.AddEmployee(w, r)
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		// Проверяем, что это числовой ID
		switch r.Method {
		case http.MethodGet:
			h.GetEmployeeByID(w, r)
		case http.MethodPut:
			h.UpdateEmployee(w, r)
		case http.MethodDelete:
			h.DeleteEmployee(w, r)
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
	return middleware.RequestIDMiddleware(nil)(mux)
}

