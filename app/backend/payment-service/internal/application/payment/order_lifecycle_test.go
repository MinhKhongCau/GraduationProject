package payment

import (
	"context"
	"errors"
	"payment-service/internal/domain/money"
	paymentdomain "payment-service/internal/domain/payment"
	vnpayadapter "payment-service/internal/infrastructure/client/vnpay"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCreateOrderPersistsAuthoritativeAmountAndTTLBoundedExpiry(t *testing.T) {
	now := time.Date(2026, 7, 19, 10, 0, 0, 0, time.UTC)
	appointmentID, payerID, expertID := uuid.New(), uuid.New(), uuid.New()
	repo := newLifecycleRepo()
	gateway := &fakePaymentGateway{paymentURL: "pay"}
	booking := lifecycleBooking(appointmentID, expertID, now.Add(time.Hour).UnixMilli(), 250000)
	usecase := lifecycleUsecase(repo, gateway, booking, now, 15*time.Minute, time.Minute)

	order, _, err := usecase.CreateOrder(context.Background(), payerID, appointmentID.String(), "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	if order.Status != paymentdomain.OrderStatusPending || order.GrossAmount != money.Money(250000) {
		t.Fatalf("unexpected stored order: status=%s amount=%d", order.Status, order.GrossAmount)
	}
	if order.ExpiresAt != now.Add(15*time.Minute).UnixMilli() {
		t.Fatalf("TTL should bound expiry: got %d", order.ExpiresAt)
	}
	if repo.createCount != 1 || gateway.generateCalls != 1 {
		t.Fatalf("expected one insert and URL, got inserts=%d urls=%d", repo.createCount, gateway.generateCalls)
	}
	if gateway.generatedExpiresAt != order.ExpiresAt || gateway.generatedCreatedAt != order.CreatedAt {
		t.Fatal("gateway did not receive persisted timestamps")
	}
}

func TestCreateOrderBookingExpiryShorterThanTTLWins(t *testing.T) {
	now := time.Date(2026, 7, 19, 10, 0, 0, 0, time.UTC)
	appointmentID, expertID := uuid.New(), uuid.New()
	bookingExpiry := now.Add(5*time.Minute + 456*time.Millisecond).UnixMilli()
	normalizedExpiry := now.Add(5 * time.Minute).UnixMilli()
	repo := newLifecycleRepo()
	order, _, err := lifecycleUsecase(repo, &fakePaymentGateway{}, lifecycleBooking(appointmentID, expertID, bookingExpiry, 1000), now, 15*time.Minute, time.Minute).
		CreateOrder(context.Background(), uuid.New(), appointmentID.String(), "ip")
	if err != nil {
		t.Fatal(err)
	}
	if order.ExpiresAt != normalizedExpiry || order.ExpiresAt%1000 != 0 || order.ExpiresAt > bookingExpiry {
		t.Fatalf("normalized booking expiry should win: got %d want %d raw=%d", order.ExpiresAt, normalizedExpiry, bookingExpiry)
	}
}

func TestCreateOrderMinimumWindowBoundaryAndTooShortWindow(t *testing.T) {
	now := time.Date(2026, 7, 19, 10, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name       string
		window     time.Duration
		wantErr    bool
		wantCreate int
	}{
		{name: "exact boundary accepted", window: time.Minute, wantCreate: 1},
		{name: "one millisecond short rejected", window: time.Minute - time.Millisecond, wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			appointmentID, expertID := uuid.New(), uuid.New()
			repo := newLifecycleRepo()
			gateway := &fakePaymentGateway{}
			usecase := lifecycleUsecase(repo, gateway, lifecycleBooking(appointmentID, expertID, now.Add(tc.window).UnixMilli(), 1000), now, 15*time.Minute, time.Minute)
			_, _, err := usecase.CreateOrder(context.Background(), uuid.New(), appointmentID.String(), "ip")
			if tc.wantErr && !errors.Is(err, ErrPaymentWindowTooShort) {
				t.Fatalf("expected short-window error, got %v", err)
			}
			if !tc.wantErr && err != nil {
				t.Fatal(err)
			}
			if repo.createCount != tc.wantCreate || (tc.wantErr && gateway.generateCalls != 0) {
				t.Fatalf("unexpected side effects: creates=%d urls=%d", repo.createCount, gateway.generateCalls)
			}
		})
	}
}

func TestCreateOrderReusesValidPendingWithoutExtendingIt(t *testing.T) {
	now := time.Date(2026, 7, 19, 10, 0, 0, 0, time.UTC)
	appointmentID, payerID, expertID := uuid.New(), uuid.New(), uuid.New()
	existing := lifecycleOrder(appointmentID, payerID, expertID, paymentdomain.OrderStatusPending, now.Add(4*time.Minute).UnixMilli())
	repo := newLifecycleRepo(existing)
	gateway := &fakePaymentGateway{paymentURL: "pay"}
	booking := lifecycleBooking(appointmentID, expertID, now.Add(10*time.Minute).UnixMilli(), existing.GrossAmount.Int64())

	order, _, err := lifecycleUsecase(repo, gateway, booking, now, 15*time.Minute, time.Minute).
		CreateOrder(context.Background(), payerID, appointmentID.String(), "ip")
	if err != nil {
		t.Fatal(err)
	}
	if order.ID != existing.ID || order.ExpiresAt != existing.ExpiresAt {
		t.Fatalf("retry changed immutable attempt: got id=%s expiry=%d", order.ID, order.ExpiresAt)
	}
	if gateway.generatedTxnRef != existing.ID.String() || gateway.generatedExpiresAt != existing.ExpiresAt {
		t.Fatal("retry URL did not reuse original transaction reference and expiry")
	}
	if repo.createCount != 0 {
		t.Fatalf("retry inserted %d orders", repo.createCount)
	}
}

