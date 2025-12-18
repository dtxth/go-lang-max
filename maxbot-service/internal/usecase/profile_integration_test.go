package usecase

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"maxbot-service/internal/domain"
	"maxbot-service/internal/infrastructure/cache"
	"maxbot-service/internal/infrastructure/maxapi"
)

// TestProfileIntegration_CompleteFlow tests the complete profile management flow
func TestProfileIntegration_CompleteFlow(t *testing.T) {
	// Setup services
	profileCache := cache.NewMockProfileCache()
	apiClient := maxapi.NewMockClient()
	
	webhookHandler := NewWebhookHandlerService(profileCache, nil)
	profileManagement := NewProfileManagementService(profileCache, apiClient)
	
	ctx := context.Background()
	userID := "test_user_integration"

	// Step 1: Initial webhook event with partial profile (Requirements 1.2, 1.3)
	t.Run("Initial webhook with first name only", func(t *testing.T) {
		event := domain.MaxWebhookEvent{
			Type: "message_new",
			Message: &domain.MessageEvent{
				From: domain.UserInfo{
					UserID:    userID,
					FirstName: "Иван",
				},
				Text: "Привет!",
			},
		}

		err := webhookHandler.HandleMaxWebhook(ctx, event)
		require.NoError(t, err)

		// Check profile was created with first name only
		profile, err := profileManagement.GetProfile(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, "Иван", profile.MaxFirstName)
		assert.Equal(t, "", profile.MaxLastName)
		assert.Equal(t, "", profile.UserProvidedName)
		assert.Equal(t, domain.SourceWebhook, profile.Source)
		assert.Equal(t, "Иван", profile.GetDisplayName())
		assert.False(t, profile.HasFullName())
	})

	// Step 2: User provides full name via message (Requirements 2.2, 2.4)
	t.Run("User provides full name via message", func(t *testing.T) {
		event := domain.MaxWebhookEvent{
			Type: "message_new",
			Message: &domain.MessageEvent{
				From: domain.UserInfo{
					UserID:    userID,
					FirstName: "Иван",
				},
				Text: "меня зовут Иван Петрович Сидоров",
			},
		}

		err := webhookHandler.HandleMaxWebhook(ctx, event)
		require.NoError(t, err)

		// Check profile was updated with user-provided name
		profile, err := profileManagement.GetProfile(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, "Иван", profile.MaxFirstName)
		assert.Equal(t, "", profile.MaxLastName)
		assert.Equal(t, "Иван Петрович Сидоров", profile.UserProvidedName)
		assert.Equal(t, domain.SourceUserInput, profile.Source)
		assert.Equal(t, "Иван Петрович Сидоров", profile.GetDisplayName()) // User input has priority
		assert.True(t, profile.HasFullName())
	})

	// Step 3: Later webhook with last name (Requirements 5.2)
	t.Run("Later webhook with last name preserves user input", func(t *testing.T) {
		event := domain.MaxWebhookEvent{
			Type: "callback_query",
			Callback: &domain.CallbackEvent{
				User: domain.UserInfo{
					UserID:    userID,
					FirstName: "Иван",
					LastName:  "Петров",
				},
				Data: "some_callback_data",
			},
		}

		err := webhookHandler.HandleMaxWebhook(ctx, event)
		require.NoError(t, err)

		// Check that user-provided name is preserved while MAX data is updated
		profile, err := profileManagement.GetProfile(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, "Иван", profile.MaxFirstName)
		assert.Equal(t, "Петров", profile.MaxLastName)
		assert.Equal(t, "Иван Петрович Сидоров", profile.UserProvidedName) // Preserved
		assert.Equal(t, domain.SourceWebhook, profile.Source) // Updated by webhook
		assert.Equal(t, "Иван Петрович Сидоров", profile.GetDisplayName()) // User input still has priority
	})

	// Step 4: API update of profile (Requirements 5.4, 5.5)
	t.Run("API update of profile", func(t *testing.T) {
		newName := "Иван Петрович"
		updates := domain.ProfileUpdates{
			UserProvidedName: &newName,
		}

		updatedProfile, err := profileManagement.UpdateProfile(ctx, userID, updates)
		require.NoError(t, err)
		assert.Equal(t, "Иван Петрович", updatedProfile.UserProvidedName)
		assert.Equal(t, "Иван Петрович", updatedProfile.GetDisplayName())
		assert.True(t, updatedProfile.HasFullName())
	})

	// Step 5: Get profile stats (Requirements 6.1, 6.3)
	t.Run("Get profile statistics", func(t *testing.T) {
		stats, err := profileManagement.GetProfileStats(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(1), stats.TotalProfiles)
		assert.Equal(t, int64(1), stats.ProfilesWithFullName)
		assert.Contains(t, stats.ProfilesBySource, domain.SourceWebhook)
	})
}

// TestProfileIntegration_NamePriority tests name priority logic (Requirements 2.3, 5.3)
func TestProfileIntegration_NamePriority(t *testing.T) {
	profileCache := cache.NewMockProfileCache()
	apiClient := maxapi.NewMockClient()
	profileManagement := NewProfileManagementService(profileCache, apiClient)
	ctx := context.Background()

	tests := []struct {
		name                string
		maxFirstName        string
		maxLastName         string
		userProvidedName    string
		expectedDisplayName string
		expectedHasFullName bool
	}{
		{
			name:                "User provided name has highest priority",
			maxFirstName:        "Иван",
			maxLastName:         "Петров",
			userProvidedName:    "Александр Сидоров",
			expectedDisplayName: "Александр Сидоров",
			expectedHasFullName: true,
		},
		{
			name:                "MAX full name when no user input",
			maxFirstName:        "Иван",
			maxLastName:         "Петров",
			userProvidedName:    "",
			expectedDisplayName: "Иван Петров",
			expectedHasFullName: true,
		},
		{
			name:                "MAX first name only",
			maxFirstName:        "Иван",
			maxLastName:         "",
			userProvidedName:    "",
			expectedDisplayName: "Иван",
			expectedHasFullName: false,
		},
		{
			name:                "Empty profile",
			maxFirstName:        "",
			maxLastName:         "",
			userProvidedName:    "",
			expectedDisplayName: "",
			expectedHasFullName: false,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID := fmt.Sprintf("test_user_%d", i)
			
			// Create profile
			profile := domain.UserProfileCache{
				UserID:           userID,
				MaxFirstName:     tt.maxFirstName,
				MaxLastName:      tt.maxLastName,
				UserProvidedName: tt.userProvidedName,
				Source:           domain.SourceWebhook,
				LastUpdated:      time.Now(),
			}

			err := profileCache.StoreProfile(ctx, userID, profile)
			require.NoError(t, err)

			// Get and verify profile
			retrievedProfile, err := profileManagement.GetProfile(ctx, userID)
			require.NoError(t, err)
			
			assert.Equal(t, tt.expectedDisplayName, retrievedProfile.GetDisplayName())
			assert.Equal(t, tt.expectedHasFullName, retrievedProfile.HasFullName())
		})
	}
}

// TestProfileIntegration_ErrorHandling tests error handling scenarios
func TestProfileIntegration_ErrorHandling(t *testing.T) {
	profileCache := cache.NewMockProfileCache()
	apiClient := maxapi.NewMockClient()
	profileManagement := NewProfileManagementService(profileCache, apiClient)
	ctx := context.Background()

	// Test invalid user ID
	t.Run("Invalid user ID", func(t *testing.T) {
		_, err := profileManagement.GetProfile(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user_id is required")
	})

	// Test invalid name validation
	t.Run("Invalid name validation", func(t *testing.T) {
		_, err := profileManagement.SetUserProvidedName(ctx, "user123", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name cannot be empty")

		_, err = profileManagement.SetUserProvidedName(ctx, "user123", "Invalid123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid characters")
	})

	// Test profile updates validation
	t.Run("Profile updates validation", func(t *testing.T) {
		invalidName := "Invalid@Name"
		updates := domain.ProfileUpdates{
			UserProvidedName: &invalidName,
		}

		_, err := profileManagement.UpdateProfile(ctx, "user123", updates)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid characters")
	})
}