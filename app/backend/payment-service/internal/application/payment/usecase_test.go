package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"payment-service/internal/domain/money"
	paymentdomain "payment-service/internal/domain/payment"
	walletdomain "payment-service/internal/domain/wallet"
	gateway "payment-service/internal/infrastructure/client/vnpay"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCreateOrderCallsPaymentEligibilityAndDerivesExpertAndAmount(t *testing.T) {
	payerID := uuid.New()
	eligibilityExpertID := uuid.New()
	appointmentID := uuid.New().String()
	repo := newFakePaymentRepo(paymentdomain.PaymentOrder{})
	paymentGateway := &fakePaymentGateway{paymentURL: "https://sandbox.vnpayment.vn/paymentv2/vpcpay.html?ok=1"}
	booking := &fakeBookingClient{eligibility: &PaymentEligibility{
		AppointmentID: appointmentID,
		ExpertID:      eligibilityExpertID.String(),
		AmountVND:     250000,
		ExpiresAt:     time.Now().Add(5 * time.Minute).UnixMilli(),
	}}
	usecase := NewUsecase(repo, repo, paymentGateway, booking)

	order, paymentURL, err := usecase.CreateOrder(context.Background(), payerID, appointmentID, "127.0.0.1")
	if err != nil {
		t.Fatalf("CreateOrder returned error: %v", err)
	}

	if booking.eligibilityCalls != 1 {
		t.Fatalf("expected booking eligibility lookup once, got %d", booking.eligibilityCalls)
	}
	if booking.lastAppointmentID != appointmentID {
		t.Fatalf("expected eligibility appointment_id %s, got %s", appointmentID, booking.lastAppointmentID)
	}
	if booking.lastPayerID != payerID {
		t.Fatalf("expected eligibility payer_id %s, got %s", payerID, booking.lastPayerID)
	}
	if repo.createCount != 1 {
		t.Fatalf("expected one payment order to be persisted, got %d", repo.createCount)
	}
	if order.ExpertID != eligibilityExpertID || repo.order.ExpertID != eligibilityExpertID {
		t.Fatalf("expected stored expert_id to be derived from booking, got order=%s repo=%s", order.ExpertID, repo.order.ExpertID)
	}
	if order.GrossAmount != money.Money(250000) || repo.order.GrossAmount != money.Money(250000) {
		t.Fatalf("expected gross amount to be derived from booking price, got order=%d repo=%d", order.GrossAmount, repo.order.GrossAmount)
	}
	if order.AppointmentID == nil || order.AppointmentID.String() != appointmentID {
		t.Fatalf("expected appointment_id %s on stored order, got %v", appointmentID, order.AppointmentID)
	}
	if paymentGateway.generateCalls != 1 {
		t.Fatalf("expected VNPay URL generation once, got %d", paymentGateway.generateCalls)
	}
	if paymentGateway.generatedAmount != 250000 {
		t.Fatalf("expected VNPay URL amount from booking price, got %d", paymentGateway.generatedAmount)
	}
	if paymentURL != paymentGateway.paymentURL {
		t.Fatalf("expected configured VNPay URL %q, got %q", paymentGateway.paymentURL, paymentURL)
	}
}

func TestCreateOrderRejectsEligibilityErrorBeforePersisting(t *testing.T) {
	payerID := uuid.New()
	appointmentID := uuid.New().String()
	repo := newFakePaymentRepo(paymentdomain.PaymentOrder{})
	usecase := NewUsecase(repo, repo, gateway.NewVNPayClient("", "", "", ""), &fakeBookingClient{err: ErrPaymentEligibilityConflict})

	_, _, err := usecase.CreateOrder(context.Background(), payerID, appointmentID, "127.0.0.1")
	if !errors.Is(err, ErrAppointmentInvalidState) {
		t.Fatalf("expected ErrAppointmentInvalidState, got %v", err)
	}
	assertNoPaymentOrderPersisted(t, repo)
}

func TestCreateOrderRejectsMissingAppointmentIDBeforePersisting(t *testing.T) {
	repo := newFakePaymentRepo(paymentdomain.PaymentOrder{})
	usecase := NewUsecase(repo, repo, gateway.NewVNPayClient("", "", "", ""), &fakeBookingClient{})

	_, _, err := usecase.CreateOrder(context.Background(), uuid.New(), "   ", "127.0.0.1")
	if !errors.Is(err, ErrInvalidCreateOrderRequest) {
		t.Fatalf("expected ErrInvalidCreateOrderRequest, got %v", err)
	}
	assertNoPaymentOrderPersisted(t, repo)
}

