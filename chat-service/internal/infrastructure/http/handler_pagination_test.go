package http

import (
	"chat-service/internal/domain"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// SetTokenInfo добавляет информацию о токене в контекст для тестов
func SetTokenInfo(ctx context.Context, tokenInfo *domain.TokenInfo) context.Context {
	return context.WithValue(ctx, tokenInfoKey, tokenInfo)
}

// mockChatServiceForPagination is a mock for testing pagination
type mockChatServiceForPagination struct {
	chats []*domain.Chat
	total int
}

func (m *mockChatServiceForPagination) GetAllChatsWithSortingAndSearch(limit, offset int, sortBy, sortOrder, search string, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
	// Simulate filtering by search
	filtered := m.chats
	if search != "" {
		filtered = []*domain.Chat{}
		for _, chat := range m.chats {
			if containsIgnoreCase(chat.Name, search) ||
				containsIgnoreCase(chat.Department, search) ||
				(chat.University != nil && containsIgnoreCase(chat.University.Name, search)) {
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

func containsIgnoreCase(str, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(str) == 0 {
		return false
	}
	
	// Simple case-insensitive contains check
	strLower := strings.ToLower(str)
	substrLower := strings.ToLower(substr)
	
	for i := 0; i <= len(strLower)-len(substrLower); i++ {
		if strLower[i:i+len(substrLower)] == substrLower {
			return true
		}
	}
	return false
}

func TestGetAllChats_WithPagination(t *testing.T) {
	// Create test data
	university := &domain.University{
		ID:   1,
		Name: "Test University",
	}
	
	chats := []*domain.Chat{
		{
			ID:           1,
			Name:         "Test Chat 1",
			URL:          "https://t.me/test1",
			MaxChatID:    "123456",
			UniversityID: &university.ID,
			University:   university,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           2,
			Name:         "Test Chat 2",
			URL:          "https://t.me/test2",
			MaxChatID:    "123457",
			UniversityID: &university.ID,
			University:   university,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}
	
	mockService := &mockChatServiceForPagination{
		chats: chats,
		total: 2,
	}
	
	handler := &Handler{
		chatService: &mockChatServiceWrapper{mockPagination: mockService},
	}
	
	// Test with pagination parameters
	req := httptest.NewRequest("GET", "/chats/all?limit=1&offset=0", nil)
	
	// Add mock token info to context
	tokenInfo := &domain.TokenInfo{
		Role:   "superadmin",
		UserID: 1,
		Valid:  true,
	}
	req = req.WithContext(SetTokenInfo(req.Context(), tokenInfo))
	
	w := httptest.NewRecorder()
	
	handler.GetAllChats(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response PaginatedChatsResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if len(response.Data) != 1 {
		t.Errorf("Expected 1 chat, got %d", len(response.Data))
	}
	
	if response.Total != 2 {
		t.Errorf("Expected total 2, got %d", response.Total)
	}
	
	if response.Limit != 1 {
		t.Errorf("Expected limit 1, got %d", response.Limit)
	}
	
	if response.Offset != 0 {
		t.Errorf("Expected offset 0, got %d", response.Offset)
	}
	
	if response.TotalPages != 2 {
		t.Errorf("Expected total pages 2, got %d", response.TotalPages)
	}
}

func TestGetAllChats_WithSearch(t *testing.T) {
	// Create test data
	university1 := &domain.University{
		ID:   1,
		Name: "МГУ",
	}
	
	university2 := &domain.University{
		ID:   2,
		Name: "СПбГУ",
	}
	
	chats := []*domain.Chat{
		{
			ID:           1,
			Name:         "Чат МГУ",
			URL:          "https://t.me/mgu",
			MaxChatID:    "123456",
			UniversityID: &university1.ID,
			University:   university1,
		},
		{
			ID:           2,
			Name:         "Чат СПбГУ",
			URL:          "https://t.me/spbgu",
			MaxChatID:    "123457",
			UniversityID: &university2.ID,
			University:   university2,
		},
	}
	
	mockService := &mockChatServiceForPagination{
		chats: chats,
		total: 2,
	}
	
	handler := &Handler{
		chatService: &mockChatServiceWrapper{mockPagination: mockService},
	}
	
	// Test with search parameter
	params := url.Values{}
	params.Add("search", "МГУ")
	params.Add("limit", "10")
	params.Add("offset", "0")
	
	req := httptest.NewRequest("GET", "/chats/all?"+params.Encode(), nil)
	
	// Add mock token info to context
	tokenInfo := &domain.TokenInfo{
		Role:   "superadmin",
		UserID: 1,
		Valid:  true,
	}
	req = req.WithContext(SetTokenInfo(req.Context(), tokenInfo))
	
	w := httptest.NewRecorder()
	
	handler.GetAllChats(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response PaginatedChatsResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if len(response.Data) != 1 {
		t.Errorf("Expected 1 chat after search, got %d", len(response.Data))
	}
	
	if response.Data[0].Name != "Чат МГУ" {
		t.Errorf("Expected chat name 'Чат МГУ', got '%s'", response.Data[0].Name)
	}
}

func TestGetAllChats_WithSorting(t *testing.T) {
	// Create test data
	university := &domain.University{
		ID:   1,
		Name: "Test University",
	}
	
	chats := []*domain.Chat{
		{
			ID:           1,
			Name:         "B Chat",
			URL:          "https://t.me/b",
			MaxChatID:    "123456",
			UniversityID: &university.ID,
			University:   university,
		},
		{
			ID:           2,
			Name:         "A Chat",
			URL:          "https://t.me/a",
			MaxChatID:    "123457",
			UniversityID: &university.ID,
			University:   university,
		},
	}
	
	mockService := &mockChatServiceForPagination{
		chats: chats,
		total: 2,
	}
	
	handler := &Handler{
		chatService: &mockChatServiceWrapper{mockPagination: mockService},
	}
	
	// Test with sorting parameters
	params := url.Values{}
	params.Add("sort_by", "name")
	params.Add("sort_order", "desc")
	params.Add("limit", "10")
	params.Add("offset", "0")
	
	req := httptest.NewRequest("GET", "/chats/all?"+params.Encode(), nil)
	
	// Add mock token info to context
	tokenInfo := &domain.TokenInfo{
		Role:   "superadmin",
		UserID: 1,
		Valid:  true,
	}
	req = req.WithContext(SetTokenInfo(req.Context(), tokenInfo))
	
	w := httptest.NewRecorder()
	
	handler.GetAllChats(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response PaginatedChatsResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if len(response.Data) != 2 {
		t.Errorf("Expected 2 chats, got %d", len(response.Data))
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

func (m *mockChatServiceWrapper) CreateOrGetUniversity(inn, kpp, name string) (*domain.University, error) {
	return nil, nil
}

func (m *mockChatServiceWrapper) CreateChat(name, url, maxChatID, source string, participantsCount int, universityID *int64, department string) (*domain.Chat, error) {
	return nil, nil
}