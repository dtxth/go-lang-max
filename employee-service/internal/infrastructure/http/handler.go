package http

import (
	"employee-service/internal/domain"
	"employee-service/internal/infrastructure/auth"
	"employee-service/internal/infrastructure/errors"
	"employee-service/internal/infrastructure/logger"
	"employee-service/internal/infrastructure/middleware"
	"employee-service/internal/usecase"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	employeeService                 domain.EmployeeServiceInterface
	batchUpdateMaxIdUseCase         *usecase.BatchUpdateMaxIdUseCase
	searchEmployeesWithRoleFilterUC *usecase.SearchEmployeesWithRoleFilterUseCase
	authClient                      *auth.AuthClient
	logger                          *logger.Logger
}

// AddEmployeeRequest представляет запрос на добавление сотрудника
type AddEmployeeRequest struct {
	Phone          string `json:"phone" example:"+79001234567" binding:"required"`
	FirstName      string `json:"first_name,omitempty" example:"Иван"`
	LastName       string `json:"last_name,omitempty" example:"Иванов"`
	MiddleName     string `json:"middle_name,omitempty" example:"Иванович"`
	INN            string `json:"inn,omitempty" example:"1234567890"`
	KPP            string `json:"kpp,omitempty" example:"123456789"`
	UniversityName string `json:"university_name,omitempty" example:"МГУ"`
	Role           string `json:"role,omitempty" example:"curator" enums:"curator,operator"`
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

// PaginatedEmployeesResponse представляет ответ с пагинацией для сотрудников
type PaginatedEmployeesResponse struct {
	Data       []*domain.Employee `json:"data"`
	Total      int                `json:"total"`
	Limit      int                `json:"limit"`
	Offset     int                `json:"offset"`
	TotalPages int                `json:"total_pages"`
}

// Employee представляет сотрудника (для Swagger)
type Employee domain.Employee

// University представляет вуз (для Swagger)
type University domain.University

func NewHandler(
	employeeService domain.EmployeeServiceInterface,
	batchUpdateMaxIdUseCase *usecase.BatchUpdateMaxIdUseCase,
	searchEmployeesWithRoleFilterUC *usecase.SearchEmployeesWithRoleFilterUseCase,
	authClient *auth.AuthClient,
	log *logger.Logger,
) *Handler {
	return &Handler{
		employeeService:                 employeeService,
		batchUpdateMaxIdUseCase:         batchUpdateMaxIdUseCase,
		searchEmployeesWithRoleFilterUC: searchEmployeesWithRoleFilterUC,
		authClient:                      authClient,
		logger:                          log,
	}
}

// SearchEmployees godoc
// @Summary      Поиск сотрудников
// @Description  Выполняет поиск сотрудников по имени, фамилии и названию вуза с применением ролевой фильтрации
// @Tags         employees
// @Accept       json
// @Produce      json
// @Param        query   query     string  false  "Поисковый запрос"
// @Param        limit   query     int     false  "Лимит результатов (по умолчанию 50, максимум 100)"
// @Param        offset  query     int     false  "Смещение для пагинации"
// @Param        Authorization  header  string  true  "Bearer token"
// @Success      200     {array}   usecase.SearchEmployeeResult
// @Failure      400     {string}  string
// @Failure      401     {string}  string
// @Failure      403     {string}  string
// @Router       /employees [get]
func (h *Handler) SearchEmployees(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	
	// Requirements 14.5: Apply role-based filtering
	// Extract JWT token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		errors.WriteError(w, errors.UnauthorizedError("missing authorization header"), requestID)
		return
	}

	// Extract token from "Bearer <token>"
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		errors.WriteError(w, errors.UnauthorizedError("invalid authorization header format"), requestID)
		return
	}

	// Validate token and get user info
	ctx := r.Context()
	tokenInfo, err := h.authClient.ValidateToken(ctx, token)
	if err != nil {
		http.Error(w, "invalid or expired token", http.StatusUnauthorized)
		return
	}

	// Extract query parameters
	query := r.URL.Query().Get("query")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	// Prepare university ID for filtering
	var universityID *int64
	if tokenInfo.UniversityId > 0 {
		uid := tokenInfo.UniversityId
		universityID = &uid
	}

	// Use the new use case with role filtering if available
	if h.searchEmployeesWithRoleFilterUC != nil {
		results, err := h.searchEmployeesWithRoleFilterUC.Execute(
			ctx,
			query,
			tokenInfo.Role,
			universityID,
			limit,
			offset,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Requirements 14.5: Return empty array for no matches
		if results == nil {
			results = []*usecase.SearchEmployeeResult{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
		return
	}

	// Fallback to old implementation if new use case not available
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
// @Description  Возвращает список всех сотрудников с пагинацией, сортировкой и поиском
// @Tags         employees
// @Accept       json
// @Produce      json
// @Param        limit      query     int     false  "Лимит результатов (по умолчанию 50, максимум 100)"
// @Param        offset     query     int     false  "Смещение для пагинации"
// @Param        sort_by    query     string  false  "Поле для сортировки (id, first_name, last_name, middle_name, phone, max_id, inn, kpp, role, university, created_at, updated_at)"
// @Param        sort_order query     string  false  "Порядок сортировки (asc, desc)"
// @Param        search     query     string  false  "Поисковый запрос по всем полям"
// @Success      200        {object}  PaginatedEmployeesResponse
// @Failure      400        {string}  string
// @Router       /employees/all [get]
func (h *Handler) GetAllEmployees(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	sortBy := r.URL.Query().Get("sort_by")
	sortOrder := r.URL.Query().Get("sort_order")
	search := r.URL.Query().Get("search")

	employees, total, err := h.employeeService.GetAllEmployeesWithSortingAndSearch(limit, offset, sortBy, sortOrder, search)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Устанавливаем значения по умолчанию для ответа
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	// Вычисляем общее количество страниц
	totalPages := (total + limit - 1) / limit
	if totalPages == 0 {
		totalPages = 1
	}

	response := PaginatedEmployeesResponse{
		Data:       employees,
		Total:      total,
		Limit:      limit,
		Offset:     offset,
		TotalPages: totalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
	h.logger.Info(r.Context(), "AddEmployee handler started", map[string]interface{}{
		"method": r.Method,
		"path":   r.URL.Path,
	})
	var req struct {
		Phone          string `json:"phone"`
		FirstName      string `json:"first_name"`
		LastName       string `json:"last_name"`
		MiddleName     string `json:"middle_name"`
		INN            string `json:"inn"`
		KPP            string `json:"kpp"`
		UniversityName string `json:"university_name"`
		Role           string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Phone == "" {
		http.Error(w, "phone is required", http.StatusBadRequest)
		return
	}

	// Логирование для отладки
	h.logger.Info(r.Context(), "AddEmployee request", map[string]interface{}{
		"phone": req.Phone,
		"first_name": req.FirstName,
		"last_name": req.LastName,
		"role": req.Role,
	})

	// Если роль указана, используем CreateEmployeeWithRole
	var employee *domain.Employee
	var err error
	
	if req.Role != "" {
		// TODO: Получить роль запрашивающего пользователя из JWT токена
		// Пока используем "superadmin" для тестирования
		requesterRole := "superadmin"
		
		employee, err = h.employeeService.CreateEmployeeWithRole(
			r.Context(),
			req.Phone,
			req.FirstName,
			req.LastName,
			req.MiddleName,
			req.INN,
			req.KPP,
			req.UniversityName,
			req.Role,
			requesterRole,
		)
	} else {
		// Используем старый метод без роли
		// Устанавливаем значения по умолчанию для пустых полей
		firstName := req.FirstName
		if firstName == "" {
			firstName = "Неизвестно"
		}
		lastName := req.LastName
		if lastName == "" {
			lastName = "Неизвестно"
		}
		universityName := req.UniversityName
		if universityName == "" {
			universityName = "Неизвестный вуз"
		}
		
		employee, err = h.employeeService.AddEmployeeByPhone(
			req.Phone,
			firstName,
			lastName,
			req.MiddleName,
			req.INN,
			req.KPP,
			universityName,
		)
	}

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

// AddEmployeeSimple - простое создание сотрудника только с телефоном
func (h *Handler) AddEmployeeSimple(w http.ResponseWriter, r *http.Request) {
	h.logger.Info(r.Context(), "AddEmployeeSimple called", map[string]interface{}{
		"method": r.Method,
	})

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

	h.logger.Info(r.Context(), "Creating employee with phone", map[string]interface{}{
		"phone": req.Phone,
	})

	// Используем старый метод без роли с дефолтными значениями
	employee, err := h.employeeService.AddEmployeeByPhone(
		req.Phone,
		"", // firstName - будет заменен на "Неизвестно"
		"", // lastName - будет заменен на "Неизвестно"
		"", // middleName
		"", // inn
		"", // kpp
		"", // universityName - будет заменен на "Неизвестный вуз"
	)

	if err != nil {
		h.logger.Info(r.Context(), "Error creating employee", map[string]interface{}{
			"error": err.Error(),
		})
		statusCode := http.StatusInternalServerError
		if err.Error() == "employee already exists" {
			statusCode = http.StatusConflict
		} else if err.Error() == "invalid phone number" {
			statusCode = http.StatusBadRequest
		}
		http.Error(w, err.Error(), statusCode)
		return
	}

	h.logger.Info(r.Context(), "Employee created successfully", map[string]interface{}{
		"employee_id": employee.ID,
	})

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
// UpdateEmployeeByMaxID обновляет данные сотрудника по MAX ID
// @Summary      Update employee by MAX ID
// @Description  Update employee first_name, last_name, and username by MAX ID
// @Tags         employees
// @Accept       json
// @Produce      json
// @Param        request  body      UpdateEmployeeByMaxIDRequest  true  "Employee update data"
// @Success      200      {object}  Employee
// @Failure      400      {string}  string
// @Failure      404      {string}  string
// @Failure      500      {string}  string
// @Router       /employees/update-by-max-id [put]
func (h *Handler) UpdateEmployeeByMaxID(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MaxID     string `json:"max_id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Username  string `json:"username"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.MaxID == "" {
		http.Error(w, "max_id is required", http.StatusBadRequest)
		return
	}

	// Получаем существующего сотрудника по MAX ID
	employee, err := h.employeeService.GetEmployeeByMaxID(req.MaxID)
	if err != nil {
		if err.Error() == "employee not found" {
			http.Error(w, "employee not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Обновляем поля (разрешаем пустые значения)
	employee.FirstName = req.FirstName
	employee.LastName = req.LastName
	// Username пока не сохраняем в employee, так как это поле auth-service

	// Сохраняем изменения
	if err := h.employeeService.UpdateEmployee(employee); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(employee)
}

type UpdateEmployeeByMaxIDRequest struct {
	MaxID     string `json:"max_id" example:"123456"`
	FirstName string `json:"first_name" example:"Андрей"`
	LastName  string `json:"last_name" example:"Иванов"`
	Username  string `json:"username" example:"testuser"`
}