func TestCreateOrderRejectsMalformedEligibilityResultBeforePersisting(t *testing.T) {
	payerID := uuid.New()
	appointmentID := uuid.New().String()
	tests := []struct {
		name        string
		eligibility PaymentEligibility
		wantErr     error
	}{
		{
			name: "appointment id mismatch",
			eligibility: PaymentEligibility{
				AppointmentID: uuid.New().String(),
				ExpertID:      uuid.New().String(),
				AmountVND:     100000,
				ExpiresAt:     time.Now().Add(5 * time.Minute).UnixMilli(),
			},
			wantErr: ErrInvalidBookingData,
		},
		{
			name: "invalid expert id",
			eligibility: PaymentEligibility{
				AppointmentID: appointmentID,
				ExpertID:      "not-a-uuid",
				AmountVND:     100000,
				ExpiresAt:     time.Now().Add(5 * time.Minute).UnixMilli(),
			},
			wantErr: ErrInvalidBookingData,
		},
		{
			name: "non-positive amount",
			eligibility: PaymentEligibility{
				AppointmentID: appointmentID,
				ExpertID:      uuid.New().String(),
				AmountVND:     0,
				ExpiresAt:     time.Now().Add(5 * time.Minute).UnixMilli(),
			},
			wantErr: ErrInvalidBookingData,
		},
		{
			name: "expired eligibility",
			eligibility: PaymentEligibility{
				AppointmentID: appointmentID,
				ExpertID:      uuid.New().String(),
				AmountVND:     100000,
				ExpiresAt:     time.Now().Add(-time.Minute).UnixMilli(),
			},
			wantErr: ErrAppointmentInvalidState,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakePaymentRepo(paymentdomain.PaymentOrder{})
			eligibility := tt.eligibility
			usecase := NewUsecase(repo, repo, gateway.NewVNPayClient("", "", "", ""), &fakeBookingClient{eligibility: &eligibility})

			_, _, err := usecase.CreateOrder(context.Background(), payerID, appointmentID, "127.0.0.1")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
			assertNoPaymentOrderPersisted(t, repo)
		})
	}
}

func TestCreateOrderRejectsMalformedAppointmentIDBeforePersisting(t *testing.T) {
	repo := newFakePaymentRepo(paymentdomain.PaymentOrder{})
	usecase := NewUsecase(repo, repo, gateway.NewVNPayClient("", "", "", ""), &fakeBookingClient{})

	_, _, err := usecase.CreateOrder(context.Background(), uuid.New(), "not-a-uuid", "127.0.0.1")
	if !errors.Is(err, ErrInvalidCreateOrderRequest) {
		t.Fatalf("expected ErrInvalidCreateOrderRequest, got %v", err)
	}
	assertNoPaymentOrderPersisted(t, repo)
}

func TestCreateOrderRejectsEmptyAppointmentIDBeforePersisting(t *testing.T) {
	tests := []struct {
		name          string
		appointmentID string
	}{
		{name: "missing", appointmentID: ""},
		{name: "whitespace", appointmentID: "   "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakePaymentRepo(paymentdomain.PaymentOrder{})
			usecase := NewUsecase(repo, repo, gateway.NewVNPayClient("", "", "", ""), &fakeBookingClient{})

			_, _, err := usecase.CreateOrder(context.Background(), uuid.New(), tt.appointmentID, "127.0.0.1")
			if !errors.Is(err, ErrInvalidCreateOrderRequest) {
				t.Fatalf("expected ErrInvalidCreateOrderRequest, got %v", err)
			}
			assertNoPaymentOrderPersisted(t, repo)
		})
	}
}

func TestCreateOrderMapsBookingNotFoundBeforePersisting(t *testing.T) {
	appointmentID := uuid.New().String()
	repo := newFakePaymentRepo(paymentdomain.PaymentOrder{})
	usecase := NewUsecase(repo, repo, gateway.NewVNPayClient("", "", "", ""), &fakeBookingClient{err: ErrAppointmentNotFound})

	_, _, err := usecase.CreateOrder(context.Background(), uuid.New(), appointmentID, "127.0.0.1")
	if !errors.Is(err, ErrBookingAppointmentNotFound) {
		t.Fatalf("expected ErrBookingAppointmentNotFound, got %v", err)
	}
	assertNoPaymentOrderPersisted(t, repo)
}

func TestProcessIPNConcurrentRequestsOnlyProcessPendingOrderOnce(t *testing.T) {
	orderID := uuid.New()
	payerID := uuid.New()
	expertID := uuid.New()
	repo := newFakePaymentRepo(paymentdomain.PaymentOrder{
		ID:               orderID,
		PayerID:          payerID,
		ExpertID:         expertID,
		GrossAmount:      money.Money(1000),
		CommissionRate:   0.15,
		CommissionAmount: money.Money(150),
		NetAmount:        money.Money(850),
		Gateway:          "VNPAY",
		Status:           paymentdomain.OrderStatusPending,
	})

	walletUsecase := &fakeWalletUsecase{delay: 50 * time.Millisecond}
	repo.addParticipant(walletUsecase)
	usecase := NewUsecase(
		repo,
		repo,
		gateway.NewVNPayClient("", "", "", ""),
		&fakeBookingClient{},
	)

	params := map[string][]string{
		"vnp_TxnRef":        {orderID.String()},
		"vnp_Amount":        {"100000"},
		"vnp_ResponseCode":  {"00"},
		"vnp_TransactionNo": {"vnp-txn-1"},
	}
	signVNPayParams(params)

	start := make(chan struct{})
	type result struct {
		alreadyProcessed bool
		err              error
	}
	results := make(chan result, 2)

	for i := 0; i < 2; i++ {
		go func() {
			<-start
			alreadyProcessed, err := usecase.ProcessIPN(context.Background(), params)
			results <- result{alreadyProcessed: alreadyProcessed, err: err}
		}()
	}
	close(start)

	var alreadyProcessedCount int
	for i := 0; i < 2; i++ {
		res := <-results
		if res.err != nil {
			t.Fatalf("ProcessIPN returned error: %v", res.err)
		}
		if res.alreadyProcessed {
			alreadyProcessedCount++
		}
	}

	if alreadyProcessedCount != 1 {
		t.Fatalf("expected exactly one alreadyProcessed response, got %d", alreadyProcessedCount)
	}
	if repo.successTransitions != 1 {
		t.Fatalf("expected exactly one PENDING -> SUCCESS transition, got %d", repo.successTransitions)
	}
	if repo.order.Status != paymentdomain.OrderStatusSuccess {
		t.Fatalf("expected final order status SUCCESS, got %s", repo.order.Status.String())
	}
	if len(repo.outboxEvents) != 0 {
		t.Fatalf("expected no outbox events for order without appointment, got %d", len(repo.outboxEvents))
	}
}

