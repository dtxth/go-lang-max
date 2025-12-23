package usecase

import (
	"context"
	"testing"

	"maxbot-service/internal/infrastructure/maxapi"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMaxBotService_GetInternalUsers(t *testing.T) {
	// Setup
	mockClient := maxapi.NewMockClient()
	service := NewMaxBotService(mockClient)
	ctx := context.Background()

	t.Run("successful request with valid phones", func(t *testing.T) {
		phones := []string{"+79991234567", "+79995678901", "+79999999999"}
		
		users, failedPhones, err := service.GetInternalUsers(ctx, phones)
		
		require.NoError(t, err)
		assert.Len(t, users, 3)
		assert.Empty(t, failedPhones)
		
		// Check first user (should have Петр Петров based on mock logic)
		user1 := users[0]
		assert.Equal(t, "Петр", user1.FirstName)
		assert.Equal(t, "Петров", user1.LastName)
		assert.Equal(t, "+79991234567", user1.PhoneNumber)
		assert.Equal(t, "petr_petrov", user1.Username)
		assert.Equal(t, "max.ru/petr_petrov", user1.Link)
		assert.False(t, user1.IsBot)
		assert.NotEmpty(t, user1.AvatarURL)
		assert.NotEmpty(t, user1.FullAvatarURL)
		
		// Check second user (should have Анна Сидорова based on mock logic)
		user2 := users[1]
		assert.Equal(t, "Анна", user2.FirstName)
		assert.Equal(t, "Сидорова", user2.LastName)
		assert.Equal(t, "+79995678901", user2.PhoneNumber)
		assert.Equal(t, "anna_sidorova", user2.Username)
		assert.Equal(t, "max.ru/anna_sidorova", user2.Link)
		
		// Check third user (should have Мария Иванова with no username)
		user3 := users[2]
		assert.Equal(t, "Мария", user3.FirstName)
		assert.Equal(t, "Иванова", user3.LastName)
		assert.Equal(t, "+79999999999", user3.PhoneNumber)
		assert.Empty(t, user3.Username)
		assert.Equal(t, "max.ru/u/abc123hash", user3.Link)
	})

	t.Run("empty phone list", func(t *testing.T) {
		users, failedPhones, err := service.GetInternalUsers(ctx, []string{})
		
		require.NoError(t, err)
		assert.Empty(t, users)
		assert.Empty(t, failedPhones)
	})

	t.Run("invalid phones", func(t *testing.T) {
		phones := []string{"invalid", "123", ""}
		
		users, failedPhones, err := service.GetInternalUsers(ctx, phones)
		
		require.NoError(t, err)
		assert.Empty(t, users)
		assert.Len(t, failedPhones, 3)
		assert.Contains(t, failedPhones, "invalid")
		assert.Contains(t, failedPhones, "123")
		assert.Contains(t, failedPhones, "")
	})

	t.Run("mixed valid and invalid phones", func(t *testing.T) {
		phones := []string{"+79991234567", "invalid", "+79995678901"}
		
		users, failedPhones, err := service.GetInternalUsers(ctx, phones)
		
		require.NoError(t, err)
		assert.Len(t, users, 2)
		assert.Len(t, failedPhones, 1)
		assert.Contains(t, failedPhones, "invalid")
		
		// Check that valid phones were processed
		assert.Equal(t, "+79991234567", users[0].PhoneNumber)
		assert.Equal(t, "+79995678901", users[1].PhoneNumber)
	})

	t.Run("batch size limit", func(t *testing.T) {
		// Create 101 phones to exceed limit
		phones := make([]string, 101)
		for i := 0; i < 101; i++ {
			phones[i] = "+7999123456" + string(rune('0'+i%10))
		}
		
		users, failedPhones, err := service.GetInternalUsers(ctx, phones)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "batch size exceeds maximum of 100 phones")
		assert.Nil(t, users)
		assert.Nil(t, failedPhones)
	})
}

func TestMaxBotService_GetInternalUsers_Integration(t *testing.T) {
	// This test demonstrates how the method would be used in practice
	mockClient := maxapi.NewMockClient()
	service := NewMaxBotService(mockClient)
	ctx := context.Background()

	// Simulate employee creation scenario
	phones := []string{
		"+79991234567", // Should get Петр Петров
		"+79995678901", // Should get Анна Сидорова  
		"+79999999999", // Should get Мария Иванова (no username)
		"+79991111111", // Should get Иван Иванов (default)
	}

	users, failedPhones, err := service.GetInternalUsers(ctx, phones)
	
	require.NoError(t, err)
	assert.Len(t, users, 4)
	assert.Empty(t, failedPhones)

	// Verify that we got detailed user information
	for _, user := range users {
		assert.NotEmpty(t, user.FirstName, "FirstName should not be empty")
		assert.NotEmpty(t, user.LastName, "LastName should not be empty")
		assert.NotEmpty(t, user.PhoneNumber, "PhoneNumber should not be empty")
		assert.NotEmpty(t, user.Link, "Link should not be empty")
		assert.NotEmpty(t, user.AvatarURL, "AvatarURL should not be empty")
		assert.NotEmpty(t, user.FullAvatarURL, "FullAvatarURL should not be empty")
		assert.Greater(t, user.UserID, int64(0), "UserID should be positive")
		assert.False(t, user.IsBot, "Should not be bot")
	}

	// Verify link generation logic
	for _, user := range users {
		if user.Username != "" {
			assert.Equal(t, "max.ru/"+user.Username, user.Link, "Link should use username format")
		} else {
			assert.Contains(t, user.Link, "max.ru/u/", "Link should use hash format for users without username")
		}
	}
}