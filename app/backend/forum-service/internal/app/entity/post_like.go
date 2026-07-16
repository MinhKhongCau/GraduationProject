package entity

import "time"

type PostLike struct {
	PostID    int64
	UserID    int64
	CreatedAt time.Time
}
