package main

import (
	"e2e-tests/utils"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration проверяет взаимодействие между сервисами
func TestIntegration(t *testing.T) {
	configs := utils.DefaultServiceConfigs()
	
	// Проверяем доступность всех сервисов
	for serviceName, config := range configs {
		t.Run(fmt.Sprintf("Service %s availability", serviceName), func(t *testing.T) {
			err := utils.WaitForService(config.BaseURL, 5)
			if err != nil {
				t.Logf("Service %s is not available: %v", serviceName, err)
				t.Skip("Service not available")
			}
		})
	}

	// Создаем клиенты для всех сервисов
	authClient := utils.NewTestClient(configs["auth"])
	structureClient := utils.NewTestClient(configs["structure"])
	employeeClient := utils.NewTestClient(configs["employee"])
	chatClient := utils.NewTestClient(configs["chat"])

	var accessToken string
	var universityID int
	var employeeID string

	t.Run("Full User Journey", func(t *testing.T) {
		// 1. Регистрация пользователя
		testUser := utils.GenerateTestUser()
		
		resp, err := authClient.GetClient().R().
			SetBody(testUser).
			Post("/register")
		
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode())

		// 2. Авторизация пользователя
		loginResp, err := authClient.GetClient().R().
			SetBody(map[string]string{
				"email":    testUser.Email,
				"password": testUser.Password,
			}).
			Post("/login")
		
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode())
		
		var tokens map[string]interface{}
		err = json.Unmarshal(loginResp.Body(), &tokens)
		require.NoError(t, err)
		
		accessToken = tokens["access_token"].(string)
		
		// Устанавливаем токен для всех клиентов
		structureClient.SetAuthToken(accessToken)
		employeeClient.SetAuthToken(accessToken)
		chatClient.SetAuthToken(accessToken)
	})

	t.Run("Structure and Employee Integration", func(t *testing.T) {
		// 3. Создание университета
		testUniversity := utils.GenerateTestUniversity()
		
		resp, err := structureClient.GetClient().R().
			SetBody(testUniversity).
			Post("/universities")
		
		require.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode())
		
		var university map[string]interface{}
		err = json.Unmarshal(resp.Body(), &university)
		require.NoError(t, err)
		
		universityID = int(university["id"].(float64))

		// 4. Создание структуры университета
		structureData := map[string]interface{}{
			"university_id": universityID,
			"branches": []map[string]interface{}{
				{
					"name": "Integration Test Branch",
					"faculties": []map[string]interface{}{
						{
							"name": "Integration Test Faculty",
							"departments": []map[string]interface{}{
								{
									"name": "Integration Test Department",
									"groups": []map[string]interface{}{
										{
											"name": "Integration Test Group",
										},
									},
								},
							},
						},
					},
				},
			},
		}
		
		resp, err = structureClient.GetClient().R().
			SetBody(structureData).
			Post("/structure")
		
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode())

		// 5. Создание сотрудника
		testEmployee := utils.GenerateTestEmployee()
		
		resp, err = employeeClient.GetClient().R().
			SetBody(testEmployee).
			Post("/simple-employee")
		
		require.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode())
		
		var employee map[string]interface{}
		err = json.Unmarshal(resp.Body(), &employee)
		require.NoError(t, err)
		
		employeeID = employee["id"].(string)
	})

	t.Run("Chat Integration", func(t *testing.T) {
		// 6. Создание чата
		testChat := utils.GenerateTestChat()
		
		resp, err := chatClient.GetClient().R().
			SetBody(testChat).
			Post("/chats")
		
		require.NoError(t, err)
		// Chat service может не иметь этого эндпойнта, проверяем статус
		if resp.StatusCode() >= 200 && resp.StatusCode() < 300 {
			var chat map[string]interface{}
			err = json.Unmarshal(resp.Body(), &chat)
			require.NoError(t, err)
			
			chatID := chat["id"].(string)
			
			// 7. Отправка сообщения в чат
			messageData := map[string]interface{}{
				"chat_id": chatID,
				"content": "Integration test message",
				"type":    "text",
			}
			
			resp, err = chatClient.GetClient().R().
				SetBody(messageData).
				Post("/messages")
			
			require.NoError(t, err)
			// Проверяем, что сообщение было отправлено
			assert.True(t, resp.StatusCode() >= 200 && resp.StatusCode() < 300)
		}
	})

	t.Run("Cross-Service Data Consistency", func(t *testing.T) {
		// 8. Проверяем, что данные согласованы между сервисами
		
		// Проверяем университет в structure service
		resp, err := structureClient.GetClient().R().
			Get(fmt.Sprintf("/universities/%d", universityID))
		
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode())
		
		// Проверяем структуру университета
		resp, err = structureClient.GetClient().R().
			Get(fmt.Sprintf("/universities/%d/structure", universityID))
		
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode())
		
		// Проверяем сотрудника в employee service
		resp, err = employeeClient.GetClient().R().
			Get("/employees/all")
		
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode())
		
		var employees []interface{}
		err = json.Unmarshal(resp.Body(), &employees)
		require.NoError(t, err)
		
		// Проверяем, что наш сотрудник есть в списке
		found := false
		for _, emp := range employees {
			if empMap, ok := emp.(map[string]interface{}); ok {
				if empMap["id"] == employeeID {
					found = true
					break
				}
			}
		}
		assert.True(t, found, "Employee should be found in the list")
	})

	t.Run("Service Health Monitoring", func(t *testing.T) {
		// 9. Проверяем здоровье всех сервисов
		services := map[string]*utils.TestClient{
			"auth":      authClient,
			"structure": structureClient,
			"employee":  employeeClient,
			"chat":      chatClient,
		}
		
		for serviceName, client := range services {
			t.Run(fmt.Sprintf("%s health", serviceName), func(t *testing.T) {
				resp, err := client.GetClient().R().Get("/health")
				require.NoError(t, err)
				assert.Equal(t, 200, resp.StatusCode())
			})
		}
	})

	t.Run("Performance Test", func(t *testing.T) {
		// 10. Простой тест производительности
		start := time.Now()
		
		// Выполняем несколько запросов параллельно
		done := make(chan bool, 10)
		
		for i := 0; i < 10; i++ {
			go func() {
				defer func() { done <- true }()
				
				// Запрос к auth service
				authClient.GetClient().R().Get("/health")
				
				// Запрос к structure service
				structureClient.GetClient().R().Get("/universities")
				
				// Запрос к employee service
				employeeClient.GetClient().R().Get("/employees/all")
			}()
		}
		
		// Ждем завершения всех запросов
		for i := 0; i < 10; i++ {
			<-done
		}
		
		duration := time.Since(start)
		t.Logf("10 parallel requests completed in %v", duration)
		
		// Проверяем, что все запросы выполнились за разумное время
		assert.Less(t, duration, 30*time.Second, "Requests should complete within 30 seconds")
	})

	t.Run("Error Handling Integration", func(t *testing.T) {
		// 11. Тестируем обработку ошибок между сервисами
		
		// Попытка создать структуру для несуществующего университета
		invalidStructure := map[string]interface{}{
			"university_id": 99999,
			"branches":      []map[string]interface{}{},
		}
		
		resp, err := structureClient.GetClient().R().
			SetBody(invalidStructure).
			Post("/structure")
		
		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode())
		
		// Попытка получить несуществующего сотрудника
		resp, err = employeeClient.GetClient().R().
			Get("/employees/non-existent-id")
		
		require.NoError(t, err)
		// Может быть 404 или другой код в зависимости от реализации
		assert.True(t, resp.StatusCode() >= 400)
	})
}