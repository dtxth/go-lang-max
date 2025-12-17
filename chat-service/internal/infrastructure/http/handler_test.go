package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchChats_Unauthorized(t *testing.T) {
	handler := NewHandler(nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/chats?query=Математика", nil)
	w := httptest.NewRecorder()

	handler.SearchChats(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestGetAllChats_Unauthorized(t *testing.T) {
	handler := NewHandler(nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/chats/all?limit=10", nil)
	w := httptest.NewRecorder()

	handler.GetAllChats(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestGetChatByID_InvalidID(t *testing.T) {
	handler := NewHandler(nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/chats/invalid", nil)
	w := httptest.NewRecorder()

	handler.GetChatByID(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestAddAdministrator_MissingPhone(t *testing.T) {
	handler := NewHandler(nil, nil, nil)

	reqBody := AddAdministratorRequest{}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/chats/1/administrators", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.AddAdministrator(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestAddAdministrator_InvalidChatID(t *testing.T) {
	handler := NewHandler(nil, nil, nil)

	reqBody := AddAdministratorRequest{
		Phone: "+79001234567",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/chats/invalid/administrators", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.AddAdministrator(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestAddAdministrator_InvalidJSON(t *testing.T) {
	handler := NewHandler(nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/chats/1/administrators", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	handler.AddAdministrator(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestAddAdministrator_InvalidPath(t *testing.T) {
	handler := NewHandler(nil, nil, nil)

	reqBody := AddAdministratorRequest{
		Phone: "+79001234567",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/chats/1/invalid", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.AddAdministrator(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestRemoveAdministrator_InvalidID(t *testing.T) {
	handler := NewHandler(nil, nil, nil)

	req := httptest.NewRequest(http.MethodDelete, "/administrators/invalid", nil)
	w := httptest.NewRecorder()

	handler.RemoveAdministrator(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestRefreshParticipantsCount_InvalidChatID(t *testing.T) {
	handler := NewHandler(nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/chats/invalid/refresh-participants", nil)
	w := httptest.NewRecorder()

	handler.RefreshParticipantsCount(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestRefreshParticipantsCount_InvalidPath(t *testing.T) {
	handler := NewHandler(nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/chats/1/invalid", nil)
	w := httptest.NewRecorder()

	handler.RefreshParticipantsCount(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}
