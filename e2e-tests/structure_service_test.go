package main

import (
	"e2e-tests/utils"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStructureService(t *testing.T) {
	// Настройка клиента
	configs := utils.DefaultServiceConfigs()
	client := utils.NewTestClient(configs["structure"])

	// Ждем доступности сервиса
	err := utils.WaitForService(configs["structure"].BaseURL, 10)
	require.NoError(t, err, "Structure service should be available")

	var testUniversity utils.TestUniversity
	var universityID int

	t.Run("Create University", func(t *testing.T) {
		testUniversity = utils.GenerateTestUniversity()
		
		resp, err := client.GetClient().R().
			SetBody(testUniversity).
			Post("/universities")
		
		require.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode())
		
		var university map[string]interface{}
		err = json.Unmarshal(resp.Body(), &university)
		require.NoError(t, err)
		assert.Equal(t, testUniversity.Name, university["name"])
		
		universityID = int(university["id"].(float64))
	})

	t.Run("Get All Universities", func(t *testing.T) {
		resp, err := client.GetClient().R().
			SetQueryParam("limit", "10").
			SetQueryParam("offset", "0").
			Get("/universities")
		
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode())
		
		var response map[string]interface{}
		err = json.Unmarshal(resp.Body(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "universities")
		assert.Contains(t, response, "total")
		assert.Contains(t, response, "limit")
		assert.Contains(t, response, "offset")
	})

	t.Run("Get University by ID", func(t *testing.T) {
		resp, err := client.GetClient().R().
			Get(fmt.Sprintf("/universities/%d", universityID))
		
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode())
		
		var university map[string]interface{}
		err = json.Unmarshal(resp.Body(), &university)
		require.NoError(t, err)
		assert.Equal(t, testUniversity.Name, university["name"])
		assert.Equal(t, float64(universityID), university["id"])
	})

	t.Run("Update University Name", func(t *testing.T) {
		newName := testUniversity.Name + " Updated"
		updateData := map[string]string{
			"name": newName,
		}
		
		resp, err := client.GetClient().R().
			SetBody(updateData).
			Put(fmt.Sprintf("/universities/%d/name", universityID))
		
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode())
		
		// Проверяем, что имя действительно обновилось
		resp, err = client.GetClient().R().
			Get(fmt.Sprintf("/universities/%d", universityID))
		
		require.NoError(t, err)
		var university map[string]interface{}
		err = json.Unmarshal(resp.Body(), &university)
		require.NoError(t, err)
		assert.Equal(t, newName, university["name"])
	})

	t.Run("Get University Structure", func(t *testing.T) {
		resp, err := client.GetClient().R().
			Get(fmt.Sprintf("/universities/%d/structure", universityID))
		
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode())
		
		var structure map[string]interface{}
		err = json.Unmarshal(resp.Body(), &structure)
		require.NoError(t, err)
		assert.Contains(t, structure, "id")
		assert.Contains(t, structure, "name")
		assert.Contains(t, structure, "type")
	})

	t.Run("Create Structure", func(t *testing.T) {
		structureData := map[string]interface{}{
			"university_id": universityID,
			"branches": []map[string]interface{}{
				{
					"name": "Test Branch",
					"faculties": []map[string]interface{}{
						{
							"name": "Test Faculty",
							"departments": []map[string]interface{}{
								{
									"name": "Test Department",
									"groups": []map[string]interface{}{
										{
											"name": "Test Group",
										},
									},
								},
							},
						},
					},
				},
			},
		}
		
		resp, err := client.GetClient().R().
			SetBody(structureData).
			Post("/structure")
		
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode())
		
		var response map[string]interface{}
		err = json.Unmarshal(resp.Body(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "message")
	})

	t.Run("Get Department Managers", func(t *testing.T) {
		resp, err := client.GetClient().R().
			Get("/departments/managers")
		
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode())
		
		var managers []interface{}
		err = json.Unmarshal(resp.Body(), &managers)
		require.NoError(t, err)
		// Массив может быть пустым, это нормально
	})

	t.Run("Assign Operator", func(t *testing.T) {
		operatorData := map[string]interface{}{
			"user_id":       "test-user-id",
			"department_id": 1,
		}
		
		resp, err := client.GetClient().R().
			SetBody(operatorData).
			Post("/departments/managers")
		
		require.NoError(t, err)
		// Может быть 201 или 400/404 если пользователь/департамент не существует
		assert.True(t, resp.StatusCode() == 201 || resp.StatusCode() == 400 || resp.StatusCode() == 404)
	})

	t.Run("Invalid University ID", func(t *testing.T) {
		resp, err := client.GetClient().R().
			Get("/universities/invalid")
		
		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode())
	})

	t.Run("Non-existent University", func(t *testing.T) {
		resp, err := client.GetClient().R().
			Get("/universities/99999")
		
		require.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode())
	})

	t.Run("Invalid Structure Data", func(t *testing.T) {
		invalidData := map[string]interface{}{
			"invalid_field": "invalid_value",
		}
		
		resp, err := client.GetClient().R().
			SetBody(invalidData).
			Post("/structure")
		
		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode())
	})

	t.Run("Update Non-existent University Name", func(t *testing.T) {
		updateData := map[string]string{
			"name": "New Name",
		}
		
		resp, err := client.GetClient().R().
			SetBody(updateData).
			Put("/universities/99999/name")
		
		require.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode())
	})
}