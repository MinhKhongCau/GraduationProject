package domain

import (
	"testing"
	"time"
)

func TestPaymentExpiryUsesEarlierBound(t *testing.T) {
	now := time.UnixMilli(1_000_000)
	if got := PaymentExpiry(now, 15*time.Minute, now.Add(5*time.Minute).UnixMilli()); got != now.Add(5*time.Minute).UnixMilli() {
		t.Fatalf("booking expiry should win, got %d", got)
	}
	if got := PaymentExpiry(now, 5*time.Minute, now.Add(15*time.Minute).UnixMilli()); got != now.Add(5*time.Minute).UnixMilli() {
		t.Fatalf("TTL expiry should win, got %d", got)
	}
}

func TestPaymentExpiryUsesVNPayWholeSecondPrecision(t *testing.T) {
	now := time.UnixMilli(1_000_123)
	bookingExpiry := now.Add(5*time.Minute + 456*time.Millisecond).UnixMilli()
	got := PaymentExpiry(now, 15*time.Minute, bookingExpiry)
	if got%1000 != 0 || got > bookingExpiry {
		t.Fatalf("expiry must be whole-second and not exceed booking: got=%d booking=%d", got, bookingExpiry)
	}
}

func TestTTLExpiryWithMillisecondsIsNormalized(t *testing.T) {
	now := time.UnixMilli(1_000_789)
	rawTTLExpiry := now.Add(15 * time.Minute).UnixMilli()
	got := PaymentExpiry(now, 15*time.Minute, now.Add(time.Hour).UnixMilli())
	if got%1000 != 0 || got > rawTTLExpiry {
		t.Fatalf("TTL expiry must be floored to a whole second: got=%d raw=%d", got, rawTTLExpiry)
	}
}

func TestExpiryBoundariesAreDeterministic(t *testing.T) {
	now := time.UnixMilli(1_000_000)
	if !HasUsablePaymentWindow(now, now.Add(time.Minute).UnixMilli(), time.Minute) {
		t.Fatal("exact minimum window must be usable")
	}
	if HasUsablePaymentWindow(now, now.Add(time.Minute-time.Millisecond).UnixMilli(), time.Minute) {
		t.Fatal("window below minimum must be rejected")
	}
	if !IsOrderExpired(now, now.UnixMilli()) {
		t.Fatal("order must expire at the exact expiry instant")
	}
}
