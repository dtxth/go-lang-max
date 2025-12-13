package config

import (
	"os"
	"time"
)

type Config struct {
	GRPCPort       string
	MaxAPIURL      string
	MaxAPIToken    string
	RequestTimeout time.Duration
	MockMode       bool
}

func Load() *Config {
	return &Config{
		GRPCPort:       getEnv("GRPC_PORT", "9095"),
		MaxAPIURL:      getEnv("MAX_API_URL", ""),
		MaxAPIToken:    getEnv("MAX_API_TOKEN", ""),
		RequestTimeout: getDurationEnv("MAX_API_TIMEOUT", 5*time.Second),
		MockMode:       getBoolEnv("MOCK_MODE", false),
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
