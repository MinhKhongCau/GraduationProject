package handlers

import (
	"profile-service/config"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// SetupTestDB khởi tạo một DB mock bằng sqlmock và gán vào config.DB toàn cục.
func SetupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Không thể khởi tạo sqlmock: %s", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("Không thể mở kết nối GORM: %s", err)
	}

	// Overwrite biến global config.DB
	config.DB = gormDB

	return gormDB, mock
}
