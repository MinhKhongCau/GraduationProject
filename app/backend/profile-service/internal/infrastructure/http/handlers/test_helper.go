package handlers

import (
	"net/http/httptest"
	"profile-service/config"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
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

// assertBaseResponse kiểm tra response theo đúng envelope chung
// { message, statusCode, timestamp, result } và statusCode khớp HTTP status.
func assertBaseResponse(t *testing.T, w *httptest.ResponseRecorder, resp map[string]interface{}) {
	t.Helper()
	assert.Equal(t, float64(w.Code), resp["statusCode"])
	assert.NotEmpty(t, resp["message"])
	assert.Regexp(t, `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$`, resp["timestamp"])
	assert.Contains(t, resp, "result")
	assert.NotContains(t, resp, "success")
	assert.NotContains(t, resp, "data")
}
