package main

import (
	"e2e-tests/utils"
	"fmt"
	"log"
	"os"
	"testing"
	"time"
)

// TestMain выполняется перед всеми тестами
func TestMain(m *testing.M) {
	log.Println("Starting E2E tests...")
	
	// Check Gateway Service availability first (single entry point)
	configs := utils.DefaultServiceConfigs()
	gatewayConfig := configs["gateway"]
	
	log.Printf("Checking Gateway Service at %s...", gatewayConfig.BaseURL)
	err := utils.WaitForService(gatewayConfig.BaseURL, 10)
	if err != nil {
		log.Printf("WARNING: Gateway Service is not available: %v", err)
		log.Printf("E2E tests require Gateway Service to be running")
		log.Printf("Please start the Gateway Service and try again")
		os.Exit(1)
	} else {
		log.Printf("✓ Gateway Service is available")
	}
	
	// Check services that are not routed through Gateway
	directServices := map[string]utils.ServiceConfig{
		"migration": configs["migration"],
		"maxbot":    configs["maxbot"],
	}
	
	log.Println("Checking direct services availability...")
	for serviceName, config := range directServices {
		log.Printf("Checking %s at %s...", serviceName, config.BaseURL)
		
		err := utils.WaitForService(config.BaseURL, 10)
		if err != nil {
			log.Printf("WARNING: Service %s is not available: %v", serviceName, err)
			log.Printf("Some tests may be skipped")
		} else {
			log.Printf("✓ Service %s is available", serviceName)
		}
		
		// Небольшая пауза между проверками
		time.Sleep(500 * time.Millisecond)
	}
	
	log.Println("Starting test execution...")
	
	// Запускаем тесты
	code := m.Run()
	
	log.Println("E2E tests completed")
	os.Exit(code)
}

// Вспомогательная функция для логирования результатов тестов
func logTestResult(t *testing.T, testName string, success bool) {
	if success {
		log.Printf("✓ %s: PASSED", testName)
	} else {
		log.Printf("✗ %s: FAILED", testName)
	}
}

// Benchmark тесты для проверки производительности
func BenchmarkAuthServiceLogin(b *testing.B) {
	configs := utils.DefaultServiceConfigs()
	client := utils.NewTestClient(configs["auth"]) // Now points to Gateway Service
	
	// Создаем тестового пользователя
	testUser := utils.GenerateTestUser()
	
	// Регистрируем пользователя
	_, err := client.GetClient().R().
		SetBody(testUser).
		Post("/register")
	
	if err != nil {
		b.Fatalf("Failed to register user: %v", err)
	}
	
	loginData := map[string]string{
		"email":    testUser.Email,
		"password": testUser.Password,
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		resp, err := client.GetClient().R().
			SetBody(loginData).
			Post("/login")
		
		if err != nil {
			b.Fatalf("Login failed: %v", err)
		}
		
		if resp.StatusCode() != 200 {
			b.Fatalf("Expected status 200, got %d", resp.StatusCode())
		}
	}
}

func BenchmarkStructureServiceGetUniversities(b *testing.B) {
	configs := utils.DefaultServiceConfigs()
	client := utils.NewTestClient(configs["structure"]) // Now points to Gateway Service
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		resp, err := client.GetClient().R().
			SetQueryParam("limit", "10").
			Get("/universities")
		
		if err != nil {
			b.Fatalf("Request failed: %v", err)
		}
		
		if resp.StatusCode() != 200 {
			b.Fatalf("Expected status 200, got %d", resp.StatusCode())
		}
	}
}

func BenchmarkEmployeeServiceGetAll(b *testing.B) {
	configs := utils.DefaultServiceConfigs()
	client := utils.NewTestClient(configs["employee"]) // Now points to Gateway Service
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		resp, err := client.GetClient().R().
			Get("/employees/all")
		
		if err != nil {
			b.Fatalf("Request failed: %v", err)
		}
		
		if resp.StatusCode() != 200 {
			b.Fatalf("Expected status 200, got %d", resp.StatusCode())
		}
	}
}

// Пример теста нагрузки
func TestLoadTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}
	
	configs := utils.DefaultServiceConfigs()
	authClient := utils.NewTestClient(configs["auth"]) // Now points to Gateway Service
	
	// Проверяем доступность Gateway Service
	err := utils.WaitForService(configs["gateway"].BaseURL, 5)
	if err != nil {
		t.Skip("Gateway service not available")
	}
	
	// Создаем тестового пользователя
	testUser := utils.GenerateTestUser()
	
	_, err = authClient.GetClient().R().
		SetBody(testUser).
		Post("/register")
	
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}
	
	loginData := map[string]string{
		"email":    testUser.Email,
		"password": testUser.Password,
	}
	
	// Выполняем 100 параллельных запросов
	concurrency := 10
	requestsPerWorker := 10
	
	done := make(chan bool, concurrency)
	errors := make(chan error, concurrency*requestsPerWorker)
	
	start := time.Now()
	
	for i := 0; i < concurrency; i++ {
		go func(workerID int) {
			defer func() { done <- true }()
			
			for j := 0; j < requestsPerWorker; j++ {
				resp, err := authClient.GetClient().R().
					SetBody(loginData).
					Post("/login")
				
				if err != nil {
					errors <- fmt.Errorf("worker %d request %d failed: %v", workerID, j, err)
					return
				}
				
				if resp.StatusCode() != 200 {
					errors <- fmt.Errorf("worker %d request %d got status %d", workerID, j, resp.StatusCode())
					return
				}
			}
		}(i)
	}
	
	// Ждем завершения всех воркеров
	for i := 0; i < concurrency; i++ {
		<-done
	}
	
	duration := time.Since(start)
	
	// Проверяем ошибки
	close(errors)
	errorCount := 0
	for err := range errors {
		t.Logf("Error: %v", err)
		errorCount++
	}
	
	totalRequests := concurrency * requestsPerWorker
	successRate := float64(totalRequests-errorCount) / float64(totalRequests) * 100
	
	t.Logf("Load test completed:")
	t.Logf("  Total requests: %d", totalRequests)
	t.Logf("  Errors: %d", errorCount)
	t.Logf("  Success rate: %.2f%%", successRate)
	t.Logf("  Duration: %v", duration)
	t.Logf("  Requests per second: %.2f", float64(totalRequests)/duration.Seconds())
	
	// Проверяем, что успешность выше 95%
	if successRate < 95.0 {
		t.Errorf("Success rate too low: %.2f%% (expected >= 95%%)", successRate)
	}
	
	// Проверяем, что все запросы выполнились за разумное время
	if duration > 30*time.Second {
		t.Errorf("Load test took too long: %v (expected <= 30s)", duration)
	}
}