package db

import (
	"embed"
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"forum-service/configs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// RunMigrations applies every pending golang-migrate migration (schema DDL
// and seed data, see internal/repository/db/migrations/) using the SQL
// files embedded into the binary. It runs once at startup, before any GORM
// query — matching the "just `docker compose up`" experience sibling
// services get from GORM AutoMigrate, while giving forum-service the
// versioned, ltree-aware migration sequence SPEC.md §6.1 requires. Safe to
// call on every boot: golang-migrate tracks applied versions in
// `schema_migrations` and treats "nothing to do" as a no-op, not an error.
func RunMigrations(cfg *configs.Config) {
	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		log.Fatalf("forum-service: failed to load embedded migrations: %v", err)
	}

	// x-multi-statement is required because pgx's default (extended) query
	// protocol only supports one statement per Exec — without it, migration
	// files with more than one ;-separated statement (all of ours) fail with
	// confusing parser errors instead of running each statement in turn.
	dsn := fmt.Sprintf("pgx5://%s:%s@%s:%s/%s?sslmode=%s&x-multi-statement=true",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSSLMode)

	m, err := migrate.NewWithSourceInstance("iofs", source, dsn)
	if err != nil {
		log.Fatalf("forum-service: failed to initialize migrator: %v", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("forum-service: migration failed: %v", err)
	}

	log.Println("forum-service: migrations up to date")
}

