package db

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"

	"forum-service/configs"
)

const EventsExchange = "forum.events"

var (
	rabbitConn *amqp.Connection
	Channel    *amqp.Channel
)

// InitRabbitMQ connects to RabbitMQ and declares the `forum.events` topic
// exchange used for SPEC.md §5's domain events. It mirrors payment-service's
// graceful-fallback behavior: if the broker is unreachable, Channel stays
// nil and PublishEvent logs to the console instead of failing the caller —
// forum-service's event publishing is best-effort, not required for
// correctness of the request it's attached to.
func InitRabbitMQ(cfg *configs.Config) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", cfg.RabbitMQUser, cfg.RabbitMQPass, cfg.RabbitMQHost, cfg.RabbitMQPort)
	log.Printf("forum-service: connecting to RabbitMQ at amqp://%s:***@%s:%s/", cfg.RabbitMQUser, cfg.RabbitMQHost, cfg.RabbitMQPort)

	var err error
	rabbitConn, err = amqp.Dial(url)
	if err != nil {
		log.Printf("forum-service: WARNING failed to connect to RabbitMQ: %v — running in console-log mode", err)
		return
	}

	Channel, err = rabbitConn.Channel()
	if err != nil {
		log.Printf("forum-service: WARNING failed to open RabbitMQ channel: %v", err)
		return
	}

	err = Channel.ExchangeDeclare(
		EventsExchange, // name
		"topic",        // type
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		log.Printf("forum-service: WARNING failed to declare exchange %q: %v", EventsExchange, err)
		Channel = nil
		return
	}

	log.Printf("forum-service: connected to RabbitMQ, declared exchange %q", EventsExchange)
}

// PublishEvent publishes payload (a JSON-encoded event body) under
// routingKey on the forum.events exchange. Failures are logged, never
// returned as fatal — callers should not roll back a domain write just
// because the notification side-channel is unavailable.
func PublishEvent(routingKey string, payload []byte) {
	if Channel == nil {
		log.Printf("forum-service: [RABBITMQ MOCK] %s: %s", routingKey, payload)
		return
	}

	err := Channel.Publish(
		EventsExchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        payload,
		},
	)
	if err != nil {
		log.Printf("forum-service: failed to publish event key=%s: %v", routingKey, err)
		return
	}

	log.Printf("forum-service: published event key=%s", routingKey)
}
