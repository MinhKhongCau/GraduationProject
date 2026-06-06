// pkg/database/postgres.go
package database

import (
	"booking-service/internal/domain"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Khai báo một biến toàn cục để chứa kết nối DB
var DB *gorm.DB

// ConnectDB khởi tạo kết nối đến PostgreSQL
func ConnectDB() {
	// Chuỗi kết nối dựa trên cấu hình docker-compose.yml của bạn
	// User: root, Pass: rootpassword, DB: booking_db, Port: 5432
	dsn := "host=localhost user=root password=rootpassword dbname=booking_db port=5432 sslmode=disable TimeZone=Asia/Ho_Chi_Minh"

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
