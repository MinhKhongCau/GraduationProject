package paymentpostgres

import (
	"context"
	"errors"
	"payment-service/internal/domain/entity"
	"payment-service/internal/domain/vo"
	apppayment "payment-service/internal/payment/application"
	"payment-service/internal/wallet"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type pgRepository struct {
	db *gorm.DB
}

type paymentUnitOfWork struct {
	db            *gorm.DB
	walletUsecase wallet.Usecase
}

type paymentTx struct {
	db            *gorm.DB
	walletUsecase wallet.Usecase
}

func NewRepository(db *gorm.DB) apppayment.Repository {
	return &pgRepository{db: db}
}

func NewUnitOfWork(db *gorm.DB, walletUsecase wallet.Usecase) apppayment.UnitOfWork {
	return &paymentUnitOfWork{
		db:            db,
		walletUsecase: walletUsecase,
	}
}

func (r *pgRepository) NewUnitOfWork(walletUsecase wallet.Usecase) apppayment.UnitOfWork {
	return &paymentUnitOfWork{
		db:            r.db,
		walletUsecase: walletUsecase,
	}
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

func (r *pgRepository) GetByIDForUpdate(tx *gorm.DB, orderID uuid.UUID) (*entity.PaymentOrder, error) {
	var order entity.PaymentOrder
	err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", orderID).
		First(&order).Error
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

func (r *pgRepository) GetByGatewayTxnRefWithTx(tx *gorm.DB, ref string) (*entity.PaymentOrder, error) {
	var order entity.PaymentOrder
	err := tx.Where("gateway_txn_ref = ?", ref).First(&order).Error
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

func (u *paymentUnitOfWork) WithinTx(ctx context.Context, fn func(tx apppayment.Tx) error) error {
	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&paymentTx{
			db:            tx,
			walletUsecase: u.walletUsecase,
		})
	})
}

func (tx *paymentTx) GetOrderForUpdate(ctx context.Context, orderID uuid.UUID) (*entity.PaymentOrder, error) {
	var order entity.PaymentOrder
	err := tx.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", orderID).
		First(&order).Error
	if err != nil {
		return nil, mapTxError(err)
	}
	return &order, nil
}

func (tx *paymentTx) GetGatewayTxnRef(ctx context.Context, ref string) (*entity.PaymentOrder, error) {
	var order entity.PaymentOrder
	err := tx.db.WithContext(ctx).Where("gateway_txn_ref = ?", ref).First(&order).Error
	if err != nil {
		return nil, mapTxError(err)
	}
	return &order, nil
}

func (tx *paymentTx) UpdateOrder(ctx context.Context, order *entity.PaymentOrder) error {
	return tx.db.WithContext(ctx).Save(order).Error
}

func (tx *paymentTx) SaveOutboxEvent(ctx context.Context, event *entity.OutboxEvent) error {
	return tx.db.WithContext(ctx).Create(event).Error
}

func (tx *paymentTx) CreditWalletPending(ctx context.Context, userID uuid.UUID, amount vo.Money, refID uuid.UUID, idempotencyKey string) error {
	return tx.walletUsecase.CreditPendingWithTx(ctx, tx.db, userID, amount, "PAYMENT_ORDER", refID, idempotencyKey)
}

func (tx *paymentTx) DebitWalletPending(ctx context.Context, userID uuid.UUID, amount vo.Money, refID uuid.UUID, idempotencyKey string) error {
	return tx.walletUsecase.DebitPendingWithTx(ctx, tx.db, userID, amount, "PAYMENT_ORDER", refID, idempotencyKey)
}

func mapTxError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apppayment.ErrTxRecordNotFound
	}
	return err
}