func TestProcessIPNCommitsPaymentWalletAndBookingOutboxAtomically(t *testing.T) {
	orderID := uuid.New()
	appointmentID := uuid.New()
	repo := newFakePaymentRepo(paymentdomain.PaymentOrder{
		ID:               orderID,
		PayerID:          uuid.New(),
		ExpertID:         uuid.New(),
		GrossAmount:      money.Money(1000),
		CommissionRate:   0.15,
		CommissionAmount: money.Money(150),
		NetAmount:        money.Money(850),
		Gateway:          "VNPAY",
		Status:           paymentdomain.OrderStatusPending,
		ExpiresAt:        time.Now().Add(5 * time.Minute).UnixMilli(),
		AppointmentID:    &appointmentID,
	})
	walletUsecase := &fakeWalletUsecase{}
	repo.addParticipant(walletUsecase)

	usecase := NewUsecase(repo, repo, gateway.NewVNPayClient("", "", "", ""), &fakeBookingClient{})

	alreadyProcessed, err := usecase.ProcessIPN(context.Background(), signedSuccessIPNParams(orderID))
	if err != nil {
		t.Fatalf("ProcessIPN returned error: %v", err)
	}
	if alreadyProcessed {
		t.Fatal("expected alreadyProcessed to be false")
	}
	if repo.order.Status != paymentdomain.OrderStatusSuccess {
		t.Fatalf("expected order status SUCCESS, got %s", repo.order.Status.String())
	}
	if walletUsecase.pendingBalance != money.Money(850) {
		t.Fatalf("expected wallet pending balance 850, got %d", walletUsecase.pendingBalance)
	}
	if len(walletUsecase.transactions) != 2 {
		t.Fatalf("expected two wallet transactions, got %d", len(walletUsecase.transactions))
	}
	assertOnlyBookingConfirmOutbox(t, repo)
}

func TestProcessIPNRollsBackWhenWalletCreditFails(t *testing.T) {
	orderID := uuid.New()
	appointmentID := uuid.New()
	repo := newFakePaymentRepo(paymentdomain.PaymentOrder{
		ID:               orderID,
		PayerID:          uuid.New(),
		ExpertID:         uuid.New(),
		GrossAmount:      money.Money(1000),
		CommissionRate:   0.15,
		CommissionAmount: money.Money(150),
		NetAmount:        money.Money(850),
		Gateway:          "VNPAY",
		Status:           paymentdomain.OrderStatusPending,
		AppointmentID:    &appointmentID,
	})
	walletUsecase := &fakeWalletUsecase{creditErr: errors.New("credit failed")}
	repo.addParticipant(walletUsecase)

	usecase := NewUsecase(repo, repo, gateway.NewVNPayClient("", "", "", ""), &fakeBookingClient{})

	alreadyProcessed, err := usecase.ProcessIPN(context.Background(), signedSuccessIPNParams(orderID))
	if err == nil {
		t.Fatal("expected wallet credit error")
	}
	if alreadyProcessed {
		t.Fatal("expected alreadyProcessed to be false")
	}
	assertRolledBackPaymentWalletAndOutbox(t, repo, walletUsecase)
}

func TestProcessIPNRollsBackCreditWhenWalletDebitFails(t *testing.T) {
	orderID := uuid.New()
	appointmentID := uuid.New()
	repo := newFakePaymentRepo(paymentdomain.PaymentOrder{
		ID:               orderID,
		PayerID:          uuid.New(),
		ExpertID:         uuid.New(),
		GrossAmount:      money.Money(1000),
		CommissionRate:   0.15,
		CommissionAmount: money.Money(150),
		NetAmount:        money.Money(850),
		Gateway:          "VNPAY",
		Status:           paymentdomain.OrderStatusPending,
		AppointmentID:    &appointmentID,
	})
	walletUsecase := &fakeWalletUsecase{debitErr: errors.New("debit failed")}
	repo.addParticipant(walletUsecase)

	usecase := NewUsecase(repo, repo, gateway.NewVNPayClient("", "", "", ""), &fakeBookingClient{})

	alreadyProcessed, err := usecase.ProcessIPN(context.Background(), signedSuccessIPNParams(orderID))
	if err == nil {
		t.Fatal("expected wallet debit error")
	}
	if alreadyProcessed {
		t.Fatal("expected alreadyProcessed to be false")
	}
	assertRolledBackPaymentWalletAndOutbox(t, repo, walletUsecase)
}

func TestProcessIPNAbortsWhenGatewayTxnLookupReturnsDatabaseError(t *testing.T) {
	orderID := uuid.New()
	repo := newFakePaymentRepo(paymentdomain.PaymentOrder{
		ID:               orderID,
		PayerID:          uuid.New(),
		ExpertID:         uuid.New(),
		GrossAmount:      money.Money(1000),
		CommissionRate:   0.15,
		CommissionAmount: money.Money(150),
		NetAmount:        money.Money(850),
		Gateway:          "VNPAY",
		Status:           paymentdomain.OrderStatusPending,
		ExpiresAt:        time.Now().Add(5 * time.Minute).UnixMilli(),
	})
	repo.gatewayLookupErr = errors.New("database unavailable")

	usecase := NewUsecase(
		repo,
		repo,
		gateway.NewVNPayClient("", "", "", ""),
		&fakeBookingClient{},
	)

	params := map[string][]string{
		"vnp_TxnRef":        {orderID.String()},
		"vnp_Amount":        {"100000"},
		"vnp_ResponseCode":  {"00"},
		"vnp_TransactionNo": {"vnp-txn-1"},
	}
	signVNPayParams(params)
	alreadyProcessed, err := usecase.ProcessIPN(context.Background(), params)
	if err == nil {
		t.Fatal("expected database lookup error")
	}
	if alreadyProcessed {
		t.Fatal("expected alreadyProcessed to be false")
	}
	if repo.successTransitions != 0 {
		t.Fatalf("expected no PENDING -> SUCCESS transition, got %d", repo.successTransitions)
	}
	if len(repo.outboxEvents) != 0 {
		t.Fatalf("expected no outbox events, got %d", len(repo.outboxEvents))
	}
}

