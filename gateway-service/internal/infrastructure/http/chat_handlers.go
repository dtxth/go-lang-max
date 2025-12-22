package http

import (
	"encoding/json"
	"net/http"
	"strings"

	chatpb "chat-service/api/proto"
)

// GetAllChatsHandler handles getting all chats with pagination
func (h *Handler) GetAllChatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed", h.getRequestID(r))
		return
	}

	requestID := h.getRequestID(r)
	ctx, cancel := h.createContextWithTimeout(r, h.config.Services.Chat.Timeout)
	defer cancel()

	page, limit, sortBy, sortOrder := h.parseQueryParams(r)

	req := &chatpb.GetAllChatsRequest{
		Page:      page,
		Limit:     limit,
		SortBy:    sortBy,
		SortOrder: sortOrder,
	}

	client := h.clientManager.GetChatClient()
	if !h.checkServiceAvailability(w, client, "Chat", requestID) {
		return
	}
	
	resp, err := client.GetAllChats(ctx, req)
	if err != nil {
		h.handleGRPCErrorLegacy(w, err, requestID)
		return
	}

	if resp.Error != "" {
		h.writeErrorResponse(w, http.StatusInternalServerError, "get_chats_failed", resp.Error, requestID)
		return
	}

	// Convert to HTTP response format
	chats := make([]map[string]interface{}, len(resp.Chats))
	for i, chat := range resp.Chats {
		chats[i] = map[string]interface{}{
			"id":                 chat.Id,
			"name":               chat.Name,
			"url":                chat.Url,
			"max_chat_id":        chat.MaxChatId,
			"participants_count": chat.ParticipantsCount,
			"university_id":      chat.UniversityId,
			"department":         chat.Department,
			"source":             chat.Source,
			"created_at":         chat.CreatedAt,
			"updated_at":         chat.UpdatedAt,
		}
	}

	response := map[string]interface{}{
		"chats": chats,
		"total": resp.Total,
		"page":  resp.Page,
		"limit": resp.Limit,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// SearchChatsHandler handles searching chats
func (h *Handler) SearchChatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed", h.getRequestID(r))
		return
	}

	requestID := h.getRequestID(r)
	ctx, cancel := h.createContextWithTimeout(r, h.config.Services.Chat.Timeout)
	defer cancel()

	query := r.URL.Query().Get("query")
	if query == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "missing_query", "Query parameter is required", requestID)
		return
	}

	page, limit, sortBy, sortOrder := h.parseQueryParams(r)

	req := &chatpb.SearchChatsRequest{
		Query:     query,
		Page:      page,
		Limit:     limit,
		SortBy:    sortBy,
		SortOrder: sortOrder,
	}

	client := h.clientManager.GetChatClient()
	resp, err := client.SearchChats(ctx, req)
	if err != nil {
		h.handleGRPCErrorLegacy(w, err, requestID)
		return
	}

	if resp.Error != "" {
		h.writeErrorResponse(w, http.StatusInternalServerError, "search_chats_failed", resp.Error, requestID)
		return
	}

	// Convert to HTTP response format
	chats := make([]map[string]interface{}, len(resp.Chats))
	for i, chat := range resp.Chats {
		chats[i] = map[string]interface{}{
			"id":                 chat.Id,
			"name":               chat.Name,
			"url":                chat.Url,
			"max_chat_id":        chat.MaxChatId,
			"participants_count": chat.ParticipantsCount,
			"university_id":      chat.UniversityId,
			"department":         chat.Department,
			"source":             chat.Source,
			"created_at":         chat.CreatedAt,
			"updated_at":         chat.UpdatedAt,
		}
	}

	response := map[string]interface{}{
		"chats": chats,
		"total": resp.Total,
		"page":  resp.Page,
		"limit": resp.Limit,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetChatByIDHandler handles getting a specific chat by ID
func (h *Handler) GetChatByIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed", h.getRequestID(r))
		return
	}

	requestID := h.getRequestID(r)
	ctx, cancel := h.createContextWithTimeout(r, h.config.Services.Chat.Timeout)
	defer cancel()

	// Extract chat ID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_path", "Chat ID is required", requestID)
		return
	}

	chatID, err := h.parseIntParam(pathParts[1])
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_chat_id", "Invalid chat ID format", requestID)
		return
	}

	req := &chatpb.GetChatByIDRequest{
		Id: chatID,
	}

	client := h.clientManager.GetChatClient()
	resp, err := client.GetChatByID(ctx, req)
	if err != nil {
		h.handleGRPCErrorLegacy(w, err, requestID)
		return
	}

	if resp.Error != "" {
		if strings.Contains(resp.Error, "not found") {
			h.writeErrorResponse(w, http.StatusNotFound, "chat_not_found", resp.Error, requestID)
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, "get_chat_failed", resp.Error, requestID)
		}
		return
	}

	// Convert to HTTP response format
	response := map[string]interface{}{
		"id":                 resp.Chat.Id,
		"name":               resp.Chat.Name,
		"url":                resp.Chat.Url,
		"max_chat_id":        resp.Chat.MaxChatId,
		"participants_count": resp.Chat.ParticipantsCount,
		"university_id":      resp.Chat.UniversityId,
		"department":         resp.Chat.Department,
		"source":             resp.Chat.Source,
		"created_at":         resp.Chat.CreatedAt,
		"updated_at":         resp.Chat.UpdatedAt,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// CreateChatHandler handles creating a new chat
func (h *Handler) CreateChatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed", h.getRequestID(r))
		return
	}

	requestID := h.getRequestID(r)
	ctx, cancel := h.createContextWithTimeout(r, h.config.Services.Chat.Timeout)
	defer cancel()

	var req chatpb.CreateChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format", requestID)
		return
	}

	client := h.clientManager.GetChatClient()
	resp, err := client.CreateChat(ctx, &req)
	if err != nil {
		h.handleGRPCErrorLegacy(w, err, requestID)
		return
	}

	if resp.Error != "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "create_chat_failed", resp.Error, requestID)
		return
	}

	// Convert to HTTP response format
	response := map[string]interface{}{
		"id":                 resp.Chat.Id,
		"name":               resp.Chat.Name,
		"url":                resp.Chat.Url,
		"max_chat_id":        resp.Chat.MaxChatId,
		"participants_count": resp.Chat.ParticipantsCount,
		"university_id":      resp.Chat.UniversityId,
		"department":         resp.Chat.Department,
		"source":             resp.Chat.Source,
		"created_at":         resp.Chat.CreatedAt,
		"updated_at":         resp.Chat.UpdatedAt,
	}

	h.writeJSONResponse(w, http.StatusCreated, response)
}

