package outbox

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"payment-service/internal/domain/entity"
	apppayment "payment-service/internal/payment/application"
	paymentdomain "payment-service/internal/payment/domain"

	"github.com/google/uuid"
)

var fixedNow = time.UnixMilli(1_700_000_000_000)

type memoryOutboxState struct {
	mu     sync.Mutex
	events map[uuid.UUID]entity.OutboxEvent
}

type memoryOutboxRepository struct {
	state              *memoryOutboxState
	failRecordAttempts int
	getCalls           atomic.Int32
	getStarted         chan struct{}
	releaseGet         chan struct{}
}

func newMemoryOutboxRepository(events ...entity.OutboxEvent) *memoryOutboxRepository {
	state := &memoryOutboxState{events: make(map[uuid.UUID]entity.OutboxEvent)}
	for _, event := range events {
		state.events[event.ID] = event
	}
	return &memoryOutboxRepository{state: state}
}

func (r *memoryOutboxRepository) GetEligibleEvents(_ context.Context, nowMillis int64, limit int) ([]entity.OutboxEvent, error) {
	r.getCalls.Add(1)
	if r.getStarted != nil {
		select {
		case r.getStarted <- struct{}{}:
		default:
		}
		<-r.releaseGet
	}
	r.state.mu.Lock()
	defer r.state.mu.Unlock()
	var events []entity.OutboxEvent
	for _, event := range r.state.events {
		if paymentdomain.IsOutboxEligible(event.Status, event.NextAttemptAt, nowMillis) {
			events = append(events, event)
		}
	}
	sort.Slice(events, func(i, j int) bool {
		if events[i].CreatedAt == events[j].CreatedAt {
			return events[i].ID.String() < events[j].ID.String()
		}
		return events[i].CreatedAt < events[j].CreatedAt
	})
	if len(events) > limit {
		events = events[:limit]
	}
	return events, nil
}

func (r *memoryOutboxRepository) RecordAttempt(_ context.Context, eventID uuid.UUID, result AttemptResult) (bool, error) {
	r.state.mu.Lock()
	defer r.state.mu.Unlock()
	if r.failRecordAttempts > 0 {
		r.failRecordAttempts--
		return false, errors.New("simulated local persistence failure")
	}
	event := r.state.events[eventID]
	if event.Status == paymentdomain.OutboxStatusDelivered || event.Status == paymentdomain.OutboxStatusDead {
		return false, nil
	}
	event.AttemptCount++
	event.LastAttemptAt = &result.AttemptedAt
	event.NextAttemptAt = result.NextAttemptAt
	event.Status = result.Status
	event.Published = result.Status == paymentdomain.OutboxStatusDelivered
	if result.LastError == "" {
		event.LastError = nil
	} else {
		message := result.LastError
		event.LastError = &message
	}
	if result.Status == paymentdomain.OutboxStatusDelivered {
		event.DeliveredAt = &result.AttemptedAt
	}
	r.state.events[eventID] = event
	return true, nil
}

func (r *memoryOutboxRepository) event(id uuid.UUID) entity.OutboxEvent {
	r.state.mu.Lock()
	defer r.state.mu.Unlock()
	return r.state.events[id]
}

type fakeBookingClient struct {
	mu         sync.Mutex
	errorsByID map[string]error
	calls      []string
	active     int
	maxActive  int
	delay      time.Duration
}

func (c *fakeBookingClient) GetPaymentEligibility(context.Context, string, uuid.UUID) (*apppayment.PaymentEligibility, error) {
	return nil, nil
}

func (c *fakeBookingClient) ConfirmAppointment(_ context.Context, appointmentID string) error {
	return c.call("SUCCESS", appointmentID)
}

func (c *fakeBookingClient) FailAppointment(_ context.Context, appointmentID string) error {
	return c.call("FAILED", appointmentID)
}

func (c *fakeBookingClient) call(status, appointmentID string) error {
	c.mu.Lock()
	c.calls = append(c.calls, status+":"+appointmentID)
	c.active++
	if c.active > c.maxActive {
		c.maxActive = c.active
	}
	err := c.errorsByID[appointmentID]
	c.mu.Unlock()
	if c.delay > 0 {
		time.Sleep(c.delay)
	}
	c.mu.Lock()
	c.active--
	c.mu.Unlock()
	return err
}

func (c *fakeBookingClient) callCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.calls)
}

