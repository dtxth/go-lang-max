package http

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"structure-service/internal/domain"
	"structure-service/internal/infrastructure/excel"
	"structure-service/internal/usecase"
)

type Handler struct {
	structureService *usecase.StructureService
}

func NewHandler(structureService *usecase.StructureService) *Handler {
	return &Handler{structureService: structureService}
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

	structure, err := h.structureService.GetStructure(universityID)
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

// ImportExcel godoc
// @Summary      Импортировать структуру из Excel
// @Description  Импортирует структуру вуза из Excel файла
// @Tags         import
// @Accept       multipart/form-data
// @Produce      json
// @Param        file  formData  file  true  "Excel файл со структурой"
// @Success      200   {object}  map[string]string
// @Failure      400   {string}  string
// @Router       /import/excel [post]
func (h *Handler) ImportExcel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Парсим multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, "failed to parse form", http.StatusBadRequest)
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

	// Импортируем структуру
	if err := h.structureService.ImportFromExcel(rows); err != nil {
		http.Error(w, "failed to import: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"message": "Structure imported successfully",
	})
}

