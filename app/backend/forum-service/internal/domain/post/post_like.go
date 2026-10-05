package post

import "time"

type PostLike struct {
	PostID    int64
	UserID    string
	CreatedAt time.Time
}
