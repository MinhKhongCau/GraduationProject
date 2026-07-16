package dto

type LikeResponse struct {
	PostID    int64 `json:"postId"`
	Liked     bool  `json:"liked"`
	LikeCount int   `json:"likeCount"`
}
