package config

import (
	"os"
)

type Config struct {
    DBUrl         string
    AccessSecret  string
    RefreshSecret string
    Port          string
    GRPCPort      string
}

func Load() *Config {
    return &Config{
        DBUrl:         os.Getenv("DATABASE_URL"),
        AccessSecret:  os.Getenv("ACCESS_SECRET"),
        RefreshSecret: os.Getenv("REFRESH_SECRET"),
        Port:          getEnv("PORT", "8080"),
        GRPCPort:      getEnv("GRPC_PORT", "9090"),
    }
}

func getEnv(key, def string) string {
    if val, ok := os.LookupEnv(key); ok {
        return val
    }
    return def
}