func testPublisher(repo Repository, client apppayment.BookingServiceClient, options PublisherOptions) *Publisher {
	options.Clock = func() time.Time { return fixedNow }
	if options.MaxAttempts == 0 {
		options.MaxAttempts = 3
	}
	if options.BaseBackoff == 0 {
		options.BaseBackoff = 5 * time.Second
	}
	if options.MaxBackoff == 0 {
		options.MaxBackoff = time.Minute
	}
	if options.BatchSize == 0 {
		options.BatchSize = 50
	}
	return NewPublisherWithOptions(repo, client, options)
}

func bookingEvent(eventType string, createdAt int64) entity.OutboxEvent {
	appointmentID := uuid.New()
	status := "SUCCESS"
	if eventType == bookingFailEvent {
		status = "FAILED"
	}
	return entity.OutboxEvent{
		ID:            uuid.New(),
		AggregateType: "PAYMENT_ORDER",
		AggregateID:   uuid.New(),
		EventType:     eventType,
		Payload: fmt.Sprintf(`{"appointment_id":%q,"order_id":%q,"status":%q}`,
			appointmentID.String(), uuid.New().String(), status),
		Status:    paymentdomain.OutboxStatusPending,
		CreatedAt: createdAt,
	}
}

func appointmentIDFromEvent(t *testing.T, event entity.OutboxEvent) string {
	t.Helper()
	payload, err := parseBookingEventPayload(event.Payload)
	if err != nil {
		t.Fatal(err)
	}
	return payload.AppointmentID
}

func TestSuccessfulBookingDeliveryMarksDelivered(t *testing.T) {
	for _, eventType := range []string{bookingConfirmEvent, bookingFailEvent} {
		t.Run(eventType, func(t *testing.T) {
			event := bookingEvent(eventType, 1)
			repo := newMemoryOutboxRepository(event)
			client := &fakeBookingClient{}
			testPublisher(repo, client, PublisherOptions{}).publishEligibleEvents(context.Background())

			stored := repo.event(event.ID)
			if stored.Status != paymentdomain.OutboxStatusDelivered || !stored.Published {
				t.Fatalf("unexpected state: %+v", stored)
			}
			if stored.AttemptCount != 1 || stored.LastAttemptAt == nil || stored.DeliveredAt == nil || *stored.DeliveredAt != fixedNow.UnixMilli() {
				t.Fatalf("attempt metadata not persisted: %+v", stored)
			}
			if stored.NextAttemptAt != nil || stored.LastError != nil {
				t.Fatalf("delivered event retained failure metadata: %+v", stored)
			}

			testPublisher(repo, client, PublisherOptions{}).publishEligibleEvents(context.Background())
			if client.callCount() != 1 {
				t.Fatalf("delivered event was sent again: calls=%d", client.callCount())
			}
		})
	}
}

func TestRetryableDeliverySchedulesPersistentBackoff(t *testing.T) {
	tests := []struct {
		name     string
		category apppayment.BookingDeliveryFailureCategory
	}{
		{name: "network", category: apppayment.BookingDeliveryNetwork},
		{name: "timeout", category: apppayment.BookingDeliveryTimeout},
		{name: "rate limited", category: apppayment.BookingDeliveryRateLimited},
		{name: "upstream", category: apppayment.BookingDeliveryUpstream},
		{name: "ambiguous 2xx response", category: apppayment.BookingDeliveryMalformedResponse},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			event := bookingEvent(bookingConfirmEvent, 1)
			appointmentID := appointmentIDFromEvent(t, event)
			repo := newMemoryOutboxRepository(event)
			client := &fakeBookingClient{errorsByID: map[string]error{
				appointmentID: &apppayment.BookingDeliveryError{Category: test.category, Retryable: true, Message: "safe transient error"},
			}}
			testPublisher(repo, client, PublisherOptions{}).publishEligibleEvents(context.Background())

			stored := repo.event(event.ID)
			expectedNext := fixedNow.Add(5 * time.Second).UnixMilli()
			if stored.Status != paymentdomain.OutboxStatusRetryWait || stored.AttemptCount != 1 || stored.NextAttemptAt == nil || *stored.NextAttemptAt != expectedNext {
				t.Fatalf("unexpected retry state: %+v", stored)
			}
			if stored.LastError == nil || *stored.LastError != "safe transient error" {
				t.Fatalf("unexpected last_error: %+v", stored.LastError)
			}

			// A new repository and worker object sees the same persisted state, but not before next_attempt_at.
			restartedRepo := &memoryOutboxRepository{state: repo.state}
			testPublisher(restartedRepo, client, PublisherOptions{}).publishEligibleEvents(context.Background())
			if client.callCount() != 1 {
				t.Fatalf("retry ran before persisted next_attempt_at: calls=%d", client.callCount())
			}
		})
	}
}

