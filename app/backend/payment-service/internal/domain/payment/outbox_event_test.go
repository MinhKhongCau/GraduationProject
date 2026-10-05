package payment

import (
	"testing"
)

func TestFreshOutboxEventStartsPendingWithZeroAttempts(t *testing.T) {
	event := &OutboxEvent{}
	if err := event.BeforeCreate(nil); err != nil {
		t.Fatal(err)
	}
	if event.Status != OutboxStatusPending || event.AttemptCount != 0 {
		t.Fatalf("unexpected fresh state: status=%s attempts=%d", event.Status, event.AttemptCount)
	}
	if event.CreatedAt == 0 {
		t.Fatal("expected created_at to be initialized")
	}
}
