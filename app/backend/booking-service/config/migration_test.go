package config

import (
	"os"
	"strings"
	"testing"
)

func TestUnavailableStatusMigrationAndPreMigrationPreserveStatusThree(t *testing.T) {
	migration := readRepositoryFile(t, "../migrate_v0_5_ef.sql")
	for _, required := range []string{
		`DROP CONSTRAINT IF EXISTS chk_booking_expert_slots_status`,
		`ADD CONSTRAINT chk_booking_expert_slots_status`,
		`CHECK (status IN (0, 1, 2, 3))`,
	} {
		if !strings.Contains(migration, required) {
			t.Fatalf("EF migration missing %q", required)
		}
	}

	preMigration := readRepositoryFile(t, "database.go")
	for _, required := range []string{`WHEN 'UNAVAILABLE' THEN 3`, `WHEN '3' THEN 3`} {
		if !strings.Contains(preMigration, required) {
			t.Fatalf("startup pre-migration would not preserve status 3: missing %q", required)
		}
	}
}

func readRepositoryFile(t *testing.T, path string) string {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(contents)
}