func TestCreateOrderReusesStillValidPendingBelowNewOrderMinimumWindow(t *testing.T) {
	now := time.Date(2026, 7, 19, 10, 0, 0, 0, time.UTC)
	appointmentID, payerID, expertID := uuid.New(), uuid.New(), uuid.New()
	existing := lifecycleOrder(appointmentID, payerID, expertID, paymentdomain.OrderStatusPending, now.Add(30*time.Second).UnixMilli())
	repo := newLifecycleRepo(existing)

	order, _, err := lifecycleUsecase(repo, &fakePaymentGateway{}, lifecycleBooking(appointmentID, expertID, existing.ExpiresAt, 1000), now, 15*time.Minute, time.Minute).
		CreateOrder(context.Background(), payerID, appointmentID.String(), "ip")
	if err != nil {
		t.Fatalf("valid retry should not be treated as a new too-short order: %v", err)
	}
	if order.ID != existing.ID || repo.createCount != 0 {
		t.Fatalf("expected existing attempt, got id=%s creates=%d", order.ID, repo.createCount)
	}
}

func TestCreateOrderExpiresPendingThenCreatesReplacement(t *testing.T) {
	now := time.Date(2026, 7, 19, 10, 0, 0, 0, time.UTC)
	appointmentID, payerID, expertID := uuid.New(), uuid.New(), uuid.New()
	old := lifecycleOrder(appointmentID, payerID, expertID, paymentdomain.OrderStatusPending, now.UnixMilli())
	repo := newLifecycleRepo(old)

	order, _, err := lifecycleUsecase(repo, &fakePaymentGateway{}, lifecycleBooking(appointmentID, expertID, now.Add(5*time.Minute).UnixMilli(), 1000), now, 15*time.Minute, time.Minute).
		CreateOrder(context.Background(), payerID, appointmentID.String(), "ip")
	if err != nil {
		t.Fatal(err)
	}
	if order.ID == old.ID || repo.orders[old.ID].Status != paymentdomain.OrderStatusExpired || repo.createCount != 1 {
		t.Fatalf("replacement policy failed: old=%s new=%s creates=%d", repo.orders[old.ID].Status, order.ID, repo.createCount)
	}
}

func TestCreateOrderRetryableHistoryDoesNotBlockReplacement(t *testing.T) {
	now := time.Date(2026, 7, 19, 10, 0, 0, 0, time.UTC)
	for _, status := range []paymentdomain.PaymentOrderStatus{paymentdomain.OrderStatusFailed, paymentdomain.OrderStatusExpired} {
		t.Run(status.String(), func(t *testing.T) {
			appointmentID, payerID, expertID := uuid.New(), uuid.New(), uuid.New()
			repo := newLifecycleRepo(
				lifecycleOrder(appointmentID, payerID, expertID, paymentdomain.OrderStatusFailed, now.Add(-time.Minute).UnixMilli()),
				lifecycleOrder(appointmentID, payerID, expertID, status, now.Add(-time.Minute).UnixMilli()),
			)
			_, _, err := lifecycleUsecase(repo, &fakePaymentGateway{}, lifecycleBooking(appointmentID, expertID, now.Add(5*time.Minute).UnixMilli(), 1000), now, 15*time.Minute, time.Minute).
				CreateOrder(context.Background(), payerID, appointmentID.String(), "ip")
			if err != nil || repo.createCount != 1 {
				t.Fatalf("retryable history blocked replacement: err=%v creates=%d", err, repo.createCount)
			}
		})
	}
}

func TestCreateOrderSuccessRejectsBeforeBookingAndGateway(t *testing.T) {
	now := time.Date(2026, 7, 19, 10, 0, 0, 0, time.UTC)
	appointmentID, payerID, expertID := uuid.New(), uuid.New(), uuid.New()
	repo := newLifecycleRepo(lifecycleOrder(appointmentID, payerID, expertID, paymentdomain.OrderStatusSuccess, now.Add(-time.Minute).UnixMilli()))
	gateway := &fakePaymentGateway{}
	booking := lifecycleBooking(appointmentID, expertID, now.Add(5*time.Minute).UnixMilli(), 1000)

	_, _, err := lifecycleUsecase(repo, gateway, booking, now, 15*time.Minute, time.Minute).
		CreateOrder(context.Background(), payerID, appointmentID.String(), "ip")
	if !errors.Is(err, ErrAppointmentAlreadyPaid) {
		t.Fatalf("expected already-paid conflict, got %v", err)
	}
	if booking.eligibilityCalls != 0 || repo.createCount != 0 || gateway.generateCalls != 0 {
		t.Fatalf("already-paid request had side effects: booking=%d creates=%d urls=%d", booking.eligibilityCalls, repo.createCount, gateway.generateCalls)
	}
}

func TestCreateOrderUnknownTerminalStateIsNotAutomaticallyRetried(t *testing.T) {
	now := time.Date(2026, 7, 19, 10, 0, 0, 0, time.UTC)
	appointmentID, payerID, expertID := uuid.New(), uuid.New(), uuid.New()
	repo := newLifecycleRepo(lifecycleOrder(appointmentID, payerID, expertID, paymentdomain.PaymentOrderStatus(99), now.Add(-time.Minute).UnixMilli()))
	gateway := &fakePaymentGateway{}

	_, _, err := lifecycleUsecase(repo, gateway, lifecycleBooking(appointmentID, expertID, now.Add(5*time.Minute).UnixMilli(), 1000), now, 15*time.Minute, time.Minute).
		CreateOrder(context.Background(), payerID, appointmentID.String(), "ip")
	if !errors.Is(err, ErrExistingOrderConflict) || repo.createCount != 0 || gateway.generateCalls != 0 {
		t.Fatalf("unexpected state was retried: err=%v creates=%d urls=%d", err, repo.createCount, gateway.generateCalls)
	}
}

