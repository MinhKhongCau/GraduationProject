package repository

import (
	"context"
	"errors"
	apppayment "payment-service/internal/application/payment"
	appwallet "payment-service/internal/application/wallet"
	"payment-service/internal/domain/money"
	paymentdomain "payment-service/internal/domain/payment"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type paymentRepository struct {
	db *gorm.DB
}

type paymentUnitOfWork struct {
	db            *gorm.DB
	walletUsecase appwallet.Usecase
}

type paymentTx struct {
	db            *gorm.DB
	walletUsecase appwallet.Usecase
}

func NewPaymentRepository(db *gorm.DB) apppayment.Repository {
	return &paymentRepository{db: db}
}

func NewUnitOfWork(db *gorm.DB, walletUsecase appwallet.Usecase) apppayment.UnitOfWork {
	return &paymentUnitOfWork{
		db:            db,
		walletUsecase: walletUsecase,
	}
}

func (r *paymentRepository) NewUnitOfWork(walletUsecase appwallet.Usecase) apppayment.UnitOfWork {
	return &paymentUnitOfWork{
		db:            r.db,
		walletUsecase: walletUsecase,
	}
}

func (r *paymentRepository) Create(order *paymentdomain.PaymentOrder) error {
	return mapWriteError(r.db.Create(order).Error)
}

func (r *paymentRepository) GetByID(orderID uuid.UUID) (*paymentdomain.PaymentOrder, error) {
	var order paymentdomain.PaymentOrder
	err := r.db.Where("id = ?", orderID).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *paymentRepository) GetPaymentOrder(ctx context.Context, orderID uuid.UUID) (*paymentdomain.PaymentOrder, error) {
	var order paymentdomain.PaymentOrder
	err := r.db.WithContext(ctx).Where("id = ?", orderID).First(&order).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apppayment.ErrPaymentOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *paymentRepository) ListPaymentOrders(ctx context.Context, filter apppayment.PaymentOrderFilter) ([]paymentdomain.PaymentOrder, int64, error) {
	query := paymentOrderFilterQuery(r.db.WithContext(ctx).Model(&paymentdomain.PaymentOrder{}), filter)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var orders []paymentdomain.PaymentOrder
	err := query.Order("created_at DESC, id DESC").Limit(filter.Page.Size).Offset(filter.Page.Offset()).Find(&orders).Error
	return orders, total, err
}

type paymentOrderTotalsRow struct {
	TotalOrders      int64
	PendingOrders    int64
	SuccessOrders    int64
	FailedOrders     int64
	ExpiredOrders    int64
	GrossAmount      int64
	CommissionAmount int64
	NetAmount        int64
}

// SummarizePaymentOrders tổng hợp theo cùng bộ lọc với ListPaymentOrders (bỏ phân trang) trong một
// câu truy vấn; tổng tiền chỉ tính đơn SUCCESS.
func (r *paymentRepository) SummarizePaymentOrders(ctx context.Context, filter apppayment.PaymentOrderFilter) (apppayment.PaymentOrderTotals, error) {
	var row paymentOrderTotalsRow
	err := paymentOrderFilterQuery(r.db.WithContext(ctx).Model(&paymentdomain.PaymentOrder{}), filter).
		Select(paymentOrderTotalsSelect,
			paymentdomain.OrderStatusPending, paymentdomain.OrderStatusSuccess, paymentdomain.OrderStatusFailed, paymentdomain.OrderStatusExpired,
			paymentdomain.OrderStatusSuccess, paymentdomain.OrderStatusSuccess, paymentdomain.OrderStatusSuccess).
		Scan(&row).Error
	if err != nil {
		return apppayment.PaymentOrderTotals{}, err
	}
	return apppayment.PaymentOrderTotals(row), nil
}

const paymentOrderTotalsSelect = `COUNT(*) AS total_orders,
	COUNT(*) FILTER (WHERE status = ?) AS pending_orders,
	COUNT(*) FILTER (WHERE status = ?) AS success_orders,
	COUNT(*) FILTER (WHERE status = ?) AS failed_orders,
	COUNT(*) FILTER (WHERE status = ?) AS expired_orders,
	COALESCE(SUM(gross_amount) FILTER (WHERE status = ?), 0) AS gross_amount,
	COALESCE(SUM(commission_amount) FILTER (WHERE status = ?), 0) AS commission_amount,
	COALESCE(SUM(net_amount) FILTER (WHERE status = ?), 0) AS net_amount`

func paymentOrderFilterQuery(query *gorm.DB, filter apppayment.PaymentOrderFilter) *gorm.DB {
	if filter.PayerID != uuid.Nil {
		query = query.Where("payer_id = ?", filter.PayerID)
	}
	if filter.ExpertID != uuid.Nil {
		query = query.Where("expert_id = ?", filter.ExpertID)
	}
	if filter.ScopeExperts {
		query = query.Where("expert_id IN ?", filter.ExpertIDs)
	}
	if filter.AppointmentID != nil {
		query = query.Where("appointment_id = ?", *filter.AppointmentID)
	}
	switch filter.Type {
	case apppayment.PaymentOrderTypeAppointment:
		query = query.Where("appointment_id IS NOT NULL")
	case apppayment.PaymentOrderTypeTopUp:
		query = query.Where("appointment_id IS NULL")
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.FulfillmentStatus != "" {
		query = query.Where("fulfillment_status = ?", filter.FulfillmentStatus)
	}
	if filter.FromMs > 0 {
		query = query.Where("created_at >= ?", filter.FromMs)
	}
	if filter.ToMs > 0 {
		query = query.Where("created_at < ?", filter.ToMs)
	}
	return query
}

// ApplyOrderReview khoá đơn thanh toán, áp dụng thay đổi của Admin và lưu hồ sơ bồi hoàn trong
// cùng transaction.
func (r *paymentRepository) ApplyOrderReview(ctx context.Context, orderID uuid.UUID, apply func(order *paymentdomain.PaymentOrder) (*paymentdomain.PaymentCompensationCase, error)) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order paymentdomain.PaymentOrder
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", orderID).First(&order).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apppayment.ErrPaymentOrderNotFound
		}
		if err != nil {
			return err
		}
		compensationCase, err := apply(&order)
		if err != nil {
			return err
		}
		if err := tx.Save(&order).Error; err != nil {
			return mapWriteError(err)
		}
		if err := tx.Create(compensationCase).Error; err != nil {
			return mapWriteError(err)
		}
		return nil
	})
}

// UpdateCompensationCase khoá hồ sơ bồi hoàn, áp dụng thay đổi và lưu lại.
func (r *paymentRepository) UpdateCompensationCase(ctx context.Context, caseID uuid.UUID, apply func(compensationCase *paymentdomain.PaymentCompensationCase) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var compensationCase paymentdomain.PaymentCompensationCase
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", caseID).First(&compensationCase).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apppayment.ErrCompensationCaseNotFound
		}
		if err != nil {
			return err
		}
		if err := apply(&compensationCase); err != nil {
			return err
		}
		return tx.Save(&compensationCase).Error
	})
}

func (r *paymentRepository) GetByIDForUpdate(tx *gorm.DB, orderID uuid.UUID) (*paymentdomain.PaymentOrder, error) {
	var order paymentdomain.PaymentOrder
	err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", orderID).
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *paymentRepository) GetByGatewayTxnRef(ref string) (*paymentdomain.PaymentOrder, error) {
	var order paymentdomain.PaymentOrder
	err := r.db.Where("gateway_txn_ref = ?", ref).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *paymentRepository) GetSuccessfulByAppointment(ctx context.Context, appointmentID uuid.UUID) (*paymentdomain.PaymentOrder, error) {
	var order paymentdomain.PaymentOrder
	err := r.db.WithContext(ctx).
		Where("appointment_id = ? AND status = ?", appointmentID, paymentdomain.OrderStatusSuccess).
		First(&order).Error
	if err != nil {
		return nil, mapTxError(err)
	}
	return &order, nil
}

func (r *paymentRepository) GetByGatewayTxnRefWithTx(tx *gorm.DB, ref string) (*paymentdomain.PaymentOrder, error) {
	var order paymentdomain.PaymentOrder
	err := tx.Where("gateway_txn_ref = ?", ref).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *paymentRepository) Update(order *paymentdomain.PaymentOrder) error {
	return r.db.Save(order).Error
}

type compensationCaseRow struct {
	paymentdomain.PaymentCompensationCase `gorm:"embedded"`
	PaymentExpertID                       uuid.UUID                          `gorm:"column:payment_expert_id"`
	PaymentPayerID                        uuid.UUID                          `gorm:"column:payment_payer_id"`
	PaymentStatus                         paymentdomain.PaymentOrderStatus   `gorm:"column:payment_status"`
	GatewayCaptureStatus                  paymentdomain.GatewayCaptureStatus `gorm:"column:payment_gateway_capture_status"`
	FulfillmentStatus                     paymentdomain.FulfillmentStatus    `gorm:"column:payment_fulfillment_status"`
}

