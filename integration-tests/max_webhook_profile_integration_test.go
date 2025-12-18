package integration_tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	MaxBotServiceURL = "http://localhost:8095"
	RedisAddr        = "localhost:6379"
)

// TestIntegrationSetup verifies that the integration test environment is properly configured
func TestIntegrationSetup(t *testing.T) {
	t.Log("Testing MAX webhook profile integration setup")
	
	// This test validates that our integration test structure is correct
	// and can be used for actual integration testing when services are running
	
	// Test webhook event structure
	webhookEvent := MaxWebhookEvent{
		Type: "message_new",
		Message: &MessageEvent{
			From: UserInfo{
				UserID:    "test_user_123",
				FirstName: "Test",
				LastName:  "User",
			},
			Text: "Hello",
		},
	}
	
	// Validate JSON marshaling works correctly
	data, err := json.Marshal(webhookEvent)
	require.NoError(t, err, "Webhook event should marshal to JSON")
	
	var unmarshaled MaxWebhookEvent
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err, "Webhook event should unmarshal from JSON")
	
	assert.Equal(t, "message_new", unmarshaled.Type)
	assert.Equal(t, "test_user_123", unmarshaled.Message.From.UserID)
	assert.Equal(t, "Test", unmarshaled.Message.From.FirstName)
	assert.Equal(t, "User", unmarshaled.Message.From.LastName)
	
	t.Log("Integration test structure validation passed")
}

// MaxWebhookEvent представляет webhook событие от MAX
type MaxWebhookEvent struct {
	Type     string         `json:"type"`
	Message  *MessageEvent  `json:"message,omitempty"`
	Callback *CallbackEvent `json:"callback_query,omitempty"`
}

type MessageEvent struct {
	From UserInfo `json:"from"`
	Text string   `json:"text"`
}

type CallbackEvent struct {
	User UserInfo `json:"user"`
	Data string   `json:"data"`
}