func TestCreateOrderUniquePendingRaceReturnsWinner(t *testing.T) {
	now := time.Date(2026, 7, 19, 10, 0, 0, 0, time.UTC)
	appointmentID, payerID, expertID := uuid.New(), uuid.New(), uuid.New()
	winner := lifecycleOrder(appointmentID, payerID, expertID, paymentdomain.OrderStatusPending, now.Add(5*time.Minute).UnixMilli())
	repo := newLifecycleRepo()
	repo.uniqueConflictWinner = &winner

	order, _, err := lifecycleUsecase(repo, &fakePaymentGateway{}, lifecycleBooking(appointmentID, expertID, now.Add(5*time.Minute).UnixMilli(), 1000), now, 15*time.Minute, time.Minute).
		CreateOrder(context.Background(), payerID, appointmentID.String(), "ip")
	if err != nil {
		t.Fatal(err)
	}
	if order.ID != winner.ID || repo.createCount != 0 {
		t.Fatalf("did not return race winner: got=%s winner=%s creates=%d", order.ID, winner.ID, repo.createCount)
	}
}

func TestConcurrentCreateOrderPersistsOnePending(t *testing.T) {
	now := time.Date(2026, 7, 19, 10, 0, 0, 0, time.UTC)
	appointmentID, payerID, expertID := uuid.New(), uuid.New(), uuid.New()
	repo := newLifecycleRepo()
	usecase := lifecycleUsecase(repo, &fakePaymentGateway{}, lifecycleBooking(appointmentID, expertID, now.Add(5*time.Minute).UnixMilli(), 1000), now, 15*time.Minute, time.Minute)

	start := make(chan struct{})
	results := make(chan uuid.UUID, 2)
	errs := make(chan error, 2)
	for range 2 {
		go func() {
			<-start
			order, _, err := usecase.CreateOrder(context.Background(), payerID, appointmentID.String(), "ip")
			if err != nil {
				errs <- err
				return
			}
			results <- order.ID
		}()
	}
	close(start)
	first, second := <-results, <-results
	if len(errs) != 0 {
		t.Fatal(<-errs)
	}
	if first != second || repo.createCount != 1 || repo.countStatus(paymentdomain.OrderStatusPending) != 1 {
		t.Fatalf("concurrent retry was not idempotent: ids=%s/%s creates=%d pending=%d", first, second, repo.createCount, repo.countStatus(paymentdomain.OrderStatusPending))
	}
}

func TestExpiredOrderSuccessExpiresReplacementAndSettlesOnce(t *testing.T) {
	now := time.Date(2026, 7, 19, 10, 0, 0, 0, time.UTC)
	appointmentID, payerID, expertID := uuid.New(), uuid.New(), uuid.New()
	old := lifecycleOrder(appointmentID, payerID, expertID, paymentdomain.OrderStatusExpired, now.Add(-time.Minute).UnixMilli())
	replacement := lifecycleOrder(appointmentID, payerID, expertID, paymentdomain.OrderStatusPending, now.Add(5*time.Minute).UnixMilli())
	repo := newLifecycleRepo(old, replacement)
	usecase := lifecycleUsecase(repo, vnpayadapter.NewVNPayClient("", "", "", ""), lifecycleBooking(appointmentID, expertID, now.Add(5*time.Minute).UnixMilli(), 1000), now, 15*time.Minute, time.Minute)

	already, err := usecase.ProcessIPN(context.Background(), signedSuccessIPNParams(old.ID))
	if err != nil || already {
		t.Fatalf("expired success was not processed: already=%v err=%v", already, err)
	}
	if repo.orders[old.ID].Status != paymentdomain.OrderStatusSuccess || repo.orders[replacement.ID].Status != paymentdomain.OrderStatusExpired {
		t.Fatalf("unexpected statuses old=%s replacement=%s", repo.orders[old.ID].Status, repo.orders[replacement.ID].Status)
	}
	if repo.walletCredits != 1 || repo.walletDebits != 1 || len(repo.outboxEvents) != 1 {
		t.Fatalf("settlement should happen once: credits=%d debits=%d outbox=%d", repo.walletCredits, repo.walletDebits, len(repo.outboxEvents))
	}
	already, err = usecase.ProcessIPN(context.Background(), signedSuccessIPNParams(old.ID))
	if err != nil || !already || repo.walletCredits != 1 || len(repo.outboxEvents) != 1 {
		t.Fatalf("duplicate expired-order success was not idempotent: already=%v err=%v credits=%d outbox=%d", already, err, repo.walletCredits, len(repo.outboxEvents))
	}

	already, err = usecase.ProcessIPN(context.Background(), signedSuccessIPNParams(replacement.ID))
	if err != nil || !already || repo.countStatus(paymentdomain.OrderStatusSuccess) != 1 || repo.walletCredits != 1 {
		t.Fatalf("replacement became a second success: already=%v err=%v success=%d credits=%d", already, err, repo.countStatus(paymentdomain.OrderStatusSuccess), repo.walletCredits)
	}
}

