package repository

import (
	"errors"
	"strings"
	"testing"

	apppayment "payment-service/internal/application/payment"
	paymentdomain "payment-service/internal/domain/payment"

	"github.com/google/uuid"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestMapWriteErrorTranslatesPaymentOrderUniqueIndexes(t *testing.T) {
	for _, tc := range []struct {
		constraint string
		want       error
	}{
		{constraint: "ux_payment_orders_active_pending_appointment", want: apppayment.ErrActivePendingOrderExists},
		{constraint: "ux_payment_orders_success_appointment", want: apppayment.ErrAppointmentAlreadyPaid},
	} {
		err := mapWriteError(&pgconn.PgError{Code: "23505", ConstraintName: tc.constraint})
		if !errors.Is(err, tc.want) {
			t.Fatalf("constraint %s mapped to %v", tc.constraint, err)
		}
	}
}

func TestCompensationReadQueryJoinsPaymentState(t *testing.T) {
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=localhost user=test dbname=test sslmode=disable"}), &gorm.Config{
		DryRun: true, DisableAutomaticPing: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	var rows []compensationCaseRow
	caseID := uuid.New()
	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return compensationCaseReadQuery(tx).Where("compensation.id = ?", caseID).Scan(&rows)
	})
	normalized := strings.ToLower(strings.Join(strings.Fields(sql), " "))
	for _, fragment := range []string{
		"from payment_compensation_cases as compensation",
		"join payment_orders as payment on payment.id = compensation.payment_order_id",
		"payment.status as payment_status",
		"payment.gateway_capture_status as payment_gateway_capture_status",
		"payment.fulfillment_status as payment_fulfillment_status",
	} {
		if !strings.Contains(normalized, fragment) {
			t.Fatalf("query missing %q: %s", fragment, normalized)
		}
	}
}

func TestPaymentOrderFilterQueryAppliesRoleScope(t *testing.T) {
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=localhost user=test dbname=test sslmode=disable"}), &gorm.Config{
		DryRun: true, DisableAutomaticPing: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	expertA, expertB := uuid.New(), uuid.New()
	render := func(filter apppayment.PaymentOrderFilter) string {
		var rows []paymentdomain.PaymentOrder
		sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
			return paymentOrderFilterQuery(tx.Model(&paymentdomain.PaymentOrder{}), filter).Find(&rows)
		})
		return strings.ToLower(strings.Join(strings.Fields(sql), " "))
	}

	admin := render(apppayment.PaymentOrderFilter{ScopeExperts: true, ExpertIDs: []uuid.UUID{expertA, expertB}, Type: apppayment.PaymentOrderTypeAppointment})
	for _, fragment := range []string{"expert_id in (", expertA.String(), expertB.String(), "appointment_id is not null"} {
		if !strings.Contains(admin, fragment) {
			t.Fatalf("admin query missing %q: %s", fragment, admin)
		}
	}
	if strings.Contains(admin, "payer_id") {
		t.Fatalf("admin query must not filter by payer when not requested: %s", admin)
	}

	expert := render(apppayment.PaymentOrderFilter{ExpertID: expertA, Type: apppayment.PaymentOrderTypeTopUp})
	if !strings.Contains(expert, "expert_id = '"+expertA.String()+"'") || !strings.Contains(expert, "appointment_id is null") {
		t.Fatalf("expert query is not scoped: %s", expert)
	}
}
