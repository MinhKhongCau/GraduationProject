package application

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"payment-service/internal/domain/entity"
	"payment-service/internal/domain/vo"
	paymentdomain "payment-service/internal/payment/domain"

	"github.com/google/uuid"
)

func TestCompensationCaseVisibilityFiltersAndExposesSafeGatewayEvidence(t *testing.T) {
	appointmentID, orderID, caseID := uuid.New(), uuid.New(), uuid.New()
	order := lifecycleOrder(appointmentID, uuid.New(), uuid.New(), entity.OrderStatusSuccess, time.Now().UnixMilli())
	order.ID = orderID
	order.GatewayCaptureStatus = paymentdomain.GatewayCaptureSucceeded
	order.FulfillmentStatus = paymentdomain.FulfillmentRefundRequired
	repo := newLifecycleRepo(order)
	repo.compensationCases = []entity.PaymentCompensationCase{
		{
			ID: caseID, PaymentOrderID: orderID, AppointmentID: appointmentID,
			Type: paymentdomain.CompensationBookingFulfillment, Status: paymentdomain.CompensationRefundRequired,
			ReasonCode: paymentdomain.CompensationReasonBookingConflict, SafeReason: "safe",
			GatewayOrderReference: orderID.String(), GatewayTransactionNumber: "vnp-transaction-number",
			GatewayResponseCode: "00", GatewayTransactionStatus: "00", GatewayPaymentDate: "20260719100000",
			AmountVND: vo.Money(250000), CreatedAt: time.Now().UnixMilli(), UpdatedAt: time.Now().UnixMilli(),
		},
		{ID: uuid.New(), PaymentOrderID: uuid.New(), AppointmentID: uuid.New(), Status: paymentdomain.CompensationManualReview},
	}
	usecase := lifecycleUsecase(repo, &fakePaymentGateway{}, &fakeBookingClient{}, time.Now(), time.Minute, time.Second)

	page, err := usecase.ListCompensationCases(context.Background(), CompensationCaseFilter{
		Status: paymentdomain.CompensationRefundRequired, AppointmentID: &appointmentID, PaymentOrderID: &orderID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 1 || page.Page != 1 || page.Size != 20 || len(page.Items) != 1 || page.Items[0].AmountVND != 250000 || !page.Items[0].MoneyPaid || page.Items[0].PaymentStatus != "SUCCESS" {
		t.Fatalf("unexpected page: %+v", page)
	}
	detail, err := usecase.GetCompensationCase(context.Background(), caseID)
	if err != nil || detail.ID != caseID {
		t.Fatalf("unexpected detail: %+v err=%v", detail, err)
	}
	encoded, err := json.Marshal(detail)
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) == "" || !strings.Contains(string(encoded), "vnp-transaction-number") || strings.Contains(string(encoded), "secure_hash") || strings.Contains(string(encoded), "raw_query") {
		t.Fatalf("admin evidence DTO is incomplete or unsafe: %s", encoded)
	}
}

func TestCompensationCaseVisibilityRejectsInvalidFilter(t *testing.T) {
	repo := newLifecycleRepo()
	usecase := lifecycleUsecase(repo, &fakePaymentGateway{}, &fakeBookingClient{}, time.Now(), time.Minute, time.Second)
	_, err := usecase.ListCompensationCases(context.Background(), CompensationCaseFilter{Status: "RESOLVED"})
	if !errors.Is(err, ErrInvalidCompensationFilter) {
		t.Fatalf("expected validation error, got %v", err)
	}
}
