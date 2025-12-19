package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
    DBUrl                   string
    AccessSecret            string
    RefreshSecret           string
    Port                    string
    GRPCPort                string
    NotificationServiceType string
    MaxBotServiceAddr       string
    MaxBotToken             string
    MinPasswordLength       int
    ResetTokenExpiration    int // in minutes
    TokenCleanupInterval    int // in minutes
}

func Load() (*Config, error) {
    minPasswordLength := getEnvInt("MIN_PASSWORD_LENGTH", 12)
    resetTokenExpiration := getEnvInt("RESET_TOKEN_EXPIRATION", 15)
    tokenCleanupInterval := getEnvInt("TOKEN_CLEANUP_INTERVAL", 60) // Default: 1 hour
    notificationServiceType := getEnv("NOTIFICATION_SERVICE_TYPE", "mock")
    
    cfg := &Config{
        DBUrl:                   os.Getenv("DATABASE_URL"),
        AccessSecret:            os.Getenv("ACCESS_SECRET"),
        RefreshSecret:           os.Getenv("REFRESH_SECRET"),
        Port:                    getEnv("PORT", "8080"),
        GRPCPort:                getEnv("GRPC_PORT", "9090"),
        NotificationServiceType: notificationServiceType,
        MaxBotServiceAddr:       getEnv("MAXBOT_SERVICE_ADDR", ""),
        MaxBotToken:             os.Getenv("MAX_BOT_TOKEN"),
        MinPasswordLength:       minPasswordLength,
        ResetTokenExpiration:    resetTokenExpiration,
        TokenCleanupInterval:    tokenCleanupInterval,
    }
    
    // Validate configuration
    if err := cfg.Validate(); err != nil {
        return nil, err
    }
    
    return cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
    if c.MinPasswordLength < 8 {
        return fmt.Errorf("MIN_PASSWORD_LENGTH must be at least 8, got %d", c.MinPasswordLength)
    }
    
    if c.ResetTokenExpiration < 1 {
        return fmt.Errorf("RESET_TOKEN_EXPIRATION must be at least 1 minute, got %d", c.ResetTokenExpiration)
    }
    
    if c.TokenCleanupInterval < 1 {
        return fmt.Errorf("TOKEN_CLEANUP_INTERVAL must be at least 1 minute, got %d", c.TokenCleanupInterval)
    }
    
    if c.NotificationServiceType != "mock" && c.NotificationServiceType != "max" {
        return fmt.Errorf("NOTIFICATION_SERVICE_TYPE must be 'mock' or 'max', got '%s'", c.NotificationServiceType)
    }
    
    if c.NotificationServiceType == "max" && c.MaxBotServiceAddr == "" {
        return fmt.Errorf("MAXBOT_SERVICE_ADDR is required when NOTIFICATION_SERVICE_TYPE is 'max'")
    }
    
    return nil
}

func getEnv(key, def string) string {
    if val, ok := os.LookupEnv(key); ok {
        return val
    }
    return def
}

func getEnvInt(key string, def int) int {
    if val, ok := os.LookupEnv(key); ok {
        if intVal, err := strconv.Atoi(val); err == nil {
            return intVal
        }
    }
    return def
}