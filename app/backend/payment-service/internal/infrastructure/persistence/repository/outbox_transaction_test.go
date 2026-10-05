package repository

import (
	"context"
	"errors"
	apppayment "payment-service/internal/application/payment"
	"payment-service/internal/domain/money"
	paymentdomain "payment-service/internal/domain/payment"
	"testing"

	"github.com/google/uuid"
)

type attemptMemoryState struct {
	event paymentdomain.OutboxEvent
	order paymentdomain.PaymentOrder
	cases []paymentdomain.PaymentCompensationCase
}

type fakeAttemptPersistence struct {
	state  *attemptMemoryState
	failAt string
}

func (s *fakeAttemptPersistence) LoadOutboxForUpdate(context.Context, uuid.UUID) (*paymentdomain.OutboxEvent, error) {
	copy := s.state.event
	return &copy, nil
}

func (s *fakeAttemptPersistence) SaveOutbox(_ context.Context, event *paymentdomain.OutboxEvent) error {
	if s.failAt == "outbox" {
		return errors.New("outbox update failed")
	}
	s.state.event = *event
	return nil
}

func (s *fakeAttemptPersistence) LoadPaymentOrderForUpdate(context.Context, uuid.UUID) (*paymentdomain.PaymentOrder, error) {
	copy := s.state.order
	return &copy, nil
}

func (s *fakeAttemptPersistence) UpdatePaymentFulfillment(_ context.Context, _ uuid.UUID, status paymentdomain.FulfillmentStatus) error {
	if s.failAt == "fulfillment" {
		return errors.New("fulfillment update failed")
	}
	s.state.order.FulfillmentStatus = status
	return nil
}

func (s *fakeAttemptPersistence) SaveCompensationCase(_ context.Context, compensationCase *paymentdomain.PaymentCompensationCase) error {
	if s.failAt == "compensation" {
		return errors.New("compensation insert failed")
	}
	for i := range s.state.cases {
		if s.state.cases[i].PaymentOrderID == compensationCase.PaymentOrderID && s.state.cases[i].ReasonCode == compensationCase.ReasonCode {
			return nil
		}
	}
	s.state.cases = append(s.state.cases, *compensationCase)
	return nil
}

func runFakeAttemptTransaction(state *attemptMemoryState, failAt string, result apppayment.OutboxAttemptResult) (bool, error) {
	snapshot := *state
	snapshot.cases = append([]paymentdomain.PaymentCompensationCase(nil), state.cases...)
	updated, err := recordAttemptWithinTx(context.Background(), &fakeAttemptPersistence{state: state, failAt: failAt}, state.event.ID, result)
	if err != nil {
		*state = snapshot
	}
	return updated, err
}

func TestDeadTransitionFailuresRollBackEveryChange(t *testing.T) {
	for _, failAt := range []string{"outbox", "fulfillment", "compensation"} {
		t.Run(failAt, func(t *testing.T) {
			state := successfulConfirmState()
			updated, err := runFakeAttemptTransaction(&state, failAt, deadAttempt(paymentdomain.BookingDeliveryConflict))
			if err == nil || updated {
				t.Fatalf("expected %s failure, updated=%v err=%v", failAt, updated, err)
			}
			if state.event.Status != paymentdomain.OutboxStatusPending || state.event.AttemptCount != 0 || state.event.TerminalReasonCode != "" {
				t.Fatalf("outbox change escaped rollback: %+v", state.event)
			}
			if state.order.FulfillmentStatus != paymentdomain.FulfillmentPending || len(state.cases) != 0 {
				t.Fatalf("payment/compensation change escaped rollback: order=%+v cases=%+v", state.order, state.cases)
			}
		})
	}
}

func TestDeliveredAndFulfillmentPersistTogether(t *testing.T) {
	for _, test := range []struct {
		name        string
		eventType   string
		payment     paymentdomain.PaymentOrderStatus
		fulfillment paymentdomain.FulfillmentStatus
	}{
		{name: "confirm", eventType: paymentdomain.BookingConfirmEvent, payment: paymentdomain.OrderStatusSuccess, fulfillment: paymentdomain.FulfillmentBookingConfirmed},
		{name: "fail", eventType: paymentdomain.BookingFailEvent, payment: paymentdomain.OrderStatusFailed, fulfillment: paymentdomain.FulfillmentBookingFailed},
	} {
		t.Run(test.name, func(t *testing.T) {
			state := successfulConfirmState()
			state.event.EventType = test.eventType
			state.order.Status = test.payment
			updated, err := runFakeAttemptTransaction(&state, "", apppayment.OutboxAttemptResult{Status: paymentdomain.OutboxStatusDelivered, AttemptedAt: 200})
			if err != nil || !updated || state.event.Status != paymentdomain.OutboxStatusDelivered || state.order.FulfillmentStatus != test.fulfillment || len(state.cases) != 0 {
				t.Fatalf("delivered state diverged: updated=%v err=%v event=%+v order=%+v cases=%+v", updated, err, state.event, state.order, state.cases)
			}
		})
	}

	state := successfulConfirmState()
	updated, err := runFakeAttemptTransaction(&state, "fulfillment", apppayment.OutboxAttemptResult{Status: paymentdomain.OutboxStatusDelivered, AttemptedAt: 200})
	if err == nil || updated || state.event.Status != paymentdomain.OutboxStatusPending || state.order.FulfillmentStatus != paymentdomain.FulfillmentPending {
		t.Fatalf("delivered outbox escaped failed fulfillment rollback: updated=%v err=%v state=%+v", updated, err, state)
	}
}

