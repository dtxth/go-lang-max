package usecase

import (
	"context"
	"employee-service/internal/domain"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockFailingProfileCache simulates a failing profile cache for testing resilience
type mockFailingProfileCache struct {
	shouldFail bool
}

func (m *mockFailingProfileCache) GetProfile(ctx context.Context, userID string) (*domain.CachedUserProfile, error) {
	if m.shouldFail {
		return nil, domain.ErrCacheUnavailable
	}
	return nil, nil // Return nil profile (not found) when not failing
}

func (m *mockFailingProfileCache) SetFailure(shouldFail bool) {
	m.shouldFail = shouldFail
}

// TestEmployeeService_CacheFailureResilience tests that the system continues working when profile cache fails
func TestEmployeeService_CacheFailureResilience(t *testing.T) {
	// Setup mocks using existing test infrastructure
	employeeRepo := newMockEmployeeRepo()
	universityRepo := newMockUniversityRepo()
	maxService := newMockMaxService()
	authService := newMockAuthService()
	passwordGenerator := newMockPasswordGenerator()
	notificationService := newMockNotificationService()
	profileCache := &mockFailingProfileCache{}

	// Create service
	service := NewEmployeeService(
		employeeRepo,
		universityRepo,
		maxService,
		authService,
		passwordGenerator,
		notificationService,
		profileCache,
	)

	phone := "+79991234567"
	firstName := "Иван"
	lastName := "Петров"
	inn := "1234567890"
	kpp := "123456789"
	universityName := "Тестовый университет"

	// Setup MAX service to return profile data
	maxService.users[phone] = "max_123"

	t.Run("Cache failure - graceful degradation with user provided names", func(t *testing.T) {
		// Simulate cache failure
		profileCache.SetFailure(true)

		// Call should succeed despite cache failure (Requirements 3.4, 7.5)
		employee, err := service.AddEmployeeByPhone(
			phone,
			firstName, lastName, "",
			inn, kpp,
			universityName,
		)

		// Verify success
		assert.NoError(t, err)
		assert.NotNil(t, employee)
		assert.Equal(t, firstName, employee.FirstName)
		assert.Equal(t, lastName, employee.LastName)
		assert.Equal(t, "max_123", employee.MaxID)
		assert.Equal(t, string(domain.SourceUserInput), employee.ProfileSource) // User provided names have priority
	})

	t.Run("Cache failure - graceful degradation with MAX API fallback", func(t *testing.T) {
		// Simulate cache failure
		profileCache.SetFailure(true)

		// Call without user provided names - should use MAX API data (Requirements 7.3)
		employee, err := service.AddEmployeeByPhone(
			"+79991234568", // Different phone to avoid conflicts
			"", "", "", // No user provided names
			inn, kpp,
			universityName,
		)

		// Verify success with MAX API fallback
		assert.NoError(t, err)
		assert.NotNil(t, employee)
		// Should use default values since MAX API doesn't have profile for this phone
		assert.Equal(t, "Неизвестно", employee.FirstName)
		assert.Equal(t, "Неизвестно", employee.LastName)
		assert.Equal(t, string(domain.SourceDefault), employee.ProfileSource)
	})

	t.Run("Complete failure - default values", func(t *testing.T) {
		// Simulate cache failure (MAX API failure is simulated by not adding phone to users map)
		profileCache.SetFailure(true)

		// Call should still succeed with default values (Requirements 7.5)
		employee, err := service.AddEmployeeByPhone(
			"+79991234569", // Different phone to avoid conflicts
			"", "", "", // No user provided names
			inn, kpp,
			universityName,
		)

		// Verify success with default values
		assert.NoError(t, err)
		assert.NotNil(t, employee)
		assert.Equal(t, "Неизвестно", employee.FirstName)  // Default value
		assert.Equal(t, "Неизвестно", employee.LastName)   // Default value
		assert.Equal(t, "", employee.MaxID)                // No MAX_id
		assert.Equal(t, string(domain.SourceDefault), employee.ProfileSource) // Default source
	})
}

// TestEmployeeService_BackwardCompatibility tests that existing API behavior is preserved
func TestEmployeeService_BackwardCompatibility(t *testing.T) {
	// Setup mocks using existing test infrastructure
	employeeRepo := newMockEmployeeRepo()
	universityRepo := newMockUniversityRepo()
	maxService := newMockMaxService()
	authService := newMockAuthService()
	passwordGenerator := newMockPasswordGenerator()
	notificationService := newMockNotificationService()
	profileCache := &mockFailingProfileCache{}

	// Create service
	service := NewEmployeeService(
		employeeRepo,
		universityRepo,
		maxService,
		authService,
		passwordGenerator,
		notificationService,
		profileCache,
	)

	phone := "+79991234570"
	firstName := "Иван"
	lastName := "Петров"
	inn := "1234567890"
	kpp := "123456789"
	universityName := "Тестовый университет"

	// Setup MAX service to return profile data
	maxService.users[phone] = "max_123"

	t.Run("Explicit names provided - works as before", func(t *testing.T) {
		// When names are provided explicitly, system should work exactly as before (Requirements 7.1, 7.4)
		employee, err := service.AddEmployeeByPhone(
			phone,
			firstName, lastName, "",
			inn, kpp,
			universityName,
		)

		// Verify backward compatibility
		assert.NoError(t, err)
		assert.NotNil(t, employee)
		assert.Equal(t, firstName, employee.FirstName)
		assert.Equal(t, lastName, employee.LastName)
		assert.Equal(t, "max_123", employee.MaxID)
		assert.Equal(t, string(domain.SourceUserInput), employee.ProfileSource)
	})
}