// GetAllAdministratorsHandler handles getting all administrators
func (h *Handler) GetAllAdministratorsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed", h.getRequestID(r))
		return
	}

	requestID := h.getRequestID(r)
	ctx, cancel := h.createContextWithTimeout(r, h.config.Services.Chat.Timeout)
	defer cancel()

	page, limit, _, _ := h.parseQueryParams(r)

	req := &chatpb.GetAllAdministratorsRequest{
		Page:  page,
		Limit: limit,
	}

	client := h.clientManager.GetChatClient()
	resp, err := client.GetAllAdministrators(ctx, req)
	if err != nil {
		h.handleGRPCErrorLegacy(w, err, requestID)
		return
	}

	if resp.Error != "" {
		h.writeErrorResponse(w, http.StatusInternalServerError, "get_administrators_failed", resp.Error, requestID)
		return
	}

	// Convert to HTTP response format
	administrators := make([]map[string]interface{}, len(resp.Administrators))
	for i, admin := range resp.Administrators {
		administrators[i] = map[string]interface{}{
			"id":         admin.Id,
			"chat_id":    admin.ChatId,
			"phone":      admin.Phone,
			"max_id":     admin.MaxId,
			"add_user":   admin.AddUser,
			"add_admin":  admin.AddAdmin,
			"created_at": admin.CreatedAt,
		}
	}

	response := map[string]interface{}{
		"administrators": administrators,
		"total":          resp.Total,
		"page":           resp.Page,
		"limit":          resp.Limit,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetAdministratorByIDHandler handles getting a specific administrator by ID
func (h *Handler) GetAdministratorByIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed", h.getRequestID(r))
		return
	}

	requestID := h.getRequestID(r)
	ctx, cancel := h.createContextWithTimeout(r, h.config.Services.Chat.Timeout)
	defer cancel()

	// Extract administrator ID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_path", "Administrator ID is required", requestID)
		return
	}

	adminID, err := h.parseIntParam(pathParts[2])
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_admin_id", "Invalid administrator ID format", requestID)
		return
	}

	req := &chatpb.GetAdministratorByIDRequest{
		Id: adminID,
	}

	client := h.clientManager.GetChatClient()
	resp, err := client.GetAdministratorByID(ctx, req)
	if err != nil {
		h.handleGRPCErrorLegacy(w, err, requestID)
		return
	}

	if resp.Error != "" {
		if strings.Contains(resp.Error, "not found") {
			h.writeErrorResponse(w, http.StatusNotFound, "administrator_not_found", resp.Error, requestID)
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, "get_administrator_failed", resp.Error, requestID)
		}
		return
	}

	// Convert to HTTP response format
	response := map[string]interface{}{
		"id":         resp.Administrator.Id,
		"chat_id":    resp.Administrator.ChatId,
		"phone":      resp.Administrator.Phone,
		"max_id":     resp.Administrator.MaxId,
		"add_user":   resp.Administrator.AddUser,
		"add_admin":  resp.Administrator.AddAdmin,
		"created_at": resp.Administrator.CreatedAt,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// AddAdministratorHandler handles adding an administrator to a chat
func (h *Handler) AddAdministratorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed", h.getRequestID(r))
		return
	}

	requestID := h.getRequestID(r)
	ctx, cancel := h.createContextWithTimeout(r, h.config.Services.Chat.Timeout)
	defer cancel()

	var req chatpb.AddAdministratorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format", requestID)
		return
	}

	client := h.clientManager.GetChatClient()
	resp, err := client.AddAdministrator(ctx, &req)
	if err != nil {
		h.handleGRPCErrorLegacy(w, err, requestID)
		return
	}

	if resp.Error != "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "add_administrator_failed", resp.Error, requestID)
		return
	}

	// Convert to HTTP response format
	response := map[string]interface{}{
		"id":         resp.Administrator.Id,
		"chat_id":    resp.Administrator.ChatId,
		"phone":      resp.Administrator.Phone,
		"max_id":     resp.Administrator.MaxId,
		"add_user":   resp.Administrator.AddUser,
		"add_admin":  resp.Administrator.AddAdmin,
		"created_at": resp.Administrator.CreatedAt,
	}

	h.writeJSONResponse(w, http.StatusCreated, response)
}

