package dao

import "time"

type PostBookmarkDAO struct {
	PostID    int64     `gorm:"column:post_id;primaryKey"`
	UserID    int64     `gorm:"column:user_id;primaryKey"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (PostBookmarkDAO) TableName() string { return "post_bookmarks" }
