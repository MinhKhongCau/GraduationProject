package schemas

import (
	"time"

	commentdomain "forum-service/internal/domain/comment"
)

type CreateCommentRequest struct {
	Content  string `json:"content" binding:"required"`
	ParentID *int64 `json:"parentId"`
}

type UpdateCommentRequest struct {
	Content string `json:"content" binding:"required"`
}

type CommentResponse struct {
	ID        int64             `json:"id"`
	PostID    int64             `json:"postId"`
	UserID    string            `json:"userId"`
	User      *AuthorDTO        `json:"user,omitempty"`
	ParentID  *int64            `json:"parentId"`
	Path      string            `json:"path"`
	Content   *string           `json:"content"`
	Deleted   bool              `json:"deleted"`
	CreatedAt time.Time         `json:"createdAt"`
	Replies   []CommentResponse `json:"replies"`
}

// NewCommentResponse renders a soft-deleted comment as a content-less
// placeholder ("deleted": true) rather than omitting it, so reply chains
// under it stay visible (SPEC.md §3.6).
func NewCommentResponse(c *commentdomain.Comment, profiles map[string]AuthorDTO) CommentResponse {
	var userDTO *AuthorDTO
	if prof, exists := profiles[c.UserID]; exists {
		userDTO = &prof
	} else {
		userDTO = &AuthorDTO{
			ID:   c.UserID,
			Name: "Community member",
			Role: "PATIENT",
		}
	}

	resp := CommentResponse{
		ID:        c.ID,
		PostID:    c.PostID,
		UserID:    c.UserID,
		User:      userDTO,
		ParentID:  c.ParentID,
		Path:      c.Path.String(),
		CreatedAt: c.CreatedAt,
		Replies:   []CommentResponse{},
	}
	if c.IsDeleted() {
		resp.Deleted = true
	} else {
		content := c.Content
		resp.Content = &content
	}
	for _, r := range c.Replies {
		resp.Replies = append(resp.Replies, NewCommentResponse(r, profiles))
	}
	return resp
}
