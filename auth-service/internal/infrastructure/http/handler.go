package http

import (
	"auth-service/internal/domain"
	"auth-service/internal/infrastructure/errors"
	"auth-service/internal/infrastructure/middleware"
	"auth-service/internal/infrastructure/phone"
	"auth-service/internal/usecase"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type Handler struct {
    auth *usecase.AuthService
}

func NewHandler(auth *usecase.AuthService) *Handler {
    return &Handler{auth: auth}
}

// GetMetrics godoc
// @Summary      Get metrics
// @Description  Returns current metrics for password operations and notifications
// @Tags         monitoring
// @Produce      json
// @Success      200  {object}  object  "Metrics snapshot"
// @Router       /metrics [get]
func (h *Handler) GetMetrics(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    if h.auth == nil || h.auth.GetMetrics() == nil {
        w.WriteHeader(http.StatusServiceUnavailable)
        json.NewEncoder(w).Encode(map[string]string{"error": "metrics not available"})
        return
    }
    
    snapshot := h.auth.GetMetrics().GetMetrics()
    
    response := map[string]interface{}{
        "user_creations":        snapshot.UserCreations,
        "password_resets":       snapshot.PasswordResets,
        "password_changes":      snapshot.PasswordChanges,
        "notifications_sent":    snapshot.NotificationsSent,
        "notifications_failed":  snapshot.NotificationsFailed,
        "tokens_generated":      snapshot.TokensGenerated,
        "tokens_used":           snapshot.TokensUsed,
        "tokens_expired":        snapshot.TokensExpired,
        "tokens_invalidated":    snapshot.TokensInvalidated,
        "maxbot_healthy":        snapshot.MaxBotHealthy,
        "last_health_check":     snapshot.LastHealthCheck,
        "notification_success_rate": h.auth.GetMetrics().GetNotificationSuccessRate(),
        "notification_failure_rate": h.auth.GetMetrics().GetNotificationFailureRate(),
    }
    
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}

// Register godoc
// @Summary      Register new user
// @Description  Creates user and stores hashed password. Provide either email or phone.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body      object{email=string,phone=string,password=string,role=string}  true  "User credentials (provide either email or phone, role is optional, defaults to operator)"
// @Success      200    {object}  domain.User
// @Failure      400    {string}  string
// @Router       /register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
    requestID := middleware.GetRequestID(r.Context())
    
    var req struct {
        Email    string `json:"email"`
        Phone    string `json:"phone"`
        Password string `json:"password"`
        Role     string `json:"role"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        errors.WriteError(w, errors.ValidationError("invalid request body").WithError(err), requestID)
        return
    }

    // Validate required fields - either email or phone must be provided
    if req.Email == "" && req.Phone == "" {
        errors.WriteError(w, errors.ValidationError("either email or phone must be provided"), requestID)
        return
    }
    if req.Password == "" {
        errors.WriteError(w, errors.MissingFieldError("password"), requestID)
        return
    }

    var user *domain.User
    var err error

    if req.Phone != "" {
        // Normalize phone number
        normalizedPhone := phone.NormalizePhone(req.Phone)
        user, err = h.auth.RegisterByPhone(normalizedPhone, req.Password, req.Role)
    } else {
        user, err = h.auth.Register(req.Email, req.Password, req.Role)
    }

    if err != nil {
        errors.WriteError(w, err, requestID)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}

// Login godoc
// @Summary      Login user
// @Description  Returns access and refresh tokens. Supports login by email or phone.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body      object{email=string,phone=string,password=string}  true  "User credentials (provide either email or phone)"
// @Success      200    {object}  domain.TokenPair
// @Failure      401    {string}  string
// @Router       /login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
    requestID := middleware.GetRequestID(r.Context())
    
    var req struct {
        Email    string `json:"email"`
        Phone    string `json:"phone"`
        Password string `json:"password"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        errors.WriteError(w, errors.ValidationError("invalid request body").WithError(err), requestID)
        return
    }


    
    // Validate that either email or phone is provided
    if req.Email == "" && req.Phone == "" {
        errors.WriteError(w, errors.MissingFieldError("email or phone"), requestID)
        return
    }
    if req.Password == "" {
        errors.WriteError(w, errors.MissingFieldError("password"), requestID)
        return
    }

    // Use phone if provided, otherwise use email
    identifier := req.Email
    if req.Phone != "" {
        // Normalize phone number to +7XXXXXXXXXX format
        normalizedPhone := phone.NormalizePhone(req.Phone)
        identifier = normalizedPhone
        

    }

    tokens, err := h.auth.LoginByIdentifier(identifier, req.Password)
    if err != nil {
        errors.WriteError(w, err, requestID)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "access_token":  tokens.AccessToken,
        "refresh_token": tokens.RefreshToken,
    })
}

