package http

import (
	"chat-service/internal/domain"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Mock service for testing pagination
type mockChatServiceForPagination struct {
	chats []*domain.Chat
}

func (m *mockChatServiceForPagination) GetAllChatsWithSortingAndSearch(limit, offset int, sortBy, sortOrder, search string, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
	// Simple mock implementation for testing
	var filtered []*domain.Chat
	
	if search == "" {
		filtered = m.chats
	} else {
		filtered = []*domain.Chat{}
		for _, chat := range m.chats {
			if containsIgnoreCase(chat.Name, search) ||
				containsIgnoreCase(chat.Department, search) {
				filtered = append(filtered, chat)
			}
		}
	}
	
	// Simulate pagination
	start := offset
	end := offset + limit
	if start > len(filtered) {
		return []*domain.Chat{}, len(filtered), nil
	}
	if end > len(filtered) {
		end = len(filtered)
	}
	
	return filtered[start:end], len(filtered), nil
}

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func TestGetAllChats_WithPagination(t *testing.T) {
	// Create test data
	universityID := int64(1)
	
	chats := []*domain.Chat{
		{
			ID:           1,
			Name:         "Test Chat 1",
			URL:          "https://t.me/test1",
			MaxChatID:    "123456",
			UniversityID: &universityID,
			Department:   "IT",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           2,
			Name:         "Test Chat 2",
			URL:          "https://t.me/test2",
			MaxChatID:    "123457",
			UniversityID: &universityID,
			Department:   "Math",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	mockService := &mockChatServiceForPagination{
		chats: chats,
	}
	
	handler := NewHandler(&mockChatServiceWrapper{mockPagination: mockService}, nil, nil)
	
	// Test pagination
	req := httptest.NewRequest("GET", "/chats?limit=1&offset=0", nil)
	
	// Add mock token info to context
	tokenInfo := &domain.TokenInfo{
		Valid:  true,
		UserID: 1,
		Role:   "superadmin",
	}
	ctx := context.WithValue(req.Context(), contextKey("tokenInfo"), tokenInfo)
	req = req.WithContext(ctx)
	
	w := httptest.NewRecorder()
	
	handler.GetAllChats(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	chatsData, ok := response["data"].([]interface{})
	if !ok {
		t.Fatal("Expected data array in response")
	}
	
	if len(chatsData) != 1 {
		t.Errorf("Expected 1 chat, got %d", len(chatsData))
	}
	
	total, ok := response["total"].(float64)
	if !ok {
		t.Fatal("Expected total in response")
	}
	
	if int(total) != 2 {
		t.Errorf("Expected total 2, got %d", int(total))
	}
}

func TestGetAllChats_WithSearch(t *testing.T) {
	// Create test data
	universityID1 := int64(1)
	universityID2 := int64(2)
	
	chats := []*domain.Chat{
		{
			ID:           1,
			Name:         "МГУ Чат",
			URL:          "https://t.me/mgu",
			MaxChatID:    "123456",
			UniversityID: &universityID1,
			Department:   "Физика",
		},
		{
			ID:           2,
			Name:         "СПбГУ Чат",
			URL:          "https://t.me/spbgu",
			MaxChatID:    "123457",
			UniversityID: &universityID2,
			Department:   "Математика",
		},
	}

	mockService := &mockChatServiceForPagination{
		chats: chats,
	}
	
	handler := NewHandler(&mockChatServiceWrapper{mockPagination: mockService}, nil, nil)
	
	// Test search
	req := httptest.NewRequest("GET", "/chats?search=МГУ&limit=10", nil)
	
	// Add mock token info to context
	tokenInfo := &domain.TokenInfo{
		Valid:  true,
		UserID: 1,
		Role:   "superadmin",
	}
	ctx := context.WithValue(req.Context(), contextKey("tokenInfo"), tokenInfo)
	req = req.WithContext(ctx)
	
	w := httptest.NewRecorder()
	
	handler.GetAllChats(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	chatsData, ok := response["data"].([]interface{})
	if !ok {
		t.Fatal("Expected data array in response")
	}
	
	if len(chatsData) != 1 {
		t.Errorf("Expected 1 chat, got %d", len(chatsData))
	}
}

func TestGetAllChats_WithSorting(t *testing.T) {
	// Create test data
	universityID := int64(1)
	
	chats := []*domain.Chat{
		{
			ID:           1,
			Name:         "B Chat",
			URL:          "https://t.me/b",
			MaxChatID:    "123456",
			UniversityID: &universityID,
			Department:   "IT",
		},
		{
			ID:           2,
			Name:         "A Chat",
			URL:          "https://t.me/a",
			MaxChatID:    "123457",
			UniversityID: &universityID,
			Department:   "Math",
		},
	}

	mockService := &mockChatServiceForPagination{
		chats: chats,
	}
	
	handler := NewHandler(&mockChatServiceWrapper{mockPagination: mockService}, nil, nil)
	
	// Test sorting
	req := httptest.NewRequest("GET", "/chats?sortBy=name&sortOrder=asc&limit=10", nil)
	
	// Add mock token info to context
	tokenInfo := &domain.TokenInfo{
		Valid:  true,
		UserID: 1,
		Role:   "superadmin",
	}
	ctx := context.WithValue(req.Context(), contextKey("tokenInfo"), tokenInfo)
	req = req.WithContext(ctx)
	
	w := httptest.NewRecorder()
	
	handler.GetAllChats(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// mockChatServiceWrapper wraps the pagination mock to satisfy the interface
type mockChatServiceWrapper struct {
	mockPagination *mockChatServiceForPagination
}

func (m *mockChatServiceWrapper) GetAllChatsWithSortingAndSearch(limit, offset int, sortBy, sortOrder, search string, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
	return m.mockPagination.GetAllChatsWithSortingAndSearch(limit, offset, sortBy, sortOrder, search, filter)
}

// Implement other required methods as no-ops for testing
func (m *mockChatServiceWrapper) SearchChats(query string, limit, offset int, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
	return nil, 0, nil
}

func (m *mockChatServiceWrapper) GetAllChats(limit, offset int, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
	return nil, 0, nil
}

func (m *mockChatServiceWrapper) GetChatByID(id int64) (*domain.Chat, error) {
	return nil, nil
}

func (m *mockChatServiceWrapper) AddAdministratorWithFlags(chatID int64, phone string, maxID string, addUser bool, addAdmin bool, skipPhoneValidation bool) (*domain.Administrator, error) {
	return nil, nil
}

func (m *mockChatServiceWrapper) GetAdministratorByID(id int64) (*domain.Administrator, error) {
	return nil, nil
}

func (m *mockChatServiceWrapper) GetAllAdministrators(query string, limit, offset int) ([]*domain.Administrator, int, error) {
	return nil, 0, nil
}

func (m *mockChatServiceWrapper) RemoveAdministrator(adminID int64) error {
	return nil
}

func (m *mockChatServiceWrapper) CreateChat(name, url, maxChatID, source string, participantsCount int, universityID *int64, department string) (*domain.Chat, error) {
	return nil, nil
}

func (m *mockChatServiceWrapper) RefreshParticipantsCount(ctx context.Context, chatID int64) (*domain.ParticipantsInfo, error) {
	return &domain.ParticipantsInfo{
		Count:     100,
		UpdatedAt: time.Now(),
		Source:    "api",
	}, nil
}