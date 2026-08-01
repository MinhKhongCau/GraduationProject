// File: internal/messaging/rabbitmq.go
package messaging

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"profile-service/config"
	"profile-service/internal/models"
)

// Naming follows RABBITMQ_CONVENTION.md: <domain>.exchange, <entity>.<action>, <service>.queue.
const (
	UserExchange           = "user.exchange"
	ProfileQueue           = "profile.queue"
	ProfileDeadLetterQueue = "profile.queue.dlq"
	UserCreatedRoutingKey  = "user.created"
	maxRetries             = 3
	retryDelay             = 2 * time.Second
	retryCountHeader       = "x-retry-count"
)

var channel *amqp.Channel

// UserCreatedEventData mirrors the "data" payload auth-service puts in the
// user.created event (see auth-service's UserCreatedEventData record).
type UserCreatedEventData struct {
	AccountID   string `json:"accountId"`
	Email       string `json:"email"`
	FullName    string `json:"fullName"`
	Role        string `json:"role"`
	DateOfBirth string `json:"dateOfBirth"`
}

// UserCreatedEvent is the standard event envelope from RABBITMQ_CONVENTION.md.
type UserCreatedEvent struct {
	EventID    string               `json:"eventId"`
	EventType  string               `json:"eventType"`
	OccurredAt string               `json:"occurredAt"`
	Source     string               `json:"source"`
	Data       UserCreatedEventData `json:"data"`
}

// ProfileRequest structure for RPC call
type ProfileRequest struct {
	CorrelationID string   `json:"correlationId"`
	AuthorIDs     []string `json:"authorIds"`
}

type AuthorProfile struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatarUrl"`
	Role      string `json:"role"`
}

type ProfileResponse struct {
	CorrelationID string          `json:"correlationId"`
	Profiles      []AuthorProfile `json:"profiles"`
}

// StartUserCreatedConsumer connects to RabbitMQ, declares the durable
// profile.queue (bound to user.exchange/user.created) plus its DLQ, and
// consumes messages with manual ACK. Each decoded event is handed to handle;
// handle must be idempotent since redelivery can happen after a crash or a retry.
// If the broker is unreachable, the consumer logs a warning and stays disabled
// rather than failing service startup — the internal REST fallback still works.
func StartUserCreatedConsumer(cfg config.RabbitMQConfig, handle func(UserCreatedEvent) error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", cfg.User, cfg.Pass, cfg.Host, cfg.Port)
	log.Printf("profile-service: connecting to RabbitMQ at amqp://%s:***@%s:%s/", cfg.User, cfg.Host, cfg.Port)

	conn, err := amqp.Dial(url)
	if err != nil {
		log.Printf("profile-service: WARNING failed to connect to RabbitMQ: %v — user.created consumer disabled", err)
		return
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Printf("profile-service: WARNING failed to open RabbitMQ channel: %v", err)
		return
	}
	channel = ch

	if err := channel.ExchangeDeclare(UserExchange, "topic", true, false, false, false, nil); err != nil {
		log.Printf("profile-service: WARNING failed to declare exchange %q: %v", UserExchange, err)
		return
	}

	if _, err := channel.QueueDeclare(ProfileQueue, true, false, false, false, nil); err != nil {
		log.Printf("profile-service: WARNING failed to declare queue %q: %v", ProfileQueue, err)
		return
	}

	if _, err := channel.QueueDeclare(ProfileDeadLetterQueue, true, false, false, false, nil); err != nil {
		log.Printf("profile-service: WARNING failed to declare DLQ %q: %v", ProfileDeadLetterQueue, err)
		return
	}

	if err := channel.QueueBind(ProfileQueue, UserCreatedRoutingKey, UserExchange, false, nil); err != nil {
		log.Printf("profile-service: WARNING failed to bind queue %q: %v", ProfileQueue, err)
		return
	}

	if err := channel.Qos(10, 0, false); err != nil {
		log.Printf("profile-service: WARNING failed to set QoS: %v", err)
	}

	msgs, err := channel.Consume(ProfileQueue, "", false, false, false, false, nil)
	if err != nil {
		log.Printf("profile-service: WARNING failed to start consuming %q: %v", ProfileQueue, err)
		return
	}

	log.Printf("profile-service: listening for %q events on queue %q", UserCreatedRoutingKey, ProfileQueue)

	// Start the profile request consumer too
	startProfileRequestConsumer()

	go func() {
		for msg := range msgs {
			handleDelivery(msg, handle)
		}
	}()
}

func startProfileRequestConsumer() {
	requestQueue := "profile.request.queue"
	forumExchange := "forum.events"
	requestRoutingKey := "profile.get_batch.request"
	responseRoutingKey := "profile.get_batch.response"

	if _, err := channel.QueueDeclare(requestQueue, true, false, false, false, nil); err != nil {
		log.Printf("profile-service: WARNING failed to declare queue %q: %v", requestQueue, err)
		return
	}

	if err := channel.QueueBind(requestQueue, requestRoutingKey, UserExchange, false, nil); err != nil {
		log.Printf("profile-service: WARNING failed to bind queue %q to %q: %v", requestQueue, requestRoutingKey, err)
		return
	}

	// Declare the forum.events exchange just in case it doesn't exist yet
	if err := channel.ExchangeDeclare(forumExchange, "topic", true, false, false, false, nil); err != nil {
		log.Printf("profile-service: WARNING failed to declare exchange %q: %v", forumExchange, err)
	}

	msgs, err := channel.Consume(requestQueue, "", false, false, false, false, nil)
	if err != nil {
		log.Printf("profile-service: WARNING failed to start consuming %q: %v", requestQueue, err)
		return
	}

	log.Printf("profile-service: listening for profile request events on queue %q", requestQueue)

	go func() {
		for msg := range msgs {
			var req ProfileRequest
			if err := json.Unmarshal(msg.Body, &req); err != nil {
				log.Printf("profile-service: error decoding profile request: %v", err)
				msg.Ack(false)
				continue
			}

			var dbProfiles []models.Profile
			if err := config.DB.Preload("PatientProfile").Preload("ExpertProfile").Where("auth_id IN ?", req.AuthorIDs).Find(&dbProfiles).Error; err != nil {
				log.Printf("profile-service: error querying profiles: %v", err)
			}

			authorProfiles := make([]AuthorProfile, 0, len(dbProfiles))
			for _, p := range dbProfiles {
				avatar := ""
				if p.Role == models.RolePatient && p.PatientProfile != nil {
					avatar = p.PatientProfile.AvatarURL
				} else if p.Role == models.RoleExpert && p.ExpertProfile != nil {
					avatar = p.ExpertProfile.AvatarURL
				}
				authorProfiles = append(authorProfiles, AuthorProfile{
					ID:        p.AuthID.String(),
					Name:      p.Name,
					AvatarURL: avatar,
					Role:      string(p.Role),
				})
			}

			resp := ProfileResponse{
				CorrelationID: req.CorrelationID,
				Profiles:      authorProfiles,
			}

			respBytes, err := json.Marshal(resp)
			if err != nil {
				log.Printf("profile-service: error encoding profile response: %v", err)
				msg.Ack(false)
				continue
			}

			err = channel.Publish(
				forumExchange,
				responseRoutingKey,
				false,
				false,
				amqp.Publishing{
					ContentType:  "application/json",
					DeliveryMode: amqp.Persistent,
					Body:         respBytes,
				},
			)
			if err != nil {
				log.Printf("profile-service: failed to publish profile response: %v", err)
			} else {
				log.Printf("profile-service: published profile response for correlationId: %s", req.CorrelationID)
			}

			msg.Ack(false)
		}
	}()
}

func handleDelivery(msg amqp.Delivery, handle func(UserCreatedEvent) error) {
	var event UserCreatedEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Printf("profile-service: malformed user.created payload, sending to DLQ: %v", err)
		deadLetter(msg)
		msg.Ack(false)
		return
	}

	if err := handle(event); err != nil {
		retryCount := currentRetryCount(msg.Headers)
		if retryCount < maxRetries {
			log.Printf("profile-service: retrying user.created event %s (attempt %d/%d): %v",
				event.EventID, retryCount+1, maxRetries, err)
			time.Sleep(retryDelay)
			republish(msg, retryCount+1)
		} else {
			log.Printf("profile-service: user.created event %s exceeded max retries, sending to DLQ: %v",
				event.EventID, err)
			deadLetter(msg)
		}
		msg.Ack(false)
		return
	}

	msg.Ack(false)
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

// republish re-queues a copy of msg onto profile.queue via the default
// exchange (every queue is implicitly bound to it under its own name),
// with the retry count incremented so handleDelivery can cap attempts.
func republish(msg amqp.Delivery, retryCount int) {
	headers := amqp.Table{}
	for k, v := range msg.Headers {
		headers[k] = v
	}
	headers[retryCountHeader] = int32(retryCount)

	err := channel.Publish("", ProfileQueue, false, false, amqp.Publishing{
		ContentType:  msg.ContentType,
		DeliveryMode: amqp.Persistent,
		Headers:      headers,
		Body:         msg.Body,
	})
	if err != nil {
		log.Printf("profile-service: failed to republish event for retry, sending to DLQ instead: %v", err)
		deadLetter(msg)
	}
}

func deadLetter(msg amqp.Delivery) {
	err := channel.Publish("", ProfileDeadLetterQueue, false, false, amqp.Publishing{
		ContentType:  msg.ContentType,
		DeliveryMode: amqp.Persistent,
		Headers:      msg.Headers,
		Body:         msg.Body,
	})
	if err != nil {
		log.Printf("profile-service: failed to publish event to DLQ %q: %v", ProfileDeadLetterQueue, err)
	}
}

// PublishSyncSeedAuthors publishes the sync seed authors event to the forum.events exchange.
func PublishSyncSeedAuthors(expertID string) error {
	if channel == nil {
		return fmt.Errorf("RabbitMQ channel is not initialized")
	}
	payload := map[string]string{
		"expertId": expertID,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return channel.Publish(
		"forum.events",
		"profile.sync_seed_authors",
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
}
