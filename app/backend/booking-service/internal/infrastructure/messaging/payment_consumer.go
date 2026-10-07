// File: internal/infrastructure/messaging/payment_consumer.go
package messaging

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	appappointment "booking-service/internal/application/appointment"
	"booking-service/internal/infrastructure/grpc/paymentpb"

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/encoding/protojson"
)

// Naming theo RABBITMQ_CONVENTION.md: <domain>.exchange/events, <entity>.<action>, <service>.queue.
const (
	PaymentExchange          = "payment.events"
	BookingQueue             = "booking.queue"
	BookingDeadLetterQueue   = "booking.queue.dlq"
	PaymentSucceededRouteKey = "payment.succeeded"
	PaymentFailedRouteKey    = "payment.failed"

	maxRetries       = 3
	retryDelay       = 2 * time.Second
	reconnectDelay   = 5 * time.Second
	retryCountHeader = "x-retry-count"
)

// PaymentResultHandler là phần usecase lịch hẹn mà consumer cần (idempotent).
type PaymentResultHandler interface {
	HandlePaymentResult(command appappointment.HandlePaymentResultCommand) error
}

type RabbitMQConfig struct {
	Host string
	Port string
	User string
	Pass string
}

// PaymentEventConsumer nhận event payment.succeeded / payment.failed (schema paymentpb) và cập nhật
// lịch hẹn. Đây là đường giao nhận thứ hai bên cạnh gRPC ApplyPaymentResult; vì
// HandlePaymentResult idempotent nên nhận trùng từ cả hai đường là an toàn.
type PaymentEventConsumer struct {
	cfg     RabbitMQConfig
	handler PaymentResultHandler
}

func NewPaymentEventConsumer(cfg RabbitMQConfig, handler PaymentResultHandler) *PaymentEventConsumer {
	return &PaymentEventConsumer{cfg: cfg, handler: handler}
}

// Start chạy consumer trong goroutine riêng và tự kết nối lại khi broker mất kết nối.
// Không làm service dừng khởi động nếu RabbitMQ chưa sẵn sàng.
func (c *PaymentEventConsumer) Start(ctx context.Context) {
	go func() {
		for {
			err := c.consume(ctx)
			if ctx.Err() != nil {
				return
			}
			log.Printf("booking-service: WARNING payment event consumer stopped: %v — reconnecting in %s", err, reconnectDelay)
			select {
			case <-ctx.Done():
				return
			case <-time.After(reconnectDelay):
			}
		}
	}()
}

func (c *PaymentEventConsumer) consume(ctx context.Context) error {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", c.cfg.User, c.cfg.Pass, c.cfg.Host, c.cfg.Port)
	conn, err := amqp.Dial(url)
	if err != nil {
		return fmt.Errorf("connect amqp://%s:***@%s:%s/: %w", c.cfg.User, c.cfg.Host, c.cfg.Port, err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("open channel: %w", err)
	}
	defer ch.Close()

	if err := declareTopology(ch); err != nil {
		return err
	}
	if err := ch.Qos(10, 0, false); err != nil {
		return fmt.Errorf("set QoS: %w", err)
	}
	msgs, err := ch.Consume(BookingQueue, "booking-service", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume %q: %w", BookingQueue, err)
	}
	log.Printf("booking-service: listening for %q/%q events on queue %q", PaymentSucceededRouteKey, PaymentFailedRouteKey, BookingQueue)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				return errors.New("delivery channel closed")
			}
			c.handleDelivery(ch, msg)
		}
	}
}

func declareTopology(ch *amqp.Channel) error {
	if err := ch.ExchangeDeclare(PaymentExchange, "topic", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare exchange %q: %w", PaymentExchange, err)
	}
	if _, err := ch.QueueDeclare(BookingQueue, true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare queue %q: %w", BookingQueue, err)
	}
	if _, err := ch.QueueDeclare(BookingDeadLetterQueue, true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare DLQ %q: %w", BookingDeadLetterQueue, err)
	}
	for _, key := range []string{PaymentSucceededRouteKey, PaymentFailedRouteKey} {
		if err := ch.QueueBind(BookingQueue, key, PaymentExchange, false, nil); err != nil {
			return fmt.Errorf("bind %q to %q: %w", BookingQueue, key, err)
		}
	}
	return nil
}

type deliveryOutcome int

const (
	outcomeAck deliveryOutcome = iota
	outcomeRetry
	outcomeDeadLetter
)

