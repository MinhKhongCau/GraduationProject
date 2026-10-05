package config

import (
	"fmt"
	"log"
	paymentdomain "payment-service/internal/domain/payment"
	walletdomain "payment-service/internal/domain/wallet"
	withdrawaldomain "payment-service/internal/domain/withdrawal"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

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

	fmt.Println("⏳ Running AutoMigrate...")
	err = database.AutoMigrate(
		&withdrawaldomain.BankAccount{},
		&walletdomain.Wallet{},
		&walletdomain.WalletTransaction{},
		&paymentdomain.PaymentOrder{},
		&paymentdomain.PaymentCompensationCase{},
		&withdrawaldomain.WithdrawalRequest{},
		&paymentdomain.OutboxEvent{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	fmt.Println("✅ Database migrated successfully!")
	DB = database
}