func TestDuplicateGatewayCaptureCreatesOneRefundCaseWithoutSettlement(t *testing.T) {
	now := time.Date(2026, 7, 19, 10, 0, 0, 0, time.UTC)
	appointmentID, payerID, expertID := uuid.New(), uuid.New(), uuid.New()
	winner := lifecycleOrder(appointmentID, payerID, expertID, paymentdomain.OrderStatusSuccess, now.Add(-time.Minute).UnixMilli())
	winner.GatewayTxnRef = "winner-vnp-txn"
	loser := lifecycleOrder(appointmentID, payerID, expertID, paymentdomain.OrderStatusExpired, now.Add(-time.Minute).UnixMilli())
	repo := newLifecycleRepo(winner, loser)
	usecase := lifecycleUsecase(repo, vnpayadapter.NewVNPayClient("", "", "", ""), lifecycleBooking(appointmentID, expertID, now.Add(5*time.Minute).UnixMilli(), 1000), now, 15*time.Minute, time.Minute)
	params := signedSuccessIPNParamsWithTxn(loser.ID, "second-vnp-txn")

	already, err := usecase.ProcessIPN(context.Background(), params)
	if err != nil || !already {
		t.Fatalf("first attempt was not stably acknowledged: already=%v err=%v", already, err)
	}
	firstPaidAt := *repo.orders[loser.ID].PaidAt
	usecase.(*paymentUsecase).clock = func() time.Time { return now.Add(time.Hour) }
	params["vnp_TransactionStatus"] = []string{"01"}
	params["vnp_PayDate"] = []string{"20260719110000"}
	signVNPayParams(params)
	already, err = usecase.ProcessIPN(context.Background(), params)
	if err != nil || !already {
		t.Fatalf("duplicate attempt was not stably acknowledged: already=%v err=%v", already, err)
	}
	if *repo.orders[loser.ID].PaidAt != firstPaidAt {
		t.Fatal("duplicate IPN rewrote the first captured-at timestamp")
	}

	stored := repo.orders[loser.ID]
	if stored.Status != paymentdomain.OrderStatusExpired || stored.GatewayCaptureStatus != paymentdomain.GatewayCaptureDuplicate || stored.FulfillmentStatus != paymentdomain.FulfillmentRefundRequired {
		t.Fatalf("duplicate gateway truth was not preserved: %+v", stored)
	}
	if stored.GatewayTxnRef != "second-vnp-txn" || stored.PaidAt == nil {
		t.Fatalf("duplicate capture audit data missing: %+v", stored)
	}
	if stored.GatewayResponseCode != "00" || stored.GatewayTransactionStatus != "00" || stored.GatewayPaymentDate != "20260719100000" {
		t.Fatalf("signed gateway evidence was not preserved immutably: %+v", stored)
	}
	if len(repo.compensationCases) != 1 {
		t.Fatalf("expected one idempotent compensation case, got %d", len(repo.compensationCases))
	}
	compensationCase := repo.compensationCases[0]
	if compensationCase.PaymentOrderID != loser.ID || compensationCase.AppointmentID != appointmentID || compensationCase.AmountVND != loser.GrossAmount || compensationCase.GatewayOrderReference != loser.ID.String() || compensationCase.GatewayTransactionNumber != "second-vnp-txn" || compensationCase.GatewayResponseCode != "00" || compensationCase.GatewayTransactionStatus != "00" || compensationCase.GatewayPaymentDate != "20260719100000" || compensationCase.CreatedAt != firstPaidAt || compensationCase.Status != paymentdomain.CompensationRefundRequired {
		t.Fatalf("unexpected compensation case: %+v", compensationCase)
	}
	if repo.walletCredits != 0 || repo.walletDebits != 0 || len(repo.outboxEvents) != 0 {
		t.Fatalf("duplicate capture produced settlement side effects: credits=%d debits=%d outbox=%d", repo.walletCredits, repo.walletDebits, len(repo.outboxEvents))
	}
}

func TestDuplicateCaptureInvalidIPNAndDifferentAppointmentCreateNoCase(t *testing.T) {
	now := time.Date(2026, 7, 19, 10, 0, 0, 0, time.UTC)
	payerID, expertID := uuid.New(), uuid.New()
	winnerAppointment, otherAppointment := uuid.New(), uuid.New()
	winner := lifecycleOrder(winnerAppointment, payerID, expertID, paymentdomain.OrderStatusSuccess, now.Add(-time.Minute).UnixMilli())
	loser := lifecycleOrder(winnerAppointment, payerID, expertID, paymentdomain.OrderStatusExpired, now.Add(-time.Minute).UnixMilli())

	for _, test := range []struct {
		name   string
		params func() map[string][]string
	}{
		{name: "invalid checksum", params: func() map[string][]string {
			params := signedSuccessIPNParamsWithTxn(loser.ID, "second-vnp-txn")
			params["vnp_SecureHash"] = []string{"invalid"}
			return params
		}},
		{name: "invalid amount", params: func() map[string][]string {
			params := signedSuccessIPNParamsWithTxn(loser.ID, "second-vnp-txn")
			params["vnp_Amount"] = []string{"invalid"}
			signVNPayParams(params)
			return params
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo := newLifecycleRepo(winner, loser)
			usecase := lifecycleUsecase(repo, vnpayadapter.NewVNPayClient("", "", "", ""), lifecycleBooking(winnerAppointment, expertID, now.Add(5*time.Minute).UnixMilli(), 1000), now, 15*time.Minute, time.Minute)
			if _, err := usecase.ProcessIPN(context.Background(), test.params()); err == nil {
				t.Fatal("expected validation error")
			}
			stored := repo.orders[loser.ID]
			if len(repo.compensationCases) != 0 || repo.walletCredits != 0 || len(repo.outboxEvents) != 0 || stored.GatewayCaptureStatus != "" || stored.GatewayTxnRef != "" || stored.GatewayResponseCode != "" {
				t.Fatal("invalid IPN created compensation or settlement side effects")
			}
		})
	}

	other := lifecycleOrder(otherAppointment, payerID, expertID, paymentdomain.OrderStatusPending, now.Add(5*time.Minute).UnixMilli())
	repo := newLifecycleRepo(winner, other)
	usecase := lifecycleUsecase(repo, vnpayadapter.NewVNPayClient("", "", "", ""), lifecycleBooking(otherAppointment, expertID, now.Add(5*time.Minute).UnixMilli(), 1000), now, 15*time.Minute, time.Minute)
	already, err := usecase.ProcessIPN(context.Background(), signedSuccessIPNParamsWithTxn(other.ID, "other-appointment-txn"))
	if err != nil || already || repo.orders[other.ID].Status != paymentdomain.OrderStatusSuccess || len(repo.compensationCases) != 0 {
		t.Fatalf("different appointment was treated as duplicate capture: already=%v err=%v order=%+v cases=%d", already, err, repo.orders[other.ID], len(repo.compensationCases))
	}
}

