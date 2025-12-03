package http

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"structure-service/internal/domain"
	"structure-service/internal/infrastructure/excel"
	"structure-service/internal/infrastructure/logger"
	"structure-service/internal/usecase"
)

type Handler struct {
	structureService              *usecase.StructureService
	getUniversityStructureUseCase *usecase.GetUniversityStructureUseCase
	assignOperatorUseCase         *usecase.AssignOperatorToDepartmentUseCase
	importStructureUseCase        *usecase.ImportStructureFromExcelUseCase
	createStructureUseCase        *usecase.CreateStructureFromRowUseCase
	departmentManagerRepo         domain.DepartmentManagerRepository
	logger                        *logger.Logger
}

func NewHandler(
	structureService *usecase.StructureService,
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
	path := strings.TrimPrefix(r.URL.Path, "/universities/")
	path = strings.TrimSuffix(path, "/structure")
	universityID, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		http.Error(w, "invalid university id", http.StatusBadRequest)
		return
	}

	structure, err := h.getUniversityStructureUseCase.Execute(r.Context(), universityID)
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
// @Description  Возвращает список всех вузов
// @Tags         universities
// @Accept       json
// @Produce      json
// @Success      200  {array}   domain.University
// @Router       /universities [get]
func (h *Handler) GetAllUniversities(w http.ResponseWriter, r *http.Request) {
	universities, err := h.structureService.GetAllUniversities()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(universities)
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
