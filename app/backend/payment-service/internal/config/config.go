package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost       string
	DBPort       string
	DBUser       string
	DBPassword   string
	DBName       string
	DBSSLMode    string
	ServerPort   string
	RedisHost    string
	RedisPort    string
	RabbitMQHost string
	RabbitMQPort string
	RabbitMQUser string
	RabbitMQPass string
}

var AppConfig *Config

func LoadConfig() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  Không tìm thấy file .env, sử dụng biến môi trường hệ thống")
	}

	AppConfig = &Config{
		DBHost:       getEnvOrDefault("DB_HOST", "localhost"),
		DBPort:       getEnvOrDefault("DB_PORT", "5432"),
		DBUser:       getEnvOrDefault("DB_USER", "admin"),
		DBPassword:   getEnvOrDefault("DB_PASSWORD", "admin"),
		DBName:       getEnvOrDefault("PAYMENT_DB_NAME", "payment_db"),
		DBSSLMode:    getEnvOrDefault("DB_SSLMODE", "disable"),
		ServerPort:   getEnvOrDefault("PORT", "8082"),
		RedisHost:    getEnvOrDefault("REDIS_HOST", "localhost"),
		RedisPort:    getEnvOrDefault("REDIS_PORT", "6379"),
		RabbitMQHost: getEnvOrDefault("RABBITMQ_HOST", "localhost"),
		RabbitMQPort: getEnvOrDefault("RABBITMQ_PORT", "5672"),
		RabbitMQUser: getEnvOrDefault("RABBITMQ_DEFAULT_USER", "admin"),
		RabbitMQPass: getEnvOrDefault("RABBITMQ_DEFAULT_PASS", "admin123"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}
