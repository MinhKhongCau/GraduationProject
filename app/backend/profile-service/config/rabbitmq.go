// File: config/rabbitmq.go
package config

import "os"

// RabbitMQConfig holds the broker connection settings used to consume
// domain events (see app/backend/RABBITMQ_CONVENTION.md).
type RabbitMQConfig struct {
	Host string
	Port string
	User string
	Pass string
}

func LoadRabbitMQConfig() RabbitMQConfig {
	return RabbitMQConfig{
		Host: getEnvOrDefault("RABBITMQ_HOST", "rabbitmq"),
		Port: getEnvOrDefault("RABBITMQ_PORT", "5672"),
		User: getEnvOrDefault("RABBITMQ_DEFAULT_USER", "admin"),
		Pass: getEnvOrDefault("RABBITMQ_DEFAULT_PASS", "admin"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}
