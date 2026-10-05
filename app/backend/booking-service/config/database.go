// config/database.go
package config

import (
	appointmentdomain "booking-service/internal/domain/appointment"
	scheduledomain "booking-service/internal/domain/schedule"
	slotdomain "booking-service/internal/domain/slot"
	timeoffdomain "booking-service/internal/domain/timeoff"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	_ "time/tzdata"
)

// Khai báo một biến toàn cục để chứa kết nối DB
var DB *gorm.DB

// ConnectDB khởi tạo kết nối đến PostgreSQL
func ConnectDB() {
	cfg := AppConfig
	host := cfg.DBHost
	port := cfg.DBPort
	user := cfg.DBUser
	password := cfg.DBPassword
	dbname := cfg.DBName
	sslmode := cfg.DBSSLMode

	log.Printf("Connecting to database: host=%s user=%s password=%s dbname=%s port=%s sslmode=%s", host, user, password, dbname, port, sslmode)

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
		&scheduledomain.TimeTemplate{},
		&scheduledomain.Availability{},
		&timeoffdomain.ExpertTimeOff{},
		&slotdomain.ExpertSlot{},
		&appointmentdomain.Appointment{},
		&appointmentdomain.MedicalRecord{},
		&appointmentdomain.Review{},
	)

	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	fmt.Println("✅ Database migrated successfully!")
	seedDefaultTemplates(database)
	DB = database
}

// seedDefaultTemplates tự động chèn các ca mẫu mặc định nếu DB trống
func seedDefaultTemplates(db *gorm.DB) {
	var count int64
	db.Model(&scheduledomain.TimeTemplate{}).Count(&count)
	if count == 0 {
		fmt.Println("🌱 Seeding default Time Templates...")
		templates := []scheduledomain.TimeTemplate{
			{
				TemplateID:          "11111111-1111-1111-1111-111111111111",
				ShiftName:           "Ca Sáng (08h-12h)",
				StartTime:           "08:00",
				EndTime:             "12:00",
				SlotDurationMinutes: 60,
				IsActive:            true,
			},
			{
				TemplateID:          "22222222-2222-2222-2222-222222222222",
				ShiftName:           "Ca Chiều (13h-17h)",
				StartTime:           "13:00",
				EndTime:             "17:00",
				SlotDurationMinutes: 60,
				IsActive:            true,
			},
		}
		if err := db.Create(&templates).Error; err != nil {
			log.Printf("⚠️  Failed to seed default templates: %v", err)
		} else {
			fmt.Println("✅ Default Time Templates seeded successfully!")
		}
	}

	// Seed Expert Availabilities for the default test expert
	var countAvail int64
	if err := db.Model(&scheduledomain.Availability{}).Count(&countAvail).Error; err == nil && countAvail == 0 {
		fmt.Println("🌱 Seeding default Expert Availabilities...")
		avails := []scheduledomain.Availability{
			{
				AvailabilityID: "33333333-3333-3333-3333-333333333333",
				ExpertID:       "ce7b23b0-6b42-4e71-a482-84a8b0839422",
				TemplateID:     "11111111-1111-1111-1111-111111111111",
				DayOfWeek:      1, // Monday
				IsEnabled:      true,
				EffectiveFrom:  1719680400000,
			},
			{
				AvailabilityID: "44444444-4444-4444-4444-444444444444",
				ExpertID:       "ce7b23b0-6b42-4e71-a482-84a8b0839422",
				TemplateID:     "11111111-1111-1111-1111-111111111111",
				DayOfWeek:      2, // Tuesday
				IsEnabled:      true,
				EffectiveFrom:  1719680400000,
			},
			{
				AvailabilityID: "55555555-5555-5555-5555-555555555555",
				ExpertID:       "ce7b23b0-6b42-4e71-a482-84a8b0839422",
				TemplateID:     "22222222-2222-2222-2222-222222222222",
				DayOfWeek:      3, // Wednesday
				IsEnabled:      true,
				EffectiveFrom:  1719680400000,
			},
		}
		if err := db.Create(&avails).Error; err != nil {
			log.Printf("⚠️  Failed to seed default availabilities: %v", err)
		} else {
			fmt.Println("✅ Default Expert Availabilities seeded successfully!")
		}
	}
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
		    WHEN 'UNAVAILABLE' THEN 3
		    WHEN '0' THEN 0
		    WHEN '1' THEN 1
		    WHEN '2' THEN 2
		    WHEN '3' THEN 3
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
		    WHEN 'COMPLETED'       THEN 3
		    WHEN '0' THEN 0
		    WHEN '1' THEN 1
		    WHEN '2' THEN 2
		    WHEN '3' THEN 3
		    ELSE 0
		  END
	`)
	db.Exec(`ALTER TABLE "Booking_Appointments" ALTER COLUMN status SET DEFAULT 0`)

	// 2b. Xoá bỏ unique constraint trên slot_id của Booking_Appointments
	db.Exec(`ALTER TABLE "Booking_Appointments" DROP CONSTRAINT IF EXISTS "uni_Booking_Appointments_slot_id"`)

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

	silentDB := db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)})

	for _, tc := range timestampColumns {
		// Bỏ qua lỗi vì nếu cột đã là bigint, việc cast sang timestamp sẽ gây lỗi (đúng như ý muốn để skip)
		silentDB.Exec(fmt.Sprintf(`
			ALTER TABLE "%s"
			  ALTER COLUMN "%s" TYPE bigint
			  USING EXTRACT(EPOCH FROM "%s"::timestamp with time zone)::bigint * 1000
		`, tc.table, tc.col, tc.col))
	}

	fmt.Println("✅ Pre-migrations completed.")
}
