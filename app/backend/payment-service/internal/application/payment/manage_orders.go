package payment

import (
	"context"
	"payment-service/internal/application/managedscope"
	"payment-service/internal/application/readquery"
	paymentdomain "payment-service/internal/domain/payment"

	"github.com/google/uuid"
)

// OrderReviewStore ghi các thao tác của Admin trong một transaction, khoá dòng trước khi gọi apply
// để quy tắc nghiệp vụ được kiểm tra trên dữ liệu mới nhất.
type OrderReviewStore interface {
	// ApplyOrderReview khoá đơn, gọi apply rồi lưu đơn đã đổi cùng hồ sơ bồi hoàn apply trả về.
	ApplyOrderReview(ctx context.Context, orderID uuid.UUID, apply func(order *paymentdomain.PaymentOrder) (*paymentdomain.PaymentCompensationCase, error)) error
	// UpdateCompensationCase khoá hồ sơ bồi hoàn, gọi apply rồi lưu lại.
	UpdateCompensationCase(ctx context.Context, caseID uuid.UUID, apply func(compensationCase *paymentdomain.PaymentCompensationCase) error) error
}

// ---------- Chuyên gia: doanh thu của chính mình ----------

func (u *paymentUsecase) ListExpertOrders(ctx context.Context, expertID uuid.UUID, filter PaymentOrderFilter) (*readquery.Page[ExpertPaymentOrderView], error) {
	if expertID == uuid.Nil {
		return nil, ErrPaymentOrderForbidden
	}
	filter = expertFilter(expertID, filter)
	return listOrders(ctx, u, filter, expertPaymentOrderView)
}

func (u *paymentUsecase) GetExpertOrder(ctx context.Context, expertID, orderID uuid.UUID) (*ExpertPaymentOrderView, error) {
	if expertID == uuid.Nil || orderID == uuid.Nil {
		return nil, ErrPaymentOrderForbidden
	}
	order, err := u.getOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order.ExpertID != expertID {
		return nil, ErrPaymentOrderForbidden
	}
	result := expertPaymentOrderView(order)
	return &result, nil
}

