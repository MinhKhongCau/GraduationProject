package entity

import "time"

type PostBookmark struct {
	PostID    int64
	UserID    int64
	CreatedAt time.Time
}