func TestProcessIPNRejectsMissingSecureHash(t *testing.T) {
	orderID := uuid.New()
	repo := newFakePaymentRepo(paymentdomain.PaymentOrder{
		ID:               orderID,
		PayerID:          uuid.New(),
		ExpertID:         uuid.New(),
		GrossAmount:      money.Money(1000),
		CommissionRate:   0.15,
		CommissionAmount: money.Money(150),
		NetAmount:        money.Money(850),
		Gateway:          "VNPAY",
		Status:           paymentdomain.OrderStatusPending,
	})

	usecase := NewUsecase(repo, repo, gateway.NewVNPayClient("", "", "", ""), &fakeBookingClient{})

	_, err := usecase.ProcessIPN(context.Background(), map[string][]string{
		"vnp_TxnRef":       {orderID.String()},
		"vnp_Amount":       {"100000"},
		"vnp_ResponseCode": {"00"},
	})
	if err == nil || !strings.Contains(err.Error(), "vnp_SecureHash") {
		t.Fatalf("expected missing secure hash error, got %v", err)
	}
	if repo.successTransitions != 0 {
		t.Fatalf("expected no PENDING -> SUCCESS transition, got %d", repo.successTransitions)
	}
}

func TestProcessIPNRejectsInvalidSecureHash(t *testing.T) {
	orderID := uuid.New()
	repo := newFakePaymentRepo(paymentdomain.PaymentOrder{
		ID:               orderID,
		PayerID:          uuid.New(),
		ExpertID:         uuid.New(),
		GrossAmount:      money.Money(1000),
		CommissionRate:   0.15,
		CommissionAmount: money.Money(150),
		NetAmount:        money.Money(850),
		Gateway:          "VNPAY",
		Status:           paymentdomain.OrderStatusPending,
	})

	usecase := NewUsecase(repo, repo, gateway.NewVNPayClient("", "", "", ""), &fakeBookingClient{})

	_, err := usecase.ProcessIPN(context.Background(), map[string][]string{
		"vnp_TxnRef":       {orderID.String()},
		"vnp_Amount":       {"100000"},
		"vnp_ResponseCode": {"00"},
		"vnp_SecureHash":   {"invalid"},
	})
	if err == nil || !strings.Contains(err.Error(), "checksum") {
		t.Fatalf("expected invalid checksum error, got %v", err)
	}
	if repo.successTransitions != 0 {
		t.Fatalf("expected no PENDING -> SUCCESS transition, got %d", repo.successTransitions)
	}
}

func TestProcessIPNRejectsMalformedAmount(t *testing.T) {
	tests := []struct {
		name   string
		amount *string
	}{
		{name: "missing amount", amount: nil},
		{name: "invalid integer", amount: stringPtr("not-a-number")},
		{name: "zero amount", amount: stringPtr("0")},
		{name: "negative amount", amount: stringPtr("-100")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderID := uuid.New()
			repo := newFakePaymentRepo(paymentdomain.PaymentOrder{
				ID:               orderID,
				PayerID:          uuid.New(),
				ExpertID:         uuid.New(),
				GrossAmount:      money.Money(1000),
				CommissionRate:   0.15,
				CommissionAmount: money.Money(150),
				NetAmount:        money.Money(850),
				Gateway:          "VNPAY",
				Status:           paymentdomain.OrderStatusPending,
			})

			usecase := NewUsecase(repo, repo, gateway.NewVNPayClient("", "", "", ""), &fakeBookingClient{})
			params := map[string][]string{
				"vnp_TxnRef":       {orderID.String()},
				"vnp_ResponseCode": {"00"},
			}
			if tt.amount != nil {
				params["vnp_Amount"] = []string{*tt.amount}
			}
			signVNPayParams(params)

			_, err := usecase.ProcessIPN(context.Background(), params)
			if err == nil || !strings.Contains(err.Error(), "vnp_Amount") {
				t.Fatalf("expected amount error, got %v", err)
			}
			if repo.successTransitions != 0 {
				t.Fatalf("expected no PENDING -> SUCCESS transition, got %d", repo.successTransitions)
			}
		})
	}
}

func TestProcessIPNFailedPaymentWithAppointmentCreatesBookingFailOutbox(t *testing.T) {
	orderID := uuid.New()
	appointmentID := uuid.New()
	repo := newFakePaymentRepo(paymentdomain.PaymentOrder{
		ID:               orderID,
		PayerID:          uuid.New(),
		ExpertID:         uuid.New(),
		GrossAmount:      money.Money(1000),
		CommissionRate:   0.15,
		CommissionAmount: money.Money(150),
		NetAmount:        money.Money(850),
		Gateway:          "VNPAY",
		Status:           paymentdomain.OrderStatusPending,
		ExpiresAt:        time.Now().Add(5 * time.Minute).UnixMilli(),
		AppointmentID:    &appointmentID,
	})
	usecase := NewUsecase(repo, repo, gateway.NewVNPayClient("", "", "", ""), &fakeBookingClient{})

	alreadyProcessed, err := usecase.ProcessIPN(context.Background(), signedFailedIPNParams(orderID))
	if err != nil {
		t.Fatalf("ProcessIPN returned error: %v", err)
	}
	if alreadyProcessed {
		t.Fatal("expected alreadyProcessed to be false")
	}
	if repo.order.Status != paymentdomain.OrderStatusFailed {
		t.Fatalf("expected order FAILED, got %s", repo.order.Status.String())
	}
	assertOnlyBookingFailOutbox(t, repo)
}

