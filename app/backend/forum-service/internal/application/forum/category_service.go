package forum

import (
	"errors"
	"strings"

	categorydomain "forum-service/internal/domain/category"
	"forum-service/internal/infrastructure/persistence/models"
	"forum-service/internal/infrastructure/persistence/repository"
)

type CategoryService struct {
	repo *repository.CategoryRepository
}

func NewCategoryService(repo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) List() ([]categorydomain.Category, error) {
	rows, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}
	result := make([]categorydomain.Category, 0, len(rows))
	for _, r := range rows {
		result = append(result, r.ToEntity())
	}
	return result, nil
}

func (s *CategoryService) GetBySlug(slug string) (*categorydomain.Category, error) {
	row, err := s.repo.FindBySlug(slug)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	e := row.ToEntity()
	return &e, nil
}

func (s *CategoryService) Create(name, slug, description string) (*categorydomain.Category, error) {
	row := models.CategoryDAO{Name: strings.TrimSpace(name), Slug: strings.TrimSpace(slug), Description: description}
	if err := s.repo.Create(&row); err != nil {
		return nil, err
	}
	e := row.ToEntity()
	return &e, nil
}

func (s *CategoryService) Update(id int64, name, slug, description string) (*categorydomain.Category, error) {
	existing, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	existing.Name = strings.TrimSpace(name)
	existing.Slug = strings.TrimSpace(slug)
	existing.Description = description
	if err := s.repo.Update(existing); err != nil {
		return nil, err
	}
	e := existing.ToEntity()
	return &e, nil
}

// Delete removes a category. Categories are hard-deleted (not soft-deleted,
// unlike posts/comments — SPEC.md §3.6 only covers those two), but the
// posts.category_id ON DELETE RESTRICT constraint blocks deletion while any
// post still references it; that FK violation is translated into
// ErrCategoryHasPosts for the handler to surface as 409 (SPEC.md §3.5).
func (s *CategoryService) Delete(id int64) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	if err := s.repo.DeleteByID(id); err != nil {
		if repository.IsForeignKeyViolation(err) {
			return ErrCategoryHasPosts
		}
		return err
	}
	return nil
}