func TestPermanentDeliveryFailuresBecomeDead(t *testing.T) {
	tests := []struct {
		name       string
		clientErr  error
		event      entity.OutboxEvent
		wantCalls  int
		wantReason string
	}{
		{name: "not found", clientErr: &apppayment.BookingDeliveryError{Category: apppayment.BookingDeliveryNotFound, Message: "HTTP 404"}, event: bookingEvent(bookingConfirmEvent, 1), wantCalls: 1},
		{name: "opposite terminal conflict", clientErr: &apppayment.BookingDeliveryError{Category: apppayment.BookingDeliveryConflict, Message: "HTTP 409"}, event: bookingEvent(bookingConfirmEvent, 1), wantCalls: 1},
		{name: "authentication", clientErr: &apppayment.BookingDeliveryError{Category: apppayment.BookingDeliveryAuthentication, Message: "HTTP 401"}, event: bookingEvent(bookingConfirmEvent, 1), wantCalls: 1},
		{name: "unsupported event", event: bookingEvent("wallet.payment.received", 1), wantCalls: 0, wantReason: "unsupported outbox event type"},
		{name: "malformed payload", event: bookingEvent(bookingConfirmEvent, 1), wantCalls: 0, wantReason: "malformed booking outbox payload"},
	}
	tests[4].event.Payload = "not-json"
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			appointmentID := ""
			if test.wantCalls > 0 {
				appointmentID = appointmentIDFromEvent(t, test.event)
			}
			repo := newMemoryOutboxRepository(test.event)
			client := &fakeBookingClient{errorsByID: map[string]error{appointmentID: test.clientErr}}
			publisher := testPublisher(repo, client, PublisherOptions{})
			publisher.publishEligibleEvents(context.Background())

			stored := repo.event(test.event.ID)
			if stored.Status != paymentdomain.OutboxStatusDead || stored.AttemptCount != 1 || stored.NextAttemptAt != nil {
				t.Fatalf("unexpected dead state: %+v", stored)
			}
			if client.callCount() != test.wantCalls {
				t.Fatalf("expected %d booking calls, got %d", test.wantCalls, client.callCount())
			}
			if test.wantReason != "" && (stored.LastError == nil || !strings.Contains(*stored.LastError, test.wantReason)) {
				t.Fatalf("expected last_error containing %q, got %+v", test.wantReason, stored.LastError)
			}
			publisher.publishEligibleEvents(context.Background())
			if client.callCount() != test.wantCalls {
				t.Fatal("dead event was retried")
			}
		})
	}
}

func TestLegacyBookingPayloadWithOnlyAppointmentIDStillDelivers(t *testing.T) {
	event := bookingEvent(bookingConfirmEvent, 1)
	appointmentID := appointmentIDFromEvent(t, event)
	event.Payload = fmt.Sprintf(`{"appointment_id":%q}`, appointmentID)
	repo := newMemoryOutboxRepository(event)
	client := &fakeBookingClient{}
	testPublisher(repo, client, PublisherOptions{}).publishEligibleEvents(context.Background())
	if repo.event(event.ID).Status != paymentdomain.OutboxStatusDelivered || client.callCount() != 1 {
		t.Fatalf("legacy booking payload was not delivered: %+v", repo.event(event.ID))
	}
}

func TestRetryableFailureAtMaxAttemptsBecomesDead(t *testing.T) {
	event := bookingEvent(bookingConfirmEvent, 1)
	event.AttemptCount = 2
	appointmentID := appointmentIDFromEvent(t, event)
	repo := newMemoryOutboxRepository(event)
	client := &fakeBookingClient{errorsByID: map[string]error{
		appointmentID: &apppayment.BookingDeliveryError{Category: apppayment.BookingDeliveryUpstream, Retryable: true, Message: "HTTP 503"},
	}}
	testPublisher(repo, client, PublisherOptions{MaxAttempts: 3}).publishEligibleEvents(context.Background())
	stored := repo.event(event.ID)
	if stored.Status != paymentdomain.OutboxStatusDead || stored.AttemptCount != 3 || stored.NextAttemptAt != nil {
		t.Fatalf("unexpected max-attempt state: %+v", stored)
	}
}

