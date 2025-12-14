package config

import (
	"os"
	"time"
)

type Config struct {
	DBUrl         string
	Port          string
	GRPCPort      string
	MaxAPI        string // URL для MAX API (опционально)
	MaxBotAddress string
	MaxBotTimeout time.Duration
	AuthAddress   string
	AuthTimeout   time.Duration
}

func Load() *Config {
	return &Config{
		DBUrl:         os.Getenv("DATABASE_URL"),
		Port:          getEnv("PORT", "8082"),
		GRPCPort:      getEnv("GRPC_PORT", "9092"),
		MaxAPI:        getEnv("MAX_API_URL", ""),
		MaxBotAddress: getEnv("MAXBOT_GRPC_ADDR", "localhost:9095"),
		MaxBotTimeout: getDurationEnv("MAXBOT_TIMEOUT", 5*time.Second),
		AuthAddress:   getEnv("AUTH_GRPC_ADDR", "localhost:9090"),
		AuthTimeout:   getDurationEnv("AUTH_TIMEOUT", 5*time.Second),
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