func TestProcessIPNFailedPaymentWithoutAppointmentCreatesNoBookingEvent(t *testing.T) {
	orderID := uuid.New()
	repo := newFakePaymentRepo(paymentdomain.PaymentOrder{
		ID:               orderID,
		PayerID:          uuid.New(),
		ExpertID:         uuid.New(),
		GrossAmount:      money.Money(1000),
		CommissionRate:   0.15,
		CommissionAmount: money.Money(150),
		NetAmount:        money.Money(850),
		Gateway:          "VNPAY",
		Status:           paymentdomain.OrderStatusPending,
		ExpiresAt:        time.Now().Add(5 * time.Minute).UnixMilli(),
	})
	usecase := NewUsecase(repo, repo, gateway.NewVNPayClient("", "", "", ""), &fakeBookingClient{})

	if _, err := usecase.ProcessIPN(context.Background(), signedFailedIPNParams(orderID)); err != nil {
		t.Fatalf("ProcessIPN returned error: %v", err)
	}
	if len(repo.outboxEvents) != 0 {
		t.Fatalf("expected no booking event, got %d", len(repo.outboxEvents))
	}
}

func TestProcessIPNInvalidChecksumCreatesNoFailureEvent(t *testing.T) {
	orderID := uuid.New()
	appointmentID := uuid.New()
	repo := newFakePaymentRepo(paymentdomain.PaymentOrder{
		ID:               orderID,
		PayerID:          uuid.New(),
		ExpertID:         uuid.New(),
		GrossAmount:      money.Money(1000),
		CommissionRate:   0.15,
		CommissionAmount: money.Money(150),
		NetAmount:        money.Money(850),
		Gateway:          "VNPAY",
		Status:           paymentdomain.OrderStatusPending,
		ExpiresAt:        time.Now().Add(5 * time.Minute).UnixMilli(),
		AppointmentID:    &appointmentID,
	})
	usecase := NewUsecase(repo, repo, gateway.NewVNPayClient("", "", "", ""), &fakeBookingClient{})

	params := signedFailedIPNParams(orderID)
	params["vnp_SecureHash"] = []string{"invalid"}
	if _, err := usecase.ProcessIPN(context.Background(), params); err == nil {
		t.Fatal("expected invalid checksum error")
	}
	if len(repo.outboxEvents) != 0 {
		t.Fatalf("expected no failure event, got %d", len(repo.outboxEvents))
	}
}

func TestProcessIPNAmountMismatchCreatesNoFailureEvent(t *testing.T) {
	orderID := uuid.New()
	appointmentID := uuid.New()
	repo := newFakePaymentRepo(paymentdomain.PaymentOrder{
		ID:               orderID,
		PayerID:          uuid.New(),
		ExpertID:         uuid.New(),
		GrossAmount:      money.Money(1000),
		CommissionRate:   0.15,
		CommissionAmount: money.Money(150),
		NetAmount:        money.Money(850),
		Gateway:          "VNPAY",
		Status:           paymentdomain.OrderStatusPending,
		ExpiresAt:        time.Now().Add(5 * time.Minute).UnixMilli(),
		AppointmentID:    &appointmentID,
	})
	usecase := NewUsecase(repo, repo, gateway.NewVNPayClient("", "", "", ""), &fakeBookingClient{})

	params := signedFailedIPNParams(orderID)
	params["vnp_Amount"] = []string{"99900"}
	signVNPayParams(params)
	if _, err := usecase.ProcessIPN(context.Background(), params); err == nil {
		t.Fatal("expected amount mismatch error")
	}
	if len(repo.outboxEvents) != 0 {
		t.Fatalf("expected no failure event, got %d", len(repo.outboxEvents))
	}
}

func TestProcessIPNDuplicateFailedPaymentDoesNotCreateDuplicateFailureEvent(t *testing.T) {
	orderID := uuid.New()
	appointmentID := uuid.New()
	repo := newFakePaymentRepo(paymentdomain.PaymentOrder{
		ID:               orderID,
		PayerID:          uuid.New(),
		ExpertID:         uuid.New(),
		GrossAmount:      money.Money(1000),
		CommissionRate:   0.15,
		CommissionAmount: money.Money(150),
		NetAmount:        money.Money(850),
		Gateway:          "VNPAY",
		Status:           paymentdomain.OrderStatusPending,
		ExpiresAt:        time.Now().Add(5 * time.Minute).UnixMilli(),
		AppointmentID:    &appointmentID,
	})
	usecase := NewUsecase(repo, repo, gateway.NewVNPayClient("", "", "", ""), &fakeBookingClient{})
	params := signedFailedIPNParams(orderID)

	alreadyProcessed, err := usecase.ProcessIPN(context.Background(), params)
	if err != nil {
		t.Fatalf("first ProcessIPN returned error: %v", err)
	}
	if alreadyProcessed {
		t.Fatal("expected first IPN to process")
	}
	alreadyProcessed, err = usecase.ProcessIPN(context.Background(), params)
	if err != nil {
		t.Fatalf("second ProcessIPN returned error: %v", err)
	}
	if !alreadyProcessed {
		t.Fatal("expected duplicate IPN to be already processed")
	}
	assertOnlyBookingFailOutbox(t, repo)
}

