package messaging

import (
	"errors"
	"fmt"
	"testing"

	appappointment "booking-service/internal/application/appointment"
	"booking-service/internal/infrastructure/grpc/paymentpb"

	"github.com/google/uuid"
	"google.golang.org/protobuf/encoding/protojson"
)

type fakeHandler struct {
	commands []appappointment.HandlePaymentResultCommand
	err      error
}

func (f *fakeHandler) HandlePaymentResult(command appappointment.HandlePaymentResultCommand) error {
	f.commands = append(f.commands, command)
	return f.err
}

func eventBody(t *testing.T, appointmentID string, status paymentpb.PaymentStatus) []byte {
	t.Helper()
	body, err := protojson.Marshal(&paymentpb.PaymentStatusChangedEvent{
		EventId:    uuid.NewString(),
		EventType:  PaymentSucceededRouteKey,
		OccurredAt: "2026-10-07T10:00:00Z",
		Source:     "payment-service",
		Data: &paymentpb.PaymentStatusChanged{
			OrderId:       uuid.NewString(),
			AppointmentId: appointmentID,
			AmountVnd:     300000,
			Status:        status,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func TestProcessAppliesPaymentResult(t *testing.T) {
	tests := []struct {
		status paymentpb.PaymentStatus
		want   appappointment.PaymentResultStatus
	}{
		{paymentpb.PaymentStatus_PAYMENT_STATUS_SUCCESS, appappointment.PaymentResultSuccess},
		{paymentpb.PaymentStatus_PAYMENT_STATUS_FAILED, appappointment.PaymentResultFailed},
	}
	for _, test := range tests {
		t.Run(test.status.String(), func(t *testing.T) {
			handler := &fakeHandler{}
			appointmentID := uuid.NewString()
			outcome, err := NewPaymentEventConsumer(RabbitMQConfig{}, handler).process(eventBody(t, appointmentID, test.status))
			if outcome != outcomeAck || err != nil {
				t.Fatalf("expected ack, got outcome=%d err=%v", outcome, err)
			}
			if len(handler.commands) != 1 || handler.commands[0].AppointmentID != appointmentID || handler.commands[0].Status != test.want {
				t.Fatalf("unexpected commands: %+v", handler.commands)
			}
		})
	}
}

func TestProcessAcceptsEventsWithUnknownFields(t *testing.T) {
	handler := &fakeHandler{}
	appointmentID := uuid.NewString()
	body := []byte(fmt.Sprintf(`{"eventId":"e1","eventType":"payment.succeeded","futureField":1,"data":{"appointmentId":%q,"status":"PAYMENT_STATUS_SUCCESS","newField":"x"}}`, appointmentID))
	if outcome, err := NewPaymentEventConsumer(RabbitMQConfig{}, handler).process(body); outcome != outcomeAck || err != nil {
		t.Fatalf("expected ack, got outcome=%d err=%v", outcome, err)
	}
	if len(handler.commands) != 1 {
		t.Fatalf("expected handler call, got %+v", handler.commands)
	}
}

func TestProcessIgnoresOrdersWithoutAppointment(t *testing.T) {
	handler := &fakeHandler{}
	outcome, err := NewPaymentEventConsumer(RabbitMQConfig{}, handler).process(eventBody(t, "", paymentpb.PaymentStatus_PAYMENT_STATUS_SUCCESS))
	if outcome != outcomeAck || err != nil || len(handler.commands) != 0 {
		t.Fatalf("expected silent ack, got outcome=%d err=%v commands=%+v", outcome, err, handler.commands)
	}
}

func TestProcessDeadLettersPermanentFailures(t *testing.T) {
	consumer := NewPaymentEventConsumer(RabbitMQConfig{}, &fakeHandler{})
	if outcome, _ := consumer.process([]byte("{not json")); outcome != outcomeDeadLetter {
		t.Fatalf("malformed body: expected dead letter, got %d", outcome)
	}
	if outcome, _ := consumer.process(eventBody(t, uuid.NewString(), paymentpb.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED)); outcome != outcomeDeadLetter {
		t.Fatalf("unspecified status: expected dead letter, got %d", outcome)
	}
	for _, permanent := range []error{
		appappointment.ErrNotFound,
		fmt.Errorf("%w: cancelled appointment cannot be confirmed", appappointment.ErrPaymentResultConflict),
	} {
		consumer := NewPaymentEventConsumer(RabbitMQConfig{}, &fakeHandler{err: permanent})
		if outcome, _ := consumer.process(eventBody(t, uuid.NewString(), paymentpb.PaymentStatus_PAYMENT_STATUS_SUCCESS)); outcome != outcomeDeadLetter {
			t.Fatalf("%v: expected dead letter, got %d", permanent, outcome)
		}
	}
}

func TestProcessRetriesTransientFailures(t *testing.T) {
	consumer := NewPaymentEventConsumer(RabbitMQConfig{}, &fakeHandler{err: errors.New("database is down")})
	if outcome, err := consumer.process(eventBody(t, uuid.NewString(), paymentpb.PaymentStatus_PAYMENT_STATUS_SUCCESS)); outcome != outcomeRetry || err == nil {
		t.Fatalf("expected retry, got outcome=%d err=%v", outcome, err)
	}
}
