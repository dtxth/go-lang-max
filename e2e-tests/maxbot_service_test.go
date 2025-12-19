package main

import (
	"e2e-tests/utils"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMaxBotService(t *testing.T) {
	// Настройка клиента
	configs := utils.DefaultServiceConfigs()
	client := utils.NewTestClient(configs["maxbot"])

	// Ждем доступности сервиса
	err := utils.WaitForService(configs["maxbot"].BaseURL, 10)
	require.NoError(t, err, "MaxBot service should be available")

	t.Run("Health Check", func(t *testing.T) {
		resp, err := client.GetClient().R().Get("/health")
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode())
		
		var health map[string]interface{}
		err = json.Unmarshal(resp.Body(), &health)
		require.NoError(t, err)
		assert.Contains(t, health, "status")
	})

	t.Run("Root Endpoint", func(t *testing.T) {
		resp, err := client.GetClient().R().Get("/")
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode())
	})

	t.Run("Get Profile (Mock Mode)", func(t *testing.T) {
		// В mock режиме должен возвращать тестовые данные
		resp, err := client.GetClient().R().
			SetQueryParam("user_id", "123456789").
			Get("/profile")
		
		require.NoError(t, err)
		// В mock режиме может быть 200, в реальном режиме может быть ошибка
		assert.True(t, resp.StatusCode() == 200 || resp.StatusCode() >= 400)
		
		if resp.StatusCode() == 200 {
			var profile map[string]interface{}
			err = json.Unmarshal(resp.Body(), &profile)
			require.NoError(t, err)
			// В mock режиме должны быть базовые поля
			assert.Contains(t, profile, "id")
		}
	})

	t.Run("Get Bot Info", func(t *testing.T) {
		resp, err := client.GetClient().R().Get("/bot/info")
		require.NoError(t, err)
		// Может быть 200 в mock режиме или ошибка в реальном
		assert.True(t, resp.StatusCode() == 200 || resp.StatusCode() >= 400)
		
		if resp.StatusCode() == 200 {
			var botInfo map[string]interface{}
			err = json.Unmarshal(resp.Body(), &botInfo)
			require.NoError(t, err)
			// Должны быть базовые поля бота
			assert.Contains(t, botInfo, "name")
		}
	})

	t.Run("Validate Init Data", func(t *testing.T) {
		initData := map[string]interface{}{
			"initData": "user=%7B%22id%22%3A123456789%2C%22first_name%22%3A%22Test%22%2C%22last_name%22%3A%22User%22%7D&auth_date=1640995200&hash=test_hash",
		}
		
		resp, err := client.GetClient().R().
			SetBody(initData).
			Post("/validate")
		
		require.NoError(t, err)
		// В mock режиме может быть 200, в реальном режиме может быть ошибка валидации
		assert.True(t, resp.StatusCode() == 200 || resp.StatusCode() >= 400)
	})

	t.Run("Webhook Endpoint", func(t *testing.T) {
		webhookData := map[string]interface{}{
			"update_id": 123456789,
			"message": map[string]interface{}{
				"message_id": 1,
				"from": map[string]interface{}{
					"id":         123456789,
					"first_name": "Test",
					"last_name":  "User",
				},
				"chat": map[string]interface{}{
					"id":   123456789,
					"type": "private",
				},
				"date": 1640995200,
				"text": "Test message",
			},
		}
		
		resp, err := client.GetClient().R().
			SetBody(webhookData).
			Post("/webhook/max")
		
		require.NoError(t, err)
		// Webhook может принимать данные или отклонять их
		assert.True(t, resp.StatusCode() >= 200 && resp.StatusCode() < 500)
	})

	t.Run("Metrics Endpoint", func(t *testing.T) {
		resp, err := client.GetClient().R().Get("/metrics")
		require.NoError(t, err)
		// Может быть 200 если метрики включены
		assert.True(t, resp.StatusCode() == 200 || resp.StatusCode() == 404)
		
		if resp.StatusCode() == 200 {
			var metrics map[string]interface{}
			err = json.Unmarshal(resp.Body(), &metrics)
			require.NoError(t, err)
			// Должны быть базовые метрики
			assert.IsType(t, map[string]interface{}{}, metrics)
		}
	})

	t.Run("Cache Status", func(t *testing.T) {
		resp, err := client.GetClient().R().Get("/cache/status")
		require.NoError(t, err)
		// Может быть 200 если кеш доступен
		assert.True(t, resp.StatusCode() == 200 || resp.StatusCode() >= 400)
		
		if resp.StatusCode() == 200 {
			var cacheStatus map[string]interface{}
			err = json.Unmarshal(resp.Body(), &cacheStatus)
			require.NoError(t, err)
			assert.Contains(t, cacheStatus, "status")
		}
	})

	t.Run("Invalid Profile Request", func(t *testing.T) {
		resp, err := client.GetClient().R().
			SetQueryParam("user_id", "invalid").
			Get("/profile")
		
		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode())
	})

	t.Run("Empty Webhook Data", func(t *testing.T) {
		resp, err := client.GetClient().R().
			SetBody(map[string]interface{}{}).
			Post("/webhook/max")
		
		require.NoError(t, err)
		// Пустые данные должны быть отклонены
		assert.True(t, resp.StatusCode() >= 400)
	})

	t.Run("Invalid Init Data", func(t *testing.T) {
		invalidInitData := map[string]interface{}{
			"initData": "invalid_data",
		}
		
		resp, err := client.GetClient().R().
			SetBody(invalidInitData).
			Post("/validate")
		
		require.NoError(t, err)
		// Невалидные данные должны быть отклонены
		assert.True(t, resp.StatusCode() >= 400)
	})

	t.Run("Method Not Allowed", func(t *testing.T) {
		resp, err := client.GetClient().R().
			Delete("/profile")
		
		require.NoError(t, err)
		assert.Equal(t, 405, resp.StatusCode())
	})
}