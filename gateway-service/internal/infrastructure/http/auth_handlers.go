package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	authpb "auth-service/api/proto"
)

// RegisterHandler handles user registration
func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed", h.getRequestID(r))
		return
	}

	requestID := h.getRequestID(r)
	ctx, cancel := h.createContextWithTimeout(r, h.config.Services.Auth.Timeout)
	defer cancel()

	var req authpb.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format", requestID)
		return
	}

	client := h.clientManager.GetAuthClient()
	if !h.checkServiceAvailability(w, client, "Auth", requestID) {
		return
	}
	
	var resp *authpb.RegisterResponse
	
	err := h.executeWithRetryAndCircuitBreaker(ctx, "auth", func(ctx context.Context) error {
		var err error
		resp, err = client.Register(ctx, &req)
		return err
	})
	
	if err != nil {
		h.handleGRPCError(w, err, requestID, "auth", "Register")
		return
	}

	if resp.Error != "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "registration_failed", resp.Error, requestID)
		return
	}

	// Convert to HTTP response format
	response := map[string]interface{}{
		"access_token":  resp.Tokens.AccessToken,
		"refresh_token": resp.Tokens.RefreshToken,
		"id":           resp.User.Id,
		"email":        resp.User.Email,
		"phone":        resp.User.Phone,
		"role":         resp.User.Role,
		"created_at":   resp.User.CreatedAt,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// LoginHandler handles user login by email
func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed", h.getRequestID(r))
		return
	}

	requestID := h.getRequestID(r)
	ctx, cancel := h.createContextWithTimeout(r, h.config.Services.Auth.Timeout)
	defer cancel()

	var req authpb.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format", requestID)
		return
	}

	client := h.clientManager.GetAuthClient()
	if !h.checkServiceAvailability(w, client, "Auth", requestID) {
		return
	}
	
	var resp *authpb.LoginResponse
	
	err := h.executeWithRetryAndCircuitBreaker(ctx, "auth", func(ctx context.Context) error {
		var err error
		resp, err = client.Login(ctx, &req)
		return err
	})
	
	if err != nil {
		h.handleGRPCError(w, err, requestID, "auth", "Login")
		return
	}

	if resp.Error != "" {
		h.writeErrorResponse(w, http.StatusUnauthorized, "login_failed", resp.Error, requestID)
		return
	}

	// Convert to HTTP response format
	response := map[string]interface{}{
		"access_token":  resp.Tokens.AccessToken,
		"refresh_token": resp.Tokens.RefreshToken,
		"id":           resp.User.Id,
		"email":        resp.User.Email,
		"phone":        resp.User.Phone,
		"role":         resp.User.Role,
		"created_at":   resp.User.CreatedAt,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// LoginByPhoneHandler handles user login by phone
func (h *Handler) LoginByPhoneHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed", h.getRequestID(r))
		return
	}

	requestID := h.getRequestID(r)
	ctx, cancel := h.createContextWithTimeout(r, h.config.Services.Auth.Timeout)
	defer cancel()

	var req authpb.LoginByPhoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format", requestID)
		return
	}

	client := h.clientManager.GetAuthClient()
	var resp *authpb.LoginResponse
	
	err := h.executeWithRetryAndCircuitBreaker(ctx, "auth", func(ctx context.Context) error {
		var err error
		resp, err = client.LoginByPhone(ctx, &req)
		return err
	})
	
	if err != nil {
		h.handleGRPCError(w, err, requestID, "auth", "LoginByPhone")
		return
	}

	if resp.Error != "" {
		h.writeErrorResponse(w, http.StatusUnauthorized, "login_failed", resp.Error, requestID)
		return
	}

	// Convert to HTTP response format
	response := map[string]interface{}{
		"access_token":  resp.Tokens.AccessToken,
		"refresh_token": resp.Tokens.RefreshToken,
		"id":           resp.User.Id,
		"email":        resp.User.Email,
		"phone":        resp.User.Phone,
		"role":         resp.User.Role,
		"created_at":   resp.User.CreatedAt,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// RefreshHandler handles token refresh
func (h *Handler) RefreshHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed", h.getRequestID(r))
		return
	}

	requestID := h.getRequestID(r)
	ctx, cancel := h.createContextWithTimeout(r, h.config.Services.Auth.Timeout)
	defer cancel()

	var req authpb.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format", requestID)
		return
	}

	client := h.clientManager.GetAuthClient()
	var resp *authpb.RefreshResponse
	
	err := h.executeWithRetryAndCircuitBreaker(ctx, "auth", func(ctx context.Context) error {
		var err error
		resp, err = client.Refresh(ctx, &req)
		return err
	})
	
	if err != nil {
		h.handleGRPCError(w, err, requestID, "auth", "Refresh")
		return
	}

	if resp.Error != "" {
		h.writeErrorResponse(w, http.StatusUnauthorized, "refresh_failed", resp.Error, requestID)
		return
	}

	// Convert to HTTP response format
	response := map[string]interface{}{
		"access_token":  resp.Tokens.AccessToken,
		"refresh_token": resp.Tokens.RefreshToken,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// LogoutHandler handles user logout
func (h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed", h.getRequestID(r))
		return
	}

	requestID := h.getRequestID(r)
	ctx, cancel := h.createContextWithTimeout(r, h.config.Services.Auth.Timeout)
	defer cancel()

	var req authpb.LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format", requestID)
		return
	}

	client := h.clientManager.GetAuthClient()
	var resp *authpb.LogoutResponse
	
	err := h.executeWithRetryAndCircuitBreaker(ctx, "auth", func(ctx context.Context) error {
		var err error
		resp, err = client.Logout(ctx, &req)
		return err
	})
	
	if err != nil {
		h.handleGRPCError(w, err, requestID, "auth", "Logout")
		return
	}

	if resp.Error != "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "logout_failed", resp.Error, requestID)
		return
	}

	response := map[string]interface{}{
		"success": resp.Success,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// AuthenticateMAXHandler handles MAX authentication
func (h *Handler) AuthenticateMAXHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed", h.getRequestID(r))
		return
	}

	requestID := h.getRequestID(r)
	ctx, cancel := h.createContextWithTimeout(r, h.config.Services.Auth.Timeout)
	defer cancel()

	var req authpb.AuthenticateMAXRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format", requestID)
		return
	}

	client := h.clientManager.GetAuthClient()
	var resp *authpb.AuthenticateMAXResponse
	
	err := h.executeWithRetryAndCircuitBreaker(ctx, "auth", func(ctx context.Context) error {
		var err error
		resp, err = client.AuthenticateMAX(ctx, &req)
		return err
	})
	
	if err != nil {
		h.handleGRPCError(w, err, requestID, "auth", "AuthenticateMAX")
		return
	}

	if resp.Error != "" {
		h.writeErrorResponse(w, http.StatusUnauthorized, "max_auth_failed", resp.Error, requestID)
		return
	}

	// Convert to HTTP response format
	response := map[string]interface{}{
		"access_token":  resp.Tokens.AccessToken,
		"refresh_token": resp.Tokens.RefreshToken,
		"id":           resp.User.Id,
		"email":        resp.User.Email,
		"phone":        resp.User.Phone,
		"role":         resp.User.Role,
		"created_at":   resp.User.CreatedAt,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetBotMeHandler handles bot info requests
func (h *Handler) GetBotMeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed", h.getRequestID(r))
		return
	}

	requestID := h.getRequestID(r)
	ctx, cancel := h.createContextWithTimeout(r, h.config.Services.Auth.Timeout)
	defer cancel()

	req := &authpb.GetBotMeRequest{}

	client := h.clientManager.GetAuthClient()
	var resp *authpb.GetBotMeResponse
	
	err := h.executeWithRetryAndCircuitBreaker(ctx, "auth", func(ctx context.Context) error {
		var err error
		resp, err = client.GetBotMe(ctx, req)
		return err
	})
	
	if err != nil {
		h.handleGRPCError(w, err, requestID, "auth", "GetBotMe")
		return
	}

	if resp.Error != "" {
		h.writeErrorResponse(w, http.StatusInternalServerError, "bot_info_failed", resp.Error, requestID)
		return
	}

	// Convert to HTTP response format
	response := map[string]interface{}{
		"id":         resp.Bot.Id,
		"username":   resp.Bot.Username,
		"first_name": resp.Bot.FirstName,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetMetricsHandler handles metrics requests
func (h *Handler) GetMetricsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed", h.getRequestID(r))
		return
	}

	requestID := h.getRequestID(r)
	ctx, cancel := h.createContextWithTimeout(r, h.config.Services.Auth.Timeout)
	defer cancel()

	req := &authpb.GetMetricsRequest{}

	client := h.clientManager.GetAuthClient()
	var resp *authpb.GetMetricsResponse
	
	err := h.executeWithRetryAndCircuitBreaker(ctx, "auth", func(ctx context.Context) error {
		var err error
		resp, err = client.GetMetrics(ctx, req)
		return err
	})
	
	if err != nil {
		h.handleGRPCError(w, err, requestID, "auth", "GetMetrics")
		return
	}

	if resp.Error != "" {
		h.writeErrorResponse(w, http.StatusInternalServerError, "metrics_failed", resp.Error, requestID)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, resp.Metrics)
}

