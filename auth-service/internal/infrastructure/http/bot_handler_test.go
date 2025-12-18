package http

import (
	"auth-service/internal/domain"
	"auth-service/internal/infrastructure/maxbot"
	"auth-service/internal/usecase"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_GetBotMe(t *testing.T) {
	// Create mock dependencies
	mockMaxBotClient := maxbot.NewMockClient()
	
	// Create auth service with minimal dependencies (we only need maxbot client for this test)
	authService := &usecase.AuthService{}
	authService.SetMaxBotClient(mockMaxBotClient)
	
	// Create handler
	handler := NewHandler(authService)
	
	// Create test request
	req := httptest.NewRequest("GET", "/bot/me", nil)
	w := httptest.NewRecorder()
	
	// Call handler
	handler.GetBotMe(w, req)
	
	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	// Check content type
	expectedContentType := "application/json"
	if contentType := w.Header().Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("Expected content type %s, got %s", expectedContentType, contentType)
	}
	
	// Parse response
	var response BotInfoResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	// Check response values
	expectedName := "Digital University Bot"
	expectedAddLink := "https://max.ru/bot/digital_university_bot"
	
	if response.Name != expectedName {
		t.Errorf("Expected name %s, got %s", expectedName, response.Name)
	}
	
	if response.AddLink != expectedAddLink {
		t.Errorf("Expected add link %s, got %s", expectedAddLink, response.AddLink)
	}
}

func TestHandler_GetBotMe_Error(t *testing.T) {
	// Create mock dependencies
	mockMaxBotClient := maxbot.NewMockClient()
	mockMaxBotClient.SetError(domain.ErrMaxBotUnavailable) // Set an error
	
	// Create auth service
	authService := &usecase.AuthService{}
	authService.SetMaxBotClient(mockMaxBotClient)
	
	// Create handler
	handler := NewHandler(authService)
	
	// Create test request
	req := httptest.NewRequest("GET", "/bot/me", nil)
	w := httptest.NewRecorder()
	
	// Call handler
	handler.GetBotMe(w, req)
	
	// Check response - should be an error
	if w.Code == http.StatusOK {
		t.Errorf("Expected error status, got %d", w.Code)
	}
}