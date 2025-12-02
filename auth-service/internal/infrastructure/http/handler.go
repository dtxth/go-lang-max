package http

import (
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
    var req struct {
        Email    string `json:"email"`
        Password string `json:"password"`
        Role     string `json:"role"`
    }
    json.NewDecoder(r.Body).Decode(&req)

    user, err := h.auth.Register(req.Email, req.Password, req.Role)
    if err != nil {
        http.Error(w, err.Error(), 400)
        return
    }

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
    var req struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }
    _ = json.NewDecoder(r.Body).Decode(&req)

    tokens, err := h.auth.Login(req.Email, req.Password)
    if err != nil {
        http.Error(w, err.Error(), http.StatusUnauthorized)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "access_token":  tokens.AccessToken,
        "refresh_token": tokens.RefreshToken,
    })
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
    var req struct {
        RefreshToken string `json:"refresh_token"`
    }
    _ = json.NewDecoder(r.Body).Decode(&req)
    if req.RefreshToken == "" {
        http.Error(w, "refresh_token required", http.StatusBadRequest)
        return
    }

    tokens, err := h.auth.Refresh(req.RefreshToken)
    if err != nil {
        http.Error(w, err.Error(), http.StatusUnauthorized)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "access_token":  tokens.AccessToken,
        "refresh_token": tokens.RefreshToken,
    })
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
    var req struct {
        RefreshToken string `json:"refresh_token"`
    }
    _ = json.NewDecoder(r.Body).Decode(&req)
    if req.RefreshToken == "" {
        http.Error(w, "refresh_token required", http.StatusBadRequest)
        return
    }

    if err := h.auth.Logout(req.RefreshToken); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "logged_out"})
}