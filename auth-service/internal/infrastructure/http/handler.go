package http

import (
	"auth-service/internal/infrastructure/errors"
	"auth-service/internal/infrastructure/middleware"
	"auth-service/internal/usecase"
	"encoding/json"
	"net/http"
)

type Handler struct {
    auth *usecase.AuthService
}

func NewHandler(auth *usecase.AuthService) *Handler {
    return &Handler{auth: auth}
}

// Register godoc
// @Summary      Register new user
// @Description  Creates user and stores hashed password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body      object{email=string,password=string,role=string}  true  "User credentials (role is optional, defaults to operator)"
// @Success      200    {object}  domain.User
// @Failure      400    {string}  string
// @Router       /register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
    requestID := middleware.GetRequestID(r.Context())
    
    var req struct {
        Email    string `json:"email"`
        Password string `json:"password"`
        Role     string `json:"role"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        errors.WriteError(w, errors.ValidationError("invalid request body").WithError(err), requestID)
        return
    }

    // Validate required fields
    if req.Email == "" {
        errors.WriteError(w, errors.MissingFieldError("email"), requestID)
        return
    }
    if req.Password == "" {
        errors.WriteError(w, errors.MissingFieldError("password"), requestID)
        return
    }

    user, err := h.auth.Register(req.Email, req.Password, req.Role)
    if err != nil {
        errors.WriteError(w, err, requestID)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}

// Login godoc
// @Summary      Login user
// @Description  Returns access and refresh tokens
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body      object{email=string,password=string}  true  "User credentials"
// @Success      200    {object}  domain.TokenPair
// @Failure      401    {string}  string
// @Router       /login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
    requestID := middleware.GetRequestID(r.Context())
    
    var req struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        errors.WriteError(w, errors.ValidationError("invalid request body").WithError(err), requestID)
        return
    }

    // Validate required fields
    if req.Email == "" {
        errors.WriteError(w, errors.MissingFieldError("email"), requestID)
        return
    }
    if req.Password == "" {
        errors.WriteError(w, errors.MissingFieldError("password"), requestID)
        return
    }

    tokens, err := h.auth.Login(req.Email, req.Password)
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