type fakePaymentRepo struct {
	mu                 sync.Mutex
	rowLock            sync.Mutex
	rowLockHeld        bool
	order              paymentdomain.PaymentOrder
	createCount        int
	outboxEvents       []paymentdomain.OutboxEvent
	compensationCases  []paymentdomain.PaymentCompensationCase
	successTransitions int
	gatewayLookupErr   error
	participants       []fakeTxParticipant
	walletUsecase      *fakeWalletUsecase
}

func newFakePaymentRepo(order paymentdomain.PaymentOrder) *fakePaymentRepo {
	return &fakePaymentRepo{
		order:         order,
		walletUsecase: &fakeWalletUsecase{},
	}
}

type fakeTxParticipant interface {
	beginTx()
	commitTx()
	rollbackTx()
}

func (r *fakePaymentRepo) addParticipant(participant fakeTxParticipant) {
	r.participants = append(r.participants, participant)
	if walletUsecase, ok := participant.(*fakeWalletUsecase); ok {
		r.walletUsecase = walletUsecase
	}
}

func (r *fakePaymentRepo) Create(order *paymentdomain.PaymentOrder) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.createCount++
	r.order = *order
	return nil
}

func (r *fakePaymentRepo) GetByID(orderID uuid.UUID) (*paymentdomain.PaymentOrder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.order.ID != orderID {
		return nil, ErrTxRecordNotFound
	}
	order := r.order
	return &order, nil
}

func (r *fakePaymentRepo) GetOrderForUpdate(ctx context.Context, orderID uuid.UUID) (*paymentdomain.PaymentOrder, error) {
	r.rowLock.Lock()
	r.mu.Lock()
	r.rowLockHeld = true
	defer r.mu.Unlock()

	if r.order.ID != orderID {
		return nil, ErrTxRecordNotFound
	}
	order := r.order
	return &order, nil
}

func (r *fakePaymentRepo) GetByGatewayTxnRef(ref string) (*paymentdomain.PaymentOrder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.gatewayLookupErr != nil {
		return nil, r.gatewayLookupErr
	}
	if r.order.GatewayTxnRef != ref {
		return nil, ErrTxRecordNotFound
	}
	order := r.order
	return &order, nil
}

func (r *fakePaymentRepo) GetSuccessfulByAppointment(ctx context.Context, appointmentID uuid.UUID) (*paymentdomain.PaymentOrder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.order.AppointmentID == nil || *r.order.AppointmentID != appointmentID || r.order.Status != paymentdomain.OrderStatusSuccess {
		return nil, ErrTxRecordNotFound
	}
	order := r.order
	return &order, nil
}

func (r *fakePaymentRepo) GetOrder(ctx context.Context, orderID uuid.UUID) (*paymentdomain.PaymentOrder, error) {
	return r.GetByID(orderID)
}

func (r *fakePaymentRepo) ListOrdersForAppointmentForUpdate(ctx context.Context, appointmentID uuid.UUID) ([]paymentdomain.PaymentOrder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.order.AppointmentID == nil || *r.order.AppointmentID != appointmentID {
		return nil, nil
	}
	return []paymentdomain.PaymentOrder{r.order}, nil
}

func (r *fakePaymentRepo) CreateOrder(ctx context.Context, order *paymentdomain.PaymentOrder) error {
	return r.Create(order)
}

func (r *fakePaymentRepo) GetGatewayTxnRef(ctx context.Context, ref string) (*paymentdomain.PaymentOrder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.gatewayLookupErr != nil {
		return nil, r.gatewayLookupErr
	}
	if r.order.GatewayTxnRef != ref {
		return nil, ErrTxRecordNotFound
	}
	order := r.order
	return &order, nil
}

func (r *fakePaymentRepo) Update(order *paymentdomain.PaymentOrder) error {
	return r.updateOrder(order)
}

func (r *fakePaymentRepo) UpdateOrder(ctx context.Context, order *paymentdomain.PaymentOrder) error {
	return r.updateOrder(order)
}

func (r *fakePaymentRepo) ExpireOtherPendingOrders(ctx context.Context, appointmentID, exceptOrderID uuid.UUID) error {
	return nil
}

func (r *fakePaymentRepo) updateOrder(order *paymentdomain.PaymentOrder) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.order.Status == paymentdomain.OrderStatusPending && order.Status == paymentdomain.OrderStatusSuccess {
		r.successTransitions++
	}
	r.order = *order
	return nil
}

func (r *fakePaymentRepo) SaveOutboxEvent(ctx context.Context, event *paymentdomain.OutboxEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.outboxEvents = append(r.outboxEvents, *event)
	return nil
}

func (r *fakePaymentRepo) SaveCompensationCase(ctx context.Context, compensationCase *paymentdomain.PaymentCompensationCase) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.compensationCases {
		existing := &r.compensationCases[i]
		if existing.PaymentOrderID == compensationCase.PaymentOrderID && existing.ReasonCode == compensationCase.ReasonCode {
			return nil
		}
	}
	r.compensationCases = append(r.compensationCases, *compensationCase)
	return nil
}

func (r *fakePaymentRepo) CreditWalletPending(ctx context.Context, userID uuid.UUID, amount money.Money, refID uuid.UUID, idempotencyKey string) error {
	return r.walletUsecase.CreditPending(ctx, userID, amount, "PAYMENT_ORDER", refID, idempotencyKey)
}

