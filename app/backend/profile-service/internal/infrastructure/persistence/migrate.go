// File: internal/infrastructure/persistence/migrate.go
package persistence

import (
	"embed"
	"errors"
	"fmt"
	"net/url"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// DBConfig là thông tin kết nối Postgres dùng cho migration.
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// RunMigrations chạy mọi migration golang-migrate còn thiếu (xem migrations/), nhúng sẵn
// trong binary. Gọi mỗi lần khởi động là an toàn: version đã chạy được lưu ở bảng
// schema_migrations, "không có gì để chạy" không bị coi là lỗi.
func RunMigrations(cfg DBConfig) error {
	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("load embedded migrations: %w", err)
	}

	sslMode := cfg.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}
	// x-multi-statement: mỗi file migration có nhiều câu lệnh phân tách bởi ";".
	dsn := (&url.URL{
		Scheme:   "pgx5",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Path:     "/" + cfg.Name,
		RawQuery: url.Values{"sslmode": {sslMode}, "x-multi-statement": {"true"}}.Encode(),
	}).String()

	m, err := migrate.NewWithSourceInstance("iofs", source, dsn)
	if err != nil {
		return fmt.Errorf("init migrator: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}
