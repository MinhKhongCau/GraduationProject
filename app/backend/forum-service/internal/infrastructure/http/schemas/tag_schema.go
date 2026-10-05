package schemas

import (
	tagdomain "forum-service/internal/domain/tag"
)

type TagResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func NewTagResponse(t tagdomain.Tag) TagResponse {
	return TagResponse{ID: t.ID, Name: t.Name, Slug: t.Slug}
}
