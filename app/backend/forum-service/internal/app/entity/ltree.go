package entity

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

// LTree represents a PostgreSQL ltree value: a dot-separated hierarchical
// label path (e.g. "1.5.12"), used for nested comments (SPEC.md §2.2/§3.1).
// It implements sql.Scanner/driver.Valuer so GORM can read/write the column
// as a plain string without a dedicated ltree Go driver.
type LTree string

func (l *LTree) Scan(value interface{}) error {
	if value == nil {
		*l = ""
		return nil
	}
	switch v := value.(type) {
	case string:
		*l = LTree(v)
	case []byte:
		*l = LTree(v)
	default:
		return fmt.Errorf("entity: cannot scan type %T into LTree", value)
	}
	return nil
}

func (l LTree) Value() (driver.Value, error) {
	return string(l), nil
}

func (l LTree) String() string {
	return string(l)
}

// Labels splits the path into its individual numeric-id labels, e.g.
// "1.5.12" -> ["1", "5", "12"].
func (l LTree) Labels() []string {
	if l == "" {
		return nil
	}
	return strings.Split(string(l), ".")
}
