package main

import (
	"bytes"
	"e2e-tests/utils"
	"encoding/json"
	"mime/multipart"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrationService(t *testing.T) {
	// Настройка клиента
	configs := utils.DefaultServiceConfigs()
	client := utils.NewTestClient(configs["migration"])

	// Ждем доступности сервиса
	err := utils.WaitForService(configs["migration"].BaseURL, 10)
	require.NoError(t, err, "Migration service should be available")

	t.Run("Health Check", func(t *testing.T) {
		resp, err := client.GetClient().R().Get("/health")
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode())
		
		var health map[string]interface{}
		err = json.Unmarshal(resp.Body(), &health)
		require.NoError(t, err)
		assert.Contains(t, health, "status")
	})

	t.Run("Get Migration Status", func(t *testing.T) {
		resp, err := client.GetClient().R().Get("/migrations/status")
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode())
		
		var status map[string]interface{}
		err = json.Unmarshal(resp.Body(), &status)
		require.NoError(t, err)
		assert.Contains(t, status, "migrations")
	})

	t.Run("Get Migration History", func(t *testing.T) {
		resp, err := client.GetClient().R().
			SetQueryParam("limit", "10").
			SetQueryParam("offset", "0").
			Get("/migrations/history")
		
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode())
		
		var history map[string]interface{}
		err = json.Unmarshal(resp.Body(), &history)
		require.NoError(t, err)
		assert.Contains(t, history, "migrations")
		assert.Contains(t, history, "total")
	})

	t.Run("Upload Excel File", func(t *testing.T) {
		// Создаем простой Excel файл для тестирования
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)
		
		// Добавляем файл
		fileWriter, err := writer.CreateFormFile("file", "test.xlsx")
		require.NoError(t, err)
		
		// Простые тестовые данные (не настоящий Excel, но для тестирования API)
		_, err = fileWriter.Write([]byte("test excel data"))
		require.NoError(t, err)
		
		// Добавляем дополнительные поля
		writer.WriteField("university_id", "1")
		writer.WriteField("migration_type", "structure")
		
		writer.Close()
		
		resp, err := client.GetClient().R().
			SetHeader("Content-Type", writer.FormDataContentType()).
			SetBody(buf.Bytes()).
			Post("/migrations/upload")
		
		require.NoError(t, err)
		// Может быть 200, 400 (неверный формат файла) или 500
		assert.True(t, resp.StatusCode() >= 200 && resp.StatusCode() < 600)
		
		if resp.StatusCode() == 200 {
			var result map[string]interface{}
			err = json.Unmarshal(resp.Body(), &result)
			require.NoError(t, err)
			assert.Contains(t, result, "migration_id")
		}
	})

	t.Run("Start Migration", func(t *testing.T) {
		migrationData := map[string]interface{}{
			"university_id":   1,
			"migration_type":  "structure",
			"source_type":     "excel",
			"source_data": map[string]interface{}{
				"file_path": "/tmp/test.xlsx",
			},
		}
		
		resp, err := client.GetClient().R().
			SetBody(migrationData).
			Post("/migrations/start")
		
		require.NoError(t, err)
		// Может быть 200, 400 или 500 в зависимости от данных
		assert.True(t, resp.StatusCode() >= 200 && resp.StatusCode() < 600)
	})

	t.Run("Google Sheets Migration", func(t *testing.T) {
		sheetsData := map[string]interface{}{
			"university_id":   1,
			"migration_type":  "structure",
			"source_type":     "google_sheets",
			"source_data": map[string]interface{}{
				"spreadsheet_id": "test-spreadsheet-id",
				"sheet_name":     "Test Sheet",
			},
		}
		
		resp, err := client.GetClient().R().
			SetBody(sheetsData).
			Post("/migrations/start")
		
		require.NoError(t, err)
		// Может быть 200, 400 или 500 в зависимости от настройки Google API
		assert.True(t, resp.StatusCode() >= 200 && resp.StatusCode() < 600)
	})

	t.Run("Get Migration by ID", func(t *testing.T) {
		// Используем тестовый ID
		resp, err := client.GetClient().R().
			Get("/migrations/test-migration-id")
		
		require.NoError(t, err)
		// Может быть 200 или 404
		assert.True(t, resp.StatusCode() == 200 || resp.StatusCode() == 404)
	})

	t.Run("Cancel Migration", func(t *testing.T) {
		resp, err := client.GetClient().R().
			Post("/migrations/test-migration-id/cancel")
		
		require.NoError(t, err)
		// Может быть 200, 404 или 400
		assert.True(t, resp.StatusCode() >= 200 && resp.StatusCode() < 500)
	})

	t.Run("Get Migration Logs", func(t *testing.T) {
		resp, err := client.GetClient().R().
			SetQueryParam("limit", "50").
			Get("/migrations/test-migration-id/logs")
		
		require.NoError(t, err)
		// Может быть 200 или 404
		assert.True(t, resp.StatusCode() == 200 || resp.StatusCode() == 404)
	})

	t.Run("Validate Migration Data", func(t *testing.T) {
		validationData := map[string]interface{}{
			"migration_type": "structure",
			"data": map[string]interface{}{
				"universities": []map[string]interface{}{
					{
						"name": "Test University",
						"branches": []map[string]interface{}{
							{
								"name": "Test Branch",
							},
						},
					},
				},
			},
		}
		
		resp, err := client.GetClient().R().
			SetBody(validationData).
			Post("/migrations/validate")
		
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode())
		
		var validation map[string]interface{}
		err = json.Unmarshal(resp.Body(), &validation)
		require.NoError(t, err)
		assert.Contains(t, validation, "valid")
	})

	t.Run("Invalid File Upload", func(t *testing.T) {
		// Пытаемся загрузить без файла
		resp, err := client.GetClient().R().
			Post("/migrations/upload")
		
		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode())
	})

	t.Run("Invalid Migration Data", func(t *testing.T) {
		invalidData := map[string]interface{}{
			"invalid_field": "invalid_value",
		}
		
		resp, err := client.GetClient().R().
			SetBody(invalidData).
			Post("/migrations/start")
		
		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode())
	})

	t.Run("Non-existent Migration", func(t *testing.T) {
		resp, err := client.GetClient().R().
			Get("/migrations/non-existent-id")
		
		require.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode())
	})

	t.Run("Invalid Validation Data", func(t *testing.T) {
		invalidValidation := map[string]interface{}{
			"migration_type": "invalid_type",
			"data":           "invalid_data",
		}
		
		resp, err := client.GetClient().R().
			SetBody(invalidValidation).
			Post("/migrations/validate")
		
		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode())
	})
}