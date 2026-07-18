package dto

import (
	"time"

	"forum-service/internal/app/entity"
)

type CreatePostRequest struct {
	CategoryID   int64    `json:"categoryId" binding:"required"`
	Title        string   `json:"title" binding:"required"`
	Summary      string   `json:"summary"`
	Content      string   `json:"content" binding:"required"`
	ThumbnailURL string   `json:"thumbnailUrl"`
	Status       string   `json:"status"`
	Tags         []string `json:"tags"`
}

type UpdatePostRequest struct {
	CategoryID   *int64    `json:"categoryId"`
	Title        *string   `json:"title"`
	Summary      *string   `json:"summary"`
	Content      *string   `json:"content"`
	ThumbnailURL *string   `json:"thumbnailUrl"`
	Tags         *[]string `json:"tags"`
}

type ChangePostStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type PostResponse struct {
	ID            int64     `json:"id"`
	Title         string    `json:"title"`
	Slug          string    `json:"slug"`
	Summary       string    `json:"summary"`
	ThumbnailURL  string    `json:"thumbnailUrl"`
	CategoryID    int64     `json:"categoryId"`
	AuthorID      string    `json:"authorId"`
	Status        string    `json:"status"`
	ViewCount     int       `json:"viewCount"`
	LikeCount     int       `json:"likeCount"`
	BookmarkCount int       `json:"bookmarkCount"`
	CommentCount  int       `json:"commentCount"`
	Tags          []string  `json:"tags"`
	CreatedAt     time.Time `json:"createdAt"`
}

type PostDetailResponse struct {
	PostResponse
	Content   string    `json:"content"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func NewPostResponse(p entity.Post) PostResponse {
	tags := p.Tags
	if tags == nil {
		tags = []string{}
	}
	return PostResponse{
		ID:            p.ID,
		Title:         p.Title,
		Slug:          p.Slug,
		Summary:       p.Summary,
		ThumbnailURL:  p.ThumbnailURL,
		CategoryID:    p.CategoryID,
		AuthorID:      p.AuthorID,
		Status:        string(p.Status),
		ViewCount:     p.ViewCount,
		LikeCount:     p.LikeCount,
		BookmarkCount: p.BookmarkCount,
		CommentCount:  p.CommentCount,
		Tags:          tags,
		CreatedAt:     p.CreatedAt,
	}
}

func NewPostDetailResponse(p entity.Post) PostDetailResponse {
	return PostDetailResponse{
		PostResponse: NewPostResponse(p),
		Content:      p.Content,
		UpdatedAt:    p.UpdatedAt,
	}
}

type PostListResponse struct {
	Items    []PostResponse `json:"items"`
	Page     int            `json:"page"`
	PageSize int            `json:"pageSize"`
	Total    int64          `json:"total"`
}
