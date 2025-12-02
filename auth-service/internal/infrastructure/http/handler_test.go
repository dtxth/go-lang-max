package http

import (
	"auth-service/internal/domain"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Mock repositories
type mockUserRepository struct {
	createFunc       func(user *domain.User) error
	getByEmailFunc   func(email string) (*domain.User, error)
	getByIDFunc      func(id int64) (*domain.User, error)
}

func (m *mockUserRepository) Create(user *domain.User) error {
	if m.createFunc != nil {
		return m.createFunc(user)
	}
	return errors.New("not implemented")
}

func (m *mockUserRepository) GetByEmail(email string) (*domain.User, error) {
	if m.getByEmailFunc != nil {
		return m.getByEmailFunc(email)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUserRepository) GetByID(id int64) (*domain.User, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(id)
	}
	return nil, errors.New("not implemented")
}

type mockRefreshTokenRepository struct {
	saveFunc   func(token *domain.RefreshToken) error
	getFunc    func(token string) (*domain.RefreshToken, error)
	deleteFunc func(token string) error
}

func (m *mockRefreshTokenRepository) Save(token *domain.RefreshToken) error {
	if m.saveFunc != nil {
		return m.saveFunc(token)
	}
	return errors.New("not implemented")
}

func (m *mockRefreshTokenRepository) Get(token string) (*domain.RefreshToken, error) {
	if m.getFunc != nil {
		return m.getFunc(token)
	}
	return nil, errors.New("not implemented")
}

func (m *mockRefreshTokenRepository) Delete(token string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(token)
	}
	return errors.New("not implemented")
}

func TestRegister_InvalidJSON(t *testing.T) {
	handler := NewHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	handler.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestRegister_MissingEmail(t *testing.T) {
	handler := NewHandler(nil)

	reqBody := map[string]string{
		"password": "password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestRegister_MissingPassword(t *testing.T) {
	handler := NewHandler(nil)

	reqBody := map[string]string{
		"email": "test@example.com",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestLogin_InvalidJSON(t *testing.T) {
	handler := NewHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	handler.Login(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestLogin_MissingEmail(t *testing.T) {
	handler := NewHandler(nil)

	reqBody := map[string]string{
		"password": "password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Login(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestLogin_MissingPassword(t *testing.T) {
	handler := NewHandler(nil)

	reqBody := map[string]string{
		"email": "test@example.com",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Login(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestRefresh_InvalidJSON(t *testing.T) {
	handler := NewHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	handler.Refresh(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestRefresh_MissingToken(t *testing.T) {
	handler := NewHandler(nil)

	reqBody := map[string]string{}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Refresh(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestLogout_InvalidJSON(t *testing.T) {
	handler := NewHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/logout", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	handler.Logout(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestLogout_MissingToken(t *testing.T) {
	handler := NewHandler(nil)

	reqBody := map[string]string{}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/logout", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Logout(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestHealth_Success(t *testing.T) {
	handler := NewHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler.Health(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response map[string]string
	json.NewDecoder(w.Body).Decode(&response)

	if response["status"] != "healthy" {
		t.Errorf("expected status healthy, got %s", response["status"])
	}
}
