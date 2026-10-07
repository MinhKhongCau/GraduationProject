package payment

import (
	"context"
	"log"

	"github.com/google/uuid"
)

// BookingStatusUnknown được trả về khi không đọc được trạng thái lịch hẹn từ booking-service.
const BookingStatusUnknown = "UNKNOWN"

// BookingStatusReader đọc trạng thái hiện tại của lịch hẹn (gRPC booking-service).
type BookingStatusReader interface {
	GetAppointmentStatus(ctx context.Context, appointmentID string) (string, error)
}

// ReturnUsecase xử lý trình duyệt quay về từ VNPay (vnp_ReturnUrl).
type ReturnUsecase interface {
	ProcessReturn(ctx context.Context, params map[string][]string) (*PaymentReturnResult, error)
}

type PaymentReturnResult struct {
	OrderID       string `json:"order_id"`
	AppointmentID string `json:"appointment_id,omitempty"`
	// PENDING | SUCCESS | FAILED | EXPIRED
	PaymentStatus string `json:"payment_status"`
	// PENDING_PAYMENT | CONFIRMED | CANCELLED | COMPLETED | UNKNOWN; rỗng khi order không gắn lịch hẹn.
	BookingStatus string `json:"booking_status,omitempty"`
	AmountVND     int64  `json:"amount_vnd"`
	ResponseCode  string `json:"response_code"`
	GatewayTxnRef string `json:"gateway_txn_ref,omitempty"`
	PaidAt        *int64 `json:"paid_at,omitempty"`
}

// ProcessReturn xác thực chữ ký các tham số VNPay gắn vào return URL rồi chốt order bằng đúng
// luồng IPN (idempotent). Nhờ vậy order vẫn được cập nhật khi IPN không tới được máy dev
// (localhost). Nếu IPN đã xử lý trước, bước chốt là no-op. Kết quả gồm trạng thái thanh toán
// và trạng thái lịch hẹn hiện tại để FE hiển thị.
func (u *paymentUsecase) ProcessReturn(ctx context.Context, params map[string][]string) (*PaymentReturnResult, error) {
	if _, err := u.ProcessIPN(ctx, params); err != nil {
		return nil, err
	}

	// ProcessIPN đã kiểm tra vnp_TxnRef là UUID hợp lệ.
	orderID := uuid.MustParse(firstIPNValue(params, "vnp_TxnRef"))
	order, err := u.repo.GetByID(orderID)
	if err != nil {
		return nil, ErrPaymentOrderNotFound
	}

	result := &PaymentReturnResult{
		OrderID:       order.ID.String(),
		PaymentStatus: order.Status.String(),
		AmountVND:     order.GrossAmount.Int64(),
		ResponseCode:  firstIPNValue(params, "vnp_ResponseCode"),
		GatewayTxnRef: order.GatewayTxnRef,
		PaidAt:        order.PaidAt,
	}
	if order.AppointmentID != nil {
		result.AppointmentID = order.AppointmentID.String()
		result.BookingStatus = u.readBookingStatus(ctx, result.AppointmentID)
	}
	return result, nil
}

func (u *paymentUsecase) readBookingStatus(ctx context.Context, appointmentID string) string {
	reader, ok := u.bookingClient.(BookingStatusReader)
	if !ok {
		return BookingStatusUnknown
	}
	status, err := reader.GetAppointmentStatus(ctx, appointmentID)
	if err != nil {
		log.Printf("[payment_return] read booking status appointment_id=%s failed: %v", appointmentID, err)
		return BookingStatusUnknown
	}
	return status
}