func TestOneFailedEventDoesNotStopLaterEventsAndProcessingIsSequential(t *testing.T) {
	first := bookingEvent(bookingConfirmEvent, 1)
	second := bookingEvent(bookingFailEvent, 2)
	repo := newMemoryOutboxRepository(first, second)
	client := &fakeBookingClient{
		errorsByID: map[string]error{
			appointmentIDFromEvent(t, first): &apppayment.BookingDeliveryError{Category: apppayment.BookingDeliveryNetwork, Retryable: true, Message: "network"},
		},
		delay: 5 * time.Millisecond,
	}
	testPublisher(repo, client, PublisherOptions{}).publishEligibleEvents(context.Background())
	if repo.event(first.ID).Status != paymentdomain.OutboxStatusRetryWait || repo.event(second.ID).Status != paymentdomain.OutboxStatusDelivered {
		t.Fatalf("batch did not continue: first=%s second=%s", repo.event(first.ID).Status, repo.event(second.ID).Status)
	}
	if client.callCount() != 2 || client.maxActive != 1 {
		t.Fatalf("expected two sequential calls, calls=%d max_active=%d", client.callCount(), client.maxActive)
	}
}

type idempotentBookingClient struct {
	fakeBookingClient
	mutations atomic.Int32
}

func (c *idempotentBookingClient) ConfirmAppointment(_ context.Context, appointmentID string) error {
	c.fakeBookingClient.call("SUCCESS", appointmentID)
	c.mutations.CompareAndSwap(0, 1)
	return nil
}

func TestCrashAfterBookingAcceptanceRetriesToIdempotentDelivery(t *testing.T) {
	event := bookingEvent(bookingConfirmEvent, 1)
	repo := newMemoryOutboxRepository(event)
	repo.failRecordAttempts = 1
	client := &idempotentBookingClient{}

	testPublisher(repo, client, PublisherOptions{}).publishEligibleEvents(context.Background())
	if repo.event(event.ID).Status != paymentdomain.OutboxStatusPending {
		t.Fatal("event was lost after local persistence failure")
	}

	restartedRepo := &memoryOutboxRepository{state: repo.state}
	testPublisher(restartedRepo, client, PublisherOptions{}).publishEligibleEvents(context.Background())
	if restartedRepo.event(event.ID).Status != paymentdomain.OutboxStatusDelivered {
		t.Fatalf("event was not delivered after restart: %+v", restartedRepo.event(event.ID))
	}
	if client.callCount() != 2 || client.mutations.Load() != 1 {
		t.Fatalf("expected two deliveries but one booking mutation, calls=%d mutations=%d", client.callCount(), client.mutations.Load())
	}
}

func TestSelectionOrderingBatchSizeAndExactRetryTime(t *testing.T) {
	first := bookingEvent(bookingConfirmEvent, 1)
	second := bookingEvent(bookingConfirmEvent, 2)
	third := bookingEvent(bookingConfirmEvent, 3)
	third.Status = paymentdomain.OutboxStatusRetryWait
	exact := fixedNow.UnixMilli()
	third.NextAttemptAt = &exact
	future := bookingEvent(bookingConfirmEvent, 0)
	future.Status = paymentdomain.OutboxStatusRetryWait
	futureAt := fixedNow.Add(time.Millisecond).UnixMilli()
	future.NextAttemptAt = &futureAt
	delivered := bookingEvent(bookingConfirmEvent, 0)
	delivered.Status = paymentdomain.OutboxStatusDelivered
	dead := bookingEvent(bookingConfirmEvent, 0)
	dead.Status = paymentdomain.OutboxStatusDead
	repo := newMemoryOutboxRepository(third, dead, second, delivered, future, first)

	events, err := repo.GetEligibleEvents(context.Background(), fixedNow.UnixMilli(), 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 || events[0].ID != first.ID || events[1].ID != second.ID {
		t.Fatalf("unexpected deterministic batch: %+v", events)
	}
	events, _ = repo.GetEligibleEvents(context.Background(), fixedNow.UnixMilli(), 10)
	if len(events) != 3 || events[2].ID != third.ID {
		t.Fatalf("exact next-attempt event was not eligible: %+v", events)
	}
}

func TestOverlappingPollsDoNotProcessSameBatch(t *testing.T) {
	repo := newMemoryOutboxRepository()
	repo.getStarted = make(chan struct{}, 1)
	repo.releaseGet = make(chan struct{})
	publisher := testPublisher(repo, &fakeBookingClient{}, PublisherOptions{})
	done := make(chan struct{})
	go func() {
		publisher.publishEligibleEvents(context.Background())
		close(done)
	}()
	<-repo.getStarted
	publisher.publishEligibleEvents(context.Background())
	close(repo.releaseGet)
	<-done
	if repo.getCalls.Load() != 1 {
		t.Fatalf("expected one repository poll, got %d", repo.getCalls.Load())
	}
}
