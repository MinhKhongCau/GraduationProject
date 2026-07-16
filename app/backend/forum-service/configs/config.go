package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	ServerPort string

	RabbitMQHost string
	RabbitMQPort string
	RabbitMQUser string
	RabbitMQPass string
}

var AppConfig *Config

func LoadConfig() {
	if err := godotenv.Load("configs/.env"); err != nil {
		log.Println("configs/.env not found, falling back to system environment variables")
	}

	AppConfig = &Config{
		DBHost:     getEnvOrDefault("DB_HOST", "localhost"),
		DBPort:     getEnvOrDefault("DB_PORT", "5432"),
		DBUser:     getEnvOrDefault("DB_USER", "admin"),
		DBPassword: getEnvOrDefault("DB_PASSWORD", "admin"),
		DBName:     getEnvOrDefault("DB_NAME", "forum_db"),
		DBSSLMode:  getEnvOrDefault("DB_SSLMODE", "disable"),

		ServerPort: getEnvOrDefault("PORT", "8084"),

		RabbitMQHost: getEnvOrDefault("RABBITMQ_HOST", "rabbitmq"),
		RabbitMQPort: getEnvOrDefault("RABBITMQ_PORT", "5672"),
		RabbitMQUser: getEnvOrDefault("RABBITMQ_DEFAULT_USER", "admin"),
		RabbitMQPass: getEnvOrDefault("RABBITMQ_DEFAULT_PASS", "admin"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}
