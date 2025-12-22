package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the Gateway Service
type Config struct {
	Server   ServerConfig
	Services ServicesConfig
	Logging  LoggingConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// ServicesConfig holds gRPC client configuration for all microservices
type ServicesConfig struct {
	Auth      ServiceConfig
	Chat      ServiceConfig
	Employee  ServiceConfig
	Structure ServiceConfig
}

// ServiceConfig holds configuration for a single microservice
type ServiceConfig struct {
	Address         string
	Timeout         time.Duration
	MaxRetries      int
	RetryDelay      time.Duration
	MaxRetryDelay   time.Duration
	BackoffMultiplier float64
	CircuitBreaker  CircuitBreakerConfig
}

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	MaxRequests      uint32
	Interval         time.Duration
	Timeout          time.Duration
	ReadyToTrip      func(counts map[string]uint64) bool
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string
	Format string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("GATEWAY_PORT", "8080"),
			ReadTimeout:  getDurationEnv("GATEWAY_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getDurationEnv("GATEWAY_WRITE_TIMEOUT", 30*time.Second),
		},
		Services: ServicesConfig{
			Auth: ServiceConfig{
				Address:           getEnv("AUTH_SERVICE_ADDRESS", "auth-service:9090"),
				Timeout:           getDurationEnv("AUTH_SERVICE_TIMEOUT", 10*time.Second),
				MaxRetries:        getIntEnv("AUTH_SERVICE_MAX_RETRIES", 3),
				RetryDelay:        getDurationEnv("AUTH_SERVICE_RETRY_DELAY", 100*time.Millisecond),
				MaxRetryDelay:     getDurationEnv("AUTH_SERVICE_MAX_RETRY_DELAY", 5*time.Second),
				BackoffMultiplier: getFloatEnv("AUTH_SERVICE_BACKOFF_MULTIPLIER", 2.0),
				CircuitBreaker: CircuitBreakerConfig{
					MaxRequests: uint32(getIntEnv("AUTH_SERVICE_CB_MAX_REQUESTS", 10)),
					Interval:    getDurationEnv("AUTH_SERVICE_CB_INTERVAL", 60*time.Second),
					Timeout:     getDurationEnv("AUTH_SERVICE_CB_TIMEOUT", 60*time.Second),
				},
			},
			Chat: ServiceConfig{
				Address:           getEnv("CHAT_SERVICE_ADDRESS", "chat-service:9092"),
				Timeout:           getDurationEnv("CHAT_SERVICE_TIMEOUT", 10*time.Second),
				MaxRetries:        getIntEnv("CHAT_SERVICE_MAX_RETRIES", 3),
				RetryDelay:        getDurationEnv("CHAT_SERVICE_RETRY_DELAY", 100*time.Millisecond),
				MaxRetryDelay:     getDurationEnv("CHAT_SERVICE_MAX_RETRY_DELAY", 5*time.Second),
				BackoffMultiplier: getFloatEnv("CHAT_SERVICE_BACKOFF_MULTIPLIER", 2.0),
				CircuitBreaker: CircuitBreakerConfig{
					MaxRequests: uint32(getIntEnv("CHAT_SERVICE_CB_MAX_REQUESTS", 10)),
					Interval:    getDurationEnv("CHAT_SERVICE_CB_INTERVAL", 60*time.Second),
					Timeout:     getDurationEnv("CHAT_SERVICE_CB_TIMEOUT", 60*time.Second),
				},
			},
			Employee: ServiceConfig{
				Address:           getEnv("EMPLOYEE_SERVICE_ADDRESS", "employee-service:9091"),
				Timeout:           getDurationEnv("EMPLOYEE_SERVICE_TIMEOUT", 10*time.Second),
				MaxRetries:        getIntEnv("EMPLOYEE_SERVICE_MAX_RETRIES", 3),
				RetryDelay:        getDurationEnv("EMPLOYEE_SERVICE_RETRY_DELAY", 100*time.Millisecond),
				MaxRetryDelay:     getDurationEnv("EMPLOYEE_SERVICE_MAX_RETRY_DELAY", 5*time.Second),
				BackoffMultiplier: getFloatEnv("EMPLOYEE_SERVICE_BACKOFF_MULTIPLIER", 2.0),
				CircuitBreaker: CircuitBreakerConfig{
					MaxRequests: uint32(getIntEnv("EMPLOYEE_SERVICE_CB_MAX_REQUESTS", 10)),
					Interval:    getDurationEnv("EMPLOYEE_SERVICE_CB_INTERVAL", 60*time.Second),
					Timeout:     getDurationEnv("EMPLOYEE_SERVICE_CB_TIMEOUT", 60*time.Second),
				},
			},
			Structure: ServiceConfig{
				Address:           getEnv("STRUCTURE_SERVICE_ADDRESS", "structure-service:9093"),
				Timeout:           getDurationEnv("STRUCTURE_SERVICE_TIMEOUT", 10*time.Second),
				MaxRetries:        getIntEnv("STRUCTURE_SERVICE_MAX_RETRIES", 3),
				RetryDelay:        getDurationEnv("STRUCTURE_SERVICE_RETRY_DELAY", 100*time.Millisecond),
				MaxRetryDelay:     getDurationEnv("STRUCTURE_SERVICE_MAX_RETRY_DELAY", 5*time.Second),
				BackoffMultiplier: getFloatEnv("STRUCTURE_SERVICE_BACKOFF_MULTIPLIER", 2.0),
				CircuitBreaker: CircuitBreakerConfig{
					MaxRequests: uint32(getIntEnv("STRUCTURE_SERVICE_CB_MAX_REQUESTS", 10)),
					Interval:    getDurationEnv("STRUCTURE_SERVICE_CB_INTERVAL", 60*time.Second),
					Timeout:     getDurationEnv("STRUCTURE_SERVICE_CB_TIMEOUT", 60*time.Second),
				},
			},
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}
}

// Helper functions for environment variable parsing
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getFloatEnv(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}