func TestDuplicateCaptureCompensationFailureRollsBackEvidence(t *testing.T) {
	now := time.Date(2026, 7, 19, 10, 0, 0, 0, time.UTC)
	appointmentID, payerID, expertID := uuid.New(), uuid.New(), uuid.New()
	winner := lifecycleOrder(appointmentID, payerID, expertID, paymentdomain.OrderStatusSuccess, now.Add(-time.Minute).UnixMilli())
	loser := lifecycleOrder(appointmentID, payerID, expertID, paymentdomain.OrderStatusExpired, now.Add(-time.Minute).UnixMilli())
	repo := newLifecycleRepo(winner, loser)
	repo.compensationErr = errors.New("compensation persistence failed")
	usecase := lifecycleUsecase(repo, vnpayadapter.NewVNPayClient("", "", "", ""), lifecycleBooking(appointmentID, expertID, now.Add(5*time.Minute).UnixMilli(), 1000), now, 15*time.Minute, time.Minute)

	already, err := usecase.ProcessIPN(context.Background(), signedSuccessIPNParamsWithTxn(loser.ID, "second-vnp-txn"))
	if err == nil || already {
		t.Fatalf("expected compensation failure, already=%v err=%v", already, err)
	}
	stored := repo.orders[loser.ID]
	if repo.listForUpdateCount == 0 || stored.Status != paymentdomain.OrderStatusExpired || stored.GatewayCaptureStatus != "" || stored.GatewayTxnRef != "" || stored.PaidAt != nil || len(repo.compensationCases) != 0 || repo.walletCredits != 0 || repo.walletDebits != 0 || len(repo.outboxEvents) != 0 {
		t.Fatalf("duplicate capture transaction did not roll back atomically: order=%+v cases=%+v", stored, repo.compensationCases)
	}
}

func TestConcurrentSuccessIndexLossReloadsWinnerWithoutSettlement(t *testing.T) {
	now := time.Date(2026, 7, 19, 10, 0, 0, 0, time.UTC)
	appointmentID, payerID, expertID := uuid.New(), uuid.New(), uuid.New()
	loser := lifecycleOrder(appointmentID, payerID, expertID, paymentdomain.OrderStatusExpired, now.Add(-time.Minute).UnixMilli())
	winner := lifecycleOrder(appointmentID, payerID, expertID, paymentdomain.OrderStatusPending, now.Add(5*time.Minute).UnixMilli())
	winnerAfterRace := winner
	winnerAfterRace.Status = paymentdomain.OrderStatusSuccess
	repo := newLifecycleRepo(loser, winner)
	repo.successConflictWinner = &winnerAfterRace
	usecase := lifecycleUsecase(repo, vnpayadapter.NewVNPayClient("", "", "", ""), lifecycleBooking(appointmentID, expertID, now.Add(5*time.Minute).UnixMilli(), 1000), now, 15*time.Minute, time.Minute)

	already, err := usecase.ProcessIPN(context.Background(), signedSuccessIPNParams(loser.ID))
	if err != nil || !already {
		t.Fatalf("unique-index loser was not acknowledged: already=%v err=%v", already, err)
	}
	if repo.withinTxCount != 2 {
		t.Fatalf("expected processing transaction plus winner reload, got %d transactions", repo.withinTxCount)
	}
	if repo.orders[loser.ID].Status != paymentdomain.OrderStatusExpired || repo.orders[winner.ID].Status != paymentdomain.OrderStatusSuccess {
		t.Fatalf("unexpected race result: loser=%s winner=%s", repo.orders[loser.ID].Status, repo.orders[winner.ID].Status)
	}
	if repo.walletCredits != 0 || repo.walletDebits != 0 || len(repo.outboxEvents) != 0 {
		t.Fatalf("losing transaction leaked side effects: credits=%d debits=%d outbox=%d", repo.walletCredits, repo.walletDebits, len(repo.outboxEvents))
	}
	if len(repo.compensationCases) != 1 || repo.orders[loser.ID].GatewayCaptureStatus != paymentdomain.GatewayCaptureDuplicate {
		t.Fatalf("unique-index loser was not recorded for refund review: order=%+v cases=%+v", repo.orders[loser.ID], repo.compensationCases)
	}
}

func TestConcurrentOldAndReplacementSuccessIPNsSettleOnce(t *testing.T) {
	now := time.Date(2026, 7, 19, 10, 0, 0, 0, time.UTC)
	appointmentID, payerID, expertID := uuid.New(), uuid.New(), uuid.New()
	old := lifecycleOrder(appointmentID, payerID, expertID, paymentdomain.OrderStatusExpired, now.Add(-time.Minute).UnixMilli())
	replacement := lifecycleOrder(appointmentID, payerID, expertID, paymentdomain.OrderStatusPending, now.Add(5*time.Minute).UnixMilli())
	repo := newLifecycleRepo(old, replacement)
	usecase := lifecycleUsecase(repo, vnpayadapter.NewVNPayClient("", "", "", ""), lifecycleBooking(appointmentID, expertID, now.Add(5*time.Minute).UnixMilli(), 1000), now, 15*time.Minute, time.Minute)

	type result struct {
		already bool
		err     error
	}
	results := make(chan result, 2)
	start := make(chan struct{})
	for _, orderID := range []uuid.UUID{old.ID, replacement.ID} {
		orderID := orderID
		go func() {
			<-start
			already, err := usecase.ProcessIPN(context.Background(), signedSuccessIPNParams(orderID))
			results <- result{already: already, err: err}
		}()
	}
	close(start)
	alreadyCount := 0
	for range 2 {
		result := <-results
		if result.err != nil {
			t.Fatalf("concurrent IPN returned error: %v", result.err)
		}
		if result.already {
			alreadyCount++
		}
	}
	if alreadyCount != 1 || repo.countStatus(paymentdomain.OrderStatusSuccess) != 1 || repo.walletCredits != 1 || repo.walletDebits != 1 || len(repo.outboxEvents) != 1 {
		t.Fatalf("concurrent success was not exactly once: already=%d success=%d credits=%d debits=%d outbox=%d", alreadyCount, repo.countStatus(paymentdomain.OrderStatusSuccess), repo.walletCredits, repo.walletDebits, len(repo.outboxEvents))
	}
}

