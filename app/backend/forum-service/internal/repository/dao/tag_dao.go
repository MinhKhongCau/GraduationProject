package dao

import "forum-service/internal/app/entity"

type TagDAO struct {
	ID   int64  `gorm:"column:id;primaryKey"`
	Name string `gorm:"column:name"`
	Slug string `gorm:"column:slug"`
}

func (TagDAO) TableName() string { return "tags" }

func (d TagDAO) ToEntity() entity.Tag {
	return entity.Tag{ID: d.ID, Name: d.Name, Slug: d.Slug}
}

type PostTagDAO struct {
	PostID int64 `gorm:"column:post_id;primaryKey"`
	TagID  int64 `gorm:"column:tag_id;primaryKey"`
}

func (PostTagDAO) TableName() string { return "post_tags" }
