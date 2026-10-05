package repository

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestV3MigrationCreatesAndBackfillsCompensationStateSafely(t *testing.T) {
	sql := readV3Migration(t)
	for _, fragment := range []string{
		"BEGIN;",
		"ADD COLUMN IF NOT EXISTS FULFILLMENT_STATUS",
		"ADD COLUMN IF NOT EXISTS GATEWAY_CAPTURE_STATUS",
		"ADD COLUMN IF NOT EXISTS GATEWAY_RESPONSE_CODE",
		"ADD COLUMN IF NOT EXISTS GATEWAY_TRANSACTION_STATUS",
		"ADD COLUMN IF NOT EXISTS GATEWAY_PAYMENT_DATE",
		"ADD COLUMN IF NOT EXISTS TERMINAL_REASON_CODE",
		"CREATE TABLE IF NOT EXISTS PAYMENT_COMPENSATION_CASES",
		"ADD COLUMN IF NOT EXISTS GATEWAY_ORDER_REFERENCE",
		"ADD COLUMN IF NOT EXISTS GATEWAY_TRANSACTION_NUMBER",
		"UX_PAYMENT_COMPENSATION_CASES_ORDER_REASON",
		"STATUS = 2 THEN 'CAPTURED'",
		"STATUS = 3 THEN 'FAILED'",
		"SET FULFILLMENT_STATUS = 'PENDING'",
		"EVENT.EVENT_TYPE = 'BOOKING.APPOINTMENT.CONFIRM'",
		"EVENT.STATUS = 'DELIVERED'",
		"SET FULFILLMENT_STATUS = 'BOOKING_CONFIRMED'",
		"SET FULFILLMENT_STATUS = 'BOOKING_FAILED'",
		"EVENT.TERMINAL_REASON_CODE IN ('CONFLICT', 'NOT_FOUND')",
		"ELSE 'MANUAL_REVIEW'",
		"ALTER COLUMN FULFILLMENT_STATUS SET NOT NULL",
		"ALTER COLUMN GATEWAY_CAPTURE_STATUS SET NOT NULL",
		"ALTER COLUMN GATEWAY_ORDER_REFERENCE SET NOT NULL",
		"COMMIT;",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("migration missing %q", fragment)
		}
	}

	addColumns := strings.Index(sql, "ADD COLUMN IF NOT EXISTS FULFILLMENT_STATUS")
	createTable := strings.Index(sql, "CREATE TABLE IF NOT EXISTS PAYMENT_COMPENSATION_CASES")
	validateDuplicates := strings.Index(sql, "CANNOT CREATE COMPENSATION IDEMPOTENCY INDEX")
	uniqueIndex := strings.Index(sql, "CREATE UNIQUE INDEX IF NOT EXISTS UX_PAYMENT_COMPENSATION_CASES_ORDER_REASON")
	backfillPending := strings.Index(sql, "SET FULFILLMENT_STATUS = 'PENDING'")
	insertCases := strings.Index(sql, "INSERT INTO PAYMENT_COMPENSATION_CASES")
	setNotNull := strings.Index(sql, "ALTER COLUMN FULFILLMENT_STATUS SET DEFAULT 'PENDING'")
	if !(addColumns >= 0 && createTable > addColumns && validateDuplicates > createTable && uniqueIndex > validateDuplicates && backfillPending > uniqueIndex && insertCases > backfillPending && setNotNull > insertCases) {
		t.Fatal("migration must add schema, validate case uniqueness, create idempotency protection, backfill, then apply defaults and NOT NULL constraints")
	}
	if strings.Contains(sql, "SET STATUS = 2") || strings.Contains(sql, "SET STATUS = 3") || strings.Contains(sql, "SET STATUS = 4") {
		t.Fatal("V3 must not rewrite payment gateway status history")
	}
	if strings.Contains(sql, "FOREIGN KEY") || strings.Contains(sql, "REFERENCES PAYMENT_") {
		t.Fatal("V3 must not introduce an unverified foreign key")
	}
}

func TestV3MigrationPreservesTrustedGatewayEvidence(t *testing.T) {
	sql := readV3Migration(t)
	for _, fragment := range []string{
		"GATEWAY_ORDER_REFERENCE",
		"GATEWAY_TRANSACTION_NUMBER",
		"GATEWAY_RESPONSE_CODE",
		"GATEWAY_TRANSACTION_STATUS",
		"GATEWAY_PAYMENT_DATE",
		"PAYMENT.ID::TEXT",
		"PAYMENT.GATEWAY_TXN_REF",
		"PAYMENT.GATEWAY_RESPONSE_CODE",
		"PAYMENT.GATEWAY_TRANSACTION_STATUS",
		"PAYMENT.GATEWAY_PAYMENT_DATE",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("migration does not preserve trusted gateway evidence %q", fragment)
		}
	}
	for _, forbidden := range []string{"VNP_SECUREHASH", "SECURE_HASH", "RAW_QUERY", "AUTHORIZATION_HEADER"} {
		if strings.Contains(sql, forbidden) {
			t.Fatalf("migration stores forbidden evidence %q", forbidden)
		}
	}
}

func TestV3MigrationKeepsUnknownSuccessfulFulfillmentPending(t *testing.T) {
	sql := readV3Migration(t)
	initialPending := strings.Index(sql, "SET FULFILLMENT_STATUS = 'PENDING'\nWHERE FULFILLMENT_STATUS IS NULL")
	deliveredBackfill := strings.Index(sql, "SET FULFILLMENT_STATUS = 'BOOKING_CONFIRMED'")
	deadBackfill := strings.Index(sql, "SET FULFILLMENT_STATUS = CASE")
	if !(initialPending >= 0 && initialPending < deliveredBackfill && deliveredBackfill < deadBackfill) {
		t.Fatal("legacy rows must start PENDING and only advance when delivered/dead outbox evidence exists")
	}
	if !strings.Contains(sql, "PAYMENT.STATUS = 2") || !strings.Contains(sql, "PAYMENT.FULFILLMENT_STATUS = 'PENDING'") {
		t.Fatal("terminal backfill must be limited to successful payments still lacking stronger fulfillment evidence")
	}
	if !strings.Contains(sql, "ON CONFLICT DO NOTHING") {
		t.Fatal("case backfill must be rerunnable")
	}
}

func readV3Migration(t *testing.T) string {
	t.Helper()
	_, filename, _, _ := runtime.Caller(0)
	content, err := os.ReadFile(filepath.Join(filepath.Dir(filename), "..", "..", "..", "..", "migrate_v3.sql"))
	if err != nil {
		t.Fatal(err)
	}
	return strings.ReplaceAll(strings.ToUpper(string(content)), "\r\n", "\n")
}