// LoginByPhone godoc
// @Summary      Login user by phone
// @Description  Returns access and refresh tokens for phone-based login
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body      object{phone=string,password=string}  true  "User credentials"
// @Success      200    {object}  domain.TokenPair
// @Failure      401    {string}  string
// @Router       /login-phone [post]
func (h *Handler) LoginByPhone(w http.ResponseWriter, r *http.Request) {
    log.Printf("[DEBUG] LoginByPhone called")
    requestID := middleware.GetRequestID(r.Context())

    
    var req struct {
        Phone    string `json:"phone"`
        Password string `json:"password"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        errors.WriteError(w, errors.ValidationError("invalid request body").WithError(err), requestID)
        return
    }

    // Validate required fields
    if req.Phone == "" {
        errors.WriteError(w, errors.MissingFieldError("phone"), requestID)
        return
    }
    if req.Password == "" {
        errors.WriteError(w, errors.MissingFieldError("password"), requestID)
        return
    }

    // Normalize phone number to +7XXXXXXXXXX format
    normalizedPhone := phone.NormalizePhone(req.Phone)
    
    // Добавим логирование для отладки
    log.Printf("[DEBUG] LoginByPhone: original=%s, normalized=%s", req.Phone, normalizedPhone)

    tokens, err := h.auth.LoginByIdentifier(normalizedPhone, req.Password)
    if err != nil {
        log.Printf("[DEBUG] LoginByPhone: LoginByIdentifier failed with error: %v", err)
        errors.WriteError(w, err, requestID)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "access_token":  tokens.AccessToken,
        "refresh_token": tokens.RefreshToken,
    })
}

// Refresh godoc
// @Summary      Refresh access token
// @Description  Returns new access and refresh tokens using refresh token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body      object{refresh_token=string}  true  "Refresh token"
// @Success      200    {object}  domain.TokenPair
// @Failure      400    {string}  string  "Invalid request"
// @Failure      401    {string}  string  "Invalid or expired refresh token"
// @Router       /refresh [post]
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
    requestID := middleware.GetRequestID(r.Context())
    
    var req struct {
        RefreshToken string `json:"refresh_token"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        errors.WriteError(w, errors.ValidationError("invalid request body").WithError(err), requestID)
        return
    }
    
    if req.RefreshToken == "" {
        errors.WriteError(w, errors.MissingFieldError("refresh_token"), requestID)
        return
    }

    tokens, err := h.auth.Refresh(req.RefreshToken)
    if err != nil {
        errors.WriteError(w, err, requestID)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "access_token":  tokens.AccessToken,
        "refresh_token": tokens.RefreshToken,
    })
}

