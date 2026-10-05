package models

import (
	tagdomain "forum-service/internal/domain/tag"
)

type TagDAO struct {
	ID   int64  `gorm:"column:id;primaryKey"`
	Name string `gorm:"column:name"`
	Slug string `gorm:"column:slug"`
}

func (TagDAO) TableName() string { return "tags" }

func (d TagDAO) ToEntity() tagdomain.Tag {
	return tagdomain.Tag{ID: d.ID, Name: d.Name, Slug: d.Slug}
}

type PostTagDAO struct {
	PostID int64 `gorm:"column:post_id;primaryKey"`
	TagID  int64 `gorm:"column:tag_id;primaryKey"`
}

func (PostTagDAO) TableName() string { return "post_tags" }
