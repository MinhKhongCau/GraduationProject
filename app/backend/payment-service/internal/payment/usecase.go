package payment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"payment-service/internal/domain/entity"
	"payment-service/internal/domain/vo"
	"payment-service/internal/payment/gateway"
	"payment-service/internal/wallet"
	"strconv"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Usecase interface {
	// CreateOrder tạo một payment order mới và trả về payment URL.
	// appointmentID là optional: nếu có giá trị, order sẽ được liên kết với lịch hẹn tương ứng.
	CreateOrder(ctx context.Context, payerID, expertID uuid.UUID, grossAmount vo.Money, gatewayName string, ipAddr string, appointmentID *string) (*entity.PaymentOrder, string, error)
	ProcessIPN(ctx context.Context, params map[string][]string) (bool, error)
}

type paymentUsecase struct {
	repo          Repository
	walletUsecase wallet.Usecase
	vnpayClient   *gateway.VNPayClient
}

func NewUsecase(repo Repository, walletUsecase wallet.Usecase, vnpayClient *gateway.VNPayClient) Usecase {
	return &paymentUsecase{
		repo:          repo,
		walletUsecase: walletUsecase,
		vnpayClient:   vnpayClient,
	}
}

func (u *paymentUsecase) CreateOrder(ctx context.Context, payerID, expertID uuid.UUID, grossAmount vo.Money, gatewayName string, ipAddr string, appointmentID *string) (*entity.PaymentOrder, string, error) {
	// Mức hoa hồng là 15%
	commissionRate := 0.15
	commissionAmount := vo.Money(float64(grossAmount) * commissionRate)
	netAmount := grossAmount.Sub(commissionAmount)

	order := &entity.PaymentOrder{
		ID:               uuid.New(),
		PayerID:          payerID,
		ExpertID:         expertID,
		GrossAmount:      grossAmount,
		CommissionRate:   commissionRate,
		CommissionAmount: commissionAmount,
		NetAmount:        netAmount,
		Gateway:          gatewayName,
		Status:           entity.OrderStatusPending,
	}

	// Gắn appointment_id nếu được cung cấp
	if appointmentID != nil && *appointmentID != "" {
		parsed, err := uuid.Parse(*appointmentID)
		if err != nil {
			return nil, "", fmt.Errorf("invalid appointment_id format: %w", err)
		}
		order.AppointmentID = &parsed
	}

	if err := u.repo.Create(order); err != nil {
		return nil, "", err
	}

	var paymentURL string
	if gatewayName == "VNPAY" {
		createDate := time.Now().Format("20060102150405")
		desc := fmt.Sprintf("Thanh toan MindCare don hang %s", order.ID.String())
		paymentURL = u.vnpayClient.GeneratePaymentURL(order.ID.String(), order.GrossAmount.Int64(), ipAddr, desc, createDate)
	} else {
		// Mock url
		paymentURL = fmt.Sprintf("http://localhost:8082/api/v1/payments/mock-checkout?order_id=%s", order.ID.String())
	}

	return order, paymentURL, nil
}

