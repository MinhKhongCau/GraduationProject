package entity

import "time"

type PostStatus string

const (
	PostStatusDraft     PostStatus = "DRAFT"
	PostStatusPublished PostStatus = "PUBLISHED"
	PostStatusArchived  PostStatus = "ARCHIVED"
)

type Post struct {
	ID            int64
	CategoryID    int64
	AuthorID      string
	Title         string
	Slug          string
	Summary       string
	Content       string
	ThumbnailURL  string
	Status        PostStatus
	ViewCount     int
	LikeCount     int
	BookmarkCount int
	CommentCount  int
	Tags          []string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}

func (p *Post) IsDeleted() bool {
	return p.DeletedAt != nil
}

func (p *Post) IsVisibleTo(userID string, isAdmin bool) bool {
	if p.IsDeleted() {
		return false
	}
	if p.Status == PostStatusPublished {
		return true
	}
	return isAdmin || p.AuthorID == userID
}
