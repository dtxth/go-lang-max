package config

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	DBUrl                    string
	Port                     string
	GRPCPort                 string
	MaxAPI                   string // URL для MAX API (опционально)
	MaxBotAddress            string
	MaxBotTimeout            time.Duration
	AuthAddress              string
	AuthTimeout              time.Duration
	RedisURL                 string
	RedisMaxRetries          int
	RedisRetryDelay          time.Duration
	RedisHealthCheckInterval time.Duration
}

// Load loads and validates the main application configuration
// Implements comprehensive validation as per Requirements 3.1, 3.2, 3.3, 3.4
func Load() *Config {
	log.Printf("Loading main application configuration from environment variables")
	
	config := &Config{
		DBUrl:                    getEnvRequired("DATABASE_URL"),
		Port:                     getEnvWithValidation("PORT", "8082", validatePort),
		GRPCPort:                 getEnvWithValidation("GRPC_PORT", "9092", validatePort),
		MaxAPI:                   getEnv("MAX_API_URL", ""),
		MaxBotAddress:            getEnvWithValidation("MAXBOT_GRPC_ADDR", "localhost:9095", validateGRPCAddress),
		MaxBotTimeout:            getDurationEnvWithValidation("MAXBOT_TIMEOUT", 5*time.Second, 1*time.Second, 60*time.Second),
		AuthAddress:              getEnvWithValidation("AUTH_GRPC_ADDR", "localhost:9090", validateGRPCAddress),
		AuthTimeout:              getDurationEnvWithValidation("AUTH_TIMEOUT", 5*time.Second, 1*time.Second, 60*time.Second),
		RedisURL:                 getEnvWithValidation("REDIS_URL", "redis://localhost:6379", validateRedisURL),
		RedisMaxRetries:          loadIntWithValidation("REDIS_MAX_RETRIES", 5, 1, 20),
		RedisRetryDelay:          getDurationEnvWithValidation("REDIS_RETRY_DELAY", 1*time.Second, 100*time.Millisecond, 30*time.Second),
		RedisHealthCheckInterval: getDurationEnvWithValidation("REDIS_HEALTH_CHECK_INTERVAL", 30*time.Second, 10*time.Second, 5*time.Minute),
	}
	
	// Validate MaxAPI URL if provided
	if config.MaxAPI != "" {
		if err := validateURL(config.MaxAPI); err != nil {
			log.Printf("CONFIG WARNING: Invalid MAX_API_URL '%s': %v, continuing without MAX API", config.MaxAPI, err)
			config.MaxAPI = ""
		}
	}
	
	// Log configuration summary for monitoring
	logMainConfigurationSummary(config)
	
	return config
}

// logMainConfigurationSummary logs a summary of the main configuration for monitoring
func logMainConfigurationSummary(config *Config) {
	log.Printf("Main application configuration loaded successfully:")
	log.Printf("  HTTP Port: %s", config.Port)
	log.Printf("  gRPC Port: %s", config.GRPCPort)
	log.Printf("  Auth Service: %s (timeout: %v)", config.AuthAddress, config.AuthTimeout)
	log.Printf("  MaxBot Service: %s (timeout: %v)", config.MaxBotAddress, config.MaxBotTimeout)
	log.Printf("  Redis URL: %s", config.RedisURL)
	log.Printf("  Redis Max Retries: %d", config.RedisMaxRetries)
	log.Printf("  Redis Retry Delay: %v", config.RedisRetryDelay)
	log.Printf("  Redis Health Check Interval: %v", config.RedisHealthCheckInterval)
	if config.MaxAPI != "" {
		log.Printf("  MAX API URL: %s", config.MaxAPI)
	} else {
		log.Printf("  MAX API URL: not configured")
	}
}

// getEnvRequired gets a required environment variable and panics if not set
func getEnvRequired(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("Required environment variable %s is not set", key))
	}
	return val
}

// getEnv gets an environment variable with a default value
func getEnv(key, def string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return def
}

// getEnvWithValidation gets an environment variable with validation
func getEnvWithValidation(key, def string, validator func(string) error) string {
	val := getEnv(key, def)
	if err := validator(val); err != nil {
		log.Printf("CONFIG WARNING: Invalid %s='%s': %v, using default '%s'", key, val, err, def)
		return def
	}
	return val
}

// getDurationEnv gets a duration environment variable with a default
func getDurationEnv(key string, def time.Duration) time.Duration {
	if val, ok := os.LookupEnv(key); ok {
		if parsed, err := time.ParseDuration(val); err == nil {
			return parsed
		}
		log.Printf("CONFIG WARNING: Invalid duration format for %s='%s', using default %v", key, val, def)
	}
	return def
}

// getDurationEnvWithValidation gets a duration environment variable with validation
func getDurationEnvWithValidation(key string, def, min, max time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	
	duration, err := time.ParseDuration(val)
	if err != nil {
		log.Printf("CONFIG WARNING: Invalid duration format for %s='%s': %v, using default %v", key, val, err, def)
		return def
	}
	
	if duration < min || duration > max {
		log.Printf("CONFIG WARNING: Duration %s='%s' out of range [%v, %v], using default %v", key, val, min, max, def)
		return def
	}
	
	return duration
}

// loadIntWithValidation loads an integer from environment variable with validation
func loadIntWithValidation(key string, def, min, max int) int {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	
	intVal, err := strconv.Atoi(val)
	if err != nil {
		log.Printf("CONFIG WARNING: Invalid integer format for %s='%s': %v, using default %d", key, val, err, def)
		return def
	}
	
	if intVal < min || intVal > max {
		log.Printf("CONFIG WARNING: Integer %s='%s' out of range [%d, %d], using default %d", key, val, min, max, def)
		return def
	}
	
	return intVal
}

// Validation functions

// validatePort validates that a string is a valid port number
func validatePort(port string) error {
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("invalid port number format")
	}
	if portNum < 1 || portNum > 65535 {
		return fmt.Errorf("port number must be between 1 and 65535")
	}
	return nil
}

// validateGRPCAddress validates that a string is a valid gRPC address
func validateGRPCAddress(addr string) error {
	if addr == "" {
		return fmt.Errorf("address cannot be empty")
	}
	
	// Check if it contains a colon (host:port format)
	if !strings.Contains(addr, ":") {
		return fmt.Errorf("address must be in host:port format")
	}
	
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return fmt.Errorf("address must be in host:port format")
	}
	
	// Validate port part
	return validatePort(parts[1])
}

// validateURL validates that a string is a valid URL
func validateURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("URL cannot be empty")
	}
	
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %v", err)
	}
	
	if parsedURL.Scheme == "" {
		return fmt.Errorf("URL must have a scheme (http/https)")
	}
	
	if parsedURL.Host == "" {
		return fmt.Errorf("URL must have a host")
	}
	
	return nil
}

// validateRedisURL validates that a string is a valid Redis URL
func validateRedisURL(redisURL string) error {
	if redisURL == "" {
		return fmt.Errorf("Redis URL cannot be empty")
	}
	
	parsedURL, err := url.Parse(redisURL)
	if err != nil {
		return fmt.Errorf("invalid Redis URL format: %v", err)
	}
	
	if parsedURL.Scheme != "redis" && parsedURL.Scheme != "rediss" {
		return fmt.Errorf("Redis URL must use redis:// or rediss:// scheme")
	}
	
	if parsedURL.Host == "" {
		return fmt.Errorf("Redis URL must have a host")
	}
	
	return nil
}
