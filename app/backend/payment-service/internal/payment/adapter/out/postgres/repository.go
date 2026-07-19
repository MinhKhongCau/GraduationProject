package paymentpostgres

import (
	"context"
	"errors"
	"payment-service/internal/domain/entity"
	"payment-service/internal/domain/vo"
	apppayment "payment-service/internal/payment/application"
	paymentdomain "payment-service/internal/payment/domain"
	"payment-service/internal/wallet"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
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
	return mapWriteError(r.db.Create(order).Error)
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

func (r *pgRepository) GetSuccessfulByAppointment(ctx context.Context, appointmentID uuid.UUID) (*entity.PaymentOrder, error) {
	var order entity.PaymentOrder
	err := r.db.WithContext(ctx).
		Where("appointment_id = ? AND status = ?", appointmentID, entity.OrderStatusSuccess).
		First(&order).Error
	if err != nil {
		return nil, mapTxError(err)
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

type compensationCaseRow struct {
	entity.PaymentCompensationCase `gorm:"embedded"`
	PaymentStatus                  entity.PaymentOrderStatus          `gorm:"column:payment_status"`
	GatewayCaptureStatus           paymentdomain.GatewayCaptureStatus `gorm:"column:payment_gateway_capture_status"`
	FulfillmentStatus              paymentdomain.FulfillmentStatus    `gorm:"column:payment_fulfillment_status"`
}

func (r *pgRepository) ListCompensationCases(ctx context.Context, filter apppayment.CompensationCaseFilter) ([]apppayment.CompensationCaseRecord, int64, error) {
	query := compensationCaseReadQuery(r.db.WithContext(ctx))
	if filter.Status != "" {
		query = query.Where("compensation.status = ?", filter.Status)
	}
	if filter.AppointmentID != nil {
		query = query.Where("compensation.appointment_id = ?", *filter.AppointmentID)
	}
	if filter.PaymentOrderID != nil {
		query = query.Where("compensation.payment_order_id = ?", *filter.PaymentOrderID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []compensationCaseRow
	err := query.
		Order("compensation.created_at DESC").
		Order("compensation.id DESC").
		Offset((filter.Page - 1) * filter.Size).
		Limit(filter.Size).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	return compensationCaseRecords(rows), total, nil
}

func (r *pgRepository) GetCompensationCase(ctx context.Context, caseID uuid.UUID) (*apppayment.CompensationCaseRecord, error) {
	var row compensationCaseRow
	err := compensationCaseReadQuery(r.db.WithContext(ctx)).Where("compensation.id = ?", caseID).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apppayment.ErrCompensationCaseNotFound
	}
	if err != nil {
		return nil, err
	}
	record := compensationCaseRecord(row)
	return &record, nil
}

func compensationCaseReadQuery(db *gorm.DB) *gorm.DB {
	return db.Table("payment_compensation_cases AS compensation").
		Select("compensation.*, payment.status AS payment_status, payment.gateway_capture_status AS payment_gateway_capture_status, payment.fulfillment_status AS payment_fulfillment_status").
		Joins("JOIN payment_orders AS payment ON payment.id = compensation.payment_order_id")
}

func compensationCaseRecords(rows []compensationCaseRow) []apppayment.CompensationCaseRecord {
	records := make([]apppayment.CompensationCaseRecord, len(rows))
	for i := range rows {
		records[i] = compensationCaseRecord(rows[i])
	}
	return records
}

func compensationCaseRecord(row compensationCaseRow) apppayment.CompensationCaseRecord {
	return apppayment.CompensationCaseRecord{
		Case: row.PaymentCompensationCase, PaymentStatus: row.PaymentStatus,
		GatewayCaptureStatus: row.GatewayCaptureStatus, FulfillmentStatus: row.FulfillmentStatus,
	}
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

func (tx *paymentTx) GetOrder(ctx context.Context, orderID uuid.UUID) (*entity.PaymentOrder, error) {
	var order entity.PaymentOrder
	err := tx.db.WithContext(ctx).Where("id = ?", orderID).First(&order).Error
	if err != nil {
		return nil, mapTxError(err)
	}
	return &order, nil
}

func (tx *paymentTx) ListOrdersForAppointmentForUpdate(ctx context.Context, appointmentID uuid.UUID) ([]entity.PaymentOrder, error) {
	var orders []entity.PaymentOrder
	err := tx.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("appointment_id = ?", appointmentID).
		Order("id").
		Find(&orders).Error
	return orders, err
}

func (tx *paymentTx) CreateOrder(ctx context.Context, order *entity.PaymentOrder) error {
	return mapWriteError(tx.db.WithContext(ctx).Create(order).Error)
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
	return mapWriteError(tx.db.WithContext(ctx).Save(order).Error)
}

func (tx *paymentTx) ExpireOtherPendingOrders(ctx context.Context, appointmentID, exceptOrderID uuid.UUID) error {
	return tx.db.WithContext(ctx).
		Model(&entity.PaymentOrder{}).
		Where("appointment_id = ? AND id <> ? AND status = ?", appointmentID, exceptOrderID, entity.OrderStatusPending).
		Update("status", entity.OrderStatusExpired).Error
}

func (tx *paymentTx) SaveOutboxEvent(ctx context.Context, event *entity.OutboxEvent) error {
	return tx.db.WithContext(ctx).Create(event).Error
}

func (tx *paymentTx) SaveCompensationCase(ctx context.Context, compensationCase *entity.PaymentCompensationCase) error {
	return tx.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "payment_order_id"}, {Name: "reason_code"}},
			DoNothing: true,
		}).
		Create(compensationCase).Error
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

func mapWriteError(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		switch pgErr.ConstraintName {
		case "ux_payment_orders_active_pending_appointment":
			return apppayment.ErrActivePendingOrderExists
		case "ux_payment_orders_success_appointment":
			return apppayment.ErrAppointmentAlreadyPaid
		}
	}
	return err
}
