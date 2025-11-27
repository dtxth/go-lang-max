package http

import (
	"employee-service/internal/domain"
	"employee-service/internal/usecase"
	"encoding/json"
	"net/http"
	"strconv"
)

type Handler struct {
	employeeService        *usecase.EmployeeService
	batchUpdateMaxIdUseCase *usecase.BatchUpdateMaxIdUseCase
}

// AddEmployeeRequest представляет запрос на добавление сотрудника
type AddEmployeeRequest struct {
	Phone          string `json:"phone" example:"+79001234567" binding:"required"`
	FirstName      string `json:"first_name" example:"Иван" binding:"required"`
	LastName       string `json:"last_name" example:"Иванов" binding:"required"`
	MiddleName     string `json:"middle_name,omitempty" example:"Иванович"`
	INN            string `json:"inn,omitempty" example:"1234567890"`
	KPP            string `json:"kpp,omitempty" example:"123456789"`
	UniversityName string `json:"university_name,omitempty" example:"МГУ"`
}

// UpdateEmployeeRequest представляет запрос на обновление сотрудника
type UpdateEmployeeRequest struct {
	FirstName    string `json:"first_name,omitempty" example:"Иван"`
	LastName     string `json:"last_name,omitempty" example:"Иванов"`
	MiddleName   string `json:"middle_name,omitempty" example:"Иванович"`
	Phone        string `json:"phone,omitempty" example:"+79001234567"`
	INN          string `json:"inn,omitempty" example:"1234567890"`
	KPP          string `json:"kpp,omitempty" example:"123456789"`
	UniversityID int64  `json:"university_id,omitempty" example:"1"`
}

// DeleteResponse представляет ответ на удаление
type DeleteResponse struct {
	Status string `json:"status" example:"deleted"`
}

// Employee представляет сотрудника (для Swagger)
type Employee domain.Employee

// University представляет вуз (для Swagger)
type University domain.University

func NewHandler(employeeService *usecase.EmployeeService, batchUpdateMaxIdUseCase *usecase.BatchUpdateMaxIdUseCase) *Handler {
	return &Handler{
		employeeService:        employeeService,
		batchUpdateMaxIdUseCase: batchUpdateMaxIdUseCase,
	}
}

// SearchEmployees godoc
// @Summary      Поиск сотрудников
// @Description  Выполняет поиск сотрудников по имени, фамилии и названию вуза
// @Tags         employees
// @Accept       json
// @Produce      json
// @Param        query   query     string  false  "Поисковый запрос"
// @Param        limit   query     int     false  "Лимит результатов (по умолчанию 50, максимум 100)"
// @Param        offset  query     int     false  "Смещение для пагинации"
// @Success      200     {array}   Employee
// @Failure      400     {string}  string
// @Router       /employees [get]
func (h *Handler) SearchEmployees(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	employees, err := h.employeeService.SearchEmployees(query, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(employees)
}

// GetAllEmployees godoc
// @Summary      Получить всех сотрудников
// @Description  Возвращает список всех сотрудников с пагинацией
// @Tags         employees
// @Accept       json
// @Produce      json
// @Param        limit   query     int     false  "Лимит результатов (по умолчанию 50, максимум 100)"
// @Param        offset  query     int     false  "Смещение для пагинации"
// @Success      200     {array}   Employee
// @Failure      400     {string}  string
// @Router       /employees/all [get]
func (h *Handler) GetAllEmployees(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	employees, err := h.employeeService.GetAllEmployees(limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(employees)
}

// GetEmployeeByID godoc
// @Summary      Получить сотрудника по ID
// @Description  Возвращает информацию о сотруднике по его ID
// @Tags         employees
// @Accept       json
// @Produce      json
// @Param        id      path      int     true   "ID сотрудника"
// @Success      200     {object}  Employee
// @Failure      404     {string}  string
// @Router       /employees/{id} [get]
func (h *Handler) GetEmployeeByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/employees/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid employee id", http.StatusBadRequest)
		return
	}

	employee, err := h.employeeService.GetEmployeeByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(employee)
}

