package database

import (
	"fmt"
	"log"
	"payment-service/internal/config"
	"payment-service/internal/domain/entity"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	cfg := config.AppConfig
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

	fmt.Println("⏳ Running AutoMigrate...")
	err = database.AutoMigrate(
		&entity.BankAccount{},
		&entity.Wallet{},
		&entity.WalletTransaction{},
		&entity.PaymentOrder{},
		&entity.PaymentCompensationCase{},
		&entity.WithdrawalRequest{},
		&entity.OutboxEvent{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	fmt.Println("✅ Database migrated successfully!")
	DB = database
}