// RemoveAdministratorHandler handles removing an administrator
func (h *Handler) RemoveAdministratorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed", h.getRequestID(r))
		return
	}

	requestID := h.getRequestID(r)
	ctx, cancel := h.createContextWithTimeout(r, h.config.Services.Chat.Timeout)
	defer cancel()

	// Extract administrator ID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_path", "Administrator ID is required", requestID)
		return
	}

	adminID, err := h.parseIntParam(pathParts[2])
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_admin_id", "Invalid administrator ID format", requestID)
		return
	}

	req := &chatpb.RemoveAdministratorRequest{
		Id: adminID,
	}

	client := h.clientManager.GetChatClient()
	resp, err := client.RemoveAdministrator(ctx, req)
	if err != nil {
		h.handleGRPCErrorLegacy(w, err, requestID)
		return
	}

	if resp.Error != "" {
		if strings.Contains(resp.Error, "not found") {
			h.writeErrorResponse(w, http.StatusNotFound, "administrator_not_found", resp.Error, requestID)
		} else {
			h.writeErrorResponse(w, http.StatusBadRequest, "remove_administrator_failed", resp.Error, requestID)
		}
		return
	}

	response := map[string]interface{}{
		"success": resp.Success,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// RefreshParticipantsCountHandler handles refreshing participants count for a chat
func (h *Handler) RefreshParticipantsCountHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed", h.getRequestID(r))
		return
	}

	requestID := h.getRequestID(r)
	ctx, cancel := h.createContextWithTimeout(r, h.config.Services.Chat.Timeout)
	defer cancel()

	var req chatpb.RefreshParticipantsCountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format", requestID)
		return
	}

	client := h.clientManager.GetChatClient()
	resp, err := client.RefreshParticipantsCount(ctx, &req)
	if err != nil {
		h.handleGRPCErrorLegacy(w, err, requestID)
		return
	}

	if resp.Error != "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "refresh_participants_failed", resp.Error, requestID)
		return
	}

	response := map[string]interface{}{
		"participants_count": resp.ParticipantsCount,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}