func (r *fakePaymentRepo) DebitWalletPending(ctx context.Context, userID uuid.UUID, amount money.Money, refID uuid.UUID, idempotencyKey string) error {
	return r.walletUsecase.DebitPending(ctx, userID, amount, "PAYMENT_ORDER", refID, idempotencyKey)
}

func (r *fakePaymentRepo) WithinTx(ctx context.Context, fn func(tx Tx) error) error {
	r.mu.Lock()
	orderSnapshot := r.order
	createCountSnapshot := r.createCount
	outboxSnapshot := append([]paymentdomain.OutboxEvent(nil), r.outboxEvents...)
	compensationSnapshot := append([]paymentdomain.PaymentCompensationCase(nil), r.compensationCases...)
	successTransitionsSnapshot := r.successTransitions
	for _, participant := range r.participants {
		participant.beginTx()
	}
	r.mu.Unlock()

	err := fn(r)

	r.mu.Lock()
	if err != nil {
		r.order = orderSnapshot
		r.createCount = createCountSnapshot
		r.outboxEvents = outboxSnapshot
		r.compensationCases = compensationSnapshot
		r.successTransitions = successTransitionsSnapshot
		for _, participant := range r.participants {
			participant.rollbackTx()
		}
	} else {
		for _, participant := range r.participants {
			participant.commitTx()
		}
	}
	held := r.rowLockHeld
	if held {
		r.rowLockHeld = false
	}
	r.mu.Unlock()
	if held {
		r.rowLock.Unlock()
	}
	return err
}

type fakeWalletUsecase struct {
	delay                time.Duration
	creditErr            error
	debitErr             error
	pendingBalance       money.Money
	transactions         []walletdomain.WalletTransaction
	pendingBalanceBefore money.Money
	transactionsBefore   []walletdomain.WalletTransaction
}

func (u *fakeWalletUsecase) GetOrCreateWallet(ctx context.Context, userID uuid.UUID) (*walletdomain.Wallet, error) {
	return &walletdomain.Wallet{ID: uuid.New(), UserID: userID}, nil
}

func (u *fakeWalletUsecase) GetWalletByID(ctx context.Context, walletID uuid.UUID) (*walletdomain.Wallet, error) {
	return &walletdomain.Wallet{ID: walletID}, nil
}

func (u *fakeWalletUsecase) GetTransactionHistory(ctx context.Context, userID uuid.UUID) ([]walletdomain.WalletTransaction, error) {
	return nil, nil
}

func (u *fakeWalletUsecase) CreditPending(ctx context.Context, userID uuid.UUID, amount money.Money, refType string, refID uuid.UUID, idempotencyKey string) error {
	if u.delay > 0 {
		time.Sleep(u.delay)
	}
	if u.creditErr != nil {
		return u.creditErr
	}
	u.pendingBalance = u.pendingBalance.Add(amount)
	u.transactions = append(u.transactions, walletdomain.WalletTransaction{
		Type:           walletdomain.TxTypePaymentReceived,
		Amount:         amount,
		ReferenceType:  refType,
		ReferenceID:    refID,
		IdempotencyKey: idempotencyKey,
	})
	return nil
}

func (u *fakeWalletUsecase) CreditAvailable(ctx context.Context, userID uuid.UUID, amount money.Money, refType string, refID uuid.UUID, idempotencyKey string) error {
	return nil
}

func (u *fakeWalletUsecase) DebitAvailable(ctx context.Context, userID uuid.UUID, amount money.Money, refType string, refID uuid.UUID, idempotencyKey string) error {
	return nil
}

func (u *fakeWalletUsecase) DebitPending(ctx context.Context, userID uuid.UUID, amount money.Money, refType string, refID uuid.UUID, idempotencyKey string) error {
	if u.debitErr != nil {
		return u.debitErr
	}
	u.pendingBalance = u.pendingBalance.Sub(amount)
	u.transactions = append(u.transactions, walletdomain.WalletTransaction{
		Type:           walletdomain.TxTypeCommissionDeducted,
		Amount:         money.Money(-amount.Int64()),
		ReferenceType:  refType,
		ReferenceID:    refID,
		IdempotencyKey: idempotencyKey,
	})
	return nil
}

func (u *fakeWalletUsecase) LockFunds(ctx context.Context, userID uuid.UUID, amount money.Money) error {
	return nil
}

func (u *fakeWalletUsecase) UnlockFunds(ctx context.Context, userID uuid.UUID, amount money.Money) error {
	return nil
}

func (u *fakeWalletUsecase) beginTx() {
	u.pendingBalanceBefore = u.pendingBalance
	u.transactionsBefore = append([]walletdomain.WalletTransaction(nil), u.transactions...)
}

func (u *fakeWalletUsecase) commitTx() {
	u.transactionsBefore = nil
}

func (u *fakeWalletUsecase) rollbackTx() {
	u.pendingBalance = u.pendingBalanceBefore
	u.transactions = append([]walletdomain.WalletTransaction(nil), u.transactionsBefore...)
	u.transactionsBefore = nil
}

type fakePaymentGateway struct {
	paymentURL         string
	generateCalls      int
	generatedTxnRef    string
	generatedAmount    int64
	generatedCreatedAt int64
	generatedExpiresAt int64
}

func (g *fakePaymentGateway) GeneratePaymentURL(txnRef string, amount int64, ipAddr, desc string, createdAt, expiresAt int64) string {
	g.generateCalls++
	g.generatedTxnRef = txnRef
	g.generatedAmount = amount
	g.generatedCreatedAt = createdAt
	g.generatedExpiresAt = expiresAt
	return g.paymentURL
}

