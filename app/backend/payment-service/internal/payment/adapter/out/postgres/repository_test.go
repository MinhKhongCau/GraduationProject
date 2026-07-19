package paymentpostgres

import (
	"errors"
	"testing"

	apppayment "payment-service/internal/payment/application"

	"github.com/jackc/pgx/v5/pgconn"
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
