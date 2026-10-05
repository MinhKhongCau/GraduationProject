package forum

import (
	"errors"

	tagdomain "forum-service/internal/domain/tag"
	"forum-service/internal/infrastructure/persistence/repository"
)

type TagService struct {
	tagRepo *repository.TagRepository
}

func NewTagService(tagRepo *repository.TagRepository) *TagService {
	return &TagService{tagRepo: tagRepo}
}

func (s *TagService) List() ([]tagdomain.Tag, error) {
	rows, err := s.tagRepo.FindAll()
	if err != nil {
		return nil, err
	}
	result := make([]tagdomain.Tag, 0, len(rows))
	for _, r := range rows {
		result = append(result, r.ToEntity())
	}
	return result, nil
}

func (s *TagService) GetBySlug(slug string) (*tagdomain.Tag, error) {
	row, err := s.tagRepo.FindBySlug(slug)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	e := row.ToEntity()
	return &e, nil
}