func TestPendingOrderSuccessAtPersistedExpiryStillSettles(t *testing.T) {
	now := time.Date(2026, 7, 19, 10, 0, 0, 0, time.UTC)
	appointmentID, payerID, expertID := uuid.New(), uuid.New(), uuid.New()
	order := lifecycleOrder(appointmentID, payerID, expertID, paymentdomain.OrderStatusPending, now.UnixMilli())
	repo := newLifecycleRepo(order)
	usecase := lifecycleUsecase(repo, vnpayadapter.NewVNPayClient("", "", "", ""), lifecycleBooking(appointmentID, expertID, now.Add(5*time.Minute).UnixMilli(), 1000), now, 15*time.Minute, time.Minute)

	already, err := usecase.ProcessIPN(context.Background(), signedSuccessIPNParams(order.ID))
	if err != nil || already || repo.orders[order.ID].Status != paymentdomain.OrderStatusSuccess || repo.walletCredits != 1 || len(repo.outboxEvents) != 1 {
		t.Fatalf("late gateway success was discarded: already=%v err=%v status=%s credits=%d outbox=%d", already, err, repo.orders[order.ID].Status, repo.walletCredits, len(repo.outboxEvents))
	}
}

func TestFailedIPNForExpiredOrderKeepsExpiredWithoutOutbox(t *testing.T) {
	now := time.Date(2026, 7, 19, 10, 0, 0, 0, time.UTC)
	appointmentID, payerID, expertID := uuid.New(), uuid.New(), uuid.New()
	order := lifecycleOrder(appointmentID, payerID, expertID, paymentdomain.OrderStatusExpired, now.Add(-time.Minute).UnixMilli())
	repo := newLifecycleRepo(order)
	usecase := lifecycleUsecase(repo, vnpayadapter.NewVNPayClient("", "", "", ""), lifecycleBooking(appointmentID, expertID, now.Add(5*time.Minute).UnixMilli(), 1000), now, 15*time.Minute, time.Minute)

	already, err := usecase.ProcessIPN(context.Background(), signedFailedIPNParams(order.ID))
	if err != nil || !already || repo.orders[order.ID].Status != paymentdomain.OrderStatusExpired || len(repo.outboxEvents) != 0 {
		t.Fatalf("expired failure policy violated: already=%v err=%v status=%s outbox=%d", already, err, repo.orders[order.ID].Status, len(repo.outboxEvents))
	}
}

func TestFailedIPNAtPersistedExpiryMarksPendingExpiredWithoutOutbox(t *testing.T) {
	now := time.Date(2026, 7, 19, 10, 0, 0, 0, time.UTC)
	appointmentID, payerID, expertID := uuid.New(), uuid.New(), uuid.New()
	order := lifecycleOrder(appointmentID, payerID, expertID, paymentdomain.OrderStatusPending, now.UnixMilli())
	repo := newLifecycleRepo(order)
	usecase := lifecycleUsecase(repo, vnpayadapter.NewVNPayClient("", "", "", ""), lifecycleBooking(appointmentID, expertID, now.Add(5*time.Minute).UnixMilli(), 1000), now, 15*time.Minute, time.Minute)

	already, err := usecase.ProcessIPN(context.Background(), signedFailedIPNParams(order.ID))
	if err != nil || !already || repo.orders[order.ID].Status != paymentdomain.OrderStatusExpired || len(repo.outboxEvents) != 0 {
		t.Fatalf("elapsed PENDING failure policy violated: already=%v err=%v status=%s outbox=%d", already, err, repo.orders[order.ID].Status, len(repo.outboxEvents))
	}
}

func TestInvalidExpiredOrderIPNHasNoSideEffects(t *testing.T) {
	now := time.Date(2026, 7, 19, 10, 0, 0, 0, time.UTC)
	appointmentID, payerID, expertID := uuid.New(), uuid.New(), uuid.New()
	order := lifecycleOrder(appointmentID, payerID, expertID, paymentdomain.OrderStatusExpired, now.Add(-time.Minute).UnixMilli())
	for _, tc := range []struct {
		name   string
		params func(uuid.UUID) map[string][]string
	}{
		{name: "invalid checksum", params: func(id uuid.UUID) map[string][]string {
			p := signedSuccessIPNParams(id)
			p["vnp_SecureHash"] = []string{"invalid"}
			return p
		}},
		{name: "invalid amount", params: func(id uuid.UUID) map[string][]string {
			p := signedSuccessIPNParams(id)
			p["vnp_Amount"] = []string{"invalid"}
			signVNPayParams(p)
			return p
		}},
		{name: "failed amount mismatch", params: func(id uuid.UUID) map[string][]string {
			p := signedFailedIPNParams(id)
			p["vnp_Amount"] = []string{"99900"}
			signVNPayParams(p)
			return p
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := newLifecycleRepo(order)
			usecase := lifecycleUsecase(repo, vnpayadapter.NewVNPayClient("", "", "", ""), lifecycleBooking(appointmentID, expertID, now.Add(5*time.Minute).UnixMilli(), 1000), now, 15*time.Minute, time.Minute)
			if _, err := usecase.ProcessIPN(context.Background(), tc.params(order.ID)); err == nil {
				t.Fatal("expected invalid IPN error")
			}
			if repo.orders[order.ID].Status != paymentdomain.OrderStatusExpired || repo.walletCredits != 0 || len(repo.outboxEvents) != 0 || len(repo.compensationCases) != 0 {
				t.Fatal("invalid IPN changed persisted state")
			}
		})
	}
}