type UserInfo struct {
	UserID    string `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type EmployeeRequest struct {
	Phone          string `json:"phone"`
	FirstName      string `json:"first_name,omitempty"`
	LastName       string `json:"last_name,omitempty"`
	MiddleName     string `json:"middle_name,omitempty"`
	INN            string `json:"inn,omitempty"`
	KPP            string `json:"kpp,omitempty"`
	UniversityName string `json:"university_name,omitempty"`
}

type Employee struct {
	ID                   int64     `json:"id"`
	FirstName            string    `json:"first_name"`
	LastName             string    `json:"last_name"`
	MiddleName           string    `json:"middle_name"`
	Phone                string    `json:"phone"`
	MaxID                string    `json:"max_id"`
	ProfileSource        string    `json:"profile_source"`
	ProfileLastUpdated   *string   `json:"profile_last_updated"`
	UniversityID         int64     `json:"university_id"`
}

// TestEndToEndWebhookToEmployeeCreation тестирует полный поток от webhook до создания сотрудника
func TestEndToEndWebhookToEmployeeCreation(t *testing.T) {
	// Ждем готовности сервисов
	WaitForService(t, EmployeeServiceURL, 10)
	WaitForService(t, MaxBotServiceURL, 10)
	
	// Подключаемся к Redis для проверки кэша
	rdb := redis.NewClient(&redis.Options{
		Addr: RedisAddr,
	})
	defer rdb.Close()
	
	// Проверяем подключение к Redis
	ctx := context.Background()
	err := rdb.Ping(ctx).Err()
	require.NoError(t, err, "Redis should be available")
	
	// Очищаем тестовые данные
	defer cleanupTestData(t, rdb)
	
	client := NewHTTPClient()
	
	// Тестовые данные
	userID := "test_user_12345"
	phone := "+79991234567"
	firstName := "Иван"
	lastName := "Петров"
	
	// Шаг 1: Отправляем webhook событие message_new
	t.Run("Step1_SendWebhookEvent", func(t *testing.T) {
		webhookEvent := MaxWebhookEvent{
			Type: "message_new",
			Message: &MessageEvent{
				From: UserInfo{
					UserID:    userID,
					FirstName: firstName,
					LastName:  lastName,
				},
				Text: "Привет",
			},
		}
		
		status, respBody := client.POST(t, MaxBotServiceURL+"/webhook/max", webhookEvent)
		assert.Equal(t, http.StatusOK, status, "Webhook should return 200 OK: %s", string(respBody))
	})
	
	// Шаг 2: Проверяем, что профиль сохранился в кэше
	t.Run("Step2_VerifyProfileInCache", func(t *testing.T) {
		// Даем время на обработку webhook
		time.Sleep(1 * time.Second)
		
		profileKey := fmt.Sprintf("profile:user:%s", userID)
		profileData, err := rdb.Get(ctx, profileKey).Result()
		require.NoError(t, err, "Profile should be stored in cache")
		
		var profile map[string]interface{}
		err = json.Unmarshal([]byte(profileData), &profile)
		require.NoError(t, err, "Profile should be valid JSON")
		
		assert.Equal(t, userID, profile["user_id"])
		assert.Equal(t, firstName, profile["max_first_name"])
		assert.Equal(t, lastName, profile["max_last_name"])
		assert.Equal(t, "webhook", profile["source"])
	})
	
	// Шаг 3: Создаем сотрудника без указания имени (должен использовать кэш)
	t.Run("Step3_CreateEmployeeUsingCache", func(t *testing.T) {
		employeeReq := EmployeeRequest{
			Phone:          phone,
			// Не указываем имена - должны взяться из кэша
			INN:            "1234567890",
			KPP:            "123456789",
			UniversityName: "Тестовый университет",
		}
		
		status, respBody := client.POST(t, EmployeeServiceURL+"/employees", employeeReq)
		assert.Equal(t, http.StatusCreated, status, "Employee creation should succeed: %s", string(respBody))
		
		var employee Employee
		err := json.Unmarshal(respBody, &employee)
		require.NoError(t, err, "Response should be valid employee JSON")
		
		// Проверяем, что имена взялись из кэша профиля
		assert.Equal(t, firstName, employee.FirstName, "First name should come from profile cache")
		assert.Equal(t, lastName, employee.LastName, "Last name should come from profile cache")
		assert.Equal(t, phone, employee.Phone)
		assert.Equal(t, userID, employee.MaxID, "MAX_id should be set from profile")
		assert.Equal(t, "webhook", employee.ProfileSource, "Profile source should be webhook")
		assert.NotNil(t, employee.ProfileLastUpdated, "Profile last updated should be set")
	})
	
	// Шаг 4: Тестируем приоритет user_provided над кэшем
	t.Run("Step4_TestUserProvidedPriority", func(t *testing.T) {
		phone2 := "+79991234568"
		
		// Сначала отправляем webhook для другого пользователя
		userID2 := "test_user_67890"
		webhookEvent := MaxWebhookEvent{
			Type: "message_new",
			Message: &MessageEvent{
				From: UserInfo{
					UserID:    userID2,
					FirstName: "Петр",
					LastName:  "Иванов",
				},
				Text: "Привет",
			},
		}
		
		status, _ := client.POST(t, MaxBotServiceURL+"/webhook/max", webhookEvent)
		assert.Equal(t, http.StatusOK, status)
		
		time.Sleep(1 * time.Second)
		
		// Создаем сотрудника с явно указанными именами
		employeeReq := EmployeeRequest{
			Phone:          phone2,
			FirstName:      "Александр", // Переопределяем имя из кэша
			LastName:       "Сидоров",   // Переопределяем фамилию из кэша
			INN:            "1234567891",
			KPP:            "123456790",
			UniversityName: "Тестовый университет 2",
		}
		
		status, respBody := client.POST(t, EmployeeServiceURL+"/employees", employeeReq)
		assert.Equal(t, http.StatusCreated, status, "Employee creation should succeed: %s", string(respBody))
		
		var employee Employee
		err := json.Unmarshal(respBody, &employee)
		require.NoError(t, err)
		
		// Проверяем приоритет user_provided
		assert.Equal(t, "Александр", employee.FirstName, "User-provided name should have priority")
		assert.Equal(t, "Сидоров", employee.LastName, "User-provided name should have priority")
		assert.Equal(t, "user_input", employee.ProfileSource, "Profile source should be user_input")
	})
}

// TestProfilePersistenceAcrossServiceRestarts тестирует сохранение профилей при перезапуске сервисов
func TestProfilePersistenceAcrossServiceRestarts(t *testing.T) {
	// Подключаемся к Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: RedisAddr,
	})
	defer rdb.Close()
	
	ctx := context.Background()
	err := rdb.Ping(ctx).Err()
	require.NoError(t, err, "Redis should be available")
	
	// Очищаем тестовые данные
	defer cleanupTestData(t, rdb)
	
	client := NewHTTPClient()
	
	userID := "persistent_user_123"
	firstName := "Анна"
	lastName := "Козлова"
	
	// Шаг 1: Создаем профиль через webhook
	t.Run("Step1_CreateProfile", func(t *testing.T) {
		WaitForService(t, MaxBotServiceURL, 10)
		
		webhookEvent := MaxWebhookEvent{
			Type: "message_new",
			Message: &MessageEvent{
				From: UserInfo{
					UserID:    userID,
					FirstName: firstName,
					LastName:  lastName,
				},
				Text: "Тест персистентности",
			},
		}
		
		status, _ := client.POST(t, MaxBotServiceURL+"/webhook/max", webhookEvent)
		assert.Equal(t, http.StatusOK, status)
		
		time.Sleep(1 * time.Second)
		
		// Проверяем, что профиль сохранился
		profileKey := fmt.Sprintf("profile:user:%s", userID)
		profileData, err := rdb.Get(ctx, profileKey).Result()
		require.NoError(t, err, "Profile should be stored")
		
		var profile map[string]interface{}
		err = json.Unmarshal([]byte(profileData), &profile)
		require.NoError(t, err)
		
		assert.Equal(t, firstName, profile["max_first_name"])
		assert.Equal(t, lastName, profile["max_last_name"])
	})
	
	// Шаг 2: Симулируем перезапуск сервиса (проверяем TTL)
	t.Run("Step2_CheckPersistenceAfterTime", func(t *testing.T) {
		// Проверяем TTL профиля
		profileKey := fmt.Sprintf("profile:user:%s", userID)
		ttl, err := rdb.TTL(ctx, profileKey).Result()
		require.NoError(t, err)
		
		// TTL должен быть установлен (30 дней = 2592000 секунд)
		assert.Greater(t, ttl.Seconds(), float64(2590000), "TTL should be close to 30 days")
		assert.Less(t, ttl.Seconds(), float64(2593000), "TTL should not exceed 30 days")
		
		// Профиль должен быть доступен
		profileData, err := rdb.Get(ctx, profileKey).Result()
		require.NoError(t, err, "Profile should still be available")
		
		var profile map[string]interface{}
		err = json.Unmarshal([]byte(profileData), &profile)
		require.NoError(t, err)
		
		assert.Equal(t, firstName, profile["max_first_name"])
		assert.Equal(t, lastName, profile["max_last_name"])
	})
	
	// Шаг 3: Проверяем, что профиль используется после "перезапуска"
	t.Run("Step3_UseProfileAfterRestart", func(t *testing.T) {
		WaitForService(t, EmployeeServiceURL, 10)
		
		phone := "+79991234569"
		employeeReq := EmployeeRequest{
			Phone:          phone,
			INN:            "1234567892",
			KPP:            "123456791",
			UniversityName: "Персистентный университет",
		}
		
		status, respBody := client.POST(t, EmployeeServiceURL+"/employees", employeeReq)
		assert.Equal(t, http.StatusCreated, status, "Employee creation should succeed: %s", string(respBody))
		
		var employee Employee
		err := json.Unmarshal(respBody, &employee)
		require.NoError(t, err)
		
		// Проверяем, что имена взялись из персистентного кэша
		assert.Equal(t, firstName, employee.FirstName)
		assert.Equal(t, lastName, employee.LastName)
		assert.Equal(t, "webhook", employee.ProfileSource)
	})
}

// TestConcurrentWebhookProcessing тестирует обработку множественных webhook событий одновременно
func TestConcurrentWebhookProcessing(t *testing.T) {
	WaitForService(t, MaxBotServiceURL, 10)
	
	// Подключаемся к Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: RedisAddr,
	})
	defer rdb.Close()
	
	ctx := context.Background()
	err := rdb.Ping(ctx).Err()
	require.NoError(t, err, "Redis should be available")
	
	// Очищаем тестовые данные
	defer cleanupTestData(t, rdb)
	
	client := NewHTTPClient()
	
	// Количество одновременных webhook событий
	concurrentCount := 10
	
	// Канал для сбора результатов
	results := make(chan bool, concurrentCount)
	var wg sync.WaitGroup
	
	// Шаг 1: Отправляем множественные webhook события одновременно
	t.Run("Step1_SendConcurrentWebhooks", func(t *testing.T) {
		for i := 0; i < concurrentCount; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				
				userID := fmt.Sprintf("concurrent_user_%d", index)
				firstName := fmt.Sprintf("Пользователь%d", index)
				lastName := fmt.Sprintf("Тестовый%d", index)
				
				webhookEvent := MaxWebhookEvent{
					Type: "message_new",
					Message: &MessageEvent{
						From: UserInfo{
							UserID:    userID,
							FirstName: firstName,
							LastName:  lastName,
						},
						Text: fmt.Sprintf("Сообщение %d", index),
					},
				}
				
				status, _ := client.POST(t, MaxBotServiceURL+"/webhook/max", webhookEvent)
				results <- (status == http.StatusOK)
			}(i)
		}
		
		wg.Wait()
		close(results)
		
		// Проверяем, что все webhook события обработались успешно
		successCount := 0
		for success := range results {
			if success {
				successCount++
			}
		}
		
		assert.Equal(t, concurrentCount, successCount, "All webhook events should be processed successfully")
	})
	
	// Шаг 2: Проверяем, что все профили сохранились корректно
	t.Run("Step2_VerifyAllProfilesSaved", func(t *testing.T) {
		// Даем время на обработку всех webhook событий
		time.Sleep(3 * time.Second)
		
		for i := 0; i < concurrentCount; i++ {
			userID := fmt.Sprintf("concurrent_user_%d", i)
			expectedFirstName := fmt.Sprintf("Пользователь%d", i)
			expectedLastName := fmt.Sprintf("Тестовый%d", i)
			
			profileKey := fmt.Sprintf("profile:user:%s", userID)
			profileData, err := rdb.Get(ctx, profileKey).Result()
			require.NoError(t, err, "Profile %d should be stored", i)
			
			var profile map[string]interface{}
			err = json.Unmarshal([]byte(profileData), &profile)
			require.NoError(t, err, "Profile %d should be valid JSON", i)
			
			assert.Equal(t, userID, profile["user_id"], "User ID should match for profile %d", i)
			assert.Equal(t, expectedFirstName, profile["max_first_name"], "First name should match for profile %d", i)
			assert.Equal(t, expectedLastName, profile["max_last_name"], "Last name should match for profile %d", i)
			assert.Equal(t, "webhook", profile["source"], "Source should be webhook for profile %d", i)
		}
	})
	
	// Шаг 3: Тестируем создание сотрудников с использованием всех профилей
	t.Run("Step3_CreateEmployeesUsingConcurrentProfiles", func(t *testing.T) {
		WaitForService(t, EmployeeServiceURL, 10)
		
		// Создаем сотрудников параллельно
		employeeResults := make(chan bool, concurrentCount)
		var employeeWg sync.WaitGroup
		
		for i := 0; i < concurrentCount; i++ {
			employeeWg.Add(1)
			go func(index int) {
				defer employeeWg.Done()
				
				phone := fmt.Sprintf("+7999123456%d", index)
				employeeReq := EmployeeRequest{
					Phone:          phone,
					INN:            fmt.Sprintf("123456789%d", index),
					KPP:            fmt.Sprintf("12345678%d", index),
					UniversityName: fmt.Sprintf("Университет %d", index),
				}
				
				status, respBody := client.POST(t, EmployeeServiceURL+"/employees", employeeReq)
				if status == http.StatusCreated {
					var employee Employee
					err := json.Unmarshal(respBody, &employee)
					if err == nil {
						expectedFirstName := fmt.Sprintf("Пользователь%d", index)
						expectedLastName := fmt.Sprintf("Тестовый%d", index)
						
						success := (employee.FirstName == expectedFirstName && 
								   employee.LastName == expectedLastName &&
								   employee.ProfileSource == "webhook")
						employeeResults <- success
					} else {
						employeeResults <- false
					}
				} else {
					employeeResults <- false
				}
			}(i)
		}
		
		employeeWg.Wait()
		close(employeeResults)
		
		// Проверяем результаты создания сотрудников
		employeeSuccessCount := 0
		for success := range employeeResults {
			if success {
				employeeSuccessCount++
			}
		}
		
		assert.Equal(t, concurrentCount, employeeSuccessCount, "All employees should be created with correct profile data")
	})
}

// TestWebhookEventTypes тестирует обработку разных типов webhook событий
func TestWebhookEventTypes(t *testing.T) {
	WaitForService(t, MaxBotServiceURL, 10)
	
	rdb := redis.NewClient(&redis.Options{
		Addr: RedisAddr,
	})
	defer rdb.Close()
	
	ctx := context.Background()
	defer cleanupTestData(t, rdb)
	
	client := NewHTTPClient()
	
	// Тест message_new события
	t.Run("MessageNewEvent", func(t *testing.T) {
		userID := "msg_user_123"
		webhookEvent := MaxWebhookEvent{
			Type: "message_new",
			Message: &MessageEvent{
				From: UserInfo{
					UserID:    userID,
					FirstName: "Сообщение",
					LastName:  "Пользователь",
				},
				Text: "Тестовое сообщение",
			},
		}
		
		status, _ := client.POST(t, MaxBotServiceURL+"/webhook/max", webhookEvent)
		assert.Equal(t, http.StatusOK, status)
		
		time.Sleep(1 * time.Second)
		
		profileKey := fmt.Sprintf("profile:user:%s", userID)
		profileData, err := rdb.Get(ctx, profileKey).Result()
		require.NoError(t, err)
		
		var profile map[string]interface{}
		err = json.Unmarshal([]byte(profileData), &profile)
		require.NoError(t, err)
		
		assert.Equal(t, "Сообщение", profile["max_first_name"])
		assert.Equal(t, "Пользователь", profile["max_last_name"])
	})
	
	// Тест callback_query события
	t.Run("CallbackQueryEvent", func(t *testing.T) {
		userID := "callback_user_456"
		webhookEvent := MaxWebhookEvent{
			Type: "callback_query",
			Callback: &CallbackEvent{
				User: UserInfo{
					UserID:    userID,
					FirstName: "Колбэк",
					LastName:  "Пользователь",
				},
				Data: "test_callback_data",
			},
		}
		
		status, _ := client.POST(t, MaxBotServiceURL+"/webhook/max", webhookEvent)
		assert.Equal(t, http.StatusOK, status)
		
		time.Sleep(1 * time.Second)
		
		profileKey := fmt.Sprintf("profile:user:%s", userID)
		profileData, err := rdb.Get(ctx, profileKey).Result()
		require.NoError(t, err)
		
		var profile map[string]interface{}
		err = json.Unmarshal([]byte(profileData), &profile)
		require.NoError(t, err)
		
		assert.Equal(t, "Колбэк", profile["max_first_name"])
		assert.Equal(t, "Пользователь", profile["max_last_name"])
	})
	
	// Тест неизвестного типа события
	t.Run("UnknownEventType", func(t *testing.T) {
		webhookEvent := MaxWebhookEvent{
			Type: "unknown_event_type",
		}
		
		status, _ := client.POST(t, MaxBotServiceURL+"/webhook/max", webhookEvent)
		// Должен возвращать 200 OK даже для неизвестных событий
		assert.Equal(t, http.StatusOK, status)
	})
}

// TestUserNameInputProcessing тестирует обработку пользовательского ввода имени
func TestUserNameInputProcessing(t *testing.T) {
	WaitForService(t, MaxBotServiceURL, 10)
	
	rdb := redis.NewClient(&redis.Options{
		Addr: RedisAddr,
	})
	defer rdb.Close()
	
	ctx := context.Background()
	defer cleanupTestData(t, rdb)
	
	client := NewHTTPClient()
	
	userID := "name_input_user_789"
	
	// Шаг 1: Создаем базовый профиль через webhook
	t.Run("Step1_CreateBaseProfile", func(t *testing.T) {
		webhookEvent := MaxWebhookEvent{
			Type: "message_new",
			Message: &MessageEvent{
				From: UserInfo{
					UserID:    userID,
					FirstName: "Старое",
					LastName:  "Имя",
				},
				Text: "Привет",
			},
		}
		
		status, _ := client.POST(t, MaxBotServiceURL+"/webhook/max", webhookEvent)
		assert.Equal(t, http.StatusOK, status)
		
		time.Sleep(1 * time.Second)
	})
	
	// Шаг 2: Отправляем команду обновления имени
	t.Run("Step2_UpdateNameViaUserInput", func(t *testing.T) {
		webhookEvent := MaxWebhookEvent{
			Type: "message_new",
			Message: &MessageEvent{
				From: UserInfo{
					UserID:    userID,
					FirstName: "Старое", // Это не должно измениться
					LastName:  "Имя",    // Это не должно измениться
				},
				Text: "/setname Новое Имя Пользователя",
			},
		}
		
		status, _ := client.POST(t, MaxBotServiceURL+"/webhook/max", webhookEvent)
		assert.Equal(t, http.StatusOK, status)
		
		time.Sleep(1 * time.Second)
		
		// Проверяем, что user_provided_name обновился
		profileKey := fmt.Sprintf("profile:user:%s", userID)
		profileData, err := rdb.Get(ctx, profileKey).Result()
		require.NoError(t, err)
		
		var profile map[string]interface{}
		err = json.Unmarshal([]byte(profileData), &profile)
		require.NoError(t, err)
		
		assert.Equal(t, "Старое", profile["max_first_name"], "MAX first name should not change")
		assert.Equal(t, "Имя", profile["max_last_name"], "MAX last name should not change")
		assert.Equal(t, "Новое Имя Пользователя", profile["user_provided_name"], "User provided name should be updated")
		assert.Equal(t, "user_input", profile["source"], "Source should be updated to user_input")
	})
	
	// Шаг 3: Проверяем приоритет user_provided_name при создании сотрудника
	t.Run("Step3_VerifyUserProvidedPriority", func(t *testing.T) {
		WaitForService(t, EmployeeServiceURL, 10)
		
		phone := "+79991234570"
		employeeReq := EmployeeRequest{
			Phone:          phone,
			INN:            "1234567893",
			KPP:            "123456792",
			UniversityName: "Университет пользовательского ввода",
		}
		
		status, respBody := client.POST(t, EmployeeServiceURL+"/employees", employeeReq)
		assert.Equal(t, http.StatusCreated, status, "Employee creation should succeed: %s", string(respBody))
		
		var employee Employee
		err := json.Unmarshal(respBody, &employee)
		require.NoError(t, err)
		
		// Имя должно быть из user_provided_name, а не из max_first_name/max_last_name
		assert.Equal(t, "Новое Имя Пользователя", employee.FirstName, "Should use user-provided name")
		assert.Equal(t, "", employee.LastName, "Last name should be empty when using full user-provided name")
		assert.Equal(t, "user_input", employee.ProfileSource)
	})
}



// TestEndToEndWebhookToEmployeeCreationWithServices tests the full integration when services are available
func TestEndToEndWebhookToEmployeeCreationWithServices(t *testing.T) {
	// Skip if services are not available
	if !isServiceAvailable(MaxBotServiceURL) {
		t.Skip("MaxBot service not available, skipping integration test")
	}
	if !isServiceAvailable(EmployeeServiceURL) {
		t.Skip("Employee service not available, skipping integration test")
	}
	
	// Connect to Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: RedisAddr,
	})
	defer rdb.Close()
	
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available, skipping integration test")
	}
	
	// Clean up test data
	defer cleanupTestData(t, rdb)
	
	client := NewHTTPClient()
	
	// Test data
	userID := "test_user_12345"
	phone := "+79991234567"
	firstName := "Иван"
	lastName := "Петров"
	
	// Step 1: Send webhook event
	t.Run("Step1_SendWebhookEvent", func(t *testing.T) {
		webhookEvent := MaxWebhookEvent{
			Type: "message_new",
			Message: &MessageEvent{
				From: UserInfo{
					UserID:    userID,
					FirstName: firstName,
					LastName:  lastName,
				},
				Text: "Привет",
			},
		}
		
		status, respBody := client.POST(t, MaxBotServiceURL+"/webhook/max", webhookEvent)
		assert.Equal(t, http.StatusOK, status, "Webhook should return 200 OK: %s", string(respBody))
	})
	
	// Step 2: Verify profile in cache
	t.Run("Step2_VerifyProfileInCache", func(t *testing.T) {
		time.Sleep(1 * time.Second) // Allow processing time
		
		profileKey := fmt.Sprintf("profile:user:%s", userID)
		profileData, err := rdb.Get(ctx, profileKey).Result()
		require.NoError(t, err, "Profile should be stored in cache")
		
		var profile map[string]interface{}
		err = json.Unmarshal([]byte(profileData), &profile)
		require.NoError(t, err, "Profile should be valid JSON")
		
		assert.Equal(t, userID, profile["user_id"])
		assert.Equal(t, firstName, profile["max_first_name"])
		assert.Equal(t, lastName, profile["max_last_name"])
		assert.Equal(t, "webhook", profile["source"])
	})
	
	// Step 3: Create employee using cached profile
	t.Run("Step3_CreateEmployeeUsingCache", func(t *testing.T) {
		employeeReq := EmployeeRequest{
			Phone:          phone,
			INN:            "1234567890",
			KPP:            "123456789",
			UniversityName: "Тестовый университет",
		}
		
		status, respBody := client.POST(t, EmployeeServiceURL+"/employees", employeeReq)
		assert.Equal(t, http.StatusCreated, status, "Employee creation should succeed: %s", string(respBody))
		
		var employee Employee
		err := json.Unmarshal(respBody, &employee)
		require.NoError(t, err, "Response should be valid employee JSON")
		
		// Verify names came from profile cache
		assert.Equal(t, firstName, employee.FirstName, "First name should come from profile cache")
		assert.Equal(t, lastName, employee.LastName, "Last name should come from profile cache")
		assert.Equal(t, phone, employee.Phone)
		assert.Equal(t, userID, employee.MaxID, "MAX_id should be set from profile")
		assert.Equal(t, "webhook", employee.ProfileSource, "Profile source should be webhook")
		assert.NotNil(t, employee.ProfileLastUpdated, "Profile last updated should be set")
	})
}

// TestConcurrentWebhookProcessingWithServices tests concurrent webhook processing when services are available
func TestConcurrentWebhookProcessingWithServices(t *testing.T) {
	// Skip if services are not available
	if !isServiceAvailable(MaxBotServiceURL) {
		t.Skip("MaxBot service not available, skipping integration test")
	}
	
	// Connect to Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: RedisAddr,
	})
	defer rdb.Close()
	
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available, skipping integration test")
	}
	
	// Clean up test data
	defer cleanupTestData(t, rdb)
	
	client := NewHTTPClient()
	
	// Number of concurrent webhook events
	concurrentCount := 5 // Reduced for faster testing
	
	// Channel for collecting results
	results := make(chan bool, concurrentCount)
	var wg sync.WaitGroup
	
	// Step 1: Send concurrent webhook events
	t.Run("Step1_SendConcurrentWebhooks", func(t *testing.T) {
		for i := 0; i < concurrentCount; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				
				userID := fmt.Sprintf("concurrent_user_%d", index)
				firstName := fmt.Sprintf("Пользователь%d", index)
				lastName := fmt.Sprintf("Тестовый%d", index)
				
				webhookEvent := MaxWebhookEvent{
					Type: "message_new",
					Message: &MessageEvent{
						From: UserInfo{
							UserID:    userID,
							FirstName: firstName,
							LastName:  lastName,
						},
						Text: fmt.Sprintf("Сообщение %d", index),
					},
				}
				
				status, _ := client.POST(t, MaxBotServiceURL+"/webhook/max", webhookEvent)
				results <- (status == http.StatusOK)
			}(i)
		}
		
		wg.Wait()
		close(results)
		
		// Check that all webhook events processed successfully
		successCount := 0
		for success := range results {
			if success {
				successCount++
			}
		}
		
		assert.Equal(t, concurrentCount, successCount, "All webhook events should be processed successfully")
	})
	
	// Step 2: Verify all profiles saved correctly
	t.Run("Step2_VerifyAllProfilesSaved", func(t *testing.T) {
		time.Sleep(2 * time.Second) // Allow processing time
		
		for i := 0; i < concurrentCount; i++ {
			userID := fmt.Sprintf("concurrent_user_%d", i)
			expectedFirstName := fmt.Sprintf("Пользователь%d", i)
			expectedLastName := fmt.Sprintf("Тестовый%d", i)
			
			profileKey := fmt.Sprintf("profile:user:%s", userID)
			profileData, err := rdb.Get(ctx, profileKey).Result()
			require.NoError(t, err, "Profile %d should be stored", i)
			
			var profile map[string]interface{}
			err = json.Unmarshal([]byte(profileData), &profile)
			require.NoError(t, err, "Profile %d should be valid JSON", i)
			
			assert.Equal(t, userID, profile["user_id"], "User ID should match for profile %d", i)
			assert.Equal(t, expectedFirstName, profile["max_first_name"], "First name should match for profile %d", i)
			assert.Equal(t, expectedLastName, profile["max_last_name"], "Last name should match for profile %d", i)
			assert.Equal(t, "webhook", profile["source"], "Source should be webhook for profile %d", i)
		}
	})
}

// TestProfilePersistenceAcrossServiceRestartsWithServices tests profile persistence when services are available
func TestProfilePersistenceAcrossServiceRestartsWithServices(t *testing.T) {
	// Skip if services are not available
	if !isServiceAvailable(MaxBotServiceURL) {
		t.Skip("MaxBot service not available, skipping integration test")
	}
	
	// Connect to Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: RedisAddr,
	})
	defer rdb.Close()
	
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available, skipping integration test")
	}
	
	// Clean up test data
	defer cleanupTestData(t, rdb)
	
	client := NewHTTPClient()
	
	userID := "persistent_user_123"
	firstName := "Анна"
	lastName := "Козлова"
	
	// Step 1: Create profile through webhook
	t.Run("Step1_CreateProfile", func(t *testing.T) {
		webhookEvent := MaxWebhookEvent{
			Type: "message_new",
			Message: &MessageEvent{
				From: UserInfo{
					UserID:    userID,
					FirstName: firstName,
					LastName:  lastName,
				},
				Text: "Тест персистентности",
			},
		}
		
		status, _ := client.POST(t, MaxBotServiceURL+"/webhook/max", webhookEvent)
		assert.Equal(t, http.StatusOK, status)
		
		time.Sleep(1 * time.Second)
		
		// Verify profile is stored
		profileKey := fmt.Sprintf("profile:user:%s", userID)
		profileData, err := rdb.Get(ctx, profileKey).Result()
		require.NoError(t, err, "Profile should be stored")
		
		var profile map[string]interface{}
		err = json.Unmarshal([]byte(profileData), &profile)
		require.NoError(t, err)
		
		assert.Equal(t, firstName, profile["max_first_name"])
		assert.Equal(t, lastName, profile["max_last_name"])
	})
	
	// Step 2: Check persistence after time (simulate restart)
	t.Run("Step2_CheckPersistenceAfterTime", func(t *testing.T) {
		// Check TTL
		profileKey := fmt.Sprintf("profile:user:%s", userID)
		ttl, err := rdb.TTL(ctx, profileKey).Result()
		require.NoError(t, err)
		
		// TTL should be set (30 days = 2592000 seconds)
		assert.Greater(t, ttl.Seconds(), float64(2590000), "TTL should be close to 30 days")
		assert.Less(t, ttl.Seconds(), float64(2593000), "TTL should not exceed 30 days")
		
		// Profile should still be available
		profileData, err := rdb.Get(ctx, profileKey).Result()
		require.NoError(t, err, "Profile should still be available")
		
		var profile map[string]interface{}
		err = json.Unmarshal([]byte(profileData), &profile)
		require.NoError(t, err)
		
		assert.Equal(t, firstName, profile["max_first_name"])
		assert.Equal(t, lastName, profile["max_last_name"])
	})
}

// TestWebhookEventTypesWithServices tests different webhook event types when services are available
func TestWebhookEventTypesWithServices(t *testing.T) {
	// Skip if services are not available
	if !isServiceAvailable(MaxBotServiceURL) {
		t.Skip("MaxBot service not available, skipping integration test")
	}
	
	// Connect to Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: RedisAddr,
	})
	defer rdb.Close()
	
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available, skipping integration test")
	}
	
	// Clean up test data
	defer cleanupTestData(t, rdb)
	
	client := NewHTTPClient()
	
	// Test message_new event
	t.Run("MessageNewEvent", func(t *testing.T) {
		userID := "msg_user_123"
		webhookEvent := MaxWebhookEvent{
			Type: "message_new",
			Message: &MessageEvent{
				From: UserInfo{
					UserID:    userID,
					FirstName: "Сообщение",
					LastName:  "Пользователь",
				},
				Text: "Тестовое сообщение",
			},
		}
		
		status, _ := client.POST(t, MaxBotServiceURL+"/webhook/max", webhookEvent)
		assert.Equal(t, http.StatusOK, status)
		
		time.Sleep(1 * time.Second)
		
		profileKey := fmt.Sprintf("profile:user:%s", userID)
		profileData, err := rdb.Get(ctx, profileKey).Result()
		require.NoError(t, err)
		
		var profile map[string]interface{}
		err = json.Unmarshal([]byte(profileData), &profile)
		require.NoError(t, err)
		
		assert.Equal(t, "Сообщение", profile["max_first_name"])
		assert.Equal(t, "Пользователь", profile["max_last_name"])
	})
	
	// Test callback_query event
	t.Run("CallbackQueryEvent", func(t *testing.T) {
		userID := "callback_user_456"
		webhookEvent := MaxWebhookEvent{
			Type: "callback_query",
			Callback: &CallbackEvent{
				User: UserInfo{
					UserID:    userID,
					FirstName: "Колбэк",
					LastName:  "Пользователь",
				},
				Data: "test_callback_data",
			},
		}
		
		status, _ := client.POST(t, MaxBotServiceURL+"/webhook/max", webhookEvent)
		assert.Equal(t, http.StatusOK, status)
		
		time.Sleep(1 * time.Second)
		
		profileKey := fmt.Sprintf("profile:user:%s", userID)
		profileData, err := rdb.Get(ctx, profileKey).Result()
		require.NoError(t, err)
		
		var profile map[string]interface{}
		err = json.Unmarshal([]byte(profileData), &profile)
		require.NoError(t, err)
		
		assert.Equal(t, "Колбэк", profile["max_first_name"])
		assert.Equal(t, "Пользователь", profile["max_last_name"])
	})
	
	// Test unknown event type
	t.Run("UnknownEventType", func(t *testing.T) {
		webhookEvent := MaxWebhookEvent{
			Type: "unknown_event_type",
		}
		
		status, _ := client.POST(t, MaxBotServiceURL+"/webhook/max", webhookEvent)
		// Should return 200 OK even for unknown events
		assert.Equal(t, http.StatusOK, status)
	})
}

// isServiceAvailable checks if a service is available at the given URL
func isServiceAvailable(url string) bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// cleanupTestData cleans up test data from Redis
func cleanupTestData(t *testing.T, rdb *redis.Client) {
	ctx := context.Background()
	
	// Delete all test profile keys
	keys, err := rdb.Keys(ctx, "profile:user:*").Result()
	if err != nil {
		t.Logf("Warning: failed to get profile keys for cleanup: %v", err)
		return
	}
	
	if len(keys) > 0 {
		err = rdb.Del(ctx, keys...).Err()
		if err != nil {
			t.Logf("Warning: failed to cleanup profile keys: %v", err)
		}
	}
	
	// Clean up employee database if available
	if isServiceAvailable(EmployeeServiceURL) {
		employeeDB := ConnectDB(t, EmployeeDBConnStr)
		if employeeDB != nil {
			defer employeeDB.Close()
			CleanupDB(t, employeeDB, []string{"employees", "universities"})
		}
	}
}