// pkg/database/postgres.go
package database

import (
	"booking-service/internal/domain"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "time/tzdata"
)

// Khai báo một biến toàn cục để chứa kết nối DB
var DB *gorm.DB

// ConnectDB khởi tạo kết nối đến PostgreSQL
func ConnectDB() {
	// Đọc cấu hình từ Env, nếu trống sẽ lấy mặc định khớp với docker-compose.dev.yml
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "admin"
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "admin"
	}
	dbname := os.Getenv("BOOKING_DB_NAME")
	if dbname == "" {
		// Mặc định dùng chung DB được khởi tạo trong docker-compose.dev.yml
		dbname = "psychology_assessment_db"
	}
	sslmode := os.Getenv("DB_SSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Ho_Chi_Minh",
		host, user, password, dbname, port, sslmode)

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
