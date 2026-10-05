package models

import (
	"time"

	postdomain "forum-service/internal/domain/post"
)

type PostDAO struct {
	ID            int64      `gorm:"column:id;primaryKey"`
	CategoryID    int64      `gorm:"column:category_id"`
	AuthorID      string     `gorm:"column:author_id"`
	Title         string     `gorm:"column:title"`
	Slug          string     `gorm:"column:slug"`
	Summary       string     `gorm:"column:summary"`
	Content       string     `gorm:"column:content"`
	ThumbnailURL  string     `gorm:"column:thumbnail_url"`
	Status        string     `gorm:"column:status"`
	ViewCount     int        `gorm:"column:view_count"`
	LikeCount     int        `gorm:"column:like_count"`
	BookmarkCount int        `gorm:"column:bookmark_count"`
	CommentCount  int        `gorm:"column:comment_count"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at"`
	DeletedAt     *time.Time `gorm:"column:deleted_at"`
}

func (PostDAO) TableName() string { return "posts" }

func (d PostDAO) ToEntity() postdomain.Post {
	return postdomain.Post{
		ID:            d.ID,
		CategoryID:    d.CategoryID,
		AuthorID:      d.AuthorID,
		Title:         d.Title,
		Slug:          d.Slug,
		Summary:       d.Summary,
		Content:       d.Content,
		ThumbnailURL:  d.ThumbnailURL,
		Status:        postdomain.PostStatus(d.Status),
		ViewCount:     d.ViewCount,
		LikeCount:     d.LikeCount,
		BookmarkCount: d.BookmarkCount,
		CommentCount:  d.CommentCount,
		CreatedAt:     d.CreatedAt,
		UpdatedAt:     d.UpdatedAt,
		DeletedAt:     d.DeletedAt,
	}
}

func PostFromEntity(e postdomain.Post) PostDAO {
	return PostDAO{
		ID:            e.ID,
		CategoryID:    e.CategoryID,
		AuthorID:      e.AuthorID,
		Title:         e.Title,
		Slug:          e.Slug,
		Summary:       e.Summary,
		Content:       e.Content,
		ThumbnailURL:  e.ThumbnailURL,
		Status:        string(e.Status),
		ViewCount:     e.ViewCount,
		LikeCount:     e.LikeCount,
		BookmarkCount: e.BookmarkCount,
		CommentCount:  e.CommentCount,
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
		DeletedAt:     e.DeletedAt,
	}
}
