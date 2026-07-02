// pkg/database/postgres.go
package database

import (
	"booking-service/internal/domain"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Khai báo một biến toàn cục để chứa kết nối DB
var DB *gorm.DB

// ConnectDB khởi tạo kết nối đến PostgreSQL
func ConnectDB() {
	// Chuỗi kết nối được xây dựng từ các biến môi trường DB_URL, DB_USER,
	// DB_PASSWORD, DB_NAME, DB_SSLMODE, DB_PORT
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Ho_Chi_Minh",
		os.Getenv("DB_URL"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_SSLMODE"),
	)

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("✅ Successfully connected to PostgreSQL!")

	// --- BẠN PASTE ĐOẠN NÀY VÀO ĐÂY ---
	fmt.Println("⏳ Running database migrations...")
	err = database.AutoMigrate(
		&domain.TimeTemplate{},
		&domain.Availability{},
		&domain.ExpertTimeOff{},
		&domain.ExpertSlot{},
		&domain.Appointment{},
		&domain.MedicalRecord{},
		&domain.Review{},
	)

	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	fmt.Println("✅ 7 Database tables migrated successfully!")
	DB = database
}
