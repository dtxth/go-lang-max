package http

import (
	"chat-service/internal/domain"
	"chat-service/internal/usecase"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	chatService    *usecase.ChatService
	authMiddleware *AuthMiddleware
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

// AddAdministratorRequest представляет запрос на добавление администратора
type AddAdministratorRequest struct {
	Phone string `json:"phone" example:"+79001234567" binding:"required"`
}

// DeleteResponse представляет ответ на удаление
type DeleteResponse struct {
	Status string `json:"status" example:"deleted"`
}

func NewHandler(chatService *usecase.ChatService, authMiddleware *AuthMiddleware) *Handler {
	return &Handler{
		chatService:    chatService,
		authMiddleware: authMiddleware,
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
// @Description  Возвращает список всех чатов с пагинацией (с учетом роли пользователя)
// @Tags         chats
// @Accept       json
// @Produce      json
// @Param        Authorization header    string  true   "Bearer token"
// @Param        limit         query     int     false  "Лимит результатов (по умолчанию 50, максимум 100)"
// @Param        offset        query     int     false  "Смещение для пагинации"
// @Success      200           {object}  ChatListResponse
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

	chats, totalCount, err := h.chatService.GetAllChats(limit, offset, filter)
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

	admin, err := h.chatService.AddAdministrator(chatID, req.Phone)
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