func (u *paymentUsecase) ProcessIPN(ctx context.Context, params map[string][]string) (bool, error) {
	// 1. Verify Checksum if vnp_SecureHash exists
	if len(params["vnp_SecureHash"]) > 0 {
		if !u.vnpayClient.VerifyChecksum(params) {
			log.Println("❌ VNPay IPN checksum verification failed")
			return false, errors.New("invalid checksum")
		}
	}

	// 2. Trích xuất thông tin giao dịch
	vnpTxnRef := ""
	if len(params["vnp_TxnRef"]) > 0 {
		vnpTxnRef = params["vnp_TxnRef"][0]
	}
	if vnpTxnRef == "" {
		return false, errors.New("missing vnp_TxnRef")
	}

	orderID, err := uuid.Parse(vnpTxnRef)
	if err != nil {
		return false, errors.New("invalid order uuid format")
	}

	vnpAmountStr := ""
	if len(params["vnp_Amount"]) > 0 {
		vnpAmountStr = params["vnp_Amount"][0]
	}
	vnpAmount, _ := strconv.ParseInt(vnpAmountStr, 10, 64)
	vnpAmount = vnpAmount / 100 // VNPay scale * 100

	vnpResponseCode := "00"
	if len(params["vnp_ResponseCode"]) > 0 {
		vnpResponseCode = params["vnp_ResponseCode"][0]
	}

	vnpTxnNo := ""
	if len(params["vnp_TransactionNo"]) > 0 {
		vnpTxnNo = params["vnp_TransactionNo"][0]
	}

	// 3. Tiến hành cập nhật Database
	var alreadyProcessed bool
	err = u.repo.WithTransaction(func(tx *gorm.DB) error {
		order, err := u.repo.GetByID(orderID)
		if err != nil {
			return err
		}

		if order.Status != entity.OrderStatusPending {
			alreadyProcessed = true
			return nil
		}

		if order.GrossAmount.Int64() != vnpAmount {
			return errors.New("amount mismatch")
		}

		// Replay attack check
		if vnpTxnNo != "" {
			existing, err := u.repo.GetByGatewayTxnRef(vnpTxnNo)
			if err == nil && existing.ID != order.ID {
				return errors.New("replay attack: transaction reference already processed")
			}
		}

		paidAt := time.Now().UnixMilli()
		if vnpResponseCode == "00" {
			order.Status = entity.OrderStatusSuccess
			order.PaidAt = &paidAt
			order.GatewayTxnRef = vnpTxnNo

			if err := u.repo.UpdateWithTx(tx, order); err != nil {
				return err
			}

			// Lưu Outbox event 1: wallet nhận tiền
			payloadMap := map[string]interface{}{
				"order_id":          order.ID,
				"payer_id":          order.PayerID,
				"expert_id":         order.ExpertID,
				"gross_amount":      order.GrossAmount.Int64(),
				"commission_amount": order.CommissionAmount.Int64(),
				"net_amount":        order.NetAmount.Int64(),
				"paid_at":           paidAt,
			}
			payloadBytes, _ := json.Marshal(payloadMap)

			outbox := &entity.OutboxEvent{
				AggregateType: "PAYMENT_ORDER",
				AggregateID:   order.ID,
				EventType:     "wallet.payment.received",
				Payload:       string(payloadBytes),
				Published:     false,
			}
			if err := u.repo.SaveOutboxEvent(tx, outbox); err != nil {
				return err
			}

			// Lưu Outbox event 2: thông báo Booking Service xác nhận lịch hẹn
			// Chỉ emit nếu order này có appointment_id.
			if order.AppointmentID != nil {
				bookingPayload, _ := json.Marshal(map[string]interface{}{
					"appointment_id": order.AppointmentID.String(),
					"order_id":       order.ID.String(),
					"status":         "SUCCESS",
				})
				bookingOutbox := &entity.OutboxEvent{
					AggregateType: "PAYMENT_ORDER",
					AggregateID:   order.ID,
					EventType:     "booking.appointment.confirm",
					Payload:       string(bookingPayload),
					Published:     false,
				}
				if err := u.repo.SaveOutboxEvent(tx, bookingOutbox); err != nil {
					return err
				}
			}

			// Thao tác với ví: credit gross, sau đó debit commission
			creditKey := fmt.Sprintf("credit_order_%s", order.ID.String())
			err = u.walletUsecase.CreditPending(ctx, order.ExpertID, order.GrossAmount, "PAYMENT_ORDER", order.ID, creditKey)
			if err != nil {
				return fmt.Errorf("wallet credit failed: %w", err)
			}

			debitKey := fmt.Sprintf("debit_order_%s", order.ID.String())
			err = u.walletUsecase.DebitPending(ctx, order.ExpertID, order.CommissionAmount, "PAYMENT_ORDER", order.ID, debitKey)
			if err != nil {
				return fmt.Errorf("wallet commission debit failed: %w", err)
			}

		} else {
			order.Status = entity.OrderStatusFailed
			if err := u.repo.UpdateWithTx(tx, order); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return false, err
	}

	return alreadyProcessed, nil
}
