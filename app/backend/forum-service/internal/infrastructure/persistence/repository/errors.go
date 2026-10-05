package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

var ErrNotFound = errors.New("dao: record not found")

// IsForeignKeyViolation reports whether err is a Postgres foreign-key
// constraint violation (SQLSTATE 23503) — e.g. deleting a category that
// still has posts referencing it via ON DELETE RESTRICT (SPEC.md §3.5).
func IsForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23503"
	}
	return false
}
