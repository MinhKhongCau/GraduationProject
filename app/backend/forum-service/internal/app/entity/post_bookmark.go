package entity

import "time"

type PostBookmark struct {
	PostID    int64
	UserID    string
	CreatedAt time.Time
}
