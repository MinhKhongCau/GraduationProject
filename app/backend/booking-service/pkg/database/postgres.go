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
// Mỗi bước đều kiểm tra kiểu cột hiện tại trước → an toàn khi chạy nhiều lần (idempotent).
func runPreMigrations(db *gorm.DB) {
	fmt.Println("🔄 Running pre-migrations (column type fixes)...")

	// 1. Chuyển Booking_Expert_Slots.status: varchar → smallint
	//    Lý do: Code dùng enum int (0=AVAILABLE, 1=LOCKED, 2=OCCUPIED)
	if isColumnType(db, "Booking_Expert_Slots", "status", "character varying") {
		fmt.Println("   → Migrating Booking_Expert_Slots.status: varchar → smallint")
		err := db.Exec(`
			ALTER TABLE "Booking_Expert_Slots"
			  ALTER COLUMN status TYPE smallint
			  USING CASE status
			    WHEN 'AVAILABLE' THEN 0
			    WHEN 'LOCKED'    THEN 1
			    WHEN 'OCCUPIED'  THEN 2
			    ELSE 0
			  END
		`).Error
		if err != nil {
			log.Fatalf("Pre-migration failed (Slots.status): %v", err)
		}
		db.Exec(`ALTER TABLE "Booking_Expert_Slots" ALTER COLUMN status SET DEFAULT 0`)
		fmt.Println("   ✅ Done.")
	}

	// 2. Chuyển Booking_Appointments.status: varchar → smallint
	//    Lý do: Code dùng enum int (0=PENDING_PAYMENT, 1=CONFIRMED, 2=CANCELLED)
	if isColumnType(db, "Booking_Appointments", "status", "character varying") {
		fmt.Println("   → Migrating Booking_Appointments.status: varchar → smallint")
		err := db.Exec(`
			ALTER TABLE "Booking_Appointments"
			  ALTER COLUMN status TYPE smallint
			  USING CASE status
			    WHEN 'PENDING_PAYMENT' THEN 0
			    WHEN 'CONFIRMED'       THEN 1
			    WHEN 'CANCELLED'       THEN 2
			    ELSE 0
			  END
		`).Error
		if err != nil {
			log.Fatalf("Pre-migration failed (Appointments.status): %v", err)
		}
		db.Exec(`ALTER TABLE "Booking_Appointments" ALTER COLUMN status SET DEFAULT 0`)
		fmt.Println("   ✅ Done.")
	}

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
	}

	for _, tc := range timestampColumns {
		if isColumnType(db, tc.table, tc.col, "timestamp with time zone") {
			fmt.Printf("   → Migrating %s.%s: timestamp → bigint (Unix ms)\n", tc.table, tc.col)
			err := db.Exec(fmt.Sprintf(`
				ALTER TABLE "%s"
				  ALTER COLUMN "%s" TYPE bigint
				  USING EXTRACT(EPOCH FROM "%s")::bigint * 1000
			`, tc.table, tc.col, tc.col)).Error
			if err != nil {
				log.Fatalf("Pre-migration failed (%s.%s): %v", tc.table, tc.col, err)
			}
			fmt.Println("   ✅ Done.")
		}
	}

	// 4. Chuyển Booking_Expert_Time_Off.start_datetime & end_datetime: timestamp → bigint (Unix ms)
	if isColumnType(db, "Booking_Expert_Time_Off", "start_datetime", "timestamp with time zone") {
		fmt.Println("   → Migrating Booking_Expert_Time_Off: timestamp → bigint (Unix ms)")
		err := db.Exec(`
			ALTER TABLE "Booking_Expert_Time_Off"
			  ALTER COLUMN start_datetime TYPE bigint
			  USING EXTRACT(EPOCH FROM start_datetime)::bigint * 1000
		`).Error
		if err != nil {
			log.Fatalf("Pre-migration failed (TimeOff.start_datetime): %v", err)
		}
		err = db.Exec(`
			ALTER TABLE "Booking_Expert_Time_Off"
			  ALTER COLUMN end_datetime TYPE bigint
			  USING EXTRACT(EPOCH FROM end_datetime)::bigint * 1000
		`).Error
		if err != nil {
			log.Fatalf("Pre-migration failed (TimeOff.end_datetime): %v", err)
		}
		fmt.Println("   ✅ Done.")
	}

	fmt.Println("✅ Pre-migrations completed.")
}

// isColumnType kiểm tra kiểu dữ liệu hiện tại của một cột trong DB.
// Trả về true nếu cột đang có kiểu dữ liệu khớp với expectedType.
// Dùng để đảm bảo pre-migration chỉ chạy khi thực sự cần, an toàn khi restart nhiều lần.
func isColumnType(db *gorm.DB, tableName, columnName, expectedType string) bool {
	var dataType string
	err := db.Raw(`
		SELECT data_type
		FROM information_schema.columns
		WHERE table_name = ? AND column_name = ?
		LIMIT 1
	`, tableName, columnName).Scan(&dataType).Error

	if err != nil || dataType == "" {
		// Bảng/cột chưa tồn tại → AutoMigrate sẽ tạo mới với đúng kiểu
		return false
	}
	return dataType == expectedType
}
