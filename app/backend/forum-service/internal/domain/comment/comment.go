package comment

import "time"

type Comment struct {
	ID        int64
	PostID    int64
	UserID    string
	ParentID  *int64
	Path      LTree
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	Replies   []*Comment
}

func (c *Comment) IsDeleted() bool {
	return c.DeletedAt != nil
}
