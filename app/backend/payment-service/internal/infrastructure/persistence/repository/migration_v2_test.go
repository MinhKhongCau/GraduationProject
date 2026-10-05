package repository

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestV2MigrationDefinesDurableOutboxState(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	path := filepath.Join(filepath.Dir(filename), "..", "..", "..", "..", "migrate_v2.sql")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ReplaceAll(strings.ToUpper(string(content)), "\r\n", "\n")
	for _, fragment := range []string{
		"BEGIN;",
		"ADD COLUMN IF NOT EXISTS STATUS",
		"ADD COLUMN IF NOT EXISTS ATTEMPT_COUNT",
		"ADD COLUMN IF NOT EXISTS NEXT_ATTEMPT_AT",
		"ADD COLUMN IF NOT EXISTS LAST_ATTEMPT_AT",
		"ADD COLUMN IF NOT EXISTS DELIVERED_AT",
		"ADD COLUMN IF NOT EXISTS LAST_ERROR",
		"STATUS IN ('PENDING', 'RETRY_WAIT', 'DELIVERED', 'DEAD')",
		"STATUS <> 'RETRY_WAIT' OR NEXT_ATTEMPT_AT IS NOT NULL",
		"STATUS <> 'DELIVERED' OR DELIVERED_AT IS NOT NULL",
		"WHERE PUBLISHED IS TRUE",
		"PUBLISHED IS NOT TRUE",
		"SET STATUS = 'PENDING'",
		"IX_PAYMENT_OUTBOX_EVENTS_DELIVERY_POLL",
		"COMMIT;",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("migration missing %q", fragment)
		}
	}
	addStatus := strings.Index(sql, "ADD COLUMN IF NOT EXISTS STATUS")
	mapDelivered := strings.Index(sql, "SET STATUS = 'DELIVERED'")
	mapPending := strings.Index(sql, "SET STATUS = 'PENDING'")
	setNotNull := strings.Index(sql, "ALTER COLUMN STATUS SET NOT NULL")
	statusConstraint := strings.Index(sql, "ADD CONSTRAINT CHK_PAYMENT_OUTBOX_EVENTS_STATUS")
	index := strings.Index(sql, "CREATE INDEX IF NOT EXISTS IX_PAYMENT_OUTBOX_EVENTS_DELIVERY_POLL")
	if !(addStatus >= 0 && addStatus < mapDelivered && mapDelivered < mapPending && mapPending < setNotNull && setNotNull < statusConstraint && statusConstraint < index) {
		t.Fatal("migration ordering must add columns, map delivered and pending rows, then apply constraints and the polling index")
	}
}

func TestV2MigrationLegacyMappingIsConservativeAndRerunnable(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	content, err := os.ReadFile(filepath.Join(filepath.Dir(filename), "..", "..", "..", "..", "migrate_v2.sql"))
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ReplaceAll(strings.ToUpper(string(content)), "\r\n", "\n")
	deliveredUpdateStart := strings.Index(sql, "UPDATE PAYMENT_OUTBOX_EVENTS\nSET STATUS = 'DELIVERED'")
	pendingUpdateStart := strings.Index(sql, "UPDATE PAYMENT_OUTBOX_EVENTS\nSET STATUS = 'PENDING'")
	if deliveredUpdateStart < 0 || pendingUpdateStart <= deliveredUpdateStart {
		t.Fatal("legacy mapping updates are missing or out of order")
	}
	deliveredUpdate := sql[deliveredUpdateStart:pendingUpdateStart]
	if !strings.Contains(deliveredUpdate, "PUBLISHED IS TRUE") || strings.Contains(deliveredUpdate, "PUBLISHED IS NOT TRUE") {
		t.Fatal("DELIVERED mapping must be restricted to legacy published=true rows")
	}
	pendingUpdateEnd := strings.Index(sql[pendingUpdateStart:], "UPDATE PAYMENT_OUTBOX_EVENTS\nSET ATTEMPT_COUNT")
	if pendingUpdateEnd < 0 {
		t.Fatal("could not locate end of PENDING mapping")
	}
	pendingUpdate := sql[pendingUpdateStart : pendingUpdateStart+pendingUpdateEnd]
	if !strings.Contains(pendingUpdate, "PUBLISHED IS NOT TRUE") {
		t.Fatal("legacy published=false or NULL rows must map conservatively to PENDING")
	}
	if strings.Count(sql, "ADD COLUMN IF NOT EXISTS") != 6 {
		t.Fatal("every V2 column addition must be rerunnable")
	}
	if !strings.Contains(sql, "CREATE INDEX IF NOT EXISTS") || strings.Count(sql, "DROP CONSTRAINT IF EXISTS") < 4 {
		t.Fatal("index and constraint operations must be rerunnable")
	}
	if !strings.Contains(deliveredUpdate, "STATUS IN ('PENDING', 'DELIVERED')") {
		t.Fatal("AutoMigrate-created PENDING columns and rerun DELIVERED rows must both be handled")
	}
}
