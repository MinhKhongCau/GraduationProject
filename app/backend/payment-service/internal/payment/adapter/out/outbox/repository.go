package outbox

import (
	"payment-service/internal/domain/entity"

	"gorm.io/gorm"
)

type Repository interface {
	GetUnpublishedEvents(limit int) ([]entity.OutboxEvent, error)
	MarkAsPublished(eventIDs []string) error
}

type pgRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &pgRepository{db: db}
}

func (r *pgRepository) GetUnpublishedEvents(limit int) ([]entity.OutboxEvent, error) {
	var events []entity.OutboxEvent
	err := r.db.Where("published = ?", false).Order("created_at ASC").Limit(limit).Find(&events).Error
	return events, err
}

func (r *pgRepository) MarkAsPublished(eventIDs []string) error {
	if len(eventIDs) == 0 {
		return nil
	}
	return r.db.Model(&entity.OutboxEvent{}).Where("id IN ?", eventIDs).Update("published", true).Error
}
