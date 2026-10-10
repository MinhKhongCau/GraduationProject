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
	BookingGRPCAddr           string // booking-service:9003
	ProfileGRPCAddr           string // profile-service:9002

	// SystemWalletUserID: chủ ví hệ thống giữ tiền bệnh nhân trả cho tới khi buổi tư vấn hoàn tất.
	SystemWalletUserID string

	PaymentOrderTTL        time.Duration
	PaymentMinUsableWindow time.Duration
	OutboxPollInterval     time.Duration
	OutboxMaxAttempts      int
	OutboxBaseBackoff      time.Duration
	OutboxMaxBackoff       time.Duration
	OutboxBatchSize        int
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
	outboxConfig, err := parseOutboxConfig(
		getEnvOrDefault("OUTBOX_POLL_INTERVAL_SECONDS", "5"),
		getEnvOrDefault("OUTBOX_MAX_ATTEMPTS", "10"),
		getEnvOrDefault("OUTBOX_BASE_BACKOFF_SECONDS", "5"),
		getEnvOrDefault("OUTBOX_MAX_BACKOFF_SECONDS", "300"),
		getEnvOrDefault("OUTBOX_BATCH_SIZE", "50"),
	)
	if err != nil {
		log.Fatalf("Invalid outbox configuration: %v", err)
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
		BookingGRPCAddr:           getEnvOrDefault("BOOKING_GRPC_ADDR", "booking-service:9003"),
		ProfileGRPCAddr:           getEnvOrDefault("PROFILE_GRPC_ADDR", "profile-service:9002"),
		SystemWalletUserID:        getEnvOrDefault("SYSTEM_WALLET_USER_ID", "00000000-0000-0000-0000-000000000001"),
		PaymentOrderTTL:           paymentOrderTTL,
		PaymentMinUsableWindow:    paymentMinUsableWindow,
		OutboxPollInterval:        outboxConfig.PollInterval,
		OutboxMaxAttempts:         outboxConfig.MaxAttempts,
		OutboxBaseBackoff:         outboxConfig.BaseBackoff,
		OutboxMaxBackoff:          outboxConfig.MaxBackoff,
		OutboxBatchSize:           outboxConfig.BatchSize,
	}
}

type outboxConfig struct {
	PollInterval time.Duration
	MaxAttempts  int
	BaseBackoff  time.Duration
	MaxBackoff   time.Duration
	BatchSize    int
}

func parseOutboxConfig(pollValue, attemptsValue, baseValue, maximumValue, batchValue string) (outboxConfig, error) {
	pollSeconds, err := parsePositiveBoundedInt("OUTBOX_POLL_INTERVAL_SECONDS", pollValue, 3600)
	if err != nil {
		return outboxConfig{}, err
	}
	maxAttempts, err := parsePositiveBoundedInt("OUTBOX_MAX_ATTEMPTS", attemptsValue, 100)
	if err != nil {
		return outboxConfig{}, err
	}
	baseSeconds, err := parsePositiveBoundedInt("OUTBOX_BASE_BACKOFF_SECONDS", baseValue, 86400)
	if err != nil {
		return outboxConfig{}, err
	}
	maximumSeconds, err := parsePositiveBoundedInt("OUTBOX_MAX_BACKOFF_SECONDS", maximumValue, 86400)
	if err != nil {
		return outboxConfig{}, err
	}
	batchSize, err := parsePositiveBoundedInt("OUTBOX_BATCH_SIZE", batchValue, 1000)
	if err != nil {
		return outboxConfig{}, err
	}
	if baseSeconds > maximumSeconds {
		return outboxConfig{}, fmt.Errorf("OUTBOX_BASE_BACKOFF_SECONDS must not exceed OUTBOX_MAX_BACKOFF_SECONDS")
	}
	return outboxConfig{
		PollInterval: time.Duration(pollSeconds) * time.Second,
		MaxAttempts:  maxAttempts,
		BaseBackoff:  time.Duration(baseSeconds) * time.Second,
		MaxBackoff:   time.Duration(maximumSeconds) * time.Second,
		BatchSize:    batchSize,
	}, nil
}

func parsePositiveBoundedInt(name, value string, maximum int) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 || parsed > maximum {
		return 0, fmt.Errorf("%s must be an integer between 1 and %d", name, maximum)
	}
	return parsed, nil
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
