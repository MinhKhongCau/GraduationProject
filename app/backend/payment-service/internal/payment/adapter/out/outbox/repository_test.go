package outbox

import (
	"strings"
	"testing"

	"payment-service/internal/domain/entity"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestEligibleEventsQueryUsesStateTimeOrderingAndBatchLimit(t *testing.T) {
	db := dryRunPostgres(t)
	var events []entity.OutboxEvent
	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return eligibleEventsQuery(tx, 1234, 17).Find(&events)
	})
	normalized := strings.ToLower(strings.Join(strings.Fields(sql), " "))
	for _, fragment := range []string{
		"status = 'pending'",
		"status = 'retry_wait'",
		"next_attempt_at <= 1234",
		"order by created_at asc,id asc",
		"limit 17",
	} {
		if !strings.Contains(normalized, fragment) {
			t.Fatalf("query missing %q: %s", fragment, normalized)
		}
	}
	if strings.Contains(normalized, "delivered") || strings.Contains(normalized, "dead") {
		t.Fatalf("terminal status leaked into polling query: %s", normalized)
	}
}

func TestRecordAttemptReloadsEventWithSelectForUpdate(t *testing.T) {
	db := dryRunPostgres(t)
	var event entity.OutboxEvent
	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return outboxEventForUpdateQuery(tx, [16]byte{1}).First(&event)
	})
	normalized := strings.ToLower(strings.Join(strings.Fields(sql), " "))
	if !strings.Contains(normalized, "for update") {
		t.Fatalf("expected SELECT FOR UPDATE, got %s", normalized)
	}
}

func dryRunPostgres(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=localhost user=test dbname=test sslmode=disable"}), &gorm.Config{
		DryRun:               true,
		DisableAutomaticPing: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	return db
}
