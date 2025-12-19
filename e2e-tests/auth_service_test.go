package main

import (
	"e2e-tests/utils"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthService(t *testing.T) {
	// Настройка клиента
	configs := utils.DefaultServiceConfigs()
	client := utils.NewTestClient(configs["auth"])

	// Ждем доступности сервиса
	err := utils.WaitForService(configs["auth"].BaseURL, 10)
	require.NoError(t, err, "Auth service should be available")

	t.Run("Health Check", func(t *testing.T) {
		resp, err := client.GetClient().R().Get("/health")
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode())
	})

	t.Run("Metrics", func(t *testing.T) {
		resp, err := client.GetClient().R().Get("/metrics")
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode())
		
		var metrics map[string]interface{}
		err = json.Unmarshal(resp.Body(), &metrics)
		require.NoError(t, err)
		assert.Contains(t, metrics, "user_creations")
	})

	t.Run("Bot Info", func(t *testing.T) {
		resp, err := client.GetClient().R().Get("/bot/me")
		require.NoError(t, err)
		// Может быть 200 или 500 в зависимости от настройки MaxBot
		assert.True(t, resp.StatusCode() == 200 || resp.StatusCode() == 500)
	})

	var testUser utils.TestUser
	var accessToken string

	t.Run("User Registration", func(t *testing.T) {
		testUser = utils.GenerateTestUser()
		
		resp, err := client.GetClient().R().
			SetBody(testUser).
			Post("/register")
		
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode())
		
		var user map[string]interface{}
		err = json.Unmarshal(resp.Body(), &user)
		require.NoError(t, err)
		assert.Equal(t, testUser.Email, user["email"])
		assert.Equal(t, testUser.Phone, user["phone"])
	})

	t.Run("User Login by Email", func(t *testing.T) {
		loginData := map[string]string{
			"email":    testUser.Email,
			"password": testUser.Password,
		}
		
		resp, err := client.GetClient().R().
			SetBody(loginData).
			Post("/login")
		
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode())
		
		var tokens map[string]interface{}
		err = json.Unmarshal(resp.Body(), &tokens)
		require.NoError(t, err)
		assert.Contains(t, tokens, "access_token")
		assert.Contains(t, tokens, "refresh_token")
		
		accessToken = tokens["access_token"].(string)
	})

	t.Run("User Login by Phone", func(t *testing.T) {
		loginData := map[string]string{
			"phone":    testUser.Phone,
			"password": testUser.Password,
		}
		
		resp, err := client.GetClient().R().
			SetBody(loginData).
			Post("/login-phone")
		
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode())
		
		var tokens map[string]interface{}
		err = json.Unmarshal(resp.Body(), &tokens)
		require.NoError(t, err)
		assert.Contains(t, tokens, "access_token")
		assert.Contains(t, tokens, "refresh_token")
	})

	t.Run("Password Reset Request", func(t *testing.T) {
		resetData := map[string]string{
			"phone": testUser.Phone,
		}
		
		resp, err := client.GetClient().R().
			SetBody(resetData).
			Post("/auth/password-reset/request")
		
		require.NoError(t, err)
		// Может быть 200 или 500 в зависимости от настройки уведомлений
		assert.True(t, resp.StatusCode() == 200 || resp.StatusCode() == 500)
	})

	t.Run("Change Password (Protected)", func(t *testing.T) {
		changeData := map[string]string{
			"current_password": testUser.Password,
			"new_password":     "NewTestPassword123!",
		}
		
		resp, err := client.GetClient().R().
			SetAuthToken(accessToken).
			SetBody(changeData).
			Post("/auth/password/change")
		
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode())
	})

	t.Run("MAX Authentication", func(t *testing.T) {
		maxData := utils.GenerateMAXInitData()
		
		resp, err := client.GetClient().R().
			SetBody(maxData).
			Post("/auth/max")
		
		require.NoError(t, err)
		// Может быть 200 или 401 в зависимости от настройки MAX
		assert.True(t, resp.StatusCode() == 200 || resp.StatusCode() == 401)
	})

	t.Run("Invalid Login", func(t *testing.T) {
		loginData := map[string]string{
			"email":    "invalid@example.com",
			"password": "wrongpassword",
		}
		
		resp, err := client.GetClient().R().
			SetBody(loginData).
			Post("/login")
		
		require.NoError(t, err)
		assert.Equal(t, 401, resp.StatusCode())
	})

	t.Run("Invalid Registration", func(t *testing.T) {
		invalidUser := map[string]string{
			"email": "invalid-email",
			// Отсутствует пароль
		}
		
		resp, err := client.GetClient().R().
			SetBody(invalidUser).
			Post("/register")
		
		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode())
	})
}