func (c *PaymentEventConsumer) handleDelivery(ch *amqp.Channel, msg amqp.Delivery) {
	outcome, err := c.process(msg.Body)
	switch outcome {
	case outcomeRetry:
		retryCount := currentRetryCount(msg.Headers)
		if retryCount < maxRetries {
			log.Printf("booking-service: retrying payment event %s (attempt %d/%d): %v", msg.MessageId, retryCount+1, maxRetries, err)
			time.Sleep(retryDelay)
			republish(ch, msg, retryCount+1)
		} else {
			log.Printf("booking-service: payment event %s exceeded max retries, sending to DLQ: %v", msg.MessageId, err)
			deadLetter(ch, msg)
		}
	case outcomeDeadLetter:
		log.Printf("booking-service: payment event %s rejected, sending to DLQ: %v", msg.MessageId, err)
		deadLetter(ch, msg)
	}
	_ = msg.Ack(false)
}

// process giải mã event và cập nhật lịch hẹn, trả về cách xử lý message.
func (c *PaymentEventConsumer) process(body []byte) (deliveryOutcome, error) {
	var event paymentpb.PaymentStatusChangedEvent
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(body, &event); err != nil {
		return outcomeDeadLetter, fmt.Errorf("malformed payment event: %w", err)
	}
	data := event.GetData()
	if data.GetAppointmentId() == "" {
		// Order không gắn lịch hẹn (VD: nạp ví) — không liên quan booking.
		return outcomeAck, nil
	}

	var result appappointment.PaymentResultStatus
	switch data.GetStatus() {
	case paymentpb.PaymentStatus_PAYMENT_STATUS_SUCCESS:
		result = appappointment.PaymentResultSuccess
	case paymentpb.PaymentStatus_PAYMENT_STATUS_FAILED:
		result = appappointment.PaymentResultFailed
	default:
		return outcomeDeadLetter, fmt.Errorf("unsupported payment status %s", data.GetStatus())
	}

	err := c.handler.HandlePaymentResult(appappointment.HandlePaymentResultCommand{
		AppointmentID: data.GetAppointmentId(),
		Status:        result,
	})
	switch {
	case err == nil:
		log.Printf("booking-service: applied %s event %s appointment_id=%s order_id=%s",
			event.GetEventType(), event.GetEventId(), data.GetAppointmentId(), data.GetOrderId())
		return outcomeAck, nil
	case errors.Is(err, appappointment.ErrNotFound),
		errors.Is(err, appappointment.ErrInvalidPaymentResultStatus),
		errors.Is(err, appappointment.ErrInvalidStatus),
		errors.Is(err, appappointment.ErrPaymentResultConflict):
		return outcomeDeadLetter, err
	default:
		return outcomeRetry, err
	}
}

func currentRetryCount(headers amqp.Table) int {
	if headers == nil {
		return 0
	}
	switch v := headers[retryCountHeader].(type) {
	case int32:
		return int(v)
	case int64:
		return int(v)
	case int:
		return v
	default:
		return 0
	}
}

// republish đưa lại bản sao message vào booking.queue qua default exchange với retry count tăng dần.
func republish(ch *amqp.Channel, msg amqp.Delivery, retryCount int) {
	headers := amqp.Table{}
	for k, v := range msg.Headers {
		headers[k] = v
	}
	headers[retryCountHeader] = int32(retryCount)

	err := ch.Publish("", BookingQueue, false, false, amqp.Publishing{
		ContentType:  msg.ContentType,
		DeliveryMode: amqp.Persistent,
		MessageId:    msg.MessageId,
		Type:         msg.Type,
		Headers:      headers,
		Body:         msg.Body,
	})
	if err != nil {
		log.Printf("booking-service: failed to republish payment event for retry, sending to DLQ instead: %v", err)
		deadLetter(ch, msg)
	}
}

func deadLetter(ch *amqp.Channel, msg amqp.Delivery) {
	err := ch.Publish("", BookingDeadLetterQueue, false, false, amqp.Publishing{
		ContentType:  msg.ContentType,
		DeliveryMode: amqp.Persistent,
		MessageId:    msg.MessageId,
		Type:         msg.Type,
		Headers:      msg.Headers,
		Body:         msg.Body,
	})
	if err != nil {
		log.Printf("booking-service: failed to publish payment event to DLQ %q: %v", BookingDeadLetterQueue, err)
	}
}
