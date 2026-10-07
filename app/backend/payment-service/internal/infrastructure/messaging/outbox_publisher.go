package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	apppayment "payment-service/internal/application/payment"
	paymentdomain "payment-service/internal/domain/payment"
	"payment-service/internal/infrastructure/grpc/paymentpb"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	bookingConfirmEvent = paymentdomain.BookingConfirmEvent
	bookingFailEvent    = paymentdomain.BookingFailEvent
	paymentSucceeded    = paymentdomain.PaymentSucceededEvent
	paymentFailed       = paymentdomain.PaymentFailedEvent
	maximumErrorLength  = 500
)

type PublisherOptions struct {
	PollInterval time.Duration
	MaxAttempts  int
	BaseBackoff  time.Duration
	MaxBackoff   time.Duration
	BatchSize    int
	Clock        func() time.Time
	// Events publish event payment.succeeded / payment.failed lên RabbitMQ.
	Events EventPublisher
}

type Publisher struct {
	repo          apppayment.OutboxRepository
	bookingClient apppayment.BookingServiceClient
	options       PublisherOptions
	runMu         sync.Mutex
}

type bookingEventPayload struct {
	AppointmentID string `json:"appointment_id"`
	OrderID       string `json:"order_id"`
	Status        string `json:"status"`
}

func NewPublisher(repo apppayment.OutboxRepository, bookingClient apppayment.BookingServiceClient) *Publisher {
	return NewPublisherWithOptions(repo, bookingClient, PublisherOptions{})
}

func NewPublisherWithOptions(repo apppayment.OutboxRepository, bookingClient apppayment.BookingServiceClient, options PublisherOptions) *Publisher {
	if options.PollInterval <= 0 {
		options.PollInterval = 5 * time.Second
	}
	if options.MaxAttempts <= 0 {
		options.MaxAttempts = 10
	}
	if options.BaseBackoff <= 0 {
		options.BaseBackoff = 5 * time.Second
	}
	if options.MaxBackoff <= 0 {
		options.MaxBackoff = 5 * time.Minute
	}
	if options.BatchSize <= 0 {
		options.BatchSize = 50
	}
	if options.Clock == nil {
		options.Clock = time.Now
	}
	return &Publisher{repo: repo, bookingClient: bookingClient, options: options}
}

func (p *Publisher) Start(ctx context.Context) {
	log.Printf("[outbox] worker started poll_interval=%s batch_size=%d max_attempts=%d",
		p.options.PollInterval, p.options.BatchSize, p.options.MaxAttempts)
	defer log.Println("[outbox] worker stopped")

	p.publishEligibleEvents(ctx)
	ticker := time.NewTicker(p.options.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.publishEligibleEvents(ctx)
		}
	}
}

func (p *Publisher) publishEligibleEvents(ctx context.Context) {
	if !p.runMu.TryLock() {
		log.Println("[outbox] poll skipped because the previous run is still active")
		return
	}
	defer p.runMu.Unlock()

	events, err := p.repo.GetEligibleEvents(ctx, p.options.Clock().UnixMilli(), p.options.BatchSize)
	if err != nil {
		log.Printf("[outbox] eligible event query failed: %v", err)
		return
	}
	if len(events) == 0 {
		return
	}
	log.Printf("[outbox] fetched batch size=%d", len(events))

	for i := range events {
		if ctx.Err() != nil {
			return
		}
		p.processEvent(ctx, &events[i])
	}
}

func (p *Publisher) processEvent(ctx context.Context, event *paymentdomain.OutboxEvent) {
	err := p.dispatch(ctx, event)
	if ctx.Err() != nil {
		return
	}

	attemptedAt := p.options.Clock()
	attemptNumber := event.AttemptCount + 1
	result := apppayment.OutboxAttemptResult{AttemptedAt: attemptedAt.UnixMilli()}
	category := "delivered"

	if err == nil {
		result.Status = paymentdomain.OutboxStatusDelivered
	} else {
		retryable, failureCategory := classifyDeliveryError(err)
		category = failureCategory
		result.LastError = boundedError(err)
		result.FailureCategory = paymentdomain.BookingDeliveryFailureCategory(failureCategory)
		if retryable && attemptNumber < p.options.MaxAttempts {
			delay := paymentdomain.OutboxRetryDelay(p.options.BaseBackoff, p.options.MaxBackoff, attemptNumber)
			nextAttemptAt := attemptedAt.Add(delay).UnixMilli()
			result.Status = paymentdomain.OutboxStatusRetryWait
			result.NextAttemptAt = &nextAttemptAt
		} else {
			result.Status = paymentdomain.OutboxStatusDead
		}
	}

	updated, persistErr := p.repo.RecordAttempt(ctx, event.ID, result)
	if persistErr != nil {
		log.Printf("[outbox] result persistence failed event_id=%s event_type=%s category=%s error=%v",
			event.ID, event.EventType, category, persistErr)
		return
	}
	if !updated {
		return
	}

	switch result.Status {
	case paymentdomain.OutboxStatusDelivered:
		log.Printf("[outbox] event delivered event_id=%s event_type=%s aggregate_id=%s attempts=%d",
			event.ID, event.EventType, event.AggregateID, attemptNumber)
	case paymentdomain.OutboxStatusRetryWait:
		log.Printf("[outbox] retry scheduled event_id=%s event_type=%s aggregate_id=%s attempts=%d category=%s next_attempt_at=%d",
			event.ID, event.EventType, event.AggregateID, attemptNumber, category, *result.NextAttemptAt)
	case paymentdomain.OutboxStatusDead:
		if category == string(apppayment.BookingDeliveryAuthentication) {
			log.Printf("[outbox] booking authentication failure event_id=%s event_type=%s attempts=%d",
				event.ID, event.EventType, attemptNumber)
		}
		if category == string(apppayment.BookingDeliveryConflict) {
			log.Printf("[outbox] permanent booking conflict event_id=%s event_type=%s attempts=%d",
				event.ID, event.EventType, attemptNumber)
		}
		log.Printf("[outbox] event marked DEAD event_id=%s event_type=%s aggregate_id=%s attempts=%d category=%s",
			event.ID, event.EventType, event.AggregateID, attemptNumber, category)
	}
}

