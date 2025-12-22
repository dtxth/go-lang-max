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
// All services now route through the Gateway Service at port 8080
func DefaultServiceConfigs() map[string]ServiceConfig {
	// Gateway Service acts as the single entry point for all microservices
	gatewayBaseURL := "http://localhost:8080"
	
	return map[string]ServiceConfig{
		"auth": {
			BaseURL: gatewayBaseURL, // Auth endpoints routed through Gateway
			Timeout: 30 * time.Second,
		},
		"employee": {
			BaseURL: gatewayBaseURL, // Employee endpoints routed through Gateway
			Timeout: 30 * time.Second,
		},
		"chat": {
			BaseURL: gatewayBaseURL, // Chat endpoints routed through Gateway
			Timeout: 30 * time.Second,
		},
		"structure": {
			BaseURL: gatewayBaseURL, // Structure endpoints routed through Gateway
			Timeout: 30 * time.Second,
		},
		"migration": {
			BaseURL: "http://localhost:8084", // Migration service remains direct (not part of Gateway)
			Timeout: 30 * time.Second,
		},
		"maxbot": {
			BaseURL: "http://localhost:8095", // MaxBot service remains direct (not part of Gateway)
			Timeout: 30 * time.Second,
		},
		"gateway": {
			BaseURL: gatewayBaseURL, // Gateway Service health checks
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