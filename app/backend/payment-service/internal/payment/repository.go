package payment

import (
	"payment-service/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(order *domain.PaymentOrder) error
	GetByID(orderID uuid.UUID) (*domain.PaymentOrder, error)
	GetByGatewayTxnRef(ref string) (*domain.PaymentOrder, error)
	Update(order *domain.PaymentOrder) error
	UpdateWithTx(tx *gorm.DB, order *domain.PaymentOrder) error
	SaveOutboxEvent(tx *gorm.DB, event *domain.OutboxEvent) error
	WithTransaction(fn func(tx *gorm.DB) error) error
}

type pgRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &pgRepository{db: db}
}

func (r *pgRepository) Create(order *domain.PaymentOrder) error {
	return r.db.Create(order).Error
}

func (r *pgRepository) GetByID(orderID uuid.UUID) (*domain.PaymentOrder, error) {
	var order domain.PaymentOrder
	err := r.db.Where("id = ?", orderID).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *pgRepository) GetByGatewayTxnRef(ref string) (*domain.PaymentOrder, error) {
	var order domain.PaymentOrder
	err := r.db.Where("gateway_txn_ref = ?", ref).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *pgRepository) Update(order *domain.PaymentOrder) error {
	return r.db.Save(order).Error
}

func (r *pgRepository) UpdateWithTx(tx *gorm.DB, order *domain.PaymentOrder) error {
	return tx.Save(order).Error
}

func (r *pgRepository) SaveOutboxEvent(tx *gorm.DB, event *domain.OutboxEvent) error {
	return tx.Create(event).Error
}

func (r *pgRepository) WithTransaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}
