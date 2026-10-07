package payment

import (
	"context"
	"errors"
	"payment-service/internal/domain/money"
	paymentdomain "payment-service/internal/domain/payment"
	gateway "payment-service/internal/infrastructure/client/vnpay"
	"testing"
	"time"

	"github.com/google/uuid"
)

type statusReadingBookingClient struct {
	fakeBookingClient
	status string
	err    error
	calls  []string
}

func (c *statusReadingBookingClient) GetAppointmentStatus(ctx context.Context, appointmentID string) (string, error) {
	c.calls = append(c.calls, appointmentID)
	return c.status, c.err
}

func newReturnTestRepo(appointmentID *uuid.UUID) *fakePaymentRepo {
	repo := newFakePaymentRepo(paymentdomain.PaymentOrder{
		ID:               uuid.New(),
		PayerID:          uuid.New(),
		ExpertID:         uuid.New(),
		GrossAmount:      money.Money(1000),
		CommissionRate:   0.15,
		CommissionAmount: money.Money(150),
		NetAmount:        money.Money(850),
		Gateway:          "VNPAY",
		Status:           paymentdomain.OrderStatusPending,
		ExpiresAt:        time.Now().Add(5 * time.Minute).UnixMilli(),
		AppointmentID:    appointmentID,
	})
	repo.addParticipant(&fakeWalletUsecase{})
	return repo
}

func TestProcessReturnSettlesOrderAndReportsBookingStatus(t *testing.T) {
	appointmentID := uuid.New()
	repo := newReturnTestRepo(&appointmentID)
	booking := &statusReadingBookingClient{status: "PENDING_PAYMENT"}
	usecase := NewUsecase(repo, repo, gateway.NewVNPayClient("", "", "", ""), booking).(ReturnUsecase)

	result, err := usecase.ProcessReturn(context.Background(), signedSuccessIPNParams(repo.order.ID))
	if err != nil {
		t.Fatalf("ProcessReturn returned error: %v", err)
	}
	if result.PaymentStatus != "SUCCESS" || result.BookingStatus != "PENDING_PAYMENT" || result.ResponseCode != "00" {
		t.Fatalf("unexpected result: %+v", result)
	}
	if result.OrderID != repo.order.ID.String() || result.AppointmentID != appointmentID.String() || result.AmountVND != 1000 {
		t.Fatalf("unexpected identifiers/amount: %+v", result)
	}
	if result.PaidAt == nil {
		t.Fatal("expected paid_at to be set")
	}
	if len(booking.calls) != 1 || booking.calls[0] != appointmentID.String() {
		t.Fatalf("expected one booking status read, got %v", booking.calls)
	}
	assertOnlyBookingConfirmOutbox(t, repo)
}

func TestProcessReturnIsIdempotentAfterIPN(t *testing.T) {
	appointmentID := uuid.New()
	repo := newReturnTestRepo(&appointmentID)
	booking := &statusReadingBookingClient{status: "CONFIRMED"}
	uc := NewUsecase(repo, repo, gateway.NewVNPayClient("", "", "", ""), booking)
	params := signedSuccessIPNParams(repo.order.ID)

	if _, err := uc.ProcessIPN(context.Background(), params); err != nil {
		t.Fatalf("ProcessIPN returned error: %v", err)
	}
	result, err := uc.(ReturnUsecase).ProcessReturn(context.Background(), params)
	if err != nil {
		t.Fatalf("ProcessReturn returned error: %v", err)
	}
	if result.PaymentStatus != "SUCCESS" || result.BookingStatus != "CONFIRMED" {
		t.Fatalf("unexpected result: %+v", result)
	}
	assertOnlyBookingConfirmOutbox(t, repo)
}

func TestProcessReturnReportsFailedPayment(t *testing.T) {
	appointmentID := uuid.New()
	repo := newReturnTestRepo(&appointmentID)
	booking := &statusReadingBookingClient{status: "CANCELLED"}
	usecase := NewUsecase(repo, repo, gateway.NewVNPayClient("", "", "", ""), booking).(ReturnUsecase)

	result, err := usecase.ProcessReturn(context.Background(), signedFailedIPNParams(repo.order.ID))
	if err != nil {
		t.Fatalf("ProcessReturn returned error: %v", err)
	}
	if result.PaymentStatus != "FAILED" || result.BookingStatus != "CANCELLED" || result.ResponseCode != "24" || result.PaidAt != nil {
		t.Fatalf("unexpected result: %+v", result)
	}
	assertOnlyBookingFailOutbox(t, repo)
}

func TestProcessReturnRejectsInvalidSignature(t *testing.T) {
	repo := newReturnTestRepo(nil)
	usecase := NewUsecase(repo, repo, gateway.NewVNPayClient("", "", "", ""), &fakeBookingClient{}).(ReturnUsecase)
	params := signedSuccessIPNParams(repo.order.ID)
	params["vnp_Amount"] = []string{"999999"}

	if _, err := usecase.ProcessReturn(context.Background(), params); err == nil {
		t.Fatal("expected checksum error")
	}
	if repo.order.Status != paymentdomain.OrderStatusPending {
		t.Fatalf("order must stay PENDING on invalid signature, got %s", repo.order.Status)
	}
}

func TestProcessReturnFallsBackToUnknownBookingStatus(t *testing.T) {
	appointmentID := uuid.New()
	repo := newReturnTestRepo(&appointmentID)
	booking := &statusReadingBookingClient{err: errors.New("booking unavailable")}
	usecase := NewUsecase(repo, repo, gateway.NewVNPayClient("", "", "", ""), booking).(ReturnUsecase)

	result, err := usecase.ProcessReturn(context.Background(), signedSuccessIPNParams(repo.order.ID))
	if err != nil {
		t.Fatalf("ProcessReturn returned error: %v", err)
	}
	if result.PaymentStatus != "SUCCESS" || result.BookingStatus != BookingStatusUnknown {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestProcessReturnOmitsBookingStatusForOrderWithoutAppointment(t *testing.T) {
	repo := newReturnTestRepo(nil)
	booking := &statusReadingBookingClient{status: "CONFIRMED"}
	usecase := NewUsecase(repo, repo, gateway.NewVNPayClient("", "", "", ""), booking).(ReturnUsecase)

	result, err := usecase.ProcessReturn(context.Background(), signedSuccessIPNParams(repo.order.ID))
	if err != nil {
		t.Fatalf("ProcessReturn returned error: %v", err)
	}
	if result.BookingStatus != "" || result.AppointmentID != "" || len(booking.calls) != 0 {
		t.Fatalf("expected no booking lookup, got result=%+v calls=%v", result, booking.calls)
	}
}
