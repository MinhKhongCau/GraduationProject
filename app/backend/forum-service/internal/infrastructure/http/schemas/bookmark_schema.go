package schemas

type BookmarkResponse struct {
	PostID        int64 `json:"postId"`
	Bookmarked    bool  `json:"bookmarked"`
	BookmarkCount int   `json:"bookmarkCount"`
}
