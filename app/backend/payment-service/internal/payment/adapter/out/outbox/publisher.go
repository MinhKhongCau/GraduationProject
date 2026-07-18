package outbox

import (
	"context"
	"encoding/json"
	"log"
	apppayment "payment-service/internal/payment/application"
	"payment-service/pkg/rabbitmq"
	"time"
)

// Publisher là Background Worker quét bảng outbox_events và dispatch từng event
// sang đúng đích tương ứng dựa trên event_type.
//
// Thiết kế dispatch theo event_type:
//   - "wallet.payment.received"    → Publish sang RabbitMQ (cho các consumer khác)
//   - "booking.appointment.confirm" → Gọi REST sang Booking Service (Internal call)
//   - "booking.appointment.fail"   → Gọi REST sang Booking Service (Internal call)
//
// Nếu dispatch thất bại, event không được đánh dấu published → Worker sẽ retry lần sau.
type Publisher struct {
	repo          Repository
	bookingClient apppayment.BookingServiceClient
}

// NewPublisher tạo Publisher mới.
// bookingClient: inject BookingServiceClient để gọi sang Booking Service.
func NewPublisher(repo Repository, bookingClient apppayment.BookingServiceClient) *Publisher {
	return &Publisher{
		repo:          repo,
		bookingClient: bookingClient,
	}
}

// Start chạy Publisher Worker vô hạn với interval 2 giây.
// Dừng lại khi ctx bị cancel (e.g. khi service shutdown).
func (p *Publisher) Start(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	log.Println("⏳ Outbox Publisher Worker started")

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping Outbox Publisher Worker")
			return
		case <-ticker.C:
			p.publishPendingEvents(ctx)
		}
	}
}

func (p *Publisher) publishPendingEvents(ctx context.Context) {
	events, err := p.repo.GetUnpublishedEvents(20)
	if err != nil {
		log.Printf("Outbox Publisher: Error fetching unpublished events: %v", err)
		return
	}

	if len(events) == 0 {
		return
	}

	var publishedIDs []string
	for _, event := range events {
		log.Printf("[OUTBOX DISPATCH] EventID: %s | Type: %s", event.ID, event.EventType)

		var dispatchErr error

		switch event.EventType {

		// ── Wallet events → publish sang RabbitMQ (các consumer khác lắng nghe) ──
		case "wallet.payment.received":
			dispatchErr = rabbitmq.PublishEvent(event.EventType, event.Payload)

		// ── Booking events → gọi REST sang Booking Service qua Internal JWT Auth ──
		case "booking.appointment.confirm":
			appointmentID := extractAppointmentID(event.Payload)
			if appointmentID == "" {
				log.Printf("[OUTBOX] booking.appointment.confirm: missing appointment_id in payload, skipping event %s", event.ID)
				publishedIDs = append(publishedIDs, event.ID.String()) // mark done để không retry mãi
				continue
			}
			dispatchErr = p.bookingClient.ConfirmAppointment(ctx, appointmentID)

		case "booking.appointment.fail":
			appointmentID := extractAppointmentID(event.Payload)
			if appointmentID == "" {
				log.Printf("[OUTBOX] booking.appointment.fail: missing appointment_id in payload, skipping event %s", event.ID)
				publishedIDs = append(publishedIDs, event.ID.String())
				continue
			}
			dispatchErr = p.bookingClient.FailAppointment(ctx, appointmentID)

		default:
			// Event type không xác định — log và bỏ qua (đánh dấu published để không retry)
			log.Printf("[OUTBOX] Unknown event_type: %q — marking as published to skip", event.EventType)
			publishedIDs = append(publishedIDs, event.ID.String())
			continue
		}

		if dispatchErr != nil {
			log.Printf("[OUTBOX] Failed to dispatch event %s (type: %s): %v — will retry next tick",
				event.ID, event.EventType, dispatchErr)
			// Không append vào publishedIDs → next tick sẽ retry
			continue
		}

		publishedIDs = append(publishedIDs, event.ID.String())
	}

	if len(publishedIDs) > 0 {
		if err := p.repo.MarkAsPublished(publishedIDs); err != nil {
			log.Printf("[OUTBOX] Failed to mark events as published: %v", err)
		}
	}
}

// extractAppointmentID parse appointment_id từ JSON payload của outbox event.
func extractAppointmentID(payload string) string {
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(payload), &m); err != nil {
		return ""
	}
	if v, ok := m["appointment_id"].(string); ok {
		return v
	}
	return ""
}
