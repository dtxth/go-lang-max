package utils

import (
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

// ServiceConfig содержит конфигурацию для подключения к сервису
type ServiceConfig struct {
	BaseURL string
	Timeout time.Duration
}

// TestClient предоставляет HTTP клиент для тестов
type TestClient struct {
	client *resty.Client
	config ServiceConfig
}

// NewTestClient создает новый тестовый HTTP клиент
func NewTestClient(config ServiceConfig) *TestClient {
	client := resty.New()
	client.SetBaseURL(config.BaseURL)
	client.SetTimeout(config.Timeout)
	client.SetHeader("Content-Type", "application/json")
	
	return &TestClient{
		client: client,
		config: config,
	}
}

// GetClient возвращает resty клиент
func (tc *TestClient) GetClient() *resty.Client {
	return tc.client
}

// SetAuthToken устанавливает токен авторизации
func (tc *TestClient) SetAuthToken(token string) {
	tc.client.SetAuthToken(token)
}

// ClearAuth очищает авторизацию
func (tc *TestClient) ClearAuth() {
	tc.client.SetAuthToken("")
}

// DefaultServiceConfigs возвращает конфигурации по умолчанию для всех сервисов
func DefaultServiceConfigs() map[string]ServiceConfig {
	return map[string]ServiceConfig{
		"auth": {
			BaseURL: "http://localhost:8080",
			Timeout: 30 * time.Second,
		},
		"employee": {
			BaseURL: "http://localhost:8081",
			Timeout: 30 * time.Second,
		},
		"chat": {
			BaseURL: "http://localhost:8082",
			Timeout: 30 * time.Second,
		},
		"structure": {
			BaseURL: "http://localhost:8083",
			Timeout: 30 * time.Second,
		},
		"migration": {
			BaseURL: "http://localhost:8084",
			Timeout: 30 * time.Second,
		},
		"maxbot": {
			BaseURL: "http://localhost:8095",
			Timeout: 30 * time.Second,
		},
	}
}

// WaitForService ждет, пока сервис станет доступным
func WaitForService(baseURL string, maxRetries int) error {
	client := resty.New()
	client.SetTimeout(5 * time.Second)
	
	for i := 0; i < maxRetries; i++ {
		resp, err := client.R().Get(fmt.Sprintf("%s/health", baseURL))
		if err == nil && resp.StatusCode() == 200 {
			return nil
		}
		
		time.Sleep(2 * time.Second)
	}
	
	return fmt.Errorf("service at %s is not available after %d retries", baseURL, maxRetries)
}