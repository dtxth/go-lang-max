package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"maxbot-service/internal/domain"
	"maxbot-service/internal/infrastructure/cache"
)

func TestWebhookHandlerService_ProcessUserNameInput(t *testing.T) {
	// Setup
	profileCache := cache.NewMockProfileCache()
	handler := NewWebhookHandlerService(profileCache, nil)
	ctx := context.Background()

	tests := []struct {
		name        string
		messageText string
		shouldProcess bool
		expectedName string
	}{
		{
			name:          "setname command",
			messageText:   "/setname Иван Петров",
			shouldProcess: true,
			expectedName:  "Иван Петров",
		},
		{
			name:          "russian name command",
			messageText:   "меня зовут Мария Сидорова",
			shouldProcess: true,
			expectedName:  "Мария Сидорова",
		},
		{
			name:          "my name command",
			messageText:   "мое имя Алексей Иванов",
			shouldProcess: true,
			expectedName:  "Алексей Иванов",
		},
		{
			name:          "regular message",
			messageText:   "Привет, как дела?",
			shouldProcess: false,
			expectedName:  "",
		},
		{
			name:          "empty name",
			messageText:   "/setname",
			shouldProcess: true,
			expectedName:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID := "test_user_" + tt.name
			
			// Process user input
			err := handler.processUserNameInput(ctx, userID, tt.messageText)
			require.NoError(t, err)

			if tt.shouldProcess && tt.expectedName != "" {
				// Check if profile was updated
				profile, err := profileCache.GetProfile(ctx, userID)
				require.NoError(t, err)
				
				if profile != nil {
					assert.Equal(t, tt.expectedName, profile.UserProvidedName)
					assert.Equal(t, domain.SourceUserInput, profile.Source)
				}
			}
		})
	}
}

func TestWebhookHandlerService_IsNameUpdateCommand(t *testing.T) {
	handler := &WebhookHandlerService{}

	tests := []struct {
		name        string
		messageText string
		expected    bool
	}{
		{"setname command", "/setname", true},
		{"russian name command", "меня зовут", true},
		{"my name command", "мое имя", true},
		{"my name command alt", "моё имя", true},
		{"case insensitive", "МЕНЯ ЗОВУТ", true},
		{"with extra spaces", "  /setname  ", true},
		{"regular message", "Привет!", false},
		{"partial match", "зовут меня", false},
		{"empty message", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.isNameUpdateCommand(tt.messageText)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWebhookHandlerService_ExtractUserNameFromMessage(t *testing.T) {
	handler := &WebhookHandlerService{}

	tests := []struct {
		name        string
		messageText string
		expected    string
	}{
		{
			name:        "setname command",
			messageText: "/setname Иван Петров",
			expected:    "Иван Петров",
		},
		{
			name:        "russian command",
			messageText: "меня зовут Мария Сидорова",
			expected:    "Мария Сидорова",
		},
		{
			name:        "my name command",
			messageText: "мое имя Алексей",
			expected:    "Алексей",
		},
		{
			name:        "with extra spaces",
			messageText: "  /setname   Владимир Владимирович  ",
			expected:    "Владимир Владимирович",
		},
		{
			name:        "empty name",
			messageText: "/setname",
			expected:    "",
		},
		{
			name:        "no command",
			messageText: "Привет!",
			expected:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.extractUserNameFromMessage(tt.messageText)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWebhookHandlerService_ValidateUserProvidedName(t *testing.T) {
	handler := &WebhookHandlerService{}

	tests := []struct {
		name      string
		input     string
		shouldErr bool
	}{
		{"valid russian name", "Иван Петров", false},
		{"valid name with hyphen", "Анна-Мария", false},
		{"valid single name", "Владимир", false},
		{"empty name", "", true},
		{"too long name", string(make([]rune, 101)), true},
		{"invalid characters - numbers", "Иван123", true},
		{"invalid characters - symbols", "Иван@Петров", true},
		{"valid with ё", "Семён Алёшин", false},
		{"english name", "John Smith", false},
		{"mixed languages", "John Иванов", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.validateUserProvidedName(tt.input)
			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWebhookHandlerService_HandleMaxWebhook_WithUserInput(t *testing.T) {
	// Setup
	profileCache := cache.NewMockProfileCache()
	handler := NewWebhookHandlerService(profileCache, nil)
	ctx := context.Background()

	// Test webhook event with user name input
	event := domain.MaxWebhookEvent{
		Type: "message_new",
		Message: &domain.MessageEvent{
			From: domain.UserInfo{
				UserID:    "user123",
				FirstName: "Иван",
			},
			Text: "меня зовут Иван Петрович",
		},
	}

	// Process webhook
	err := handler.HandleMaxWebhook(ctx, event)
	require.NoError(t, err)

	// Check that both webhook profile and user input were processed
	profile, err := profileCache.GetProfile(ctx, "user123")
	require.NoError(t, err)
	require.NotNil(t, profile)

	// Should have both MAX data and user-provided name
	assert.Equal(t, "Иван", profile.MaxFirstName)
	assert.Equal(t, "Иван Петрович", profile.UserProvidedName)
	assert.Equal(t, domain.SourceUserInput, profile.Source) // Last update wins
	assert.Equal(t, "Иван Петрович", profile.GetDisplayName()) // User-provided has priority
}