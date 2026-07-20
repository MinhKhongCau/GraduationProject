package vnpay

import (
	"net/url"
	"testing"
	"time"
)

func TestGeneratePaymentURLUsesPersistedExpiryInVietnamTimezone(t *testing.T) {
	client := NewVNPayClient("TMN", "SECRET", "https://example.test/pay", "https://example.test/return")
	createdAt := time.Date(2026, 7, 19, 3, 4, 5, 0, time.UTC).UnixMilli()
	expiresAt := time.Date(2026, 7, 19, 3, 19, 5, 0, time.UTC).UnixMilli()

	result := client.GeneratePaymentURL("order-1", 1000, "127.0.0.1", "payment", createdAt, expiresAt)
	parsed, err := url.Parse(result)
	if err != nil {
		t.Fatal(err)
	}
	query := parsed.Query()
	if query.Get("vnp_CreateDate") != "20260719100405" {
		t.Fatalf("unexpected create date: %s", query.Get("vnp_CreateDate"))
	}
	if query.Get("vnp_ExpireDate") != "20260719101905" {
		t.Fatalf("unexpected expire date: %s", query.Get("vnp_ExpireDate"))
	}
	location, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		t.Fatal(err)
	}
	parsedExpiry, err := time.ParseInLocation("20060102150405", query.Get("vnp_ExpireDate"), location)
	if err != nil {
		t.Fatal(err)
	}
	if parsedExpiry.UnixMilli() != expiresAt {
		t.Fatalf("signed VNPay expiry %d differs from persisted expiry %d", parsedExpiry.UnixMilli(), expiresAt)
	}
	if !client.VerifyChecksum(query) {
		t.Fatal("adding vnp_ExpireDate changed checksum correctness")
	}
}

func TestGeneratePaymentURLDoesNotExtendReusedExpiry(t *testing.T) {
	client := NewVNPayClient("TMN", "SECRET", "https://example.test/pay", "https://example.test/return")
	createdAt := time.Date(2026, 7, 19, 3, 4, 5, 0, time.UTC).UnixMilli()
	expiresAt := createdAt + int64(5*time.Minute/time.Millisecond)

	first, _ := url.Parse(client.GeneratePaymentURL("order-1", 1000, "127.0.0.1", "payment", createdAt, expiresAt))
	second, _ := url.Parse(client.GeneratePaymentURL("order-1", 1000, "127.0.0.1", "payment", createdAt, expiresAt))
	if first.Query().Get("vnp_ExpireDate") != second.Query().Get("vnp_ExpireDate") {
		t.Fatal("reused order expiry changed between URL generations")
	}
}
