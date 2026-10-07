package messaging

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	paymentdomain "payment-service/internal/domain/payment"
	"payment-service/internal/infrastructure/grpc/paymentpb"

	"github.com/google/uuid"
	"google.golang.org/protobuf/encoding/protojson"
)

type publishedMessage struct {
	routingKey string
	messageID  string
	body       []byte
}

type fakeEventPublisher struct {
	mu       sync.Mutex
	err      error
	messages []publishedMessage
}

func (p *fakeEventPublisher) Publish(_ context.Context, routingKey, messageID string, body []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.messages = append(p.messages, publishedMessage{routingKey: routingKey, messageID: messageID, body: body})
	return p.err
}

func paymentStatusEvent(t *testing.T, eventType string, status paymentdomain.PaymentOrderStatus) (paymentdomain.OutboxEvent, *paymentdomain.PaymentOrder) {
	t.Helper()
	appointmentID := uuid.New()
	paidAt := int64(1_700_000_000_500)
	order := &paymentdomain.PaymentOrder{
		ID:                  uuid.New(),
		PayerID:             uuid.New(),
		ExpertID:            uuid.New(),
		AppointmentID:       &appointmentID,
		GrossAmount:         300000,
		Gateway:             "VNPAY",
		GatewayTxnRef:       "14512345",
		GatewayResponseCode: "00",
		Status:              status,
	}
	if status == paymentdomain.OrderStatusSuccess {
		order.PaidAt = &paidAt
	}
	payload, err := paymentdomain.PaymentStatusOutboxPayload(order)
	if err != nil {
		t.Fatal(err)
	}
	return paymentdomain.OutboxEvent{
		ID:            uuid.New(),
		AggregateType: "PAYMENT_ORDER",
		AggregateID:   order.ID,
		EventType:     eventType,
		Payload:       payload,
		Status:        paymentdomain.OutboxStatusPending,
		CreatedAt:     1_700_000_000_000,
	}, order
}

func TestPaymentStatusEventIsPublishedToRabbitMQ(t *testing.T) {
	tests := []struct {
		eventType string
		status    paymentdomain.PaymentOrderStatus
		want      paymentpb.PaymentStatus
	}{
		{eventType: paymentSucceeded, status: paymentdomain.OrderStatusSuccess, want: paymentpb.PaymentStatus_PAYMENT_STATUS_SUCCESS},
		{eventType: paymentFailed, status: paymentdomain.OrderStatusFailed, want: paymentpb.PaymentStatus_PAYMENT_STATUS_FAILED},
	}
	for _, test := range tests {
		t.Run(test.eventType, func(t *testing.T) {
			event, order := paymentStatusEvent(t, test.eventType, test.status)
			repo := newMemoryOutboxRepository(event)
			booking := &fakeBookingClient{}
			events := &fakeEventPublisher{}
			testPublisher(repo, booking, PublisherOptions{Events: events}).publishEligibleEvents(context.Background())

			if stored := repo.event(event.ID); stored.Status != paymentdomain.OutboxStatusDelivered || !stored.Published {
				t.Fatalf("unexpected outbox state: %+v", stored)
			}
			if booking.callCount() != 0 {
				t.Fatalf("payment status event must not call booking-service, calls=%d", booking.callCount())
			}
			if len(events.messages) != 1 {
				t.Fatalf("expected one published message, got %d", len(events.messages))
			}
			msg := events.messages[0]
			if msg.routingKey != test.eventType || msg.messageID != event.ID.String() {
				t.Fatalf("unexpected routing key/message id: %q %q", msg.routingKey, msg.messageID)
			}

			var decoded paymentpb.PaymentStatusChangedEvent
			if err := protojson.Unmarshal(msg.body, &decoded); err != nil {
				t.Fatalf("message is not a PaymentStatusChangedEvent: %v", err)
			}
			if decoded.GetEventId() != event.ID.String() || decoded.GetEventType() != test.eventType || decoded.GetSource() != "payment-service" {
				t.Fatalf("unexpected envelope: %+v", &decoded)
			}
			if decoded.GetOccurredAt() != time.UnixMilli(event.CreatedAt).UTC().Format(time.RFC3339) {
				t.Fatalf("unexpected occurredAt: %s", decoded.GetOccurredAt())
			}
			data := decoded.GetData()
			if data.GetOrderId() != order.ID.String() || data.GetAppointmentId() != order.AppointmentID.String() ||
				data.GetAmountVnd() != 300000 || data.GetStatus() != test.want || data.GetGatewayTxnRef() != "14512345" {
				t.Fatalf("unexpected data: %+v", data)
			}
		})
	}
}

func TestPaymentStatusEventEnvelopeUsesConventionKeys(t *testing.T) {
	event, _ := paymentStatusEvent(t, paymentSucceeded, paymentdomain.OrderStatusSuccess)
	body, err := paymentStatusMessage(&event)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{`"eventId"`, `"eventType"`, `"occurredAt"`, `"source"`, `"data"`, `"appointmentId"`} {
		if !strings.Contains(string(body), key) {
			t.Fatalf("expected key %s in %s", key, body)
		}
	}
}

func TestPaymentStatusEventPublishFailureIsRetried(t *testing.T) {
	event, _ := paymentStatusEvent(t, paymentSucceeded, paymentdomain.OrderStatusSuccess)
	repo := newMemoryOutboxRepository(event)
	events := &fakeEventPublisher{err: errors.New("rabbitmq unavailable")}
	testPublisher(repo, &fakeBookingClient{}, PublisherOptions{Events: events}).publishEligibleEvents(context.Background())

	stored := repo.event(event.ID)
	if stored.Status != paymentdomain.OutboxStatusRetryWait || stored.NextAttemptAt == nil || stored.Published {
		t.Fatalf("expected retry wait, got %+v", stored)
	}
}

func TestPaymentStatusEventWithoutPublisherIsRetried(t *testing.T) {
	event, _ := paymentStatusEvent(t, paymentFailed, paymentdomain.OrderStatusFailed)
	repo := newMemoryOutboxRepository(event)
	testPublisher(repo, &fakeBookingClient{}, PublisherOptions{}).publishEligibleEvents(context.Background())

	if stored := repo.event(event.ID); stored.Status != paymentdomain.OutboxStatusRetryWait {
		t.Fatalf("expected retry wait, got %+v", stored)
	}
}

func TestMalformedPaymentStatusPayloadIsDead(t *testing.T) {
	event, _ := paymentStatusEvent(t, paymentSucceeded, paymentdomain.OrderStatusSuccess)
	event.Payload = "{not json"
	repo := newMemoryOutboxRepository(event)
	events := &fakeEventPublisher{}
	testPublisher(repo, &fakeBookingClient{}, PublisherOptions{Events: events}).publishEligibleEvents(context.Background())

	if stored := repo.event(event.ID); stored.Status != paymentdomain.OutboxStatusDead {
		t.Fatalf("expected DEAD, got %+v", stored)
	}
	if len(events.messages) != 0 {
		t.Fatalf("malformed payload must not be published")
	}
}