// AddEmployee godoc
// @Summary      Добавить сотрудника
// @Description  Добавляет нового сотрудника по номеру телефона. Автоматически получает MAX_id и создает/находит вуз
// @Tags         employees
// @Accept       json
// @Produce      json
// @Param        input   body      AddEmployeeRequest  true  "Данные сотрудника"
// @Success      201     {object}  Employee
// @Failure      400     {string}  string
// @Failure      409     {string}  string
// @Router       /employees [post]
func (h *Handler) AddEmployee(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Phone          string `json:"phone"`
		FirstName      string `json:"first_name"`
		LastName       string `json:"last_name"`
		MiddleName     string `json:"middle_name"`
		INN            string `json:"inn"`
		KPP            string `json:"kpp"`
		UniversityName string `json:"university_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Phone == "" {
		http.Error(w, "phone is required", http.StatusBadRequest)
		return
	}

	if req.FirstName == "" || req.LastName == "" {
		http.Error(w, "first_name and last_name are required", http.StatusBadRequest)
		return
	}

	employee, err := h.employeeService.AddEmployeeByPhone(
		req.Phone,
		req.FirstName,
		req.LastName,
		req.MiddleName,
		req.INN,
		req.KPP,
		req.UniversityName,
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
}

// UpdateEmployee godoc
// @Summary      Обновить сотрудника
// @Description  Обновляет данные сотрудника
// @Tags         employees
// @Accept       json
// @Produce      json
// @Param        id      path      int                true  "ID сотрудника"
// @Param        input   body      UpdateEmployeeRequest  true  "Обновленные данные сотрудника"
// @Success      200     {object}  Employee
// @Failure      400     {string}  string
// @Failure      404     {string}  string
// @Router       /employees/{id} [put]
func (h *Handler) UpdateEmployee(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/employees/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid employee id", http.StatusBadRequest)
		return
	}

	var req struct {
		FirstName    string `json:"first_name"`
		LastName     string `json:"last_name"`
		MiddleName   string `json:"middle_name"`
		Phone        string `json:"phone"`
		INN          string `json:"inn"`
		KPP          string `json:"kpp"`
		UniversityID int64  `json:"university_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Получаем существующего сотрудника
	employee, err := h.employeeService.GetEmployeeByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Обновляем поля
	if req.FirstName != "" {
		employee.FirstName = req.FirstName
	}
	if req.LastName != "" {
		employee.LastName = req.LastName
	}
	if req.MiddleName != "" {
		employee.MiddleName = req.MiddleName
	}
	if req.Phone != "" {
		employee.Phone = req.Phone
	}
	if req.INN != "" {
		employee.INN = req.INN
	}
	if req.KPP != "" {
		employee.KPP = req.KPP
	}
	if req.UniversityID > 0 {
		employee.UniversityID = req.UniversityID
	}

	if err := h.employeeService.UpdateEmployee(employee); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Получаем обновленного сотрудника
	updatedEmployee, err := h.employeeService.GetEmployeeByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedEmployee)
}

// DeleteEmployee godoc
// @Summary      Удалить сотрудника
// @Description  Удаляет сотрудника по ID
// @Tags         employees
// @Accept       json
// @Produce      json
// @Param        id      path      int     true   "ID сотрудника"
// @Success      200     {object}  DeleteResponse
// @Failure      404     {string}  string
// @Router       /employees/{id} [delete]
func (h *Handler) DeleteEmployee(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/employees/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid employee id", http.StatusBadRequest)
		return
	}

	if err := h.employeeService.DeleteEmployee(id); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

// BatchUpdateMaxID godoc
// @Summary      Trigger batch MAX_id update
// @Description  Starts a batch update job to retrieve MAX_id for all employees without it
// @Tags         employees
// @Accept       json
// @Produce      json
// @Success      200     {object}  domain.BatchUpdateResult
// @Failure      500     {string}  string
// @Router       /employees/batch-update-maxid [post]
func (h *Handler) BatchUpdateMaxID(w http.ResponseWriter, r *http.Request) {
	if h.batchUpdateMaxIdUseCase == nil {
		http.Error(w, "batch update service not available", http.StatusServiceUnavailable)
		return
	}
	
	result, err := h.batchUpdateMaxIdUseCase.StartBatchUpdate()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetBatchStatus godoc
// @Summary      Get batch update status
// @Description  Retrieves the status of a specific batch update job
// @Tags         employees
// @Accept       json
// @Produce      json
// @Param        id      path      int     true   "Batch job ID"
// @Success      200     {object}  domain.BatchUpdateJob
// @Failure      400     {string}  string
// @Failure      404     {string}  string
// @Router       /employees/batch-status/{id} [get]
func (h *Handler) GetBatchStatus(w http.ResponseWriter, r *http.Request) {
	if h.batchUpdateMaxIdUseCase == nil {
		http.Error(w, "batch update service not available", http.StatusServiceUnavailable)
		return
	}
	
	// Extract ID from path
	path := r.URL.Path
	idStr := path[len("/employees/batch-status/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid batch job id", http.StatusBadRequest)
		return
	}
	
	job, err := h.batchUpdateMaxIdUseCase.GetBatchJobStatus(id)
	if err != nil {
		http.Error(w, "batch job not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

// GetAllBatchJobs godoc
// @Summary      List all batch jobs
// @Description  Retrieves all batch update jobs with pagination
// @Tags         employees
// @Accept       json
// @Produce      json
// @Param        limit   query     int     false  "Limit results (default 50, max 100)"
// @Param        offset  query     int     false  "Offset for pagination"
// @Success      200     {array}   domain.BatchUpdateJob
// @Failure      500     {string}  string
// @Router       /employees/batch-status [get]
func (h *Handler) GetAllBatchJobs(w http.ResponseWriter, r *http.Request) {
	if h.batchUpdateMaxIdUseCase == nil {
		http.Error(w, "batch update service not available", http.StatusServiceUnavailable)
		return
	}
	
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	
	jobs, err := h.batchUpdateMaxIdUseCase.GetAllBatchJobs(limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}
