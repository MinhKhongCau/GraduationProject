// File: config/database.go
package config

import (
	"fmt"
	"log"
	"os"

	"profile-service/internal/models" // Đổi 'profile-service' thành tên module của bạn nếu khác

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		os.Getenv("DB_URL"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("PORT"),
		os.Getenv("DB_SSLMODE"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Lỗi kết nối Database:", err)
	}

	log.Println("✅ Kết nối Database thành công!")

	// Chạy AutoMigrate để tự động tạo bảng
	err = db.AutoMigrate(
		&models.Patient{},
		&models.MedicalHistory{},
		&models.Expert{},
		&models.Specialization{},
	)

	if err != nil {
		log.Fatal("❌ Lỗi AutoMigrate:", err)
	}

	log.Println("✅ AutoMigrate hoàn tất: Các bảng đã được tạo!")
	DB = db
}
