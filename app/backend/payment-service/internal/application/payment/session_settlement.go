package payment

import (
	"context"
	"errors"
	"fmt"
	"payment-service/internal/domain/money"
	paymentdomain "payment-service/internal/domain/payment"
	walletdomain "payment-service/internal/domain/wallet"

	"github.com/google/uuid"
)

// DefaultSystemWalletUserID là chủ sở hữu ví hệ thống khi SYSTEM_WALLET_USER_ID không được cấu hình.
var DefaultSystemWalletUserID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

var (
	ErrSessionOrderNotFound   = errors.New("no successful payment order for appointment")
	ErrSessionNotSettleable   = errors.New("payment order is under review or refund and cannot be paid out")
	ErrInvalidSettlementInput = errors.New("appointment id is required")
)

// SettlementUsecase trả thù lao cho chuyên gia khi buổi tư vấn hoàn tất (booking-service gọi).
type SettlementUsecase interface {
	SettleCompletedSession(ctx context.Context, appointmentID uuid.UUID) (*SessionSettlement, error)
}

type SessionSettlement struct {
	OrderID          uuid.UUID `json:"order_id"`
	AppointmentID    uuid.UUID `json:"appointment_id"`
	ExpertID         uuid.UUID `json:"expert_id"`
	GrossAmount      int64     `json:"gross_amount"`
	CommissionAmount int64     `json:"commission_amount"`
	NetAmount        int64     `json:"net_amount"`
	AlreadySettled   bool      `json:"already_settled"`
}

func NewSettlementUsecase(uow UnitOfWork, systemWalletUserID uuid.UUID) SettlementUsecase {
	if systemWalletUserID == uuid.Nil {
		systemWalletUserID = DefaultSystemWalletUserID
	}
	return &paymentUsecase{uow: uow, systemWalletUserID: systemWalletUserID}
}

func systemEscrowKey(orderID uuid.UUID) string {
	return fmt.Sprintf("escrow_order_%s", orderID.String())
}

// SettleCompletedSession chuyển tiền của một lịch hẹn đã hoàn tất:
//   - ví hệ thống: Pending -gross, Available +commission (hệ thống giữ hoa hồng)
//   - ví chuyên gia: Available +net
//
// Idempotent: order.Released đánh dấu đã chi trả, gọi lại trả AlreadySettled=true.
func (u *paymentUsecase) SettleCompletedSession(ctx context.Context, appointmentID uuid.UUID) (*SessionSettlement, error) {
	if appointmentID == uuid.Nil {
		return nil, ErrInvalidSettlementInput
	}
	var result *SessionSettlement
	err := u.uow.WithinTx(ctx, func(tx Tx) error {
		orders, err := tx.ListOrdersForAppointmentForUpdate(ctx, appointmentID)
		if err != nil {
			return err
		}
		var order *paymentdomain.PaymentOrder
		for i := range orders {
			if orders[i].Status == paymentdomain.OrderStatusSuccess {
				order = &orders[i]
				break
			}
		}
		if order == nil {
			return ErrSessionOrderNotFound
		}

		result = &SessionSettlement{
			OrderID:          order.ID,
			AppointmentID:    appointmentID,
			ExpertID:         order.ExpertID,
			GrossAmount:      order.GrossAmount.Int64(),
			CommissionAmount: order.CommissionAmount.Int64(),
			NetAmount:        order.NetAmount.Int64(),
		}
		if order.Released {
			result.AlreadySettled = true
			return nil
		}
		if order.FulfillmentStatus == paymentdomain.FulfillmentManualReview || order.FulfillmentStatus == paymentdomain.FulfillmentRefundRequired {
			return ErrSessionNotSettleable
		}

		// Đơn thanh toán trước khi chuyển sang mô hình ví hệ thống đã ghi net_amount vào Pending
		// của chuyên gia: chỉ cần chuyển sang Available như worker giải phóng tiền cũ.
		legacy, err := tx.HasWalletTransaction(ctx, fmt.Sprintf("credit_order_%s", order.ID.String()))
		if err != nil {
			return err
		}
		if legacy {
			if err := tx.AdjustWallet(ctx, order.ExpertID, order.NetAmount, money.Money(-order.NetAmount.Int64()), walletdomain.TxTypeAdjustment, order.ID, "release_"+order.ID.String()); err != nil {
				return fmt.Errorf("release legacy expert pending failed: %w", err)
			}
		} else {
			systemKey := fmt.Sprintf("payout_system_%s", order.ID.String())
			if err := tx.AdjustWallet(ctx, u.systemWalletUserID, order.CommissionAmount, money.Money(-order.GrossAmount.Int64()), walletdomain.TxTypeSessionPayout, order.ID, systemKey); err != nil {
				return fmt.Errorf("system wallet payout failed: %w", err)
			}
			expertKey := fmt.Sprintf("payout_expert_%s", order.ID.String())
			if err := tx.AdjustWallet(ctx, order.ExpertID, order.NetAmount, 0, walletdomain.TxTypeSessionPayout, order.ID, expertKey); err != nil {
				return fmt.Errorf("expert wallet payout failed: %w", err)
			}
		}

		order.Released = true
		return tx.UpdateOrder(ctx, order)
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