// Logout godoc
// @Summary      Logout user
// @Description  Invalidates refresh token and logs out user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body      object{refresh_token=string}  true  "Refresh token"
// @Success      200    {object}  object{status=string}  "Successfully logged out"
// @Failure      400    {string}  string  "Invalid request"
// @Failure      401    {string}  string  "Invalid refresh token"
// @Router       /logout [post]
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
    requestID := middleware.GetRequestID(r.Context())
    
    var req struct {
        RefreshToken string `json:"refresh_token"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        errors.WriteError(w, errors.ValidationError("invalid request body").WithError(err), requestID)
        return
    }
    
    if req.RefreshToken == "" {
        errors.WriteError(w, errors.MissingFieldError("refresh_token"), requestID)
        return
    }

    if err := h.auth.Logout(req.RefreshToken); err != nil {
        errors.WriteError(w, err, requestID)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "logged_out"})
}

// RequestPasswordReset godoc
// @Summary      Request password reset
// @Description  Generates a reset token and sends it to the user's phone via MAX Messenger
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body      object{phone=string}  true  "User phone number"
// @Success      200    {object}  object{success=bool,message=string}
// @Failure      400    {string}  string
// @Router       /auth/password-reset/request [post]
func (h *Handler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
    requestID := middleware.GetRequestID(r.Context())
    
    var req struct {
        Phone string `json:"phone"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        errors.WriteError(w, errors.ValidationError("invalid request body").WithError(err), requestID)
        return
    }

    if req.Phone == "" {
        errors.WriteError(w, errors.MissingFieldError("phone"), requestID)
        return
    }

    if err := h.auth.RequestPasswordReset(req.Phone); err != nil {
        errors.WriteError(w, err, requestID)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "message": "Password reset token sent to your phone",
    })
}

// ResetPassword godoc
// @Summary      Reset password with token
// @Description  Validates reset token and updates user password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body      object{token=string,new_password=string}  true  "Reset token and new password"
// @Success      200    {object}  object{success=bool,message=string}
// @Failure      400    {string}  string
// @Failure      401    {string}  string
// @Router       /auth/password-reset/confirm [post]
func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
    requestID := middleware.GetRequestID(r.Context())
    
    var req struct {
        Token       string `json:"token"`
        NewPassword string `json:"new_password"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        errors.WriteError(w, errors.ValidationError("invalid request body").WithError(err), requestID)
        return
    }

    if req.Token == "" {
        errors.WriteError(w, errors.MissingFieldError("token"), requestID)
        return
    }
    if req.NewPassword == "" {
        errors.WriteError(w, errors.MissingFieldError("new_password"), requestID)
        return
    }

    if err := h.auth.ResetPassword(req.Token, req.NewPassword); err != nil {
        errors.WriteError(w, err, requestID)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "message": "Password successfully reset",
    })
}

// ChangePassword godoc
// @Summary      Change password
// @Description  Allows authenticated user to change their password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer token"
// @Param        input          body      object{current_password=string,new_password=string}  true  "Current and new password"
// @Success      200            {object}  object{success=bool,message=string}
// @Failure      400            {string}  string
// @Failure      401            {string}  string
// @Router       /auth/password/change [post]
func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
    requestID := middleware.GetRequestID(r.Context())
    
    // Extract user ID from context (set by auth middleware)
    userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
    if !ok || userID == 0 {
        errors.WriteError(w, errors.UnauthorizedError("authentication required"), requestID)
        return
    }
    
    var req struct {
        CurrentPassword string `json:"current_password"`
        NewPassword     string `json:"new_password"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        errors.WriteError(w, errors.ValidationError("invalid request body").WithError(err), requestID)
        return
    }

    if req.CurrentPassword == "" {
        errors.WriteError(w, errors.MissingFieldError("current_password"), requestID)
        return
    }
    if req.NewPassword == "" {
        errors.WriteError(w, errors.MissingFieldError("new_password"), requestID)
        return
    }

    if err := h.auth.ChangePassword(userID, req.CurrentPassword, req.NewPassword); err != nil {
        errors.WriteError(w, err, requestID)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "message": "Password successfully changed",
    })
}

// Health godoc
// @Summary      Health check
// @Description  Returns service health status
// @Tags         health
// @Produce      json
// @Success      200  {object}  object{status=string}  "Service is healthy"
// @Router       /health [get]
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// BotInfoResponse represents the response for /bot/me endpoint
type BotInfoResponse struct {
    Name    string `json:"name" example:"MAX Bot"`                    // Bot name
    AddLink string `json:"add_link" example:"https://max.ru/add-bot"` // Link to add the bot
}

