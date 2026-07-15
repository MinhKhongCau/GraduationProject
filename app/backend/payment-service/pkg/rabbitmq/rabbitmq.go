package rabbitmq

import (
	"fmt"
	"log"
	"payment-service/internal/config"

	amqp "github.com/rabbitmq/amqp091-go"
)

var Conn *amqp.Connection
var Channel *amqp.Channel

func InitRabbitMQ() {
	cfg := config.AppConfig
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", cfg.RabbitMQUser, cfg.RabbitMQPass, cfg.RabbitMQHost, cfg.RabbitMQPort)
	log.Printf("Connecting to RabbitMQ: amqp://%s:***@%s:%s/", cfg.RabbitMQUser, cfg.RabbitMQHost, cfg.RabbitMQPort)

	var err error
	Conn, err = amqp.Dial(url)
	if err != nil {
		log.Printf("⚠️  Warning: Failed to connect to RabbitMQ: %v. Running in mock/console log mode.", err)
		return
	}

	Channel, err = Conn.Channel()
	if err != nil {
		log.Printf("⚠️  Warning: Failed to open RabbitMQ channel: %v", err)
		return
	}

	// Declare exchange
	err = Channel.ExchangeDeclare(
		"payment.events", // name
		"topic",          // type
		true,             // durable
		false,            // auto-deleted
		false,            // internal
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		log.Printf("⚠️  Warning: Failed to declare exchange: %v", err)
		return
	}

	fmt.Println("✅ Successfully connected to RabbitMQ and declared exchange 'payment.events'")
}

func PublishEvent(routingKey string, payload string) error {
	if Channel == nil {
		log.Printf("[RABBITMQ MOCK] PublishEvent to routing key '%s' with payload: %s", routingKey, payload)
		return nil
	}

	err := Channel.Publish(
		"payment.events", // exchange
		routingKey,       // routing key
		false,            // mandatory
		false,            // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(payload),
		},
	)
	if err != nil {
		log.Printf("❌ Failed to publish message to RabbitMQ: %v. Log: key=%s payload=%s", err, routingKey, payload)
		return err
	}

	log.Printf("✅ Published event to RabbitMQ: key=%s", routingKey)
	return nil
}
