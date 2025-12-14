package http

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"structure-service/internal/domain"
	"structure-service/internal/infrastructure/excel"
	"structure-service/internal/infrastructure/logger"
	"structure-service/internal/usecase"
)

// PaginatedUniversitiesResponse представляет ответ с пагинацией для университетов
type PaginatedUniversitiesResponse struct {
	Data       []*domain.University `json:"data"`
	Total      int                  `json:"total"`
	Limit      int                  `json:"limit"`
	Offset     int                  `json:"offset"`
	TotalPages int                  `json:"total_pages"`
}

type Handler struct {
	structureService              domain.StructureServiceInterface
	getUniversityStructureUseCase *usecase.GetUniversityStructureUseCase
	assignOperatorUseCase         *usecase.AssignOperatorToDepartmentUseCase
	importStructureUseCase        *usecase.ImportStructureFromExcelUseCase
	createStructureUseCase        *usecase.CreateStructureFromRowUseCase
	departmentManagerRepo         domain.DepartmentManagerRepository
	logger                        *logger.Logger
}

func NewHandler(
	structureService domain.StructureServiceInterface,
	getUniversityStructureUseCase *usecase.GetUniversityStructureUseCase,
	assignOperatorUseCase *usecase.AssignOperatorToDepartmentUseCase,
	importStructureUseCase *usecase.ImportStructureFromExcelUseCase,
	createStructureUseCase *usecase.CreateStructureFromRowUseCase,
	departmentManagerRepo domain.DepartmentManagerRepository,
	log *logger.Logger,
) *Handler {
	return &Handler{
		structureService:              structureService,
		getUniversityStructureUseCase: getUniversityStructureUseCase,
		assignOperatorUseCase:         assignOperatorUseCase,
		importStructureUseCase:        importStructureUseCase,
		createStructureUseCase:        createStructureUseCase,
		departmentManagerRepo:         departmentManagerRepo,
		logger:                        log,
	}
}

// GetStructure godoc
// @Summary      Получить структуру вуза
// @Description  Возвращает иерархическую структуру вуза (университет -> филиал -> факультет -> группа -> чат)
// @Tags         structure
// @Accept       json
// @Produce      json
// @Param        university_id  path      int  true  "ID вуза"
// @Success      200            {object}  domain.StructureNode
// @Failure      404            {string}  string
// @Router       /universities/{university_id}/structure [get]
func (h *Handler) GetStructure(w http.ResponseWriter, r *http.Request) {
	log.Printf("=== GetStructure handler called ===")
	path := strings.TrimPrefix(r.URL.Path, "/universities/")
	path = strings.TrimSuffix(path, "/structure")
	universityID, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		http.Error(w, "invalid university id", http.StatusBadRequest)
		return
	}
	log.Printf("=== Parsed university ID: %d ===", universityID)

	log.Printf("=== Handler: calling getUniversityStructureUseCase.Execute for university %d ===", universityID)
	structure, err := h.getUniversityStructureUseCase.Execute(r.Context(), universityID)
	log.Printf("=== Handler: got structure with ChatCount: %v ===", structure.ChatCount)
	if err != nil {
		if err == domain.ErrUniversityNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(structure)
}

// GetAllUniversities godoc
// @Summary      Получить все вузы
// @Description  Возвращает список всех вузов с пагинацией, сортировкой и поиском
// @Tags         universities
// @Accept       json
// @Produce      json
// @Param        limit      query     int     false  "Лимит результатов (по умолчанию 50, максимум 100)"
// @Param        offset     query     int     false  "Смещение для пагинации"
// @Param        sort_by    query     string  false  "Поле для сортировки (id, name, inn, kpp, foiv, created_at, updated_at)"
// @Param        sort_order query     string  false  "Порядок сортировки (asc, desc)"
// @Param        search     query     string  false  "Поисковый запрос по всем полям"
// @Success      200        {object}  PaginatedUniversitiesResponse
// @Failure      400        {string}  string
// @Router       /universities [get]
func (h *Handler) GetAllUniversities(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	sortBy := r.URL.Query().Get("sort_by")
	sortOrder := r.URL.Query().Get("sort_order")
	search := r.URL.Query().Get("search")

	universities, total, err := h.structureService.GetAllUniversitiesWithSortingAndSearch(limit, offset, sortBy, sortOrder, search)
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

	response := PaginatedUniversitiesResponse{
		Data:       universities,
		Total:      total,
		Limit:      limit,
		Offset:     offset,
		TotalPages: totalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetUniversity godoc
// @Summary      Получить вуз по ID
// @Description  Возвращает информацию о вузе
// @Tags         universities
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID вуза"
// @Success      200  {object}  domain.University
// @Failure      404  {string}  string
// @Router       /universities/{id} [get]
func (h *Handler) GetUniversity(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/universities/")
	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		http.Error(w, "invalid university id", http.StatusBadRequest)
		return
	}

	university, err := h.structureService.GetUniversity(id)
	if err != nil {
		if err == domain.ErrUniversityNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(university)
}

// CreateUniversity godoc
// @Summary      Создать вуз
// @Description  Создает новый вуз
// @Tags         universities
// @Accept       json
// @Produce      json
// @Param        input  body      domain.University  true  "Данные вуза"
// @Success      201    {object}  domain.University
// @Failure      400    {string}  string
// @Router       /universities [post]
func (h *Handler) CreateUniversity(w http.ResponseWriter, r *http.Request) {
	var university domain.University
	if err := json.NewDecoder(r.Body).Decode(&university); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.structureService.CreateUniversity(&university); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(university)
}

// CreateStructure godoc
// @Summary      Создать полную структуру
// @Description  Создает или находит все элементы структуры (университет, филиал, факультет, группа)
// @Tags         structure
// @Accept       json
// @Produce      json
// @Param        input  body      usecase.CreateStructureRequest  true  "Данные структуры"
// @Success      200    {object}  usecase.CreateStructureResponse
// @Failure      400    {string}  string
// @Router       /structure [post]
func (h *Handler) CreateStructure(w http.ResponseWriter, r *http.Request) {
	var req usecase.CreateStructureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	response, err := h.createStructureUseCase.Execute(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ImportExcel godoc
// @Summary      Импортировать структуру из Excel
// @Description  Импортирует структуру вуза из Excel файла
// @Tags         import
// @Accept       multipart/form-data
// @Produce      json
// @Param        file  formData  file  true  "Excel файл со структурой"
// @Success      200   {object}  domain.ImportResult
// @Failure      400   {string}  string
// @Router       /import/excel [post]
func (h *Handler) ImportExcel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Парсим multipart form с лимитом 50 MB
	err := r.ParseMultipartForm(50 << 20) // 50 MB
	if err != nil {
		http.Error(w, "failed to parse form or file too large (max 50MB)", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file not found", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Проверяем расширение файла
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".xlsx") &&
		!strings.HasSuffix(strings.ToLower(header.Filename), ".xls") {
		http.Error(w, "invalid file format, expected .xlsx or .xls", http.StatusBadRequest)
		return
	}

	// Читаем файл
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "failed to read file", http.StatusInternalServerError)
		return
	}

	// Парсим Excel
	rows, err := excel.ParseExcel(fileBytes)
	if err != nil {
		http.Error(w, "failed to parse excel: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Валидация: проверяем, что есть хотя бы одна строка
	if len(rows) == 0 {
		http.Error(w, "excel file contains no data rows", http.StatusBadRequest)
		return
	}

	// Импортируем структуру
	result, err := h.importStructureUseCase.Execute(rows)
	if err != nil {
		http.Error(w, "failed to import: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}


// AssignOperatorRequest представляет запрос на назначение оператора
type AssignOperatorRequest struct {
	EmployeeID int64  `json:"employee_id"`
	BranchID   *int64 `json:"branch_id,omitempty"`
	FacultyID  *int64 `json:"faculty_id,omitempty"`
	AssignedBy *int64 `json:"assigned_by,omitempty"`
}

// AssignOperator godoc
// @Summary      Назначить оператора на подразделение
// @Description  Назначает оператора на филиал или факультет
// @Tags         department-managers
// @Accept       json
// @Produce      json
// @Param        input  body      AssignOperatorRequest  true  "Данные назначения"
// @Success      201    {object}  domain.DepartmentManager
// @Failure      400    {string}  string
// @Router       /departments/managers [post]
func (h *Handler) AssignOperator(w http.ResponseWriter, r *http.Request) {
	var req AssignOperatorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	dm, err := h.assignOperatorUseCase.Execute(req.EmployeeID, req.BranchID, req.FacultyID, req.AssignedBy)
	if err != nil {
		if err == domain.ErrEmployeeNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err == domain.ErrInvalidDepartment {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dm)
}

// RemoveOperator godoc
// @Summary      Удалить назначение оператора
// @Description  Удаляет назначение оператора на подразделение
// @Tags         department-managers
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID назначения"
// @Success      204  {string}  string
// @Failure      404  {string}  string
// @Router       /departments/managers/{id} [delete]
func (h *Handler) RemoveOperator(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/departments/managers/")
	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.departmentManagerRepo.DeleteDepartmentManager(id); err != nil {
		if err == domain.ErrDepartmentManagerNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetAllDepartmentManagers godoc
// @Summary      Получить все назначения операторов
// @Description  Возвращает список всех назначений операторов на подразделения
// @Tags         department-managers
// @Accept       json
// @Produce      json
// @Success      200  {array}   domain.DepartmentManager
// @Router       /departments/managers [get]
func (h *Handler) GetAllDepartmentManagers(w http.ResponseWriter, r *http.Request) {
	managers, err := h.departmentManagerRepo.GetAllDepartmentManagers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(managers)
}

// LinkGroupToChatRequest represents request to link group to chat
type LinkGroupToChatRequest struct {
	ChatID int64 `json:"chat_id"`
}

// UpdateNameRequest represents request to update name
type UpdateNameRequest struct {
	Name string `json:"name"`
}

// LinkGroupToChat godoc
// @Summary      Связать группу с чатом
// @Description  Обновляет chat_id для группы
// @Tags         groups
// @Accept       json
// @Produce      json
// @Param        group_id  path      int                      true  "ID группы"
// @Param        input     body      LinkGroupToChatRequest  true  "ID чата"
// @Success      200       {string}  string
// @Failure      400       {string}  string
// @Router       /groups/{group_id}/chat [put]
func (h *Handler) LinkGroupToChat(w http.ResponseWriter, r *http.Request) {
	// Extract group ID from URL
	path := strings.TrimPrefix(r.URL.Path, "/groups/")
	path = strings.TrimSuffix(path, "/chat")
	groupID, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		http.Error(w, "invalid group id", http.StatusBadRequest)
		return
	}

	var req LinkGroupToChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Get group
	group, err := h.structureService.GetGroupByID(groupID)
	if err != nil {
		http.Error(w, "group not found", http.StatusNotFound)
		return
	}

	// Update chat_id
	group.ChatID = &req.ChatID
	if err := h.structureService.UpdateGroup(group); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"group linked to chat"}`))
}

// UpdateUniversityName godoc
// @Summary      Обновить название университета
// @Description  Обновляет название университета по ID
// @Tags         universities
// @Accept       json
// @Produce      json
// @Param        id     path      int                 true  "ID университета"
// @Param        input  body      UpdateNameRequest  true  "Новое название"
// @Success      200    {string}  string
// @Failure      400    {string}  string
// @Failure      404    {string}  string
// @Router       /universities/{id}/name [put]
func (h *Handler) UpdateUniversityName(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/universities/")
	path = strings.TrimSuffix(path, "/name")
	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		http.Error(w, "invalid university id", http.StatusBadRequest)
		return
	}

	var req UpdateNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		http.Error(w, "name cannot be empty", http.StatusBadRequest)
		return
	}

	if err := h.structureService.UpdateUniversityName(id, req.Name); err != nil {
		if err == domain.ErrUniversityNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"university name updated successfully"}`))
}

// UpdateBranchName godoc
// @Summary      Обновить название филиала
// @Description  Обновляет название филиала по ID
// @Tags         branches
// @Accept       json
// @Produce      json
// @Param        id     path      int                 true  "ID филиала"
// @Param        input  body      UpdateNameRequest  true  "Новое название"
// @Success      200    {string}  string
// @Failure      400    {string}  string
// @Failure      404    {string}  string
// @Router       /branches/{id}/name [put]
func (h *Handler) UpdateBranchName(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/branches/")
	path = strings.TrimSuffix(path, "/name")
	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		http.Error(w, "invalid branch id", http.StatusBadRequest)
		return
	}

	var req UpdateNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		http.Error(w, "name cannot be empty", http.StatusBadRequest)
		return
	}

	if err := h.structureService.UpdateBranchName(id, req.Name); err != nil {
		if err == domain.ErrBranchNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"branch name updated successfully"}`))
}

// UpdateFacultyName godoc
// @Summary      Обновить название факультета
// @Description  Обновляет название факультета по ID
// @Tags         faculties
// @Accept       json
// @Produce      json
// @Param        id     path      int                 true  "ID факультета"
// @Param        input  body      UpdateNameRequest  true  "Новое название"
// @Success      200    {string}  string
// @Failure      400    {string}  string
// @Failure      404    {string}  string
// @Router       /faculties/{id}/name [put]
func (h *Handler) UpdateFacultyName(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/faculties/")
	path = strings.TrimSuffix(path, "/name")
	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		http.Error(w, "invalid faculty id", http.StatusBadRequest)
		return
	}

	var req UpdateNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		http.Error(w, "name cannot be empty", http.StatusBadRequest)
		return
	}

	if err := h.structureService.UpdateFacultyName(id, req.Name); err != nil {
		if err == domain.ErrFacultyNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"faculty name updated successfully"}`))
}

// UpdateGroupName godoc
// @Summary      Обновить название группы
// @Description  Обновляет номер группы по ID
// @Tags         groups
// @Accept       json
// @Produce      json
// @Param        id     path      int                 true  "ID группы"
// @Param        input  body      UpdateNameRequest  true  "Новый номер группы"
// @Success      200    {string}  string
// @Failure      400    {string}  string
// @Failure      404    {string}  string
// @Router       /groups/{id}/name [put]
func (h *Handler) UpdateGroupName(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/groups/")
	path = strings.TrimSuffix(path, "/name")
	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		http.Error(w, "invalid group id", http.StatusBadRequest)
		return
	}

	var req UpdateNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		http.Error(w, "name cannot be empty", http.StatusBadRequest)
		return
	}

	if err := h.structureService.UpdateGroupName(id, req.Name); err != nil {
		if err == domain.ErrGroupNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"group name updated successfully"}`))
}
