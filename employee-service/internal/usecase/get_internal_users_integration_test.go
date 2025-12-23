package usecase

import (
	"context"
	"employee-service/internal/domain"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateEmployeeWithRole_UsesGetInternalUsers(t *testing.T) {
	// Setup
	employeeRepo := newMockEmployeeRepo()
	universityRepo := newMockUniversityRepo()
	maxService := newMockMaxService()
	authService := newMockAuthService()
	passwordGenerator := newMockPasswordGenerator()
	notificationService := newMockNotificationService()

	// Add test data to mock service
	maxService.users["+79991234567"] = "max_123"

	uc := NewCreateEmployeeWithRoleUseCase(
		employeeRepo,
		universityRepo,
		maxService,
		authService,
		passwordGenerator,
		notificationService,
		nil, // profileCache
	)

	ctx := context.Background()

	// Test: Create employee - should use GetInternalUsers for names
	employee, err := uc.Execute(
		ctx,
		"+79991234567", // Phone with "1234" should get "Петр Петров" from mock
		"", "", "",     // Empty names - should be filled from GetInternalUsers
		"1234567890", "123456789", "Test University",
		"operator",
		"superadmin",
	)

	// Verify
	require.NoError(t, err)
	require.NotNil(t, employee)

	// Should get names from GetInternalUsers (Петр Петров for phones with "1234")
	assert.Equal(t, "Петр", employee.FirstName)
	assert.Equal(t, "Петров", employee.LastName)
	assert.Equal(t, "+79991234567", employee.Phone)
	assert.Equal(t, "max_123", employee.MaxID) // MaxID from old method for compatibility
	assert.Equal(t, "operator", employee.Role)
}

func TestCreateEmployeeWithRole_FallbackToOldMethod(t *testing.T) {
	// Setup with a mock that fails GetInternalUsers
	employeeRepo := newMockEmployeeRepo()
	universityRepo := newMockUniversityRepo()
	authService := newMockAuthService()
	passwordGenerator := newMockPasswordGenerator()
	notificationService := newMockNotificationService()

	// Create a mock that fails GetInternalUsers but works with GetUserProfileByPhone
	maxService := &mockMaxServiceWithFailingInternalUsers{
		users: map[string]string{"+79995678901": "max_456"},
	}

	uc := NewCreateEmployeeWithRoleUseCase(
		employeeRepo,
		universityRepo,
		maxService,
		authService,
		passwordGenerator,
		notificationService,
		nil, // profileCache
	)

	ctx := context.Background()

	// Test: Create employee - should fallback to old method
	employee, err := uc.Execute(
		ctx,
		"+79995678901",
		"", "", "", // Empty names
		"1234567890", "123456789", "Test University",
		"curator",
		"superadmin",
	)

	// Verify
	require.NoError(t, err)
	require.NotNil(t, employee)

	// Should get names from fallback method (GetUserProfileByPhone)
	assert.Equal(t, "Анна", employee.FirstName)   // From GetUserProfileByPhone mock
	assert.Equal(t, "Сидорова", employee.LastName) // From GetUserProfileByPhone mock
	assert.Equal(t, "+79995678901", employee.Phone)
	assert.Equal(t, "max_456", employee.MaxID)
	assert.Equal(t, "curator", employee.Role)
}

// mockMaxServiceWithFailingInternalUsers simulates GetInternalUsers failure
type mockMaxServiceWithFailingInternalUsers struct {
	users map[string]string
}

func (m *mockMaxServiceWithFailingInternalUsers) GetMaxIDByPhone(phone string) (string, error) {
	if maxID, ok := m.users[phone]; ok {
		return maxID, nil
	}
	return "", domain.ErrMaxIDNotFound
}

func (m *mockMaxServiceWithFailingInternalUsers) BatchGetMaxIDByPhone(phones []string) (map[string]string, error) {
	result := make(map[string]string)
	for _, phone := range phones {
		if maxID, ok := m.users[phone]; ok {
			result[phone] = maxID
		}
	}
	return result, nil
}

func (m *mockMaxServiceWithFailingInternalUsers) GetUserProfileByPhone(phone string) (*domain.UserProfile, error) {
	if maxID, ok := m.users[phone]; ok {
		profile := &domain.UserProfile{
			MaxID:     maxID,
			Phone:     phone,
			FirstName: "Анна",
			LastName:  "Сидорова",
		}
		return profile, nil
	}
	return nil, domain.ErrMaxIDNotFound
}

func (m *mockMaxServiceWithFailingInternalUsers) GetInternalUsers(phones []string) ([]*domain.InternalUser, []string, error) {
	// Always fail to test fallback
	return nil, phones, domain.ErrMaxIDNotFound
}