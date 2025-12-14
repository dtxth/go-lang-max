package config

import (
	"fmt"
	"os"
)

// Config holds the application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Services ServicesConfig
	Google   GoogleConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// ServicesConfig holds external services configuration
type ServicesConfig struct {
	ChatServiceURL       string
	StructureServiceURL  string
	ChatServiceGRPC      string
	StructureServiceGRPC string
}

// GoogleConfig holds Google API configuration
type GoogleConfig struct {
	CredentialsPath string
	SpreadsheetID   string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8084"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5436"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "migration_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Services: ServicesConfig{
			ChatServiceURL:       getEnv("CHAT_SERVICE_URL", "http://localhost:8082"),
			StructureServiceURL:  getEnv("STRUCTURE_SERVICE_URL", "http://localhost:8083"),
			ChatServiceGRPC:      getEnv("CHAT_SERVICE_GRPC", "chat-service:9092"),
			StructureServiceGRPC: getEnv("STRUCTURE_SERVICE_GRPC", "structure-service:9093"),
		},
		Google: GoogleConfig{
			CredentialsPath: getEnv("GOOGLE_CREDENTIALS_PATH", ""),
			SpreadsheetID:   getEnv("GOOGLE_SPREADSHEET_ID", ""),
		},
	}
}

// GetDSN returns the database connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
