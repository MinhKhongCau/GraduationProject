package messaging

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"

	"forum-service/config"
)

const EventsExchange = "forum.events"

type AuthorProfile struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatarUrl"`
	Role      string `json:"role"`
}

type ProfileRequest struct {
	CorrelationID string   `json:"correlationId"`
	AuthorIDs     []string `json:"authorIds"`
}

type ProfileResponse struct {
	CorrelationID string          `json:"correlationId"`
	Profiles      []AuthorProfile `json:"profiles"`
}

var (
	rabbitConn      *amqp.Connection
	Channel         *amqp.Channel
	pendingRequests = make(map[string]chan []AuthorProfile)
	pendingMu       sync.Mutex
)

// InitRabbitMQ connects to RabbitMQ and declares the `forum.events` topic
// exchange used for SPEC.md §5's domain events. It mirrors payment-service's
// graceful-fallback behavior: if the broker is unreachable, Channel stays
// nil and PublishEvent logs to the console instead of failing the caller —
// forum-service's event publishing is best-effort, not required for
// correctness of the request it's attached to.
func InitRabbitMQ(cfg *config.Config) {
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

	// Start consumers
	startProfileResponseConsumer()
	startSyncSeedAuthorsConsumer()
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

func startProfileResponseConsumer() {
	responseQueue := "forum.profile_response.queue"
	responseRoutingKey := "profile.get_batch.response"

	if _, err := Channel.QueueDeclare(responseQueue, true, false, false, false, nil); err != nil {
		log.Printf("forum-service: WARNING failed to declare response queue %q: %v", responseQueue, err)
		return
	}

	if err := Channel.QueueBind(responseQueue, responseRoutingKey, EventsExchange, false, nil); err != nil {
		log.Printf("forum-service: WARNING failed to bind response queue %q to exchange %q: %v", responseQueue, EventsExchange, err)
		return
	}

	msgs, err := Channel.Consume(responseQueue, "", false, false, false, false, nil)
	if err != nil {
		log.Printf("forum-service: WARNING failed to start consuming response queue %q: %v", responseQueue, err)
		return
	}

	log.Printf("forum-service: listening for profile responses on queue %q", responseQueue)

	go func() {
		for msg := range msgs {
			var resp ProfileResponse
			if err := json.Unmarshal(msg.Body, &resp); err != nil {
				log.Printf("forum-service: error decoding profile response: %v", err)
				msg.Ack(false)
				continue
			}

			pendingMu.Lock()
			ch, exists := pendingRequests[resp.CorrelationID]
			if exists {
				delete(pendingRequests, resp.CorrelationID)
				select {
				case ch <- resp.Profiles:
				default:
				}
			}
			pendingMu.Unlock()

			msg.Ack(false)
		}
	}()
}

// FetchProfiles requests profile data for multiple authorIds over RabbitMQ (RPC-like)
// and waits for the response up to a timeout.
func FetchProfiles(authorIDs []string) (map[string]AuthorProfile, error) {
	if Channel == nil || len(authorIDs) == 0 {
		return map[string]AuthorProfile{}, nil
	}

	correlationID := uuid.NewString()
	ch := make(chan []AuthorProfile, 1)

	pendingMu.Lock()
	pendingRequests[correlationID] = ch
	pendingMu.Unlock()

	defer func() {
		pendingMu.Lock()
		delete(pendingRequests, correlationID)
		pendingMu.Unlock()
	}()

	req := ProfileRequest{
		CorrelationID: correlationID,
		AuthorIDs:     authorIDs,
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshalling profile request: %w", err)
	}

	err = Channel.Publish(
		"user.exchange",
		"profile.get_batch.request",
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         reqBytes,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error publishing profile request: %w", err)
	}

	select {
	case profiles := <-ch:
		result := make(map[string]AuthorProfile)
		for _, p := range profiles {
			result[p.ID] = p
		}
		return result, nil
	case <-time.After(3 * time.Second):
		return nil, fmt.Errorf("profile request timed out after 3 seconds")
	}
}

type SyncSeedAuthorsPayload struct {
	ExpertID string `json:"expertId"`
}

func startSyncSeedAuthorsConsumer() {
	syncQueue := "forum.sync_seed_authors.queue"
	syncRoutingKey := "profile.sync_seed_authors"

	if _, err := Channel.QueueDeclare(syncQueue, true, false, false, false, nil); err != nil {
		log.Printf("forum-service: WARNING failed to declare sync queue %q: %v", syncQueue, err)
		return
	}

	if err := Channel.QueueBind(syncQueue, syncRoutingKey, EventsExchange, false, nil); err != nil {
		log.Printf("forum-service: WARNING failed to bind sync queue %q to exchange %q: %v", syncQueue, EventsExchange, err)
		return
	}

	msgs, err := Channel.Consume(syncQueue, "", false, false, false, false, nil)
	if err != nil {
		log.Printf("forum-service: WARNING failed to start consuming sync queue %q: %v", syncQueue, err)
		return
	}

	log.Printf("forum-service: listening for sync seed authors events on queue %q", syncQueue)

	go func() {
		for msg := range msgs {
			var payload SyncSeedAuthorsPayload
			if err := json.Unmarshal(msg.Body, &payload); err != nil {
				log.Printf("forum-service: error decoding sync seed authors payload: %v", err)
				msg.Ack(false)
				continue
			}

			if payload.ExpertID == "" {
				log.Printf("forum-service: received empty expertId in sync seed authors payload")
				msg.Ack(false)
				continue
			}

			if config.DB == nil {
				log.Printf("forum-service: DB is not initialized, cannot update seed authors")
				msg.Nack(false, true)
				time.Sleep(1 * time.Second)
				continue
			}

			err = config.DB.Exec("UPDATE posts SET author_id = ? WHERE author_id IN (?, ?, ?)",
				payload.ExpertID,
				"00000000-0000-0000-0000-0000000003e9",
				"00000000-0000-0000-0000-0000000003ea",
				"00000000-0000-0000-0000-0000000003eb",
			).Error

			if err != nil {
				log.Printf("forum-service: failed to update seed post authors via RabbitMQ event: %v", err)
				msg.Nack(false, true)
			} else {
				log.Printf("forum-service: successfully updated seed post authors to expert %s via RabbitMQ event", payload.ExpertID)
				msg.Ack(false)
			}
		}
	}()
}
