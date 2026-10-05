package schemas

import (
	"time"

	postdomain "forum-service/internal/domain/post"
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

type AuthorDTO struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatarUrl"`
	Role      string `json:"role"`
}

type PostResponse struct {
	ID            int64      `json:"id"`
	Title         string     `json:"title"`
	Slug          string     `json:"slug"`
	Summary       string     `json:"summary"`
	ThumbnailURL  string     `json:"thumbnailUrl"`
	CategoryID    int64      `json:"categoryId"`
	AuthorID      string     `json:"authorId"`
	Author        *AuthorDTO `json:"author,omitempty"`
	Status        string     `json:"status"`
	ViewCount     int        `json:"viewCount"`
	LikeCount     int        `json:"likeCount"`
	BookmarkCount int        `json:"bookmarkCount"`
	CommentCount  int        `json:"commentCount"`
	Tags          []string   `json:"tags"`
	CreatedAt     time.Time  `json:"createdAt"`
}

type PostDetailResponse struct {
	PostResponse
	Content   string    `json:"content"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func NewPostResponse(p postdomain.Post, author *AuthorDTO) PostResponse {
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
		Author:        author,
		Status:        string(p.Status),
		ViewCount:     p.ViewCount,
		LikeCount:     p.LikeCount,
		BookmarkCount: p.BookmarkCount,
		CommentCount:  p.CommentCount,
		Tags:          tags,
		CreatedAt:     p.CreatedAt,
	}
}

func NewPostDetailResponse(p postdomain.Post, author *AuthorDTO) PostDetailResponse {
	return PostDetailResponse{
		PostResponse: NewPostResponse(p, author),
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
