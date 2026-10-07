package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost          string
	DBPort          string
	DBUser          string
	DBPassword      string
	DBName          string
	DBSSLMode       string
	ServerPort      string
	RollingSlotDays int

	// ─── Internal M2M Auth ──────────────────────────────────────────────────
	// Dùng để xin token từ Auth Service khi gọi service khác nội bộ.
	AuthServiceInternalURL string // http://auth-service:8080
	InternalClientID       string // "booking-service"
	InternalClientSecret   string // plain text secret (chỉ trong ENV, không commit)

	// gRPC tới profile-service (truy vấn chuyên gia + hồ sơ người khám khi đặt lịch)
	ProfileGRPCAddr string // profile-service:9002
}

var AppConfig *Config

func LoadConfig() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  Không tìm thấy file .env, sử dụng biến môi trường hệ thống")
	}

	rollingDays, err := ParseRollingSlotDays(getEnvOrDefault("ROLLING_SLOT_DAYS", "30"))
	if err != nil {
		log.Fatalf("invalid ROLLING_SLOT_DAYS: %v", err)
	}
	AppConfig = &Config{
		DBHost:          getEnvOrDefault("DB_HOST", "localhost"),
		DBPort:          getEnvOrDefault("DB_PORT", "5432"),
		DBUser:          getEnvOrDefault("DB_USER", "admin"),
		DBPassword:      getEnvOrDefault("DB_PASSWORD", "admin"),
		DBName:          getEnvOrDefault("BOOKING_DB_NAME", "booking_db"),
		DBSSLMode:       getEnvOrDefault("DB_SSLMODE", "disable"),
		ServerPort:      getEnvOrDefault("PORT", "8083"),
		RollingSlotDays: rollingDays,

		// Internal M2M Auth
		AuthServiceInternalURL: getEnvOrDefault("AUTH_SERVICE_INTERNAL_URL", "http://auth-service:8080"),
		InternalClientID:       getEnvOrDefault("INTERNAL_CLIENT_ID", "booking-service"),
		InternalClientSecret:   getEnvOrDefault("INTERNAL_CLIENT_SECRET", ""),

		ProfileGRPCAddr: getEnvOrDefault("PROFILE_GRPC_ADDR", "profile-service:9002"),
	}
}

func ParseRollingSlotDays(value string) (int, error) {
	days, err := strconv.Atoi(value)
	if err != nil || days < 1 || days > 365 {
		return 0, fmt.Errorf("must be an integer between 1 and 365")
	}
	return days, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}
