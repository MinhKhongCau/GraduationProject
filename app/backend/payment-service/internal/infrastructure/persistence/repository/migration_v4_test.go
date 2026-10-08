package repository

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	paymentdomain "payment-service/internal/domain/payment"
)

func TestV4MigrationIndexesScopedQueriesAndAllowsAdminReview(t *testing.T) {
	sql := readV4Migration(t)
	for _, fragment := range []string{
		"BEGIN;",
		"IX_PAYMENT_ORDERS_EXPERT_CREATED",
		"ON PAYMENT_ORDERS (EXPERT_ID, CREATED_AT DESC)",
		"IX_PAYMENT_ORDERS_PAYER_CREATED",
		"ON PAYMENT_ORDERS (PAYER_ID, CREATED_AT DESC)",
		"IX_PAYMENT_WALLETS_USER_ID",
		"CHK_PAYMENT_COMPENSATION_CASES_RESOLVED_TIME",
		"COMMIT;",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("migration missing %q", fragment)
		}
	}
	// Mọi giá trị domain mới phải được CHECK constraint chấp nhận.
	for _, value := range []string{
		string(paymentdomain.CompensationAdminReview),
		string(paymentdomain.CompensationResolved),
		string(paymentdomain.CompensationReasonAdminManualReview),
		string(paymentdomain.CompensationReasonAdminRefundRequest),
	} {
		if !strings.Contains(sql, "'"+value+"'") {
			t.Fatalf("migration CHECK constraints do not allow %q", value)
		}
	}
	if strings.Contains(sql, "UPDATE PAYMENT_ORDERS") || strings.Contains(sql, "DELETE FROM") {
		t.Fatal("V4 must not rewrite payment history")
	}
}

func readV4Migration(t *testing.T) string {
	t.Helper()
	_, filename, _, _ := runtime.Caller(0)
	content, err := os.ReadFile(filepath.Join(filepath.Dir(filename), "..", "..", "..", "..", "migrate_v4.sql"))
	if err != nil {
		t.Fatal(err)
	}
	return strings.ReplaceAll(strings.ToUpper(string(content)), "\r\n", "\n")
}
