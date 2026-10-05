package payment

import (
	"math"
	"testing"
	"time"
)

func TestOutboxRetryDelay(t *testing.T) {
	tests := []struct {
		name     string
		attempt  int
		expected time.Duration
	}{
		{name: "first uses base", attempt: 1, expected: 5 * time.Second},
		{name: "second doubles", attempt: 2, expected: 10 * time.Second},
		{name: "third doubles", attempt: 3, expected: 20 * time.Second},
		{name: "caps at maximum", attempt: 20, expected: 5 * time.Minute},
		{name: "overflow safe", attempt: math.MaxInt, expected: 5 * time.Minute},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := OutboxRetryDelay(5*time.Second, 5*time.Minute, test.attempt); got != test.expected {
				t.Fatalf("expected %s, got %s", test.expected, got)
			}
		})
	}
}

func TestIsOutboxEligible(t *testing.T) {
	now := int64(1000)
	past, exact, future := int64(999), now, int64(1001)
	tests := []struct {
		name   string
		status OutboxStatus
		next   *int64
		want   bool
	}{
		{name: "pending", status: OutboxStatusPending, want: true},
		{name: "retry past", status: OutboxStatusRetryWait, next: &past, want: true},
		{name: "retry exact", status: OutboxStatusRetryWait, next: &exact, want: true},
		{name: "retry future", status: OutboxStatusRetryWait, next: &future, want: false},
		{name: "retry missing time", status: OutboxStatusRetryWait, want: false},
		{name: "delivered", status: OutboxStatusDelivered, want: false},
		{name: "dead", status: OutboxStatusDead, want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := IsOutboxEligible(test.status, test.next, now); got != test.want {
				t.Fatalf("expected %v, got %v", test.want, got)
			}
		})
	}
}
