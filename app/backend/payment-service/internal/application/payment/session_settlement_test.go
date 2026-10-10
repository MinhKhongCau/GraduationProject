package payment

import (
	"context"
	"errors"
	"testing"
	"time"

	"payment-service/internal/domain/money"
	paymentdomain "payment-service/internal/domain/payment"
	walletdomain "payment-service/internal/domain/wallet"

	"github.com/google/uuid"
)

func paidSessionOrder(appointmentID uuid.UUID) paymentdomain.PaymentOrder {
	paidAt := time.Now().UnixMilli()
	return paymentdomain.PaymentOrder{
		ID:                uuid.New(),
		PayerID:           uuid.New(),
		ExpertID:          uuid.New(),
		AppointmentID:     &appointmentID,
		GrossAmount:       money.Money(1000),
		CommissionRate:    0.15,
		CommissionAmount:  money.Money(150),
		NetAmount:         money.Money(850),
		Gateway:           "VNPAY",
		Status:            paymentdomain.OrderStatusSuccess,
		FulfillmentStatus: paymentdomain.FulfillmentBookingConfirmed,
		PaidAt:            &paidAt,
	}
}

func newSettlementFixture(order paymentdomain.PaymentOrder) (*fakePaymentRepo, *fakeWalletUsecase, SettlementUsecase) {
	repo := newFakePaymentRepo(order)
	walletUsecase := &fakeWalletUsecase{}
	repo.addParticipant(walletUsecase)
	return repo, walletUsecase, NewSettlementUsecase(repo, uuid.Nil)
}

func TestSettleCompletedSessionPaysExpertNetAndKeepsCommission(t *testing.T) {
	appointmentID := uuid.New()
	order := paidSessionOrder(appointmentID)
	repo, walletUsecase, usecase := newSettlementFixture(order)

	settlement, err := usecase.SettleCompletedSession(context.Background(), appointmentID)
	if err != nil {
		t.Fatalf("SettleCompletedSession returned error: %v", err)
	}
	if settlement.AlreadySettled || settlement.NetAmount != 850 || settlement.ExpertID != order.ExpertID {
		t.Fatalf("unexpected settlement: %+v", settlement)
	}
	if !repo.order.Released {
		t.Fatal("expected order to be marked released")
	}
	want := []fakeWalletAdjustment{
		{UserID: DefaultSystemWalletUserID, Available: 150, Pending: -1000, Type: walletdomain.TxTypeSessionPayout, IdempotencyKey: "payout_system_" + order.ID.String()},
		{UserID: order.ExpertID, Available: 850, Pending: 0, Type: walletdomain.TxTypeSessionPayout, IdempotencyKey: "payout_expert_" + order.ID.String()},
	}
	if len(walletUsecase.adjustments) != len(want) {
		t.Fatalf("expected %d wallet adjustments, got %+v", len(want), walletUsecase.adjustments)
	}
	for i := range want {
		if walletUsecase.adjustments[i] != want[i] {
			t.Fatalf("adjustment %d = %+v, want %+v", i, walletUsecase.adjustments[i], want[i])
		}
	}
}

func TestSettleCompletedSessionIsIdempotent(t *testing.T) {
	appointmentID := uuid.New()
	_, walletUsecase, usecase := newSettlementFixture(paidSessionOrder(appointmentID))

	if _, err := usecase.SettleCompletedSession(context.Background(), appointmentID); err != nil {
		t.Fatalf("first settle failed: %v", err)
	}
	settlement, err := usecase.SettleCompletedSession(context.Background(), appointmentID)
	if err != nil {
		t.Fatalf("second settle failed: %v", err)
	}
	if !settlement.AlreadySettled {
		t.Fatal("expected second settle to report AlreadySettled")
	}
	if len(walletUsecase.adjustments) != 2 {
		t.Fatalf("expected payout to happen once, got %d adjustments", len(walletUsecase.adjustments))
	}
}

func TestSettleCompletedSessionReleasesLegacyExpertPending(t *testing.T) {
	appointmentID := uuid.New()
	order := paidSessionOrder(appointmentID)
	_, walletUsecase, usecase := newSettlementFixture(order)
	// Đơn cũ: tiền đã ghi vào Pending của chuyên gia lúc thanh toán.
	walletUsecase.transactions = append(walletUsecase.transactions, walletdomain.WalletTransaction{IdempotencyKey: "credit_order_" + order.ID.String()})

	if _, err := usecase.SettleCompletedSession(context.Background(), appointmentID); err != nil {
		t.Fatalf("SettleCompletedSession returned error: %v", err)
	}
	want := fakeWalletAdjustment{UserID: order.ExpertID, Available: 850, Pending: -850, Type: walletdomain.TxTypeAdjustment, IdempotencyKey: "release_" + order.ID.String()}
	if len(walletUsecase.adjustments) != 1 || walletUsecase.adjustments[0] != want {
		t.Fatalf("expected only the legacy pending release, got %+v", walletUsecase.adjustments)
	}
}

func TestSettleCompletedSessionRejectsUnpaidOrReviewedOrders(t *testing.T) {
	appointmentID := uuid.New()

	pending := paidSessionOrder(appointmentID)
	pending.Status = paymentdomain.OrderStatusPending
	_, _, usecase := newSettlementFixture(pending)
	if _, err := usecase.SettleCompletedSession(context.Background(), appointmentID); !errors.Is(err, ErrSessionOrderNotFound) {
		t.Fatalf("expected ErrSessionOrderNotFound, got %v", err)
	}

	refund := paidSessionOrder(appointmentID)
	refund.FulfillmentStatus = paymentdomain.FulfillmentRefundRequired
	repo, walletUsecase, usecase := newSettlementFixture(refund)
	if _, err := usecase.SettleCompletedSession(context.Background(), appointmentID); !errors.Is(err, ErrSessionNotSettleable) {
		t.Fatalf("expected ErrSessionNotSettleable, got %v", err)
	}
	if repo.order.Released || len(walletUsecase.adjustments) != 0 {
		t.Fatal("refund-required order must not be paid out")
	}
}

func TestSettleCompletedSessionRollsBackWhenExpertCreditFails(t *testing.T) {
	appointmentID := uuid.New()
	repo, walletUsecase, usecase := newSettlementFixture(paidSessionOrder(appointmentID))
	walletUsecase.adjustErr = errors.New("wallet down")

	if _, err := usecase.SettleCompletedSession(context.Background(), appointmentID); err == nil {
		t.Fatal("expected settlement error")
	}
	if repo.order.Released || len(walletUsecase.adjustments) != 0 {
		t.Fatal("failed settlement must not mark the order released")
	}
}
