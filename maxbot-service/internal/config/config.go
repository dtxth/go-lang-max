package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	GRPCPort       string
	HTTPPort       string
	MaxAPIURL      string
	MaxAPIToken    string
	RequestTimeout time.Duration
	MockMode       bool
	
	// Redis configuration for profile cache
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	ProfileTTL    time.Duration
	
	// Webhook configuration
	WebhookSecret string
	
	// Monitoring configuration
	MonitoringEnabled              bool
	ProfileQualityAlertThreshold   float64
	WebhookErrorAlertThreshold     float64
}

func Load() *Config {
	return &Config{
		GRPCPort:       getEnv("GRPC_PORT", "9095"),
		HTTPPort:       getEnv("HTTP_PORT", "8095"),
		MaxAPIURL:      getEnv("MAX_API_URL", ""),
		MaxAPIToken:    getEnv("MAX_BOT_TOKEN", ""),
		RequestTimeout: getDurationEnv("MAX_API_TIMEOUT", 5*time.Second),
		MockMode:       getBoolEnv("MOCK_MODE", false),
		
		// Redis configuration
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getIntEnv("REDIS_DB", 0),
		ProfileTTL:    getDurationEnv("PROFILE_TTL", 30*24*time.Hour), // 30 days
		
		// Webhook configuration
		WebhookSecret: getEnv("WEBHOOK_SECRET", ""),
		
		// Monitoring configuration
		MonitoringEnabled:              getBoolEnv("MONITORING_ENABLED", true),
		ProfileQualityAlertThreshold:   getFloatEnv("PROFILE_QUALITY_ALERT_THRESHOLD", 0.8),
		WebhookErrorAlertThreshold:     getFloatEnv("WEBHOOK_ERROR_ALERT_THRESHOLD", 0.05),
	}
}

func getEnv(key, def string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return def
}

func getDurationEnv(key string, def time.Duration) time.Duration {
	if val, ok := os.LookupEnv(key); ok {
		if parsed, err := time.ParseDuration(val); err == nil {
			return parsed
		}
	}
	return def
}

func getBoolEnv(key string, def bool) bool {
	if val, ok := os.LookupEnv(key); ok {
		return val == "true" || val == "1" || val == "yes"
	}
	return def
}

func getIntEnv(key string, def int) int {
	if val, ok := os.LookupEnv(key); ok {
		if parsed, err := strconv.Atoi(val); err == nil {
			return parsed
		}
	}
	return def
}

func getFloatEnv(key string, def float64) float64 {
	if val, ok := os.LookupEnv(key); ok {
		if parsed, err := strconv.ParseFloat(val, 64); err == nil {
			return parsed
		}
	}
	return def
}
