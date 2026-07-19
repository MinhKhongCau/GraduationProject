package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

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

	// ─── Internal M2M Auth ──────────────────────────────────────────────────
	// Dùng để xin token từ Auth Service khi gọi service khác nội bộ.
	AuthServiceInternalURL string // http://auth-service:8080
	InternalClientID       string // "payment-service"
	InternalClientSecret   string // plain text secret (chỉ trong ENV, không commit)

	// ─── URLs của các service khác (internal Docker network) ─────────────────
	BookingServiceInternalURL string // http://booking-service:8083

	PaymentOrderTTL        time.Duration
	PaymentMinUsableWindow time.Duration
}

var AppConfig *Config

func LoadConfig() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  Không tìm thấy file .env, sử dụng biến môi trường hệ thống")
	}

	paymentOrderTTL, paymentMinUsableWindow, err := parsePaymentLifecycleConfig(
		getEnvOrDefault("PAYMENT_ORDER_TTL_MINUTES", "15"),
		getEnvOrDefault("PAYMENT_MIN_USABLE_WINDOW_SECONDS", "60"),
	)
	if err != nil {
		log.Fatalf("Invalid payment lifecycle configuration: %v", err)
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

		// Internal M2M Auth
		AuthServiceInternalURL: getEnvOrDefault("AUTH_SERVICE_INTERNAL_URL", "http://auth-service:8080"),
		InternalClientID:       getEnvOrDefault("INTERNAL_CLIENT_ID", "payment-service"),
		InternalClientSecret:   getEnvOrDefault("INTERNAL_CLIENT_SECRET", ""),

		// URLs service khác
		BookingServiceInternalURL: getEnvOrDefault("BOOKING_SERVICE_INTERNAL_URL", "http://booking-service:8083"),
		PaymentOrderTTL:           paymentOrderTTL,
		PaymentMinUsableWindow:    paymentMinUsableWindow,
	}
}

func parsePaymentLifecycleConfig(ttlMinutesValue, minimumWindowSecondsValue string) (time.Duration, time.Duration, error) {
	ttlMinutes, err := strconv.Atoi(ttlMinutesValue)
	if err != nil || ttlMinutes <= 0 || ttlMinutes > 60 {
		return 0, 0, fmt.Errorf("PAYMENT_ORDER_TTL_MINUTES must be an integer between 1 and 60")
	}
	minimumWindowSeconds, err := strconv.Atoi(minimumWindowSecondsValue)
	if err != nil || minimumWindowSeconds <= 0 {
		return 0, 0, fmt.Errorf("PAYMENT_MIN_USABLE_WINDOW_SECONDS must be a positive integer")
	}

	ttl := time.Duration(ttlMinutes) * time.Minute
	minimumWindow := time.Duration(minimumWindowSeconds) * time.Second
	if minimumWindow > ttl {
		return 0, 0, fmt.Errorf("PAYMENT_MIN_USABLE_WINDOW_SECONDS must not exceed PAYMENT_ORDER_TTL_MINUTES")
	}
	return ttl, minimumWindow, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}
