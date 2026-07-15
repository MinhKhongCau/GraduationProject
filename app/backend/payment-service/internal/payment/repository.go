package payment

import (
	"payment-service/internal/domain/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(order *entity.PaymentOrder) error
	GetByID(orderID uuid.UUID) (*entity.PaymentOrder, error)
	GetByGatewayTxnRef(ref string) (*entity.PaymentOrder, error)
	Update(order *entity.PaymentOrder) error
	UpdateWithTx(tx *gorm.DB, order *entity.PaymentOrder) error
	SaveOutboxEvent(tx *gorm.DB, event *entity.OutboxEvent) error
	WithTransaction(fn func(tx *gorm.DB) error) error
}

type pgRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &pgRepository{db: db}
}

func (r *pgRepository) Create(order *entity.PaymentOrder) error {
	return r.db.Create(order).Error
}

func (r *pgRepository) GetByID(orderID uuid.UUID) (*entity.PaymentOrder, error) {
	var order entity.PaymentOrder
	err := r.db.Where("id = ?", orderID).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *pgRepository) GetByGatewayTxnRef(ref string) (*entity.PaymentOrder, error) {
	var order entity.PaymentOrder
	err := r.db.Where("gateway_txn_ref = ?", ref).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *pgRepository) Update(order *entity.PaymentOrder) error {
	return r.db.Save(order).Error
}

func (r *pgRepository) UpdateWithTx(tx *gorm.DB, order *entity.PaymentOrder) error {
	return tx.Save(order).Error
}

func (r *pgRepository) SaveOutboxEvent(tx *gorm.DB, event *entity.OutboxEvent) error {
	return tx.Create(event).Error
}

func (r *pgRepository) WithTransaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}
