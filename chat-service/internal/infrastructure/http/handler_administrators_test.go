package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"chat-service/internal/domain"
)

type mockChatServiceForAdministrators struct {
	expectedLimit  int
	expectedOffset int
	expectedQuery  string
}

func (m *mockChatServiceForAdministrators) GetAllAdministrators(query string, limit, offset int) ([]*domain.Administrator, int, error) {
	// Проверяем, что переданы правильные параметры
	if limit != m.expectedLimit {
		panic("Unexpected limit")
	}
	if offset != m.expectedOffset {
		panic("Unexpected offset")
	}
	if query != m.expectedQuery {
		panic("Unexpected query")
	}

	// Возвращаем тестовые данные
	admins := []*domain.Administrator{
		{
			ID:     1,
			ChatID: 1,
			Phone:  "+79991234567",
			MaxID:  "test_max_id",
		},
	}
	return admins, 142734, nil
}

// Заглушки для других методов интерфейса
func (m *mockChatServiceForAdministrators) SearchChats(query string, limit, offset int, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
	return nil, 0, nil
}

func (m *mockChatServiceForAdministrators) GetAllChats(limit, offset int, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
	return nil, 0, nil
}

func (m *mockChatServiceForAdministrators) GetAllChatsWithSortingAndSearch(limit, offset int, sortBy, sortOrder, search string, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
	return nil, 0, nil
}

func (m *mockChatServiceForAdministrators) GetChatByID(id int64) (*domain.Chat, error) {
	return nil, nil
}

func (m *mockChatServiceForAdministrators) GetAdministratorByID(id int64) (*domain.Administrator, error) {
	return nil, nil
}

func (m *mockChatServiceForAdministrators) RemoveAdministrator(adminID int64) error {
	return nil
}

func (m *mockChatServiceForAdministrators) CreateChat(name, url, maxChatID, source string, participantsCount int, universityID *int64, department string) (*domain.Chat, error) {
	return nil, nil
}

func (m *mockChatServiceForAdministrators) RefreshParticipantsCount(ctx context.Context, chatID int64) (*domain.ParticipantsInfo, error) {
	return nil, nil
}

func (m *mockChatServiceForAdministrators) AddAdministratorWithFlags(chatID int64, phone string, maxID string, addUser bool, addAdmin bool, skipPhoneValidation bool) (*domain.Administrator, error) {
	return nil, nil
}

func TestGetAllAdministrators_DefaultLimit(t *testing.T) {
	// Создаем мок сервиса, который ожидает дефолтный лимит 50
	mockService := &mockChatServiceForAdministrators{
		expectedLimit:  50,
		expectedOffset: 0,
		expectedQuery:  "",
	}

	handler := &Handler{
		chatService: mockService,
	}

	// Создаем запрос без параметров limit и offset
	req := httptest.NewRequest("GET", "/administrators", nil)
	w := httptest.NewRecorder()

	// Выполняем запрос
	handler.GetAllAdministrators(w, req)

	// Проверяем статус ответа
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Проверяем содержимое ответа
	var response AdministratorListResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Проверяем, что лимит установлен правильно
	if response.Limit != 50 {
		t.Errorf("Expected limit 50, got %d", response.Limit)
	}

	// Проверяем, что данные возвращаются
	if len(response.Administrators) != 1 {
		t.Errorf("Expected 1 administrator, got %d", len(response.Administrators))
	}

	if response.TotalCount != 142734 {
		t.Errorf("Expected total count 142734, got %d", response.TotalCount)
	}
}

func TestGetAllAdministrators_CustomLimit(t *testing.T) {
	// Создаем мок сервиса, который ожидает кастомный лимит 25
	mockService := &mockChatServiceForAdministrators{
		expectedLimit:  25,
		expectedOffset: 10,
		expectedQuery:  "test",
	}

	handler := &Handler{
		chatService: mockService,
	}

	// Создаем запрос с параметрами
	req := httptest.NewRequest("GET", "/administrators?limit=25&offset=10&query=test", nil)
	w := httptest.NewRecorder()

	// Выполняем запрос
	handler.GetAllAdministrators(w, req)

	// Проверяем статус ответа
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Проверяем содержимое ответа
	var response AdministratorListResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Проверяем, что лимит установлен правильно
	if response.Limit != 25 {
		t.Errorf("Expected limit 25, got %d", response.Limit)
	}
}

func TestGetAllAdministrators_MaxLimit(t *testing.T) {
	// Создаем мок сервиса, который ожидает максимальный лимит 100
	mockService := &mockChatServiceForAdministrators{
		expectedLimit:  100,
		expectedOffset: 0,
		expectedQuery:  "",
	}

	handler := &Handler{
		chatService: mockService,
	}

	// Создаем запрос с лимитом больше максимального
	req := httptest.NewRequest("GET", "/administrators?limit=200", nil)
	w := httptest.NewRecorder()

	// Выполняем запрос
	handler.GetAllAdministrators(w, req)

	// Проверяем статус ответа
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Проверяем содержимое ответа
	var response AdministratorListResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Проверяем, что лимит ограничен максимальным значением
	if response.Limit != 100 {
		t.Errorf("Expected limit 100, got %d", response.Limit)
	}
}