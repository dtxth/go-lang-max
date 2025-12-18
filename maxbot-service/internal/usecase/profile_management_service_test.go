package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"maxbot-service/internal/domain"
	"maxbot-service/internal/infrastructure/cache"
	"maxbot-service/internal/infrastructure/maxapi"
)

func TestProfileManagementService_GetProfile(t *testing.T) {
	// Setup
	profileCache := cache.NewMockProfileCache()
	apiClient := maxapi.NewMockClient()
	service := NewProfileManagementService(profileCache, apiClient)
	ctx := context.Background()

	// Test getting non-existent profile
	profile, err := service.GetProfile(ctx, "user123")
	require.NoError(t, err)
	assert.Equal(t, "user123", profile.UserID)
	assert.Equal(t, domain.SourceDefault, profile.Source)

	// Store a profile first
	testProfile := domain.UserProfileCache{
		UserID:       "user123",
		MaxFirstName: "Иван",
		MaxLastName:  "Петров",
		Source:       domain.SourceWebhook,
		LastUpdated:  time.Now(),
	}
	err = profileCache.StoreProfile(ctx, "user123", testProfile)
	require.NoError(t, err)

	// Test getting existing profile
	profile, err = service.GetProfile(ctx, "user123")
	require.NoError(t, err)
	assert.Equal(t, "user123", profile.UserID)
	assert.Equal(t, "Иван", profile.MaxFirstName)
	assert.Equal(t, "Петров", profile.MaxLastName)
	assert.Equal(t, domain.SourceWebhook, profile.Source)
}

func TestProfileManagementService_SetUserProvidedName(t *testing.T) {
	// Setup
	profileCache := cache.NewMockProfileCache()
	apiClient := maxapi.NewMockClient()
	service := NewProfileManagementService(profileCache, apiClient)
	ctx := context.Background()

	// Test setting user-provided name
	profile, err := service.SetUserProvidedName(ctx, "user123", "Иван Петрович")
	require.NoError(t, err)
	assert.Equal(t, "user123", profile.UserID)
	assert.Equal(t, "Иван Петрович", profile.UserProvidedName)
	assert.Equal(t, domain.SourceUserInput, profile.Source)
	assert.Equal(t, "Иван Петрович", profile.GetDisplayName())

	// Test validation error
	_, err = service.SetUserProvidedName(ctx, "user123", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name cannot be empty")

	// Test invalid characters
	_, err = service.SetUserProvidedName(ctx, "user123", "Test123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid characters")
}

func TestProfileManagementService_UpdateProfile(t *testing.T) {
	// Setup
	profileCache := cache.NewMockProfileCache()
	apiClient := maxapi.NewMockClient()
	service := NewProfileManagementService(profileCache, apiClient)
	ctx := context.Background()

	// Store initial profile
	testProfile := domain.UserProfileCache{
		UserID:       "user123",
		MaxFirstName: "Иван",
		Source:       domain.SourceWebhook,
		LastUpdated:  time.Now(),
	}
	err := profileCache.StoreProfile(ctx, "user123", testProfile)
	require.NoError(t, err)

	// Test updating profile
	lastName := "Петров"
	updates := domain.ProfileUpdates{
		MaxLastName: &lastName,
	}

	profile, err := service.UpdateProfile(ctx, "user123", updates)
	require.NoError(t, err)
	assert.Equal(t, "user123", profile.UserID)
	assert.Equal(t, "Иван", profile.MaxFirstName)
	assert.Equal(t, "Петров", profile.MaxLastName)
	assert.Equal(t, "Иван Петров", profile.GetDisplayName())
}

func TestProfileManagementService_GetProfileStats(t *testing.T) {
	// Setup
	profileCache := cache.NewMockProfileCache()
	apiClient := maxapi.NewMockClient()
	service := NewProfileManagementService(profileCache, apiClient)
	ctx := context.Background()

	// Store some test profiles
	profiles := []domain.UserProfileCache{
		{
			UserID:       "user1",
			MaxFirstName: "Иван",
			MaxLastName:  "Петров",
			Source:       domain.SourceWebhook,
			LastUpdated:  time.Now(),
		},
		{
			UserID:           "user2",
			UserProvidedName: "Мария Сидорова",
			Source:           domain.SourceUserInput,
			LastUpdated:      time.Now(),
		},
		{
			UserID:       "user3",
			MaxFirstName: "Алексей",
			Source:       domain.SourceWebhook,
			LastUpdated:  time.Now(),
		},
	}

	for _, profile := range profiles {
		err := profileCache.StoreProfile(ctx, profile.UserID, profile)
		require.NoError(t, err)
	}

	// Test getting stats
	stats, err := service.GetProfileStats(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(3), stats.TotalProfiles)
	assert.Equal(t, int64(2), stats.ProfilesWithFullName) // user1 and user2 have full names
	assert.Equal(t, int64(2), stats.ProfilesBySource[domain.SourceWebhook])
	assert.Equal(t, int64(1), stats.ProfilesBySource[domain.SourceUserInput])
}

func TestProfileManagementService_Validation(t *testing.T) {
	// Setup
	profileCache := cache.NewMockProfileCache()
	apiClient := maxapi.NewMockClient()
	service := NewProfileManagementService(profileCache, apiClient)

	// Test validateUserProvidedName
	tests := []struct {
		name      string
		input     string
		shouldErr bool
	}{
		{"valid name", "Иван Петров", false},
		{"valid name with hyphen", "Анна-Мария", false},
		{"empty name", "", true},
		{"too long name", string(make([]rune, 101)), true},
		{"invalid characters", "Test123", true},
		{"valid russian name", "Владимир Владимирович", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateUserProvidedName(tt.input)
			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}