package dao

import (
	"time"

	"forum-service/internal/app/entity"
)

type CategoryDAO struct {
	ID          int64     `gorm:"column:id;primaryKey"`
	Name        string    `gorm:"column:name"`
	Slug        string    `gorm:"column:slug"`
	Description string    `gorm:"column:description"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

func (CategoryDAO) TableName() string { return "categories" }

func (d CategoryDAO) ToEntity() entity.Category {
	return entity.Category{
		ID:          d.ID,
		Name:        d.Name,
		Slug:        d.Slug,
		Description: d.Description,
		CreatedAt:   d.CreatedAt,
	}
}

func CategoryFromEntity(e entity.Category) CategoryDAO {
	return CategoryDAO{
		ID:          e.ID,
		Name:        e.Name,
		Slug:        e.Slug,
		Description: e.Description,
		CreatedAt:   e.CreatedAt,
	}
}