func lifecycleUsecase(repo *lifecycleRepo, gateway PaymentGateway, booking *fakeBookingClient, now time.Time, ttl, minimum time.Duration) Usecase {
	return NewUsecaseWithOptions(repo, repo, gateway, booking, Options{OrderTTL: ttl, MinimumWindow: minimum, Clock: func() time.Time { return now }})
}

func lifecycleBooking(appointmentID, expertID uuid.UUID, expiresAt, amount int64) *fakeBookingClient {
	return &fakeBookingClient{eligibility: &PaymentEligibility{AppointmentID: appointmentID.String(), ExpertID: expertID.String(), AmountVND: amount, ExpiresAt: expiresAt}}
}

func lifecycleOrder(appointmentID, payerID, expertID uuid.UUID, status paymentdomain.PaymentOrderStatus, expiresAt int64) paymentdomain.PaymentOrder {
	return paymentdomain.PaymentOrder{
		ID: uuid.New(), PayerID: payerID, ExpertID: expertID, AppointmentID: &appointmentID,
		GrossAmount: 1000, CommissionRate: 0.15, CommissionAmount: 150, NetAmount: 850,
		Gateway: "VNPAY", Status: status, CreatedAt: expiresAt - int64(time.Minute/time.Millisecond), ExpiresAt: expiresAt,
	}
}

type lifecycleRepo struct {
	mu                    sync.Mutex
	orders                map[uuid.UUID]paymentdomain.PaymentOrder
	createCount           int
	withinTxCount         int
	uniqueConflictWinner  *paymentdomain.PaymentOrder
	successConflictWinner *paymentdomain.PaymentOrder
	preserveWinner        *paymentdomain.PaymentOrder
	walletCredits         int
	walletDebits          int
	outboxEvents          []paymentdomain.OutboxEvent
	paymentEvents         []paymentdomain.OutboxEvent
	compensationCases     []paymentdomain.PaymentCompensationCase
	compensationErr       error
	listForUpdateCount    int
}

func newLifecycleRepo(orders ...paymentdomain.PaymentOrder) *lifecycleRepo {
	r := &lifecycleRepo{orders: make(map[uuid.UUID]paymentdomain.PaymentOrder)}
	for _, order := range orders {
		r.orders[order.ID] = order
	}
	return r
}

func (r *lifecycleRepo) Create(order *paymentdomain.PaymentOrder) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.orders[order.ID] = *order
	r.createCount++
	return nil
}

func (r *lifecycleRepo) GetByID(orderID uuid.UUID) (*paymentdomain.PaymentOrder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.get(orderID)
}

func (r *lifecycleRepo) GetByGatewayTxnRef(ref string) (*paymentdomain.PaymentOrder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.gateway(ref)
}

func (r *lifecycleRepo) GetSuccessfulByAppointment(ctx context.Context, appointmentID uuid.UUID) (*paymentdomain.PaymentOrder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, order := range r.orders {
		if order.AppointmentID != nil && *order.AppointmentID == appointmentID && order.Status == paymentdomain.OrderStatusSuccess {
			copy := order
			return &copy, nil
		}
	}
	return nil, ErrTxRecordNotFound
}

func (r *lifecycleRepo) Update(order *paymentdomain.PaymentOrder) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.orders[order.ID] = *order
	return nil
}

func (r *lifecycleRepo) ListCompensationCases(ctx context.Context, filter CompensationCaseFilter) ([]CompensationCaseRecord, int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var matches []CompensationCaseRecord
	for i := range r.compensationCases {
		compensationCase := r.compensationCases[i]
		if filter.Status != "" && compensationCase.Status != filter.Status {
			continue
		}
		if filter.AppointmentID != nil && compensationCase.AppointmentID != *filter.AppointmentID {
			continue
		}
		if filter.PaymentOrderID != nil && compensationCase.PaymentOrderID != *filter.PaymentOrderID {
			continue
		}
		order := r.orders[compensationCase.PaymentOrderID]
		matches = append(matches, CompensationCaseRecord{
			Case: compensationCase, PaymentStatus: order.Status,
			GatewayCaptureStatus: order.GatewayCaptureStatus, FulfillmentStatus: order.FulfillmentStatus,
		})
	}
	return matches, int64(len(matches)), nil
}

func (r *lifecycleRepo) GetCompensationCase(ctx context.Context, caseID uuid.UUID) (*CompensationCaseRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.compensationCases {
		if r.compensationCases[i].ID == caseID {
			copy := r.compensationCases[i]
			order := r.orders[copy.PaymentOrderID]
			return &CompensationCaseRecord{
				Case: copy, PaymentStatus: order.Status,
				GatewayCaptureStatus: order.GatewayCaptureStatus, FulfillmentStatus: order.FulfillmentStatus,
			}, nil
		}
	}
	return nil, ErrCompensationCaseNotFound
}

func (r *lifecycleRepo) WithinTx(ctx context.Context, fn func(tx Tx) error) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.withinTxCount++
	snapshot := cloneOrders(r.orders)
	createCount, credits, debits := r.createCount, r.walletCredits, r.walletDebits
	outbox := append([]paymentdomain.OutboxEvent(nil), r.outboxEvents...)
	paymentEvents := append([]paymentdomain.OutboxEvent(nil), r.paymentEvents...)
	compensationCases := append([]paymentdomain.PaymentCompensationCase(nil), r.compensationCases...)
	err := fn(r)
	if err != nil {
		r.orders = snapshot
		r.createCount, r.walletCredits, r.walletDebits = createCount, credits, debits
		r.outboxEvents = outbox
		r.paymentEvents = paymentEvents
		r.compensationCases = compensationCases
		if r.preserveWinner != nil {
			r.orders[r.preserveWinner.ID] = *r.preserveWinner
			r.preserveWinner = nil
		}
	}
	return err
}

