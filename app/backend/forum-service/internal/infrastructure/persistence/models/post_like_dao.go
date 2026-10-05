package models

import "time"

type PostLikeDAO struct {
	PostID    int64     `gorm:"column:post_id;primaryKey"`
	UserID    string    `gorm:"column:user_id;primaryKey"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (PostLikeDAO) TableName() string { return "post_likes" }