func (g *fakePaymentGateway) VerifyChecksum(params map[string][]string) bool {
	return true
}

type fakeBookingClient struct {
	eligibility       *PaymentEligibility
	err               error
	eligibilityCalls  int
	lastAppointmentID string
	lastPayerID       uuid.UUID
}

func (c *fakeBookingClient) GetPaymentEligibility(ctx context.Context, appointmentID string, payerID uuid.UUID) (*PaymentEligibility, error) {
	c.eligibilityCalls++
	c.lastAppointmentID = appointmentID
	c.lastPayerID = payerID
	if c.err != nil {
		return nil, c.err
	}
	if c.eligibility != nil {
		return c.eligibility, nil
	}
	return &PaymentEligibility{
		AppointmentID: appointmentID,
		ExpertID:      uuid.New().String(),
		AmountVND:     1000,
		ExpiresAt:     time.Now().Add(5 * time.Minute).UnixMilli(),
	}, nil
}

func (c *fakeBookingClient) ConfirmAppointment(ctx context.Context, appointmentID string) error {
	return nil
}

func (c *fakeBookingClient) FailAppointment(ctx context.Context, appointmentID string) error {
	return nil
}

func signedSuccessIPNParams(orderID uuid.UUID) map[string][]string {
	return signedSuccessIPNParamsWithTxn(orderID, "vnp-txn-1")
}

func signedSuccessIPNParamsWithTxn(orderID uuid.UUID, gatewayTxnRef string) map[string][]string {
	params := map[string][]string{
		"vnp_TxnRef":            {orderID.String()},
		"vnp_Amount":            {"100000"},
		"vnp_ResponseCode":      {"00"},
		"vnp_TransactionNo":     {gatewayTxnRef},
		"vnp_TransactionStatus": {"00"},
		"vnp_PayDate":           {"20260719100000"},
	}
	signVNPayParams(params)
	return params
}

func signedFailedIPNParams(orderID uuid.UUID) map[string][]string {
	params := map[string][]string{
		"vnp_TxnRef":        {orderID.String()},
		"vnp_Amount":        {"100000"},
		"vnp_ResponseCode":  {"24"},
		"vnp_TransactionNo": {"vnp-txn-failed-1"},
	}
	signVNPayParams(params)
	return params
}

func assertOnlyBookingConfirmOutbox(t *testing.T, repo *fakePaymentRepo) {
	t.Helper()

	if len(repo.outboxEvents) != 1 {
		t.Fatalf("expected exactly one outbox event, got %d", len(repo.outboxEvents))
	}
	if repo.outboxEvents[0].EventType != "booking.appointment.confirm" {
		t.Fatalf("expected booking.appointment.confirm outbox event, got %q", repo.outboxEvents[0].EventType)
	}
	for _, event := range repo.outboxEvents {
		if event.EventType == "wallet.payment.received" {
			t.Fatal("did not expect wallet.payment.received outbox event")
		}
	}
}

func assertOnlyBookingFailOutbox(t *testing.T, repo *fakePaymentRepo) {
	t.Helper()

	if len(repo.outboxEvents) != 1 {
		t.Fatalf("expected exactly one outbox event, got %d", len(repo.outboxEvents))
	}
	if repo.outboxEvents[0].EventType != "booking.appointment.fail" {
		t.Fatalf("expected booking.appointment.fail outbox event, got %q", repo.outboxEvents[0].EventType)
	}
}

func assertNoPaymentOrderPersisted(t *testing.T, repo *fakePaymentRepo) {
	t.Helper()

	if repo.createCount != 0 {
		t.Fatalf("expected no payment order to be persisted, got %d creates", repo.createCount)
	}
}

func assertRolledBackPaymentWalletAndOutbox(t *testing.T, repo *fakePaymentRepo, walletUsecase *fakeWalletUsecase) {
	t.Helper()

	if repo.order.Status != paymentdomain.OrderStatusPending {
		t.Fatalf("expected order status to roll back to PENDING, got %s", repo.order.Status.String())
	}
	if repo.successTransitions != 0 {
		t.Fatalf("expected no committed PENDING -> SUCCESS transition, got %d", repo.successTransitions)
	}
	if len(repo.outboxEvents) != 0 {
		t.Fatalf("expected outbox events to roll back, got %d", len(repo.outboxEvents))
	}
	if walletUsecase.pendingBalance != 0 {
		t.Fatalf("expected wallet pending balance to roll back to 0, got %d", walletUsecase.pendingBalance)
	}
	if len(walletUsecase.transactions) != 0 {
		t.Fatalf("expected wallet transactions to roll back, got %d", len(walletUsecase.transactions))
	}
}

func signVNPayParams(params map[string][]string) {
	cleanParams := make(map[string]string)
	for k, vals := range params {
		if len(vals) == 0 {
			continue
		}
		if k == "vnp_SecureHash" || k == "vnp_SecureHashType" {
			continue
		}
		cleanParams[k] = vals[0]
	}

	keys := make([]string, 0, len(cleanParams))
	for k := range cleanParams {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var queryBuilder strings.Builder
	for i, k := range keys {
		if i > 0 {
			queryBuilder.WriteString("&")
		}
		queryBuilder.WriteString(fmt.Sprintf("%s=%s", k, url.QueryEscape(cleanParams[k])))
	}

	mac := hmac.New(sha512.New, []byte("MOCK_SECRET"))
	mac.Write([]byte(queryBuilder.String()))
	params["vnp_SecureHash"] = []string{hex.EncodeToString(mac.Sum(nil))}
}

func stringPtr(s string) *string {
	return &s
}
