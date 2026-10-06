// File: config/database.go
package config

import (
	"fmt"
	"log"
	"os"

	"profile-service/internal/infrastructure/persistence"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	dbCfg := persistence.DBConfig{
		Host:     os.Getenv("DB_URL"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Name:     os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
	}

	// Schema được quản lý bằng migration có version (internal/infrastructure/persistence/migrations),
	// không dùng GORM AutoMigrate nữa để các thay đổi như đổi tên/xoá cột có backfill dữ liệu.
	if err := persistence.RunMigrations(dbCfg); err != nil {
		log.Fatal("❌ Lỗi chạy migration:", err)
	}
	log.Println("✅ Migration hoàn tất: schema đã được cập nhật!")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		dbCfg.Host, dbCfg.User, dbCfg.Password, dbCfg.Name, dbCfg.Port, dbCfg.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Lỗi kết nối Database:", err)
	}

	log.Println("✅ Kết nối Database thành công!")
	DB = db

	// Chèn dữ liệu khởi tạo (Admin mặc định + dữ liệu mẫu) nếu chưa tồn tại
	SeedInitialData(db)
}
