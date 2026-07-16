package dao

import (
	"errors"

	"gorm.io/gorm"
)

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) FindAll() ([]CategoryDAO, error) {
	var rows []CategoryDAO
	if err := r.db.Order("name ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *CategoryRepository) FindBySlug(slug string) (*CategoryDAO, error) {
	var row CategoryDAO
	if err := r.db.Where("slug = ?", slug).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (r *CategoryRepository) FindByID(id int64) (*CategoryDAO, error) {
	var row CategoryDAO
	if err := r.db.Where("id = ?", id).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (r *CategoryRepository) Create(row *CategoryDAO) error {
	return r.db.Create(row).Error
}

func (r *CategoryRepository) Update(row *CategoryDAO) error {
	return r.db.Model(&CategoryDAO{}).Where("id = ?", row.ID).Updates(map[string]interface{}{
		"name":        row.Name,
		"slug":        row.Slug,
		"description": row.Description,
	}).Error
}

func (r *CategoryRepository) DeleteByID(id int64) error {
	return r.db.Where("id = ?", id).Delete(&CategoryDAO{}).Error
}
