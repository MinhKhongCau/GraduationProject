package models

import (
	"time"

	commentdomain "forum-service/internal/domain/comment"
)

type CommentDAO struct {
	ID        int64               `gorm:"column:id;primaryKey"`
	PostID    int64               `gorm:"column:post_id"`
	UserID    string              `gorm:"column:user_id"`
	ParentID  *int64              `gorm:"column:parent_id"`
	Path      commentdomain.LTree `gorm:"column:path"`
	Content   string              `gorm:"column:content"`
	CreatedAt time.Time           `gorm:"column:created_at"`
	UpdatedAt time.Time           `gorm:"column:updated_at"`
	DeletedAt *time.Time          `gorm:"column:deleted_at"`
}

func (CommentDAO) TableName() string { return "comments" }

func (d CommentDAO) ToEntity() commentdomain.Comment {
	return commentdomain.Comment{
		ID:        d.ID,
		PostID:    d.PostID,
		UserID:    d.UserID,
		ParentID:  d.ParentID,
		Path:      d.Path,
		Content:   d.Content,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
		DeletedAt: d.DeletedAt,
	}
}
