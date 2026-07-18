package payment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"payment-service/internal/booking/client"
	"payment-service/internal/domain/entity"
	"payment-service/internal/domain/vo"
	"payment-service/internal/wallet"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Usecase interface {
	// CreateOrder creates a VNPay payment order for a booking appointment.
	CreateOrder(ctx context.Context, payerID uuid.UUID, appointmentID string, ipAddr string) (*entity.PaymentOrder, string, error)
	ProcessIPN(ctx context.Context, params map[string][]string) (bool, error)
}

type PaymentGateway interface {
	GeneratePaymentURL(txnRef string, amount int64, ipAddr, desc, createDate string) string
	VerifyChecksum(params map[string][]string) bool
}

type paymentUsecase struct {
	repo          Repository
	walletUsecase wallet.Usecase
	vnpayClient   PaymentGateway
	bookingClient client.BookingServiceClient
}

func NewUsecase(repo Repository, walletUsecase wallet.Usecase, vnpayClient PaymentGateway, bookingClient client.BookingServiceClient) Usecase {
	return &paymentUsecase{
		repo:          repo,
		walletUsecase: walletUsecase,
		vnpayClient:   vnpayClient,
		bookingClient: bookingClient,
	}
}

func (u *paymentUsecase) CreateOrder(ctx context.Context, payerID uuid.UUID, appointmentID string, ipAddr string) (*entity.PaymentOrder, string, error) {
	appointmentIDValue := strings.TrimSpace(appointmentID)
	if appointmentIDValue == "" {
		return nil, "", fmt.Errorf("%w: appointment_id is required", ErrInvalidCreateOrderRequest)
	}

	parsedApptID, err := uuid.Parse(appointmentIDValue)
	if err != nil {
		return nil, "", fmt.Errorf("%w: invalid appointment_id format: %w", ErrInvalidCreateOrderRequest, err)
	}

	eligibility, err := u.bookingClient.GetPaymentEligibility(ctx, appointmentIDValue, payerID)
	if err != nil {
		if errors.Is(err, client.ErrAppointmentNotFound) {
			return nil, "", fmt.Errorf("%w: %s", ErrBookingAppointmentNotFound, appointmentIDValue)
		}
		if errors.Is(err, client.ErrPaymentEligibilityForbidden) {
			return nil, "", fmt.Errorf("%w: appointment %s", ErrAppointmentOwnership, appointmentIDValue)
		}
		if errors.Is(err, client.ErrPaymentEligibilityConflict) {
			return nil, "", fmt.Errorf("%w: appointment %s", ErrAppointmentInvalidState, appointmentIDValue)
		}
		if errors.Is(err, client.ErrInvalidBookingPrice) {
			return nil, "", fmt.Errorf("%w: appointment price is invalid", ErrInvalidBookingData)
		}
		return nil, "", fmt.Errorf("failed to verify payment eligibility: %w", err)
	}

	if !strings.EqualFold(strings.TrimSpace(eligibility.AppointmentID), appointmentIDValue) {
		return nil, "", fmt.Errorf("%w: eligibility appointment_id mismatch", ErrInvalidBookingData)
	}
	expertID, err := uuid.Parse(strings.TrimSpace(eligibility.ExpertID))
	if err != nil {
		return nil, "", fmt.Errorf("%w: invalid eligibility expert_id: %w", ErrInvalidBookingData, err)
	}
	if expertID == uuid.Nil {
		return nil, "", fmt.Errorf("%w: missing eligibility expert_id", ErrInvalidBookingData)
	}
	if eligibility.AmountVND <= 0 {
		return nil, "", fmt.Errorf("%w: eligibility amount_vnd must be greater than zero", ErrInvalidBookingData)
	}
	if eligibility.ExpiresAt <= time.Now().UnixMilli() {
		return nil, "", fmt.Errorf("%w: eligibility has expired", ErrAppointmentInvalidState)
	}

	grossAmount := vo.Money(eligibility.AmountVND)

	// Mức hoa hồng là 15% - Tính toán DỰA TRÊN grossAmount đã được xác thực
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
		Gateway:          "VNPAY",
		Status:           entity.OrderStatusPending,
		AppointmentID:    &parsedApptID,
	}

	if err := u.repo.Create(order); err != nil {
		return nil, "", err
	}

	// VNPay requires Vietnam-local timestamp format.
	// VNPay yêu cầu định dạng thời gian theo múi giờ Việt Nam (GMT+7)
	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	var createDate string
	if err == nil {
		createDate = time.Now().In(loc).Format("20060102150405")
	} else {
		createDate = time.Now().Add(7 * time.Hour).Format("20060102150405")
	}
	desc := fmt.Sprintf("Thanh toan MindCare don hang %s", order.ID.String())
	paymentURL := u.vnpayClient.GeneratePaymentURL(order.ID.String(), order.GrossAmount.Int64(), ipAddr, desc, createDate)
	return order, paymentURL, nil
}