func TestDeadTransitionIsTypedIdempotentAndUsesPersistedPaymentEvidence(t *testing.T) {
	state := successfulConfirmState()
	result := deadAttempt(paymentdomain.BookingDeliveryConflict)
	updated, err := runFakeAttemptTransaction(&state, "", result)
	if err != nil || !updated {
		t.Fatalf("first terminal transition failed: updated=%v err=%v", updated, err)
	}
	if state.event.Status != paymentdomain.OutboxStatusDead || state.event.TerminalReasonCode != paymentdomain.BookingDeliveryConflict || state.order.FulfillmentStatus != paymentdomain.FulfillmentRefundRequired || len(state.cases) != 1 {
		t.Fatalf("unexpected terminal state: %+v", state)
	}
	compensationCase := state.cases[0]
	if compensationCase.Status != paymentdomain.CompensationRefundRequired || compensationCase.ReasonCode != paymentdomain.CompensationReasonBookingConflict || compensationCase.AmountVND != state.order.GrossAmount || compensationCase.GatewayOrderReference != state.order.ID.String() || compensationCase.GatewayTransactionNumber != state.order.GatewayTxnRef || compensationCase.GatewayResponseCode != state.order.GatewayResponseCode || compensationCase.GatewayTransactionStatus != state.order.GatewayTransactionStatus || compensationCase.GatewayPaymentDate != state.order.GatewayPaymentDate {
		t.Fatalf("case did not use typed result and persisted payment evidence: %+v", compensationCase)
	}

	updated, err = runFakeAttemptTransaction(&state, "", result)
	if err != nil || updated || state.event.AttemptCount != 1 || len(state.cases) != 1 {
		t.Fatalf("repeated terminal handling was not idempotent: updated=%v err=%v state=%+v", updated, err, state)
	}
}

func TestRetryExhaustionCreatesManualReview(t *testing.T) {
	state := successfulConfirmState()
	updated, err := runFakeAttemptTransaction(&state, "", deadAttempt(paymentdomain.BookingDeliveryUpstream))
	if err != nil || !updated || state.order.Status != paymentdomain.OrderStatusSuccess || state.order.FulfillmentStatus != paymentdomain.FulfillmentManualReview || len(state.cases) != 1 || state.cases[0].Status != paymentdomain.CompensationManualReview || state.cases[0].ReasonCode != paymentdomain.CompensationReasonBookingDeliveryRetries {
		t.Fatalf("retry exhaustion classification failed: updated=%v err=%v state=%+v", updated, err, state)
	}
}

func successfulConfirmState() attemptMemoryState {
	appointmentID := uuid.New()
	orderID := uuid.New()
	return attemptMemoryState{
		event: paymentdomain.OutboxEvent{
			ID: uuid.New(), AggregateType: "PAYMENT_ORDER", AggregateID: orderID,
			EventType: paymentdomain.BookingConfirmEvent, Status: paymentdomain.OutboxStatusPending,
		},
		order: paymentdomain.PaymentOrder{
			ID: orderID, AppointmentID: &appointmentID, Status: paymentdomain.OrderStatusSuccess,
			GrossAmount: money.Money(250000), FulfillmentStatus: paymentdomain.FulfillmentPending,
			GatewayTxnRef: "vnp-transaction-number", GatewayResponseCode: "00",
			GatewayTransactionStatus: "00", GatewayPaymentDate: "20260719100000",
		},
	}
}

func deadAttempt(category paymentdomain.BookingDeliveryFailureCategory) apppayment.OutboxAttemptResult {
	return apppayment.OutboxAttemptResult{
		Status: paymentdomain.OutboxStatusDead, AttemptedAt: 200,
		LastError: "safe diagnostic", FailureCategory: category,
	}
}
