package payment

import "encoding/json"

// Event fact gửi lên RabbitMQ (exchange payment.events) qua outbox.
const (
	PaymentSucceededEvent = "payment.succeeded"
	PaymentFailedEvent    = "payment.failed"
)

func BookingOutboxPayload(appointmentID, orderID, status string) (string, error) {
	payload, err := json.Marshal(map[string]interface{}{
		"appointment_id": appointmentID,
		"order_id":       orderID,
		"status":         status,
	})
	return string(payload), err
}

// PaymentStatusPayload là payload lưu trong outbox cho event payment.succeeded / payment.failed.
// Outbox publisher chuyển nó sang paymentpb.PaymentStatusChangedEvent khi publish.
type PaymentStatusPayload struct {
	OrderID             string `json:"order_id"`
	AppointmentID       string `json:"appointment_id,omitempty"`
	PayerID             string `json:"payer_id"`
	ExpertID            string `json:"expert_id"`
	AmountVND           int64  `json:"amount_vnd"`
	Status              string `json:"status"` // SUCCESS | FAILED
	Gateway             string `json:"gateway"`
	GatewayTxnRef       string `json:"gateway_txn_ref,omitempty"`
	GatewayResponseCode string `json:"gateway_response_code,omitempty"`
	PaidAt              int64  `json:"paid_at,omitempty"`
}

func PaymentStatusOutboxPayload(order *PaymentOrder) (string, error) {
	payload := PaymentStatusPayload{
		OrderID:             order.ID.String(),
		PayerID:             order.PayerID.String(),
		ExpertID:            order.ExpertID.String(),
		AmountVND:           order.GrossAmount.Int64(),
		Status:              order.Status.String(),
		Gateway:             order.Gateway,
		GatewayTxnRef:       order.GatewayTxnRef,
		GatewayResponseCode: order.GatewayResponseCode,
	}
	if order.AppointmentID != nil {
		payload.AppointmentID = order.AppointmentID.String()
	}
	if order.PaidAt != nil {
		payload.PaidAt = *order.PaidAt
	}
	raw, err := json.Marshal(payload)
	return string(raw), err
}