func (u *paymentUsecase) ProcessIPN(ctx context.Context, params map[string][]string) (bool, error) {
	// 1. VNPay IPN must always be authenticated by secure hash.
	if len(params["vnp_SecureHash"]) == 0 || params["vnp_SecureHash"][0] == "" {
		log.Println("VNPay IPN missing secure hash")
		return false, errors.New("missing vnp_SecureHash")
	}
	if !u.vnpayClient.VerifyChecksum(params) {
		log.Println("VNPay IPN checksum verification failed")
		return false, errors.New("invalid checksum")
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
	if vnpAmountStr == "" {
		return false, errors.New("missing vnp_Amount")
	}
	vnpAmount, err := strconv.ParseInt(vnpAmountStr, 10, 64)
	if err != nil {
		return false, fmt.Errorf("invalid vnp_Amount: %w", err)
	}
	if vnpAmount <= 0 {
		return false, errors.New("invalid vnp_Amount: amount must be greater than zero")
	}
	vnpAmount = vnpAmount / 100 // VNPay scale * 100

	vnpResponseCode := ""
	if len(params["vnp_ResponseCode"]) > 0 {
		vnpResponseCode = params["vnp_ResponseCode"][0]
	}
	if vnpResponseCode == "" {
		return false, errors.New("missing vnp_ResponseCode")
	}

	vnpTxnNo := ""
	if len(params["vnp_TransactionNo"]) > 0 {
		vnpTxnNo = params["vnp_TransactionNo"][0]
	}

	// 3. Tiến hành cập nhật Database
	var alreadyProcessed bool
	err = u.repo.WithTransaction(func(tx *gorm.DB) error {
		order, err := u.repo.GetByIDForUpdate(tx, orderID)
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
			existing, err := u.repo.GetByGatewayTxnRefWithTx(tx, vnpTxnNo)
			if err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}
			} else if existing.ID != order.ID {
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

			creditKey := fmt.Sprintf("credit_order_%s", order.ID.String())
			err = u.walletUsecase.CreditPendingWithTx(ctx, tx, order.ExpertID, order.GrossAmount, "PAYMENT_ORDER", order.ID, creditKey)
			if err != nil {
				return fmt.Errorf("wallet credit failed: %w", err)
			}

			debitKey := fmt.Sprintf("debit_order_%s", order.ID.String())
			err = u.walletUsecase.DebitPendingWithTx(ctx, tx, order.ExpertID, order.CommissionAmount, "PAYMENT_ORDER", order.ID, debitKey)
			if err != nil {
				return fmt.Errorf("wallet commission debit failed: %w", err)
			}

			// Lưu Outbox event: thông báo Booking Service xác nhận lịch hẹn.
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

		} else {
			order.Status = entity.OrderStatusFailed
			if err := u.repo.UpdateWithTx(tx, order); err != nil {
				return err
			}

			if order.AppointmentID != nil {
				bookingPayload, _ := json.Marshal(map[string]interface{}{
					"appointment_id": order.AppointmentID.String(),
					"order_id":       order.ID.String(),
					"status":         "FAILED",
				})
				bookingOutbox := &entity.OutboxEvent{
					AggregateType: "PAYMENT_ORDER",
					AggregateID:   order.ID,
					EventType:     "booking.appointment.fail",
					Payload:       string(bookingPayload),
					Published:     false,
				}
				if err := u.repo.SaveOutboxEvent(tx, bookingOutbox); err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		return false, err
	}

	return alreadyProcessed, nil
}
