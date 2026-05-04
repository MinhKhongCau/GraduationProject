// File: config/database.go
package config

import (
	"log"
	"os"

	"profile-service/internal/models" // Đổi 'profile-service' thành tên module của bạn nếu khác

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		log.Fatal("Lỗi: Không tìm thấy biến môi trường DB_URL")
	}

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
