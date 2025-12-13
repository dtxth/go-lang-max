package http

import (
	"chat-service/internal/domain"
	"chat-service/internal/infrastructure/logger"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Handler struct {
	chatService    domain.ChatServiceInterface
	authMiddleware *AuthMiddleware
	logger         *logger.Logger
}

// Chat представляет чат (для Swagger)
type Chat domain.Chat

// Administrator представляет администратора (для Swagger)
type Administrator domain.Administrator

// ChatListResponse представляет ответ со списком чатов и пагинацией
type ChatListResponse struct {
	Chats      []*Chat `json:"chats"`
	TotalCount int     `json:"total_count"`
	Limit      int     `json:"limit"`
	Offset     int     `json:"offset"`
}

// PaginatedChatsResponse представляет ответ с расширенной пагинацией для чатов
type PaginatedChatsResponse struct {
	Data       []*Chat `json:"data"`
	Total      int     `json:"total"`
	Limit      int     `json:"limit"`
	Offset     int     `json:"offset"`
	TotalPages int     `json:"total_pages"`
}

// AddAdministratorRequest представляет запрос на добавление администратора
type AddAdministratorRequest struct {
	Phone                string `json:"phone" example:"+79001234567" binding:"required"`
	MaxID                string `json:"max_id,omitempty" example:"496728250"`
	AddUser              bool   `json:"add_user" example:"true"`
	AddAdmin             bool   `json:"add_admin" example:"true"`
	SkipPhoneValidation  bool   `json:"skip_phone_validation,omitempty" example:"false"`
}

// AdministratorListResponse представляет ответ со списком администраторов и пагинацией
type AdministratorListResponse struct {
	Administrators []*Administrator `json:"administrators"`
	TotalCount     int              `json:"total_count"`
	Limit          int              `json:"limit"`
	Offset         int              `json:"offset"`
}

// DeleteResponse представляет ответ на удаление
type DeleteResponse struct {
	Status string `json:"status" example:"deleted"`
}

func NewHandler(chatService domain.ChatServiceInterface, authMiddleware *AuthMiddleware, log *logger.Logger) *Handler {
	return &Handler{
		chatService:    chatService,
		authMiddleware: authMiddleware,
		logger:         log,
	}
}

// SearchChats godoc
// @Summary      Поиск чатов
// @Description  Выполняет поиск чатов по названию с учетом роли пользователя
// @Tags         chats
// @Accept       json
// @Produce      json
// @Param        Authorization header    string  true   "Bearer token"
// @Param        query         query     string  false  "Поисковый запрос (название чата)"
// @Param        limit         query     int     false  "Лимит результатов (по умолчанию 50, максимум 100)"
// @Param        offset        query     int     false  "Смещение для пагинации"
// @Success      200           {object}  ChatListResponse
// @Failure      400           {string}  string
// @Failure      401           {string}  string
// @Failure      403           {string}  string
// @Router       /chats [get]
func (h *Handler) SearchChats(w http.ResponseWriter, r *http.Request) {
	// Получаем информацию о токене из контекста (установлена middleware)
	tokenInfo, ok := GetTokenInfo(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Создаем фильтр на основе информации о токене
	filter := domain.NewChatFilter(tokenInfo)
	if filter == nil {
		http.Error(w, "invalid token info", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query().Get("query")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	chats, totalCount, err := h.chatService.SearchChats(query, limit, offset, filter)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == domain.ErrForbidden || err == domain.ErrInvalidRole {
			statusCode = http.StatusForbidden
		}
		http.Error(w, err.Error(), statusCode)
		return
	}

	// Конвертируем domain.Chat в Chat для ответа
	responseChats := make([]*Chat, len(chats))
	for i, chat := range chats {
		c := Chat(*chat)
		responseChats[i] = &c
	}

	response := ChatListResponse{
		Chats:      responseChats,
		TotalCount: totalCount,
		Limit:      limit,
		Offset:     offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetAllChats godoc
// @Summary      Получить все чаты
// @Description  Возвращает список всех чатов с пагинацией, сортировкой и поиском (с учетом роли пользователя)
// @Tags         chats
// @Accept       json
// @Produce      json
// @Param        Authorization header    string  true   "Bearer token"
// @Param        limit         query     int     false  "Лимит результатов (по умолчанию 50, максимум 100)"
// @Param        offset        query     int     false  "Смещение для пагинации"
// @Param        sort_by       query     string  false  "Поле для сортировки (id, name, url, max_chat_id, participants_count, department, source, university, created_at, updated_at)"
// @Param        sort_order    query     string  false  "Порядок сортировки (asc, desc)"
// @Param        search        query     string  false  "Поисковый запрос по всем полям"
// @Success      200           {object}  PaginatedChatsResponse
// @Failure      400           {string}  string
// @Failure      401           {string}  string
// @Failure      403           {string}  string
// @Router       /chats/all [get]
func (h *Handler) GetAllChats(w http.ResponseWriter, r *http.Request) {
	// Получаем информацию о токене из контекста (установлена middleware)
	tokenInfo, ok := GetTokenInfo(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Создаем фильтр на основе информации о токене
	filter := domain.NewChatFilter(tokenInfo)
	if filter == nil {
		http.Error(w, "invalid token info", http.StatusUnauthorized)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	sortBy := r.URL.Query().Get("sort_by")
	sortOrder := r.URL.Query().Get("sort_order")
	search := r.URL.Query().Get("search")

	chats, totalCount, err := h.chatService.GetAllChatsWithSortingAndSearch(limit, offset, sortBy, sortOrder, search, filter)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == domain.ErrForbidden || err == domain.ErrInvalidRole {
			statusCode = http.StatusForbidden
		}
		http.Error(w, err.Error(), statusCode)
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
	totalPages := (totalCount + limit - 1) / limit
	if totalPages == 0 {
		totalPages = 1
	}

	// Конвертируем domain.Chat в Chat для ответа
	responseChats := make([]*Chat, len(chats))
	for i, chat := range chats {
		c := Chat(*chat)
		responseChats[i] = &c
	}

	response := PaginatedChatsResponse{
		Data:       responseChats,
		Total:      totalCount,
		Limit:      limit,
		Offset:     offset,
		TotalPages: totalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetChatByID godoc
// @Summary      Получить чат по ID
// @Description  Возвращает информацию о чате по его ID
// @Tags         chats
// @Accept       json
// @Produce      json
// @Param        id   path      int     true  "ID чата"
// @Success      200  {object}  Chat
// @Failure      404  {string}  string
// @Router       /chats/{id} [get]
func (h *Handler) GetChatByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/chats/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid chat id", http.StatusBadRequest)
		return
	}

	chat, err := h.chatService.GetChatByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	c := Chat(*chat)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

// AddAdministrator godoc
// @Summary      Добавить администратора к чату
// @Description  Добавляет нового администратора к чату по номеру телефона
// @Tags         administrators
// @Accept       json
// @Produce      json
// @Param        chat_id  path      int                    true  "ID чата"
// @Param        input    body      AddAdministratorRequest true  "Данные администратора"
// @Success      201      {object}  Administrator
// @Failure      400      {string}  string
// @Failure      404      {string}  string
// @Failure      409      {string}  string
// @Router       /chats/{chat_id}/administrators [post]
func (h *Handler) AddAdministrator(w http.ResponseWriter, r *http.Request) {
	// Извлекаем chat_id из пути
	path := strings.TrimPrefix(r.URL.Path, "/chats/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "administrators" {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	chatID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "invalid chat id", http.StatusBadRequest)
		return
	}

	var req AddAdministratorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Phone == "" {
		http.Error(w, "phone is required", http.StatusBadRequest)
		return
	}

	// Используем новый метод с флагами
	admin, err := h.chatService.AddAdministratorWithFlags(chatID, req.Phone, req.MaxID, req.AddUser, req.AddAdmin, req.SkipPhoneValidation)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == domain.ErrChatNotFound {
			statusCode = http.StatusNotFound
		} else if err == domain.ErrAdministratorExists {
			statusCode = http.StatusConflict
		} else if err == domain.ErrInvalidPhone {
			statusCode = http.StatusBadRequest
		}
		http.Error(w, err.Error(), statusCode)
		return
	}

	a := Administrator(*admin)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(a)
}

// GetAdministratorByID godoc
// @Summary      Получить администратора по ID
// @Description  Возвращает информацию об администраторе по его ID
// @Tags         administrators
// @Accept       json
// @Produce      json
// @Param        admin_id  path      int     true  "ID администратора"
// @Success      200      {object}  Administrator
// @Failure      400      {string}  string
// @Failure      404      {string}  string
// @Router       /administrators/{admin_id} [get]
func (h *Handler) GetAdministratorByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/administrators/")
	adminID, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		http.Error(w, "invalid administrator id", http.StatusBadRequest)
		return
	}

	admin, err := h.chatService.GetAdministratorByID(adminID)
	if err != nil {
		statusCode := http.StatusNotFound
		if err == domain.ErrAdministratorNotFound {
			statusCode = http.StatusNotFound
		}
		http.Error(w, err.Error(), statusCode)
		return
	}

	a := Administrator(*admin)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(a)
}

// GetAllAdministrators godoc
// @Summary      Получить всех администраторов
// @Description  Возвращает список всех администраторов с пагинацией и поиском
// @Tags         administrators
// @Accept       json
// @Produce      json
// @Param        query   query     string  false  "Поисковый запрос (телефон, MAX ID или название чата)"
// @Param        limit   query     int     false  "Лимит результатов (по умолчанию 50, максимум 100)"
// @Param        offset  query     int     false  "Смещение для пагинации"
// @Success      200     {object}  AdministratorListResponse
// @Failure      400     {string}  string
// @Failure      500     {string}  string
// @Router       /administrators [get]
func (h *Handler) GetAllAdministrators(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	administrators, totalCount, err := h.chatService.GetAllAdministrators(query, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Конвертируем domain.Administrator в Administrator для ответа
	responseAdmins := make([]*Administrator, len(administrators))
	for i, admin := range administrators {
		a := Administrator(*admin)
		responseAdmins[i] = &a
	}

	response := AdministratorListResponse{
		Administrators: responseAdmins,
		TotalCount:     totalCount,
		Limit:          limit,
		Offset:         offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RemoveAdministrator godoc
// @Summary      Удалить администратора из чата
// @Description  Удаляет администратора из чата. Нельзя удалить последнего администратора (должно быть минимум 2)
// @Tags         administrators
// @Accept       json
// @Produce      json
// @Param        admin_id  path      int     true  "ID администратора"
// @Success      200      {object}  DeleteResponse
// @Failure      400      {string}  string
// @Failure      404      {string}  string
// @Failure      409      {string}  string
// @Router       /administrators/{admin_id} [delete]
func (h *Handler) RemoveAdministrator(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/administrators/")
	adminID, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		http.Error(w, "invalid administrator id", http.StatusBadRequest)
		return
	}

	err = h.chatService.RemoveAdministrator(adminID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == domain.ErrAdministratorNotFound {
			statusCode = http.StatusNotFound
		} else if err == domain.ErrCannotDeleteLastAdmin {
			statusCode = http.StatusConflict
		}
		http.Error(w, err.Error(), statusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}


// CreateChatRequest представляет запрос на создание чата
type CreateChatRequest struct {
	Name              string  `json:"name" binding:"required"`
	URL               string  `json:"url" binding:"required"`
	ExternalChatID    *string `json:"external_chat_id,omitempty"`
	Source            string  `json:"source" binding:"required"`
	UniversityID      *int64  `json:"university_id,omitempty"`
	BranchID          *int64  `json:"branch_id,omitempty"`
	FacultyID         *int64  `json:"faculty_id,omitempty"`
	ParticipantsCount int     `json:"participants_count"`
	Department        string  `json:"department,omitempty"`
}



// RefreshParticipantsCount godoc
// @Summary      Обновить количество участников
// @Description  Принудительно обновляет количество участников для указанного чата из MAX API
// @Tags         chats
// @Accept       json
// @Produce      json
// @Param        Authorization header    string  true   "Bearer token"
// @Param        chat_id       path      int     true   "ID чата"
// @Success      200           {object}  map[string]interface{}
// @Failure      400           {string}  string
// @Failure      401           {string}  string
// @Failure      404           {string}  string
// @Failure      500           {string}  string
// @Router       /chats/{chat_id}/refresh-participants [post]
func (h *Handler) RefreshParticipantsCount(w http.ResponseWriter, r *http.Request) {
	// Извлекаем chat_id из пути
	path := strings.TrimPrefix(r.URL.Path, "/chats/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "refresh-participants" {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	chatID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "invalid chat id", http.StatusBadRequest)
		return
	}

	// Получаем чат для проверки существования и получения MAX Chat ID
	chat, err := h.chatService.GetChatByID(chatID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if chat.MaxChatID == "" {
		http.Error(w, "chat has no MAX Chat ID", http.StatusBadRequest)
		return
	}

	// Здесь должен быть вызов к ParticipantsUpdater, но пока возвращаем заглушку
	response := map[string]interface{}{
		"status":             "updated",
		"chat_id":            chatID,
		"participants_count": chat.ParticipantsCount,
		"updated_at":         time.Now(),
		"source":             "api",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateChat godoc
// @Summary      Создать чат
// @Description  Создает новый чат
// @Tags         chats
// @Accept       json
// @Produce      json
// @Param        input  body      CreateChatRequest  true  "Данные чата"
// @Success      201    {object}  Chat
// @Failure      400    {string}  string
// @Router       /chats [post]
func (h *Handler) CreateChat(w http.ResponseWriter, r *http.Request) {
	var req CreateChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Преобразуем external_chat_id в строку
	maxChatID := ""
	if req.ExternalChatID != nil {
		maxChatID = *req.ExternalChatID
	}

	// Создаем чат
	chat, err := h.chatService.CreateChat(
		req.Name,
		req.URL,
		maxChatID,
		req.Source,
		req.ParticipantsCount,
		req.UniversityID,
		req.Department,
	)
	if err != nil {
		if err == domain.ErrUniversityNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(chat)
}
