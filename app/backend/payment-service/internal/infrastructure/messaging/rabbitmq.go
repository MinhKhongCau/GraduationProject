package messaging

import (
	"context"
	"errors"
	"fmt"
	"log"
	"payment-service/config"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Naming theo RABBITMQ_CONVENTION.md; exchange payment.events đã có sẵn trong rabbitmq/definitions.json.
const PaymentExchange = "payment.events"

// EventPublisher publish một event lên exchange payment.events. Lỗi trả về luôn được coi là
// retryable để outbox thử lại.
type EventPublisher interface {
	Publish(ctx context.Context, routingKey, messageID string, body []byte) error
}

// RabbitPublisher giữ một connection/channel ở chế độ publisher confirm và tự kết nối lại
// khi broker mất kết nối. Publish chỉ trả về nil khi broker đã ack message.
type RabbitPublisher struct {
	url  string
	mu   sync.Mutex
	conn *amqp.Connection
	ch   *amqp.Channel
}

var _ EventPublisher = (*RabbitPublisher)(nil)

// NewRabbitPublisher thử kết nối ngay để log sớm; nếu broker chưa sẵn sàng, lần Publish sau sẽ kết nối lại.
func NewRabbitPublisher(cfg *config.Config) *RabbitPublisher {
	p := &RabbitPublisher{
		url: fmt.Sprintf("amqp://%s:%s@%s:%s/", cfg.RabbitMQUser, cfg.RabbitMQPass, cfg.RabbitMQHost, cfg.RabbitMQPort),
	}
	log.Printf("Connecting to RabbitMQ: amqp://%s:***@%s:%s/", cfg.RabbitMQUser, cfg.RabbitMQHost, cfg.RabbitMQPort)

	p.mu.Lock()
	defer p.mu.Unlock()
	if _, err := p.channel(); err != nil {
		log.Printf("⚠️  Warning: RabbitMQ not available yet: %v. Outbox will retry publishing.", err)
	} else {
		log.Printf("✅ Successfully connected to RabbitMQ and declared exchange '%s'", PaymentExchange)
	}
	return p
}

func (p *RabbitPublisher) Publish(ctx context.Context, routingKey, messageID string, body []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	ch, err := p.channel()
	if err != nil {
		return fmt.Errorf("rabbitmq unavailable: %w", err)
	}

	confirmation, err := ch.PublishWithDeferredConfirmWithContext(ctx, PaymentExchange, routingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		MessageId:    messageID,
		Type:         routingKey,
		AppId:        "payment-service",
		Timestamp:    time.Now(),
		Body:         body,
	})
	if err != nil {
		p.reset()
		return fmt.Errorf("rabbitmq publish %s: %w", routingKey, err)
	}
	acked, err := confirmation.WaitContext(ctx)
	if err != nil {
		return fmt.Errorf("rabbitmq confirm %s: %w", routingKey, err)
	}
	if !acked {
		return errors.New("rabbitmq broker nacked " + routingKey)
	}
	log.Printf("✅ Published event to RabbitMQ: key=%s message_id=%s", routingKey, messageID)
	return nil
}

func (p *RabbitPublisher) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.reset()
}

// channel trả về channel đang mở, kết nối lại nếu cần. Gọi khi đang giữ p.mu.
func (p *RabbitPublisher) channel() (*amqp.Channel, error) {
	if p.ch != nil && !p.ch.IsClosed() && p.conn != nil && !p.conn.IsClosed() {
		return p.ch, nil
	}
	p.reset()

	conn, err := amqp.Dial(p.url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}
	if err := ch.Confirm(false); err != nil {
		conn.Close()
		return nil, err
	}
	if err := ch.ExchangeDeclare(PaymentExchange, "topic", true, false, false, false, nil); err != nil {
		conn.Close()
		return nil, err
	}
	p.conn, p.ch = conn, ch
	return ch, nil
}

func (p *RabbitPublisher) reset() {
	if p.ch != nil {
		_ = p.ch.Close()
	}
	if p.conn != nil {
		_ = p.conn.Close()
	}
	p.conn, p.ch = nil, nil
}
