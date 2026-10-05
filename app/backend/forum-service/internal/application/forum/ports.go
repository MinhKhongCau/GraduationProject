package forum

import "time"

// EventPublisher is the outbound port for the domain events defined in
// SPEC.md §5. The RabbitMQ implementation (forum.events exchange) lives in
// internal/infrastructure/messaging; publishing is best-effort.
type EventPublisher interface {
	PostCreated(evt PostCreatedEvent)
	CommentCreated(evt CommentCreatedEvent)
	PostLiked(evt PostLikedEvent)
}

type PostCreatedEvent struct {
	PostID      int64     `json:"postId"`
	AuthorID    string    `json:"authorId"`
	CategoryID  int64     `json:"categoryId"`
	Title       string    `json:"title"`
	PublishedAt time.Time `json:"publishedAt"`
}

type CommentCreatedEvent struct {
	CommentID      int64   `json:"commentId"`
	PostID         int64   `json:"postId"`
	UserID         string  `json:"userId"`
	PostAuthorID   string  `json:"postAuthorId"`
	ParentID       *int64  `json:"parentId"`
	ParentAuthorID *string `json:"parentAuthorId"`
}

type PostLikedEvent struct {
	PostID       int64  `json:"postId"`
	UserID       string `json:"userId"`
	PostAuthorID string `json:"postAuthorId"`
}