func (p *Publisher) dispatch(ctx context.Context, event *paymentdomain.OutboxEvent) error {
	switch event.EventType {
	case bookingConfirmEvent:
	case bookingFailEvent:
	case paymentSucceeded, paymentFailed:
		return p.publishPaymentStatus(ctx, event)
	default:
		return &permanentEventError{category: "unsupported_event_type", message: fmt.Sprintf("unsupported outbox event type %q", event.EventType)}
	}

	payload, err := parseBookingEventPayload(event.Payload)
	if err != nil {
		return err
	}
	if event.EventType == bookingConfirmEvent {
		return p.bookingClient.ConfirmAppointment(ctx, payload.AppointmentID)
	}
	return p.bookingClient.FailAppointment(ctx, payload.AppointmentID)
}

// publishPaymentStatus chuyển payload outbox sang paymentpb.PaymentStatusChangedEvent và publish
// lên RabbitMQ. event_id = id outbox event nên ổn định qua các lần retry.
func (p *Publisher) publishPaymentStatus(ctx context.Context, event *paymentdomain.OutboxEvent) error {
	if p.options.Events == nil {
		return errors.New("rabbitmq event publisher is not configured")
	}
	body, err := paymentStatusMessage(event)
	if err != nil {
		return err
	}
	return p.options.Events.Publish(ctx, event.EventType, event.ID.String(), body)
}

func paymentStatusMessage(event *paymentdomain.OutboxEvent) ([]byte, error) {
	var payload paymentdomain.PaymentStatusPayload
	if err := json.Unmarshal([]byte(event.Payload), &payload); err != nil {
		return nil, &permanentEventError{category: "malformed_payload", message: "malformed payment status outbox payload"}
	}
	status := paymentpb.PaymentStatus_PAYMENT_STATUS_FAILED
	if payload.Status == paymentdomain.OrderStatusSuccess.String() {
		status = paymentpb.PaymentStatus_PAYMENT_STATUS_SUCCESS
	}
	message := &paymentpb.PaymentStatusChangedEvent{
		EventId:    event.ID.String(),
		EventType:  event.EventType,
		OccurredAt: time.UnixMilli(event.CreatedAt).UTC().Format(time.RFC3339),
		Source:     "payment-service",
		Data: &paymentpb.PaymentStatusChanged{
			OrderId:             payload.OrderID,
			AppointmentId:       payload.AppointmentID,
			PayerId:             payload.PayerID,
			ExpertId:            payload.ExpertID,
			AmountVnd:           payload.AmountVND,
			Status:              status,
			Gateway:             payload.Gateway,
			GatewayTxnRef:       payload.GatewayTxnRef,
			GatewayResponseCode: payload.GatewayResponseCode,
			PaidAt:              payload.PaidAt,
		},
	}
	body, err := protojson.Marshal(message)
	if err != nil {
		return nil, &permanentEventError{category: "malformed_payload", message: "payment status event could not be encoded"}
	}
	return body, nil
}

func parseBookingEventPayload(raw string) (*bookingEventPayload, error) {
	var payload bookingEventPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, &permanentEventError{category: "malformed_payload", message: "malformed booking outbox payload"}
	}
	payload.AppointmentID = strings.TrimSpace(payload.AppointmentID)
	if _, err := uuid.Parse(payload.AppointmentID); err != nil {
		return nil, &permanentEventError{category: "malformed_payload", message: "booking outbox payload has invalid appointment_id"}
	}
	return &payload, nil
}

type permanentEventError struct {
	category string
	message  string
}

func (e *permanentEventError) Error() string { return e.message }

func classifyDeliveryError(err error) (bool, string) {
	var eventErr *permanentEventError
	if errors.As(err, &eventErr) {
		return false, eventErr.category
	}
	var bookingErr *apppayment.BookingDeliveryError
	if errors.As(err, &bookingErr) {
		return bookingErr.Retryable, string(bookingErr.Category)
	}
	return true, "unknown"
}

func boundedError(err error) string {
	message := strings.Join(strings.Fields(err.Error()), " ")
	runes := []rune(message)
	if len(runes) > maximumErrorLength {
		message = string(runes[:maximumErrorLength])
	}
	return message
}