func (u *paymentUsecase) SummarizeExpertOrders(ctx context.Context, expertID uuid.UUID, filter PaymentOrderFilter) (*ExpertOrderSummary, error) {
	if expertID == uuid.Nil {
		return nil, ErrPaymentOrderForbidden
	}
	filter = expertFilter(expertID, filter)
	totals, err := u.summarize(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &ExpertOrderSummary{TotalOrders: totals.TotalOrders, SuccessOrders: totals.SuccessOrders, GrossTotal: totals.GrossAmount, CommissionTotal: totals.CommissionAmount, NetTotal: totals.NetAmount, FromMs: filter.FromMs, ToMs: filter.ToMs}, nil
}

// expertFilter ép bộ lọc về đúng chuyên gia đang đăng nhập, bỏ mọi tiêu chí phạm vi khác từ request.
func expertFilter(expertID uuid.UUID, filter PaymentOrderFilter) PaymentOrderFilter {
	filter.ExpertID = expertID
	filter.PayerID = uuid.Nil
	filter.ScopeExperts = false
	filter.ExpertIDs = nil
	return filter
}

// ---------- Admin: giao dịch của các chuyên gia mình quản lý ----------

func (u *paymentUsecase) ListAdminOrders(ctx context.Context, adminID uuid.UUID, filter PaymentOrderFilter) (*readquery.Page[AdminPaymentOrderView], error) {
	filter, err := u.adminFilter(ctx, adminID, filter)
	if err != nil {
		return nil, err
	}
	return listOrders(ctx, u, filter, adminPaymentOrderView)
}

func (u *paymentUsecase) GetAdminOrder(ctx context.Context, adminID, orderID uuid.UUID) (*AdminPaymentOrderView, error) {
	order, err := u.managedOrder(ctx, adminID, orderID)
	if err != nil {
		return nil, err
	}
	result := adminPaymentOrderView(order)
	return &result, nil
}

func (u *paymentUsecase) SummarizeAdminOrders(ctx context.Context, adminID uuid.UUID, filter PaymentOrderFilter) (*AdminOrderSummary, error) {
	filter, err := u.adminFilter(ctx, adminID, filter)
	if err != nil {
		return nil, err
	}
	totals, err := u.summarize(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &AdminOrderSummary{
		TotalOrders: totals.TotalOrders, PendingOrders: totals.PendingOrders, SuccessOrders: totals.SuccessOrders,
		FailedOrders: totals.FailedOrders, ExpiredOrders: totals.ExpiredOrders,
		GrossTotal: totals.GrossAmount, CommissionTotal: totals.CommissionAmount, NetTotal: totals.NetAmount,
		ManagedExperts: len(filter.ExpertIDs), FromMs: filter.FromMs, ToMs: filter.ToMs,
	}, nil
}

// ReviewOrder: Admin đưa đơn đã thanh toán của chuyên gia mình quản lý vào diện kiểm tra thủ công
// (MANUAL_REVIEW) hoặc yêu cầu hoàn tiền (REFUND_REQUIRED); tạo hồ sơ bồi hoàn tương ứng.
func (u *paymentUsecase) ReviewOrder(ctx context.Context, adminID, orderID uuid.UUID, action paymentdomain.CompensationStatus, note string) (*CompensationCase, error) {
	if _, err := u.managedOrder(ctx, adminID, orderID); err != nil {
		return nil, err
	}
	store, ok := u.repo.(OrderReviewStore)
	if !ok {
		return nil, ErrOrderReviewStoreUnavailable
	}
	now := u.clock().UnixMilli()
	var created *paymentdomain.PaymentCompensationCase
	err := store.ApplyOrderReview(ctx, orderID, func(order *paymentdomain.PaymentOrder) (*paymentdomain.PaymentCompensationCase, error) {
		fulfillment, plan, err := paymentdomain.PlanAdminReview(order, action, note)
		if err != nil {
			return nil, err
		}
		order.FulfillmentStatus = fulfillment
		created = newCompensationCase(order, plan, now)
		return created, nil
	})
	if err != nil {
		return nil, err
	}
	return u.GetCompensationCase(ctx, adminID, created.ID)
}

// ResolveCompensationCase: Admin đóng hồ sơ bồi hoàn thuộc chuyên gia mình quản lý.
func (u *paymentUsecase) ResolveCompensationCase(ctx context.Context, adminID, caseID uuid.UUID, note string) (*CompensationCase, error) {
	if _, err := u.managedCompensationCase(ctx, adminID, caseID); err != nil {
		return nil, err
	}
	store, ok := u.repo.(OrderReviewStore)
	if !ok {
		return nil, ErrOrderReviewStoreUnavailable
	}
	now := u.clock().UnixMilli()
	err := store.UpdateCompensationCase(ctx, caseID, func(compensationCase *paymentdomain.PaymentCompensationCase) error {
		return compensationCase.Resolve(note, now)
	})
	if err != nil {
		return nil, err
	}
	return u.GetCompensationCase(ctx, adminID, caseID)
}

// adminFilter giới hạn bộ lọc trong phạm vi chuyên gia adminID quản lý. filter.ExpertID (nếu có)
// phải thuộc phạm vi đó.
func (u *paymentUsecase) adminFilter(ctx context.Context, adminID uuid.UUID, filter PaymentOrderFilter) (PaymentOrderFilter, error) {
	var requested *uuid.UUID
	if filter.ExpertID != uuid.Nil {
		requested = &filter.ExpertID
	}
	scope, err := managedscope.Resolve(ctx, u.managedExperts, adminID, requested)
	if err != nil {
		return filter, err
	}
	filter.ExpertID = uuid.Nil
	filter.ScopeExperts = true
	filter.ExpertIDs = scope.ExpertIDs()
	return filter, nil
}

func (u *paymentUsecase) managedOrder(ctx context.Context, adminID, orderID uuid.UUID) (*paymentdomain.PaymentOrder, error) {
	if orderID == uuid.Nil {
		return nil, ErrPaymentOrderForbidden
	}
	order, err := u.getOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if _, err := managedscope.Resolve(ctx, u.managedExperts, adminID, &order.ExpertID); err != nil {
		return nil, err
	}
	return order, nil
}

func newCompensationCase(order *paymentdomain.PaymentOrder, plan paymentdomain.CompensationPlan, nowMs int64) *paymentdomain.PaymentCompensationCase {
	return &paymentdomain.PaymentCompensationCase{
		ID:                       uuid.New(),
		PaymentOrderID:           order.ID,
		AppointmentID:            *order.AppointmentID,
		Type:                     plan.Type,
		Status:                   plan.Status,
		ReasonCode:               plan.ReasonCode,
		SafeReason:               plan.SafeReason,
		GatewayOrderReference:    order.ID.String(),
		GatewayTransactionNumber: order.GatewayTxnRef,
		GatewayResponseCode:      order.GatewayResponseCode,
		GatewayTransactionStatus: order.GatewayTransactionStatus,
		GatewayPaymentDate:       order.GatewayPaymentDate,
		AmountVND:                order.GrossAmount,
		CreatedAt:                nowMs,
		UpdatedAt:                nowMs,
	}
}