func (r *paymentRepository) ListCompensationCases(ctx context.Context, filter apppayment.CompensationCaseFilter) ([]apppayment.CompensationCaseRecord, int64, error) {
	query := compensationCaseReadQuery(r.db.WithContext(ctx))
	if filter.Status != "" {
		query = query.Where("compensation.status = ?", filter.Status)
	}
	if filter.ReasonCode != "" {
		query = query.Where("compensation.reason_code = ?", filter.ReasonCode)
	}
	if filter.AppointmentID != nil {
		query = query.Where("compensation.appointment_id = ?", *filter.AppointmentID)
	}
	if filter.PaymentOrderID != nil {
		query = query.Where("compensation.payment_order_id = ?", *filter.PaymentOrderID)
	}
	// Phạm vi chuyên gia Admin quản lý luôn được áp dụng (rỗng = không có hồ sơ nào).
	query = query.Where("payment.expert_id IN ?", filter.ExpertIDs)
	if filter.FromMs > 0 {
		query = query.Where("compensation.created_at >= ?", filter.FromMs)
	}
	if filter.ToMs > 0 {
		query = query.Where("compensation.created_at < ?", filter.ToMs)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []compensationCaseRow
	err := query.
		Order("compensation.created_at DESC").
		Order("compensation.id DESC").
		Offset(filter.Page * filter.Size).
		Limit(filter.Size).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	return compensationCaseRecords(rows), total, nil
}

func (r *paymentRepository) GetCompensationCase(ctx context.Context, caseID uuid.UUID) (*apppayment.CompensationCaseRecord, error) {
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
		Select("compensation.*, payment.expert_id AS payment_expert_id, payment.payer_id AS payment_payer_id, payment.status AS payment_status, payment.gateway_capture_status AS payment_gateway_capture_status, payment.fulfillment_status AS payment_fulfillment_status").
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
		Case: row.PaymentCompensationCase, ExpertID: row.PaymentExpertID, PayerID: row.PaymentPayerID, PaymentStatus: row.PaymentStatus,
		GatewayCaptureStatus: row.GatewayCaptureStatus, FulfillmentStatus: row.FulfillmentStatus,
	}
}

func (r *paymentRepository) UpdateWithTx(tx *gorm.DB, order *paymentdomain.PaymentOrder) error {
	return tx.Save(order).Error
}

func (r *paymentRepository) SaveOutboxEvent(tx *gorm.DB, event *paymentdomain.OutboxEvent) error {
	return tx.Create(event).Error
}

func (r *paymentRepository) WithTransaction(fn func(tx *gorm.DB) error) error {
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

func (tx *paymentTx) GetOrderForUpdate(ctx context.Context, orderID uuid.UUID) (*paymentdomain.PaymentOrder, error) {
	var order paymentdomain.PaymentOrder
	err := tx.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", orderID).
		First(&order).Error
	if err != nil {
		return nil, mapTxError(err)
	}
	return &order, nil
}

func (tx *paymentTx) GetOrder(ctx context.Context, orderID uuid.UUID) (*paymentdomain.PaymentOrder, error) {
	var order paymentdomain.PaymentOrder
	err := tx.db.WithContext(ctx).Where("id = ?", orderID).First(&order).Error
	if err != nil {
		return nil, mapTxError(err)
	}
	return &order, nil
}

func (tx *paymentTx) ListOrdersForAppointmentForUpdate(ctx context.Context, appointmentID uuid.UUID) ([]paymentdomain.PaymentOrder, error) {
	var orders []paymentdomain.PaymentOrder
	err := tx.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("appointment_id = ?", appointmentID).
		Order("id").
		Find(&orders).Error
	return orders, err
}

func (tx *paymentTx) CreateOrder(ctx context.Context, order *paymentdomain.PaymentOrder) error {
	return mapWriteError(tx.db.WithContext(ctx).Create(order).Error)
}

func (tx *paymentTx) GetGatewayTxnRef(ctx context.Context, ref string) (*paymentdomain.PaymentOrder, error) {
	var order paymentdomain.PaymentOrder
	err := tx.db.WithContext(ctx).Where("gateway_txn_ref = ?", ref).First(&order).Error
	if err != nil {
		return nil, mapTxError(err)
	}
	return &order, nil
}

func (tx *paymentTx) UpdateOrder(ctx context.Context, order *paymentdomain.PaymentOrder) error {
	return mapWriteError(tx.db.WithContext(ctx).Save(order).Error)
}

func (tx *paymentTx) ExpireOtherPendingOrders(ctx context.Context, appointmentID, exceptOrderID uuid.UUID) error {
	return tx.db.WithContext(ctx).
		Model(&paymentdomain.PaymentOrder{}).
		Where("appointment_id = ? AND id <> ? AND status = ?", appointmentID, exceptOrderID, paymentdomain.OrderStatusPending).
		Update("status", paymentdomain.OrderStatusExpired).Error
}

func (tx *paymentTx) SaveOutboxEvent(ctx context.Context, event *paymentdomain.OutboxEvent) error {
	return tx.db.WithContext(ctx).Create(event).Error
}

func (tx *paymentTx) SaveCompensationCase(ctx context.Context, compensationCase *paymentdomain.PaymentCompensationCase) error {
	return tx.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "payment_order_id"}, {Name: "reason_code"}},
			DoNothing: true,
		}).
		Create(compensationCase).Error
}

func (tx *paymentTx) CreditWalletPending(ctx context.Context, userID uuid.UUID, amount money.Money, refID uuid.UUID, idempotencyKey string) error {
	return tx.walletUsecase.CreditPendingWithTx(ctx, tx.db, userID, amount, "PAYMENT_ORDER", refID, idempotencyKey)
}

func (tx *paymentTx) DebitWalletPending(ctx context.Context, userID uuid.UUID, amount money.Money, refID uuid.UUID, idempotencyKey string) error {
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
		case "ux_payment_compensation_cases_order_reason":
			return apppayment.ErrCompensationCaseExists
		}
	}
	return err
}