func (r *lifecycleRepo) GetOrder(ctx context.Context, orderID uuid.UUID) (*paymentdomain.PaymentOrder, error) {
	return r.get(orderID)
}

func (r *lifecycleRepo) GetOrderForUpdate(ctx context.Context, orderID uuid.UUID) (*paymentdomain.PaymentOrder, error) {
	return r.get(orderID)
}

func (r *lifecycleRepo) ListOrdersForAppointmentForUpdate(ctx context.Context, appointmentID uuid.UUID) ([]paymentdomain.PaymentOrder, error) {
	r.listForUpdateCount++
	var orders []paymentdomain.PaymentOrder
	for _, order := range r.orders {
		if order.AppointmentID != nil && *order.AppointmentID == appointmentID {
			orders = append(orders, order)
		}
	}
	sort.Slice(orders, func(i, j int) bool { return orders[i].ID.String() < orders[j].ID.String() })
	return orders, nil
}

func (r *lifecycleRepo) CreateOrder(ctx context.Context, order *paymentdomain.PaymentOrder) error {
	if r.uniqueConflictWinner != nil {
		winner := *r.uniqueConflictWinner
		r.uniqueConflictWinner = nil
		r.preserveWinner = &winner
		return ErrActivePendingOrderExists
	}
	for _, existing := range r.orders {
		if existing.AppointmentID != nil && order.AppointmentID != nil && *existing.AppointmentID == *order.AppointmentID && existing.Status == paymentdomain.OrderStatusPending {
			return ErrActivePendingOrderExists
		}
	}
	r.orders[order.ID] = *order
	r.createCount++
	return nil
}

func (r *lifecycleRepo) GetGatewayTxnRef(ctx context.Context, ref string) (*paymentdomain.PaymentOrder, error) {
	return r.gateway(ref)
}

func (r *lifecycleRepo) UpdateOrder(ctx context.Context, order *paymentdomain.PaymentOrder) error {
	if order.Status == paymentdomain.OrderStatusSuccess && order.AppointmentID != nil {
		if r.successConflictWinner != nil {
			winner := *r.successConflictWinner
			r.successConflictWinner = nil
			r.preserveWinner = &winner
			return ErrAppointmentAlreadyPaid
		}
		for _, existing := range r.orders {
			if existing.ID != order.ID && existing.AppointmentID != nil && *existing.AppointmentID == *order.AppointmentID && existing.Status == paymentdomain.OrderStatusSuccess {
				return ErrAppointmentAlreadyPaid
			}
		}
	}
	r.orders[order.ID] = *order
	return nil
}

func (r *lifecycleRepo) ExpireOtherPendingOrders(ctx context.Context, appointmentID, exceptOrderID uuid.UUID) error {
	for id, order := range r.orders {
		if id != exceptOrderID && order.AppointmentID != nil && *order.AppointmentID == appointmentID && order.Status == paymentdomain.OrderStatusPending {
			order.Status = paymentdomain.OrderStatusExpired
			r.orders[id] = order
		}
	}
	return nil
}

func (r *lifecycleRepo) SaveOutboxEvent(ctx context.Context, event *paymentdomain.OutboxEvent) error {
	if isPaymentStatusEvent(event) {
		r.paymentEvents = append(r.paymentEvents, *event)
		return nil
	}
	r.outboxEvents = append(r.outboxEvents, *event)
	return nil
}

func (r *lifecycleRepo) SaveCompensationCase(ctx context.Context, compensationCase *paymentdomain.PaymentCompensationCase) error {
	if r.compensationErr != nil {
		return r.compensationErr
	}
	for i := range r.compensationCases {
		existing := &r.compensationCases[i]
		if existing.PaymentOrderID == compensationCase.PaymentOrderID && existing.ReasonCode == compensationCase.ReasonCode {
			return nil
		}
	}
	r.compensationCases = append(r.compensationCases, *compensationCase)
	return nil
}

func (r *lifecycleRepo) CreditWalletPending(ctx context.Context, userID uuid.UUID, amount money.Money, refID uuid.UUID, idempotencyKey string) error {
	r.walletCredits++
	return nil
}

func (r *lifecycleRepo) DebitWalletPending(ctx context.Context, userID uuid.UUID, amount money.Money, refID uuid.UUID, idempotencyKey string) error {
	r.walletDebits++
	return nil
}

func (r *lifecycleRepo) get(orderID uuid.UUID) (*paymentdomain.PaymentOrder, error) {
	order, ok := r.orders[orderID]
	if !ok {
		return nil, ErrTxRecordNotFound
	}
	copy := order
	return &copy, nil
}

func (r *lifecycleRepo) gateway(ref string) (*paymentdomain.PaymentOrder, error) {
	for _, order := range r.orders {
		if order.GatewayTxnRef == ref {
			copy := order
			return &copy, nil
		}
	}
	return nil, ErrTxRecordNotFound
}

func (r *lifecycleRepo) countStatus(status paymentdomain.PaymentOrderStatus) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	count := 0
	for _, order := range r.orders {
		if order.Status == status {
			count++
		}
	}
	return count
}

func cloneOrders(source map[uuid.UUID]paymentdomain.PaymentOrder) map[uuid.UUID]paymentdomain.PaymentOrder {
	copy := make(map[uuid.UUID]paymentdomain.PaymentOrder, len(source))
	for id, order := range source {
		copy[id] = order
	}
	return copy
}