// MaxAuthRequest represents the request for MAX Mini App authentication
type MaxAuthRequest struct {
    InitData string `json:"init_data" binding:"required"`
}

// MaxAuthResponse represents the response for MAX Mini App authentication
type MaxAuthResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
}

// AuthenticateMAX godoc
// @Summary      Authenticate MAX Mini App user
// @Description  Validates MAX Mini App initData and returns JWT tokens
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body      MaxAuthRequest  true  "MAX initData"
// @Success      200    {object}  MaxAuthResponse  "Authentication successful"
// @Failure      400    {string}  string  "Invalid request"
// @Failure      401    {string}  string  "Authentication failed"
// @Failure      500    {string}  string  "Internal server error"
// @Router       /auth/max [post]
func (h *Handler) AuthenticateMAX(w http.ResponseWriter, r *http.Request) {
    requestID := middleware.GetRequestID(r.Context())
    
    var req MaxAuthRequest
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        errors.WriteError(w, errors.ValidationError("invalid request body").WithError(err), requestID)
        return
    }

    // Validate required fields
    if req.InitData == "" {
        errors.WriteError(w, errors.MissingFieldError("init_data"), requestID)
        return
    }

    // Authenticate using MAX initData
    tokens, err := h.auth.AuthenticateMAX(req.InitData)
    if err != nil {
        errors.WriteError(w, err, requestID)
        return
    }

    // Create response
    response := MaxAuthResponse{
        AccessToken:  tokens.AccessToken,
        RefreshToken: tokens.RefreshToken,
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}

// GetBotMe godoc
// @Summary      Get bot information
// @Description  Get bot name and add bot link from MaxBot service
// @Tags         bot
// @Accept       json
// @Produce      json
// @Success      200  {object}  BotInfoResponse  "Bot information"
// @Failure      500  {object}  object{error=string,message=string}  "Internal server error"
// @Router       /bot/me [get]
func (h *Handler) GetBotMe(w http.ResponseWriter, r *http.Request) {
    requestID := middleware.GetRequestID(r.Context())
    
    // Get bot info from auth service (which will call maxbot service)
    botInfo, err := h.auth.GetBotInfo(r.Context())
    if err != nil {
        errors.WriteError(w, err, requestID)
        return
    }

    // Create response
    response := BotInfoResponse{
        Name:    botInfo.Name,
        AddLink: botInfo.AddLink,
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}

// ValidateTokenResponse represents the response for token validation
type ValidateTokenResponse struct {
    UserID int64 `json:"user_id"`
}

// ValidateToken godoc
// @Summary      Validate JWT token
// @Description  Validates JWT token and returns user ID for other services
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer token"
// @Success      200            {object}  ValidateTokenResponse  "Token is valid"
// @Failure      401            {string}  string  "Invalid or expired token"
// @Router       /validate-token [get]
func (h *Handler) ValidateToken(w http.ResponseWriter, r *http.Request) {
    requestID := middleware.GetRequestID(r.Context())
    
    // Extract token from Authorization header
    authHeader := r.Header.Get("Authorization")
    if authHeader == "" {
        errors.WriteError(w, errors.UnauthorizedError("missing authorization header"), requestID)
        return
    }

    // Check for Bearer token format
    parts := strings.Split(authHeader, " ")
    if len(parts) != 2 || parts[0] != "Bearer" {
        errors.WriteError(w, errors.UnauthorizedError("invalid authorization header format"), requestID)
        return
    }

    token := parts[1]

    // Validate token
    userID, _, _, _, err := h.auth.ValidateTokenWithContext(token)
    if err != nil {
        errors.WriteError(w, errors.UnauthorizedError("invalid or expired token"), requestID)
        return
    }

    // Return user ID
    response := ValidateTokenResponse{
        UserID: userID,
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}