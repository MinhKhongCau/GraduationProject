package database

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestV1PaymentLifecycleMigration(t *testing.T) {
	sql := strings.ToLower(readPaymentServiceFile(t, "migrate_v1.sql"))
	for _, required := range []string{
		"begin;",
		"add column if not exists expires_at bigint",
		"where status = 1",
		"and expires_at is null",
		"set status = 4",
		"drop constraint if exists payment_orders_status_check",
		"drop constraint if exists chk_payment_orders_status",
		"check (status in (1, 2, 3, 4))",
		"chk_payment_orders_pending_expiry",
		"ux_payment_orders_active_pending_appointment",
		"where appointment_id is not null and status = 1",
		"ux_payment_orders_success_appointment",
		"where appointment_id is not null and status = 2",
		"ix_payment_orders_appointment_id",
		"having count(*) > 1",
		"do $$",
		"raise exception",
		"duplicate success orders",
		"commit;",
	} {
		if !strings.Contains(sql, required) {
			t.Fatalf("V1 migration missing %q", required)
		}
	}
	if strings.Contains(sql, "set status = 4\nwhere status = 2") || strings.Contains(sql, "set status = 4\nwhere status = 3") {
		t.Fatal("migration must not expire historical SUCCESS or FAILED rows")
	}
	assertSQLOrder(t, sql,
		"add column if not exists expires_at bigint",
		"add constraint chk_payment_orders_status",
		"and expires_at is null",
		"a partially deployed v1 application",
		"do $$",
		"create unique index if not exists ux_payment_orders_active_pending_appointment",
		"create unique index if not exists ux_payment_orders_success_appointment",
		"add constraint chk_payment_orders_pending_expiry",
		"commit;",
	)
}

func TestStartupMigrationDoesNotRewritePaymentLifecycleState(t *testing.T) {
	source := strings.ToLower(readPaymentServiceFile(t, filepath.Join("pkg", "database", "postgres.go")))
	if !strings.Contains(source, "&entity.paymentorder{}") {
		t.Fatal("PaymentOrder must remain registered with AutoMigrate")
	}
	for _, forbidden := range []string{"update payment_orders", "drop index", "drop constraint", "orderstatusexpired"} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("startup migration contains lifecycle mutation %q", forbidden)
		}
	}
}

func readPaymentServiceFile(t *testing.T, relativePath string) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate migration test")
	}
	root := filepath.Join(filepath.Dir(filename), "..", "..")
	contents, err := os.ReadFile(filepath.Join(root, relativePath))
	if err != nil {
		t.Fatal(err)
	}
	return string(contents)
}

func assertSQLOrder(t *testing.T, sql string, fragments ...string) {
	t.Helper()
	previous := -1
	for _, fragment := range fragments {
		position := strings.Index(sql, fragment)
		if position < 0 {
			t.Fatalf("SQL missing ordered fragment %q", fragment)
		}
		if position <= previous {
			t.Fatalf("SQL fragment %q is out of order", fragment)
		}
		previous = position
	}
}
