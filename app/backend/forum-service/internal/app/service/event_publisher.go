package service

import (
	"encoding/json"
	"log"
	"time"

	"forum-service/internal/repository/db"
)

// EventPublisher publishes the domain events defined in SPEC.md §5 to the
// forum.events RabbitMQ exchange for the (not-yet-implemented) Notification
// Service to consume. Publishing is best-effort — see db.PublishEvent.
type EventPublisher struct{}

func NewEventPublisher() *EventPublisher {
	return &EventPublisher{}
}

type PostCreatedEvent struct {
	PostID      int64     `json:"postId"`
	AuthorID    int64     `json:"authorId"`
	CategoryID  int64     `json:"categoryId"`
	Title       string    `json:"title"`
	PublishedAt time.Time `json:"publishedAt"`
}

type CommentCreatedEvent struct {
	CommentID      int64  `json:"commentId"`
	PostID         int64  `json:"postId"`
	UserID         int64  `json:"userId"`
	PostAuthorID   int64  `json:"postAuthorId"`
	ParentID       *int64 `json:"parentId"`
	ParentAuthorID *int64 `json:"parentAuthorId"`
}

type PostLikedEvent struct {
	PostID       int64 `json:"postId"`
	UserID       int64 `json:"userId"`
	PostAuthorID int64 `json:"postAuthorId"`
}

func (p *EventPublisher) PostCreated(evt PostCreatedEvent) {
	p.publish("forum.post.created", evt)
}

func (p *EventPublisher) CommentCreated(evt CommentCreatedEvent) {
	p.publish("forum.comment.created", evt)
}

func (p *EventPublisher) PostLiked(evt PostLikedEvent) {
	p.publish("forum.post.liked", evt)
}

func (p *EventPublisher) publish(routingKey string, evt interface{}) {
	payload, err := json.Marshal(evt)
	if err != nil {
		log.Printf("forum-service: failed to marshal event %s: %v", routingKey, err)
		return
	}
	db.PublishEvent(routingKey, payload)
}
