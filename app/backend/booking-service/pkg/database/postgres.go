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

	// Chạy pre-migration để xử lý các thay đổi kiểu dữ liệu mà AutoMigrate không tự làm được.
	// Hàm này an toàn để chạy nhiều lần: nó kiểm tra kiểu cột hiện tại trước khi ALTER.
	runPreMigrations(database)

	fmt.Println("⏳ Running AutoMigrate...")
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

	fmt.Println("✅ Database migrated successfully!")
	DB = database
}

// runPreMigrations xử lý các thay đổi kiểu cột mà AutoMigrate không thể tự cast.
func runPreMigrations(db *gorm.DB) {
	fmt.Println("🔄 Running pre-migrations (column type fixes)...")

	// 1. Ép kiểu an toàn (idempotent) cho Booking_Expert_Slots.status
	// Phải DROP DEFAULT trước, nếu không Postgres sẽ báo lỗi không thể ép kiểu (SQLSTATE 42804)
	db.Exec(`ALTER TABLE "Booking_Expert_Slots" ALTER COLUMN status DROP DEFAULT`)
	db.Exec(`
		ALTER TABLE "Booking_Expert_Slots"
		  ALTER COLUMN status TYPE smallint
		  USING CASE status::text
		    WHEN 'AVAILABLE' THEN 0
		    WHEN 'LOCKED'    THEN 1
		    WHEN 'OCCUPIED'  THEN 2
		    WHEN '0' THEN 0
		    WHEN '1' THEN 1
		    WHEN '2' THEN 2
		    ELSE 0
		  END
	`)
	db.Exec(`ALTER TABLE "Booking_Expert_Slots" ALTER COLUMN status SET DEFAULT 0`)

	// 2. Ép kiểu an toàn cho Booking_Appointments.status
	db.Exec(`ALTER TABLE "Booking_Appointments" ALTER COLUMN status DROP DEFAULT`)
	db.Exec(`
		ALTER TABLE "Booking_Appointments"
		  ALTER COLUMN status TYPE smallint
		  USING CASE status::text
		    WHEN 'PENDING_PAYMENT' THEN 0
		    WHEN 'CONFIRMED'       THEN 1
		    WHEN 'CANCELLED'       THEN 2
		    WHEN '0' THEN 0
		    WHEN '1' THEN 1
		    WHEN '2' THEN 2
		    ELSE 0
		  END
	`)
	db.Exec(`ALTER TABLE "Booking_Appointments" ALTER COLUMN status SET DEFAULT 0`)

	// 3. Chuyển các cột int64 (thời gian) đang là timestamp with time zone trên DB cũ sang bigint (Unix ms)
	timestampColumns := []struct {
		table string
		col   string
	}{
		{"Booking_Expert_Slots", "start_time"},
		{"Booking_Expert_Slots", "end_time"},
		{"Booking_Expert_Slots", "locked_expires_at"},
		{"Booking_Appointments", "created_at"},
		{"Booking_Medical_Records", "created_at"},
		{"Booking_Reviews", "created_at"},
		{"Booking_Expert_Time_Off", "start_datetime"},
		{"Booking_Expert_Time_Off", "end_datetime"},
	}

	for _, tc := range timestampColumns {
		// Bỏ qua lỗi vì nếu cột đã là bigint, việc cast sang timestamp sẽ gây lỗi (đúng như ý muốn để skip)
		db.Exec(fmt.Sprintf(`
			ALTER TABLE "%s"
			  ALTER COLUMN "%s" TYPE bigint
			  USING EXTRACT(EPOCH FROM "%s"::timestamp with time zone)::bigint * 1000
		`, tc.table, tc.col, tc.col))
	}

	fmt.Println("✅ Pre-migrations completed.")
}
