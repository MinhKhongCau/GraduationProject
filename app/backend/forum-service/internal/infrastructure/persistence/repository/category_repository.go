package repository

import (
	"errors"

	"gorm.io/gorm"

	"forum-service/internal/infrastructure/persistence/models"
)

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) FindAll() ([]models.CategoryDAO, error) {
	var rows []models.CategoryDAO
	if err := r.db.Order("name ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *CategoryRepository) FindBySlug(slug string) (*models.CategoryDAO, error) {
	var row models.CategoryDAO
	if err := r.db.Where("slug = ?", slug).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (r *CategoryRepository) FindByID(id int64) (*models.CategoryDAO, error) {
	var row models.CategoryDAO
	if err := r.db.Where("id = ?", id).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (r *CategoryRepository) Create(row *models.CategoryDAO) error {
	return r.db.Create(row).Error
}

func (r *CategoryRepository) Update(row *models.CategoryDAO) error {
	return r.db.Model(&models.CategoryDAO{}).Where("id = ?", row.ID).Updates(map[string]interface{}{
		"name":        row.Name,
		"slug":        row.Slug,
		"description": row.Description,
	}).Error
}

func (r *CategoryRepository) DeleteByID(id int64) error {
	return r.db.Where("id = ?", id).Delete(&models.CategoryDAO{}).Error
}
