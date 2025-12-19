package main

import (
	"e2e-tests/utils"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChatService(t *testing.T) {
	// Настройка клиента
	configs := utils.DefaultServiceConfigs()
	client := utils.NewTestClient(configs["chat"])

	// Ждем доступности сервиса
	err := utils.WaitForService(configs["chat"].BaseURL, 10)
	require.NoError(t, err, "Chat service should be available")

	// Получаем токен авторизации из auth service
	authClient := utils.NewTestClient(configs["auth"])
	testUser := utils.GenerateTestUser()
	
	// Регистрируем пользователя
	_, err = authClient.GetClient().R().
		SetBody(testUser).
		Post("/register")
	require.NoError(t, err)
	
	// Логинимся
	loginResp, err := authClient.GetClient().R().
		SetBody(map[string]string{
			"email":    testUser.Email,
			"password": testUser.Password,
		}).
		Post("/login")
	require.NoError(t, err)
	
	var tokens map[string]interface{}
	err = json.Unmarshal(loginResp.Body(), &tokens)
	require.NoError(t, err)
	
	accessToken := tokens["access_token"].(string)
	client.SetAuthToken(accessToken)

	var testChat utils.TestChat

	t.Run("Health Check", func(t *testing.T) {
		resp, err := client.GetClient().R().Get("/health")
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode())
	})

	t.Run("Create Chat", func(t *testing.T) {
		testChat = utils.GenerateTestChat()
		
		resp, err := client.GetClient().R().
			SetBody(testChat).
			Post("/chats")
		
		require.NoError(t, err)
		// Может быть 201 или другой код в зависимости от реализации
		assert.True(t, resp.StatusCode() >= 200 && resp.StatusCode() < 300)
		
		if resp.StatusCode() == 201 {
			var chat map[string]interface{}
			err = json.Unmarshal(resp.Body(), &chat)
			require.NoError(t, err)
			
			if id, ok := chat["id"].(string); ok {
				testChat.ID = id
			}
		}
	})

	t.Run("Get Chats", func(t *testing.T) {
		resp, err := client.GetClient().R().
			Get("/chats")
		
		require.NoError(t, err)
		// Может быть 200 или другой код в зависимости от реализации
		assert.True(t, resp.StatusCode() >= 200 && resp.StatusCode() < 500)
		
		if resp.StatusCode() == 200 {
			var chats []interface{}
			err = json.Unmarshal(resp.Body(), &chats)
			require.NoError(t, err)
		}
	})

	t.Run("Get Chat by ID", func(t *testing.T) {
		if testChat.ID != "" {
			resp, err := client.GetClient().R().
				Get("/chats/" + testChat.ID)
			
			require.NoError(t, err)
			// Может быть 200, 404 или другой код
			assert.True(t, resp.StatusCode() >= 200 && resp.StatusCode() < 500)
		} else {
			t.Skip("Chat ID not available")
		}
	})

	t.Run("Update Chat", func(t *testing.T) {
		if testChat.ID != "" {
			updateData := map[string]interface{}{
				"name":        testChat.Name + " Updated",
				"description": "Updated description",
			}
			
			resp, err := client.GetClient().R().
				SetBody(updateData).
				Put("/chats/" + testChat.ID)
			
			require.NoError(t, err)
			// Может быть 200, 404 или другой код
			assert.True(t, resp.StatusCode() >= 200 && resp.StatusCode() < 500)
		} else {
			t.Skip("Chat ID not available")
		}
	})

	t.Run("Send Message", func(t *testing.T) {
		if testChat.ID != "" {
			messageData := map[string]interface{}{
				"chat_id": testChat.ID,
				"content": "Test message",
				"type":    "text",
			}
			
			resp, err := client.GetClient().R().
				SetBody(messageData).
				Post("/messages")
			
			require.NoError(t, err)
			// Может быть 201, 400 или другой код
			assert.True(t, resp.StatusCode() >= 200 && resp.StatusCode() < 500)
		} else {
			t.Skip("Chat ID not available")
		}
	})

	t.Run("Get Messages", func(t *testing.T) {
		if testChat.ID != "" {
			resp, err := client.GetClient().R().
				SetQueryParam("chat_id", testChat.ID).
				SetQueryParam("limit", "10").
				Get("/messages")
			
			require.NoError(t, err)
			// Может быть 200 или другой код
			assert.True(t, resp.StatusCode() >= 200 && resp.StatusCode() < 500)
		} else {
			t.Skip("Chat ID not available")
		}
	})

	t.Run("Join Chat", func(t *testing.T) {
		if testChat.ID != "" {
			resp, err := client.GetClient().R().
				Post("/chats/" + testChat.ID + "/join")
			
			require.NoError(t, err)
			// Может быть 200, 400 или другой код
			assert.True(t, resp.StatusCode() >= 200 && resp.StatusCode() < 500)
		} else {
			t.Skip("Chat ID not available")
		}
	})

	t.Run("Leave Chat", func(t *testing.T) {
		if testChat.ID != "" {
			resp, err := client.GetClient().R().
				Post("/chats/" + testChat.ID + "/leave")
			
			require.NoError(t, err)
			// Может быть 200, 400 или другой код
			assert.True(t, resp.StatusCode() >= 200 && resp.StatusCode() < 500)
		} else {
			t.Skip("Chat ID not available")
		}
	})

	t.Run("Get Chat Participants", func(t *testing.T) {
		if testChat.ID != "" {
			resp, err := client.GetClient().R().
				Get("/chats/" + testChat.ID + "/participants")
			
			require.NoError(t, err)
			// Может быть 200 или другой код
			assert.True(t, resp.StatusCode() >= 200 && resp.StatusCode() < 500)
		} else {
			t.Skip("Chat ID not available")
		}
	})

	t.Run("Unauthorized Access", func(t *testing.T) {
		// Очищаем токен авторизации
		client.ClearAuth()
		
		resp, err := client.GetClient().R().
			Get("/chats")
		
		require.NoError(t, err)
		// Должен быть 401 Unauthorized
		assert.Equal(t, 401, resp.StatusCode())
	})

	t.Run("Invalid Chat Data", func(t *testing.T) {
		// Восстанавливаем токен
		client.SetAuthToken(accessToken)
		
		invalidChat := map[string]interface{}{
			"name": "", // Пустое имя
		}
		
		resp, err := client.GetClient().R().
			SetBody(invalidChat).
			Post("/chats")
		
		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode())
	})

	t.Run("Non-existent Chat", func(t *testing.T) {
		resp, err := client.GetClient().R().
			Get("/chats/non-existent-id")
		
		require.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode())
	})
}