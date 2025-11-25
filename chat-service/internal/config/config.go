package config

import "os"

type Config struct {
	DBUrl     string
	Port      string
	GRPCPort  string
	MaxAPI    string // URL для MAX API (опционально)
}

func Load() *Config {
	return &Config{
		DBUrl:    os.Getenv("DATABASE_URL"),
		Port:     getEnv("PORT", "8082"),
		GRPCPort: getEnv("GRPC_PORT", "9092"),
		MaxAPI:   getEnv("MAX_API_URL", ""),
	}
}

func getEnv(key, def string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return def
}

