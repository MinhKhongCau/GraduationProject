package db

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"forum-service/configs"
)

var DB *gorm.DB

// ConnectDB opens the GORM/Postgres connection. Schema is entirely owned by
// golang-migrate (see migrate.go) — AutoMigrate is intentionally not used
// here (SPEC.md §6.1): comments.path is an `ltree` column with no native
// GORM type, and the `ltree` extension must exist before that table does.
func ConnectDB(cfg *configs.Config) *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode)

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("forum-service: failed to connect to database: %v", err)
	}

	DB = database
	return DB
}
