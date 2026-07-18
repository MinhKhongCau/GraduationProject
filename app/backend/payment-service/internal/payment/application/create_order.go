package application

import (
	"context"
	"errors"
	"fmt"
	"payment-service/internal/booking/client"
	"payment-service/internal/domain/entity"
	"payment-service/internal/domain/vo"
	paymentdomain "payment-service/internal/payment/domain"
	"strings"
	"time"

	"github.com/google/uuid"
)

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

	// Má»©c hoa há»“ng lÃ  15% - TÃ­nh toÃ¡n Dá»°A TRÃŠN grossAmount Ä‘Ã£ Ä‘Æ°á»£c xÃ¡c thá»±c
	commission := paymentdomain.CalculateCommission(grossAmount)

	order := &entity.PaymentOrder{
		ID:               uuid.New(),
		PayerID:          payerID,
		ExpertID:         expertID,
		GrossAmount:      grossAmount,
		CommissionRate:   commission.Rate,
		CommissionAmount: commission.Amount,
		NetAmount:        commission.NetAmount,
		Gateway:          "VNPAY",
		Status:           entity.OrderStatusPending,
		AppointmentID:    &parsedApptID,
	}

	if err := u.repo.Create(order); err != nil {
		return nil, "", err
	}

	// VNPay requires Vietnam-local timestamp format.
	// VNPay yÃªu cáº§u Ä‘á»‹nh dáº¡ng thá»i gian theo mÃºi giá» Viá»‡t Nam (GMT+7)
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
