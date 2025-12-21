package http

import (
	"employee-service/internal/infrastructure/middleware"
	"encoding/json"
	"net/http"
	"strings"

	httpSwagger "github.com/swaggo/http-swagger"
)

func (h *Handler) Router() http.Handler {
	mux := http.NewServeMux()

	// Simple employee creation endpoint - используем другой путь
	mux.HandleFunc("/simple-employee", func(w http.ResponseWriter, r *http.Request) {
		h.logger.Info(r.Context(), "simple-employee route hit", map[string]interface{}{
			"method": r.Method,
			"path":   r.URL.Path,
		})
		if r.Method == http.MethodPost {
			h.AddEmployeeSimple(w, r)
		} else {
			h.logger.Info(r.Context(), "Method not allowed", map[string]interface{}{
				"method": r.Method,
				"expected": "POST",
			})
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

	// Create employee with phone only
	mux.HandleFunc("/create-employee", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Phone string `json:"phone"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.Phone == "" {
			http.Error(w, "phone is required", http.StatusBadRequest)
			return
		}

		// Создаем сотрудника с минимальными данными
		employee, err := h.employeeService.AddEmployeeByPhone(
			req.Phone,
			"Неизвестно", // firstName
			"Неизвестно", // lastName  
			"",           // middleName
			"",           // inn
			"",           // kpp
			"Неизвестный вуз", // universityName
		)

		if err != nil {
			statusCode := http.StatusInternalServerError
			if err.Error() == "employee already exists" {
				statusCode = http.StatusConflict
			} else if err.Error() == "invalid phone number" {
				statusCode = http.StatusBadRequest
			}
			http.Error(w, err.Error(), statusCode)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(employee)
	})

	// Wrap with CORS middleware (отключен) и request ID middleware
	return middleware.RequestIDMiddleware(h.logger)(middleware.CORSMiddleware(mux))
}