// HealthHandler handles health check requests for the Gateway Service
func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed", h.getRequestID(r))
		return
	}

	ctx, cancel := h.createContextWithTimeout(r, 5*time.Second)
	defer cancel()

	// Check health of all gRPC connections
	healthStatus := h.clientManager.HealthCheck(ctx)
	
	// Determine overall health status
	allHealthy := true
	for _, status := range healthStatus {
		if status != "healthy" {
			allHealthy = false
			break
		}
	}

	// Build response
	response := map[string]interface{}{
		"status":   "ok",
		"services": healthStatus,
	}

	// If any service is unhealthy, return 503 Service Unavailable
	statusCode := http.StatusOK
	if !allHealthy {
		response["status"] = "degraded"
		statusCode = http.StatusServiceUnavailable
	}

	h.writeJSONResponse(w, statusCode, response)
}

// RequestPasswordResetHandler handles password reset requests
func (h *Handler) RequestPasswordResetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed", h.getRequestID(r))
		return
	}

	requestID := h.getRequestID(r)
	ctx, cancel := h.createContextWithTimeout(r, h.config.Services.Auth.Timeout)
	defer cancel()

	var req authpb.RequestPasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format", requestID)
		return
	}

	client := h.clientManager.GetAuthClient()
	var resp *authpb.RequestPasswordResetResponse
	
	err := h.executeWithRetryAndCircuitBreaker(ctx, "auth", func(ctx context.Context) error {
		var err error
		resp, err = client.RequestPasswordReset(ctx, &req)
		return err
	})
	
	if err != nil {
		h.handleGRPCError(w, err, requestID, "auth", "RequestPasswordReset")
		return
	}

	if resp.Error != "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "password_reset_failed", resp.Error, requestID)
		return
	}

	response := map[string]interface{}{
		"success": resp.Success,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// ResetPasswordHandler handles password reset with token
func (h *Handler) ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed", h.getRequestID(r))
		return
	}

	requestID := h.getRequestID(r)
	ctx, cancel := h.createContextWithTimeout(r, h.config.Services.Auth.Timeout)
	defer cancel()

	var req authpb.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format", requestID)
		return
	}

	client := h.clientManager.GetAuthClient()
	var resp *authpb.ResetPasswordResponse
	
	err := h.executeWithRetryAndCircuitBreaker(ctx, "auth", func(ctx context.Context) error {
		var err error
		resp, err = client.ResetPassword(ctx, &req)
		return err
	})
	
	if err != nil {
		h.handleGRPCError(w, err, requestID, "auth", "ResetPassword")
		return
	}

	if resp.Error != "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "password_reset_failed", resp.Error, requestID)
		return
	}

	response := map[string]interface{}{
		"success": resp.Success,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// ChangePasswordHandler handles password change for authenticated users
func (h *Handler) ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed", h.getRequestID(r))
		return
	}

	requestID := h.getRequestID(r)
	ctx, cancel := h.createContextWithTimeout(r, h.config.Services.Auth.Timeout)
	defer cancel()

	// Extract user ID from auth token (simplified - in real implementation would validate token)
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		h.writeErrorResponse(w, http.StatusUnauthorized, "missing_auth", "Authorization header required", requestID)
		return
	}

	var reqBody struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format", requestID)
		return
	}

	// For now, we'll use a placeholder user ID - in real implementation, extract from validated token
	req := &authpb.ChangePasswordRequest{
		UserId:          1, // This should be extracted from validated token
		CurrentPassword: reqBody.CurrentPassword,
		NewPassword:     reqBody.NewPassword,
	}

	client := h.clientManager.GetAuthClient()
	var resp *authpb.ChangePasswordResponse
	
	err := h.executeWithRetryAndCircuitBreaker(ctx, "auth", func(ctx context.Context) error {
		var err error
		resp, err = client.ChangePassword(ctx, req)
		return err
	})
	
	if err != nil {
		h.handleGRPCError(w, err, requestID, "auth", "ChangePassword")
		return
	}

	if resp.Error != "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "password_change_failed", resp.Error, requestID)
		return
	}

	response := map[string]interface{}{
		"success": resp.Success,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}