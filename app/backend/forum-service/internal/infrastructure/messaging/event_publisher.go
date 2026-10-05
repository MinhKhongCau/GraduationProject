package messaging

import (
	"encoding/json"
	"log"

	"forum-service/internal/application/forum"
)

// EventPublisher publishes the domain events defined in SPEC.md §5 to the
// forum.events RabbitMQ exchange for the (not-yet-implemented) Notification
// Service to consume. Publishing is best-effort — see PublishEvent.
// It implements forum.EventPublisher.
type EventPublisher struct{}

var _ forum.EventPublisher = (*EventPublisher)(nil)

func NewEventPublisher() *EventPublisher {
	return &EventPublisher{}
}

func (p *EventPublisher) PostCreated(evt forum.PostCreatedEvent) {
	p.publish("forum.post.created", evt)
}

func (p *EventPublisher) CommentCreated(evt forum.CommentCreatedEvent) {
	p.publish("forum.comment.created", evt)
}

func (p *EventPublisher) PostLiked(evt forum.PostLikedEvent) {
	p.publish("forum.post.liked", evt)
}

func (p *EventPublisher) publish(routingKey string, evt interface{}) {
	payload, err := json.Marshal(evt)
	if err != nil {
		log.Printf("forum-service: failed to marshal event %s: %v", routingKey, err)
		return
	}
	PublishEvent(routingKey, payload)
}
