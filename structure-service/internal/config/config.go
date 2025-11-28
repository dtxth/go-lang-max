package config

import "os"

type Config struct {
	DBUrl           string
	Port            string
	GRPCPort        string
	ChatService     string // Адрес chat-service gRPC
	EmployeeService string // Адрес employee-service gRPC
}

func Load() *Config {
	return &Config{
		DBUrl:           os.Getenv("DATABASE_URL"),
		Port:            getEnv("PORT", "8083"),
		GRPCPort:        getEnv("GRPC_PORT", "9093"),
		ChatService:     getEnv("CHAT_SERVICE_GRPC", "localhost:9092"),
		EmployeeService: getEnv("EMPLOYEE_SERVICE_GRPC", "localhost:9091"),
	}
}

func getEnv(key, def string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return def
}

