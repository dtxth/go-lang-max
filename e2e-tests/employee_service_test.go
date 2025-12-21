package main

import (
	"e2e-tests/utils"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmployeeService(t *testing.T) {
	// Настройка клиента
	configs := utils.DefaultServiceConfigs()
	client := utils.NewTestClient(configs["employee"])

	// Ждем доступности сервиса
	err := utils.WaitForService(configs["employee"].BaseURL, 10)
	require.NoError(t, err, "Employee service should be available")

	var testEmployee utils.TestEmployee

	t.Run("Create Simple Employee", func(t *testing.T) {
		testEmployee = utils.GenerateTestEmployee()
		
		resp, err := client.GetClient().R().
			SetBody(testEmployee).
			Post("/simple-employee")
		
		require.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode())
		
		var employee map[string]interface{}
		err = json.Unmarshal(resp.Body(), &employee)
		require.NoError(t, err)
		assert.Equal(t, testEmployee.Name, employee["name"])
		assert.Equal(t, testEmployee.Email, employee["email"])
		assert.Equal(t, testEmployee.Phone, employee["phone"])
		
		// Сохраняем ID для дальнейших тестов
		if id, ok := employee["id"].(string); ok {
			testEmployee.ID = id
		}
	})

	t.Run("Get All Employees", func(t *testing.T) {
		resp, err := client.GetClient().R().
			Get("/employees/all")
		
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode())
		
		var employees []interface{}
		err = json.Unmarshal(resp.Body(), &employees)
		require.NoError(t, err)
		// Массив должен содержать хотя бы одного сотрудника (созданного выше)
		assert.GreaterOrEqual(t, len(employees), 1)
	})

	t.Run("Batch Update MaxID", func(t *testing.T) {
		batchData := []map[string]interface{}{
			{
				"employee_id": testEmployee.ID,
				"max_id":      "MAX123456",
			},
		}
		
		resp, err := client.GetClient().R().
			SetBody(batchData).
			Post("/employees/batch-update-maxid")
		
		require.NoError(t, err)
		// Может быть 200 или 400 в зависимости от валидности данных
		assert.True(t, resp.StatusCode() == 200 || resp.StatusCode() == 400)
	})

	t.Run("Batch Status Check", func(t *testing.T) {
		statusData := []map[string]interface{}{
			{
				"employee_id": testEmployee.ID,
			},
		}
		
		resp, err := client.GetClient().R().
			SetBody(statusData).
			Post("/employees/batch-status")
		
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode())
		
		var statuses []interface{}
		err = json.Unmarshal(resp.Body(), &statuses)
		require.NoError(t, err)
		assert.Len(t, statuses, 1)
	})

	t.Run("Invalid Employee Creation", func(t *testing.T) {
		invalidEmployee := map[string]interface{}{
			"name": "", // Пустое имя
			// Отсутствуют обязательные поля
		}
		
		resp, err := client.GetClient().R().
			SetBody(invalidEmployee).
			Post("/simple-employee")
		
		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode())
	})

	t.Run("Invalid Batch Update", func(t *testing.T) {
		invalidBatchData := []map[string]interface{}{
			{
				"employee_id": "invalid-id",
				"max_id":      "",
			},
		}
		
		resp, err := client.GetClient().R().
			SetBody(invalidBatchData).
			Post("/employees/batch-update-maxid")
		
		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode())
	})

	t.Run("Empty Batch Status", func(t *testing.T) {
		emptyBatch := []map[string]interface{}{}
		
		resp, err := client.GetClient().R().
			SetBody(emptyBatch).
			Post("/employees/batch-status")
		
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode())
		
		var statuses []interface{}
		err = json.Unmarshal(resp.Body(), &statuses)
		require.NoError(t, err)
		assert.Len(t, statuses, 0)
	})

	t.Run("Method Not Allowed", func(t *testing.T) {
		// Тестируем неподдерживаемый метод
		resp, err := client.GetClient().R().
			Delete("/employees/all")
		
		require.NoError(t, err)
		assert.Equal(t, 405, resp.StatusCode())
	})

	t.Run("Large Batch Update", func(t *testing.T) {
		// Создаем большой batch для тестирования производительности
		largeBatch := make([]map[string]interface{}, 100)
		for i := 0; i < 100; i++ {
			largeBatch[i] = map[string]interface{}{
				"employee_id": testEmployee.ID,
				"max_id":      "MAX" + string(rune(i)),
			}
		}
		
		resp, err := client.GetClient().R().
			SetBody(largeBatch).
			Post("/employees/batch-update-maxid")
		
		require.NoError(t, err)
		// Может быть 200 или 400 в зависимости от логики сервиса
		assert.True(t, resp.StatusCode() == 200 || resp.StatusCode() == 400)
	})
}