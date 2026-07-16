package service

import (
	"errors"

	"forum-service/internal/app/entity"
	"forum-service/internal/repository/dao"
)

type TagService struct {
	tagRepo *dao.TagRepository
}

func NewTagService(tagRepo *dao.TagRepository) *TagService {
	return &TagService{tagRepo: tagRepo}
}

func (s *TagService) List() ([]entity.Tag, error) {
	rows, err := s.tagRepo.FindAll()
	if err != nil {
		return nil, err
	}
	result := make([]entity.Tag, 0, len(rows))
	for _, r := range rows {
		result = append(result, r.ToEntity())
	}
	return result, nil
}

func (s *TagService) GetBySlug(slug string) (*entity.Tag, error) {
	row, err := s.tagRepo.FindBySlug(slug)
	if err != nil {
		if errors.Is(err, dao.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	e := row.ToEntity()
	return &e, nil
}
