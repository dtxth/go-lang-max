package usecase

import (
	"context"
	"testing"

	"maxbot-service/internal/domain"
)

// MockMaxAPIClient is a mock implementation of domain.MaxAPIClient for testing
type MockMaxAPIClient struct {
	existingPhones map[string]string // normalized phone -> max_id
}

func NewMockMaxAPIClient() *MockMaxAPIClient {
	return &MockMaxAPIClient{
		existingPhones: make(map[string]string),
	}
}

func (m *MockMaxAPIClient) AddExistingPhone(phone, maxID string) {
	m.existingPhones[phone] = maxID
}

func (m *MockMaxAPIClient) GetMaxIDByPhone(ctx context.Context, phone string) (string, error) {
	if maxID, exists := m.existingPhones[phone]; exists {
		return maxID, nil
	}
	return "", domain.ErrMaxIDNotFound
}

func (m *MockMaxAPIClient) ValidatePhone(phone string) (bool, string, error) {
	// Simple validation for testing
	if len(phone) >= 10 {
		return true, phone, nil
	}
	return false, "", domain.ErrInvalidPhone
}

func (m *MockMaxAPIClient) SendMessage(ctx context.Context, chatID, userID int64, text string) (string, error) {
	return "", nil
}

func (m *MockMaxAPIClient) SendNotification(ctx context.Context, phone, text string) error {
	return nil
}

func (m *MockMaxAPIClient) GetChatInfo(ctx context.Context, chatID int64) (*domain.ChatInfo, error) {
	return nil, nil
}

func (m *MockMaxAPIClient) GetChatMembers(ctx context.Context, chatID int64, limit int, marker int64) (*domain.ChatMembersList, error) {
	return nil, nil
}

func (m *MockMaxAPIClient) GetChatAdmins(ctx context.Context, chatID int64) ([]*domain.ChatMember, error) {
	return nil, nil
}

func (m *MockMaxAPIClient) CheckPhoneNumbers(ctx context.Context, phones []string) ([]string, error) {
	return nil, nil
}

func (m *MockMaxAPIClient) BatchGetUsersByPhone(ctx context.Context, phones []string) ([]*domain.UserPhoneMapping, error) {
	return nil, nil
}

func (m *MockMaxAPIClient) GetMe(ctx context.Context) (*domain.BotInfo, error) {
	return &domain.BotInfo{
		Name:    "Test Mock Bot",
		AddLink: "https://max.ru/test-bot",
	}, nil
}

func (m *MockMaxAPIClient) GetUserProfileByPhone(ctx context.Context, phone string) (*domain.UserProfile, error) {
	if maxID, exists := m.existingPhones[phone]; exists {
		return &domain.UserProfile{
			MaxID:     maxID,
			Phone:     phone,
			FirstName: "Test",
			LastName:  "User",
		}, nil
	}
	return nil, domain.ErrMaxIDNotFound
}

func TestBatchGetUsersByPhoneUseCase_Execute(t *testing.T) {
	mockClient := NewMockMaxAPIClient()
	mockClient.AddExistingPhone("+79991234567", "+79991234567")
	mockClient.AddExistingPhone("+79991234568", "+79991234568")

	uc := NewBatchGetUsersByPhoneUseCase(mockClient)
	ctx := context.Background()

	t.Run("Empty batch", func(t *testing.T) {
		result, err := uc.Execute(ctx, []string{})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("Expected empty result, got %d mappings", len(result))
		}
	})

	t.Run("Batch exceeds maximum size", func(t *testing.T) {
		phones := make([]string, 101)
		for i := 0; i < 101; i++ {
			phones[i] = "+79991234567"
		}
		_, err := uc.Execute(ctx, phones)
		if err == nil {
			t.Error("Expected error for batch size > 100")
		}
	})

	t.Run("Mixed found and not found phones", func(t *testing.T) {
		phones := []string{
			"89991234567",  // exists
			"89991234568",  // exists
			"89991234569",  // does not exist
		}

		result, err := uc.Execute(ctx, phones)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if len(result) != 3 {
			t.Errorf("Expected 3 mappings, got %d", len(result))
		}

		// Check first phone (found)
		if !result[0].Found {
			t.Error("Expected first phone to be found")
		}
		if result[0].MaxID == "" {
			t.Error("Expected MAX_id for first phone")
		}

		// Check second phone (found)
		if !result[1].Found {
			t.Error("Expected second phone to be found")
		}
		if result[1].MaxID == "" {
			t.Error("Expected MAX_id for second phone")
		}

		// Check third phone (not found)
		if result[2].Found {
			t.Error("Expected third phone to not be found")
		}
		if result[2].MaxID != "" {
			t.Error("Expected empty MAX_id for third phone")
		}
	})

	t.Run("All invalid phones", func(t *testing.T) {
		phones := []string{
			"123",
			"abc",
			"",
		}

		result, err := uc.Execute(ctx, phones)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if len(result) != 0 {
			t.Errorf("Expected empty result for all invalid phones, got %d mappings", len(result))
		}
	})

	t.Run("Batch at maximum size", func(t *testing.T) {
		phones := make([]string, 100)
		for i := 0; i < 100; i++ {
			phones[i] = "+79991234567"
		}

		result, err := uc.Execute(ctx, phones)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if len(result) != 100 {
			t.Errorf("Expected 100 mappings, got %d", len(result))
		}
	})
}

func TestBatchGetUsersByPhoneUseCase_PhoneNormalization(t *testing.T) {
	mockClient := NewMockMaxAPIClient()
	mockClient.AddExistingPhone("+79991234567", "+79991234567")

	uc := NewBatchGetUsersByPhoneUseCase(mockClient)
	ctx := context.Background()

	t.Run("Phones are normalized before lookup", func(t *testing.T) {
		phones := []string{
			"89991234567",           // should normalize to +79991234567
			"+7 (999) 123-45-67",    // should normalize to +79991234567
			"9991234567",            // should normalize to +79991234567
		}

		result, err := uc.Execute(ctx, phones)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// All phones should normalize to the same number and be found
		for i, mapping := range result {
			if !mapping.Found {
				t.Errorf("Expected phone %d to be found after normalization", i)
			}
			if mapping.MaxID != "+79991234567" {
				t.Errorf("Expected MAX_id +79991234567, got %s for phone %d", mapping.MaxID, i)
			}
			// Original phone should be preserved
			if mapping.Phone != phones[i] {
				t.Errorf("Expected original phone %s, got %s", phones[i], mapping.Phone)
			}
		}
	})
}
