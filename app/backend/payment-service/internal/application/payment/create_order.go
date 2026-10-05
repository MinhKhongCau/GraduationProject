package payment

import (
	"context"
	"errors"
	"fmt"
	"payment-service/internal/domain/money"
	paymentdomain "payment-service/internal/domain/payment"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (u *paymentUsecase) CreateOrder(ctx context.Context, payerID uuid.UUID, appointmentID string, ipAddr string) (*paymentdomain.PaymentOrder, string, error) {
	appointmentIDValue := strings.TrimSpace(appointmentID)
	if appointmentIDValue == "" {
		return nil, "", fmt.Errorf("%w: appointment_id is required", ErrInvalidCreateOrderRequest)
	}

	parsedApptID, err := uuid.Parse(appointmentIDValue)
	if err != nil {
		return nil, "", fmt.Errorf("%w: invalid appointment_id format: %w", ErrInvalidCreateOrderRequest, err)
	}

	if _, err := u.repo.GetSuccessfulByAppointment(ctx, parsedApptID); err == nil {
		return nil, "", ErrAppointmentAlreadyPaid
	} else if !errors.Is(err, ErrTxRecordNotFound) {
		return nil, "", fmt.Errorf("check successful payment order: %w", err)
	}

	eligibility, err := u.bookingClient.GetPaymentEligibility(ctx, appointmentIDValue, payerID)
	if err != nil {
		return nil, "", mapPaymentEligibilityError(err, appointmentIDValue)
	}

	expertID, grossAmount, err := validatePaymentEligibility(eligibility, appointmentIDValue)
	if err != nil {
		return nil, "", err
	}

	now := u.clock()
	if eligibility.ExpiresAt <= now.UnixMilli() {
		return nil, "", fmt.Errorf("%w: eligibility has expired", ErrAppointmentInvalidState)
	}
	expiresAt := paymentdomain.PaymentExpiry(now, u.orderTTL, eligibility.ExpiresAt)
	hasUsableWindow := paymentdomain.HasUsablePaymentWindow(now, expiresAt, u.minimumWindow)

	commission := paymentdomain.CalculateCommission(grossAmount)
	newOrder := &paymentdomain.PaymentOrder{
		PayerID:              payerID,
		ExpertID:             expertID,
		GrossAmount:          grossAmount,
		CommissionRate:       commission.Rate,
		CommissionAmount:     commission.Amount,
		NetAmount:            commission.NetAmount,
		Gateway:              "VNPAY",
		Status:               paymentdomain.OrderStatusPending,
		GatewayCaptureStatus: paymentdomain.GatewayCapturePending,
		FulfillmentStatus:    paymentdomain.FulfillmentPending,
		AppointmentID:        &parsedApptID,
		CreatedAt:            now.UnixMilli(),
		ExpiresAt:            expiresAt,
	}

	selectedOrder, err := u.createOrReuseOrder(ctx, newOrder, now.UnixMilli(), hasUsableWindow)
	if errors.Is(err, ErrActivePendingOrderExists) {
		selectedOrder, err = u.reloadConcurrentWinner(ctx, newOrder, now.UnixMilli())
	}
	if err != nil {
		return nil, "", err
	}

	desc := fmt.Sprintf("Thanh toan MindCare don hang %s", selectedOrder.ID.String())
	paymentURL := u.vnpayClient.GeneratePaymentURL(
		selectedOrder.ID.String(),
		selectedOrder.GrossAmount.Int64(),
		ipAddr,
		desc,
		selectedOrder.CreatedAt,
		selectedOrder.ExpiresAt,
	)
	return selectedOrder, paymentURL, nil
}

func (u *paymentUsecase) createOrReuseOrder(ctx context.Context, newOrder *paymentdomain.PaymentOrder, nowMs int64, hasUsableWindow bool) (*paymentdomain.PaymentOrder, error) {
	var selected *paymentdomain.PaymentOrder
	var decisionErr error
	err := u.uow.WithinTx(ctx, func(tx Tx) error {
		orders, err := tx.ListOrdersForAppointmentForUpdate(ctx, *newOrder.AppointmentID)
		if err != nil {
			return err
		}

		var activePending *paymentdomain.PaymentOrder
		for i := range orders {
			order := &orders[i]
			switch order.Status {
			case paymentdomain.OrderStatusSuccess:
				return ErrAppointmentAlreadyPaid
			case paymentdomain.OrderStatusPending:
				if paymentdomain.IsOrderExpired(timeFromMillis(nowMs), order.ExpiresAt) {
					order.Status = paymentdomain.OrderStatusExpired
					if err := tx.UpdateOrder(ctx, order); err != nil {
						return err
					}
					continue
				}
				if activePending != nil {
					return fmt.Errorf("%w: multiple active pending orders", ErrExistingOrderConflict)
				}
				activePending = order
			case paymentdomain.OrderStatusFailed, paymentdomain.OrderStatusExpired:
				// Historical retryable attempts do not block a replacement.
			default:
				return fmt.Errorf("%w: unsupported order status %d", ErrExistingOrderConflict, order.Status)
			}
		}

		if activePending != nil {
			if err := validateReusableOrder(activePending, newOrder); err != nil {
				return err
			}
			copy := *activePending
			selected = &copy
			return nil
		}
		if !hasUsableWindow {
			decisionErr = fmt.Errorf("%w: booking lock expires too soon", ErrPaymentWindowTooShort)
			return nil
		}

		newOrder.ID = uuid.New()
		if err := tx.CreateOrder(ctx, newOrder); err != nil {
			return err
		}
		copy := *newOrder
		selected = &copy
		return nil
	})
	if err != nil {
		return nil, err
	}
	return selected, decisionErr
}

func (u *paymentUsecase) reloadConcurrentWinner(ctx context.Context, expected *paymentdomain.PaymentOrder, nowMs int64) (*paymentdomain.PaymentOrder, error) {
	var winner *paymentdomain.PaymentOrder
	err := u.uow.WithinTx(ctx, func(tx Tx) error {
		orders, err := tx.ListOrdersForAppointmentForUpdate(ctx, *expected.AppointmentID)
		if err != nil {
			return err
		}
		for i := range orders {
			order := &orders[i]
			if order.Status == paymentdomain.OrderStatusSuccess {
				return ErrAppointmentAlreadyPaid
			}
			if order.Status == paymentdomain.OrderStatusPending && order.ExpiresAt > nowMs {
				if err := validateReusableOrder(order, expected); err != nil {
					return err
				}
				copy := *order
				winner = &copy
				return nil
			}
		}
		return ErrActivePendingOrderExists
	})
	return winner, err
}

func validateReusableOrder(existing, expected *paymentdomain.PaymentOrder) error {
	if existing.AppointmentID == nil || expected.AppointmentID == nil ||
		*existing.AppointmentID != *expected.AppointmentID ||
		existing.PayerID != expected.PayerID ||
		existing.ExpertID != expected.ExpertID ||
		existing.GrossAmount != expected.GrossAmount ||
		existing.Gateway != "VNPAY" || existing.ExpiresAt <= 0 {
		return ErrExistingOrderConflict
	}
	return nil
}

func validatePaymentEligibility(eligibility *PaymentEligibility, appointmentID string) (uuid.UUID, money.Money, error) {
	if eligibility == nil {
		return uuid.Nil, 0, fmt.Errorf("%w: missing eligibility response", ErrInvalidBookingData)
	}
	if !strings.EqualFold(strings.TrimSpace(eligibility.AppointmentID), appointmentID) {
		return uuid.Nil, 0, fmt.Errorf("%w: eligibility appointment_id mismatch", ErrInvalidBookingData)
	}
	expertID, err := uuid.Parse(strings.TrimSpace(eligibility.ExpertID))
	if err != nil {
		return uuid.Nil, 0, fmt.Errorf("%w: invalid eligibility expert_id: %w", ErrInvalidBookingData, err)
	}
	if expertID == uuid.Nil {
		return uuid.Nil, 0, fmt.Errorf("%w: missing eligibility expert_id", ErrInvalidBookingData)
	}
	if eligibility.AmountVND <= 0 {
		return uuid.Nil, 0, fmt.Errorf("%w: eligibility amount_vnd must be greater than zero", ErrInvalidBookingData)
	}
	return expertID, money.Money(eligibility.AmountVND), nil
}

func mapPaymentEligibilityError(err error, appointmentID string) error {
	switch {
	case errors.Is(err, ErrAppointmentNotFound):
		return fmt.Errorf("%w: %s", ErrBookingAppointmentNotFound, appointmentID)
	case errors.Is(err, ErrPaymentEligibilityForbidden):
		return fmt.Errorf("%w: appointment %s", ErrAppointmentOwnership, appointmentID)
	case errors.Is(err, ErrPaymentEligibilityConflict):
		return fmt.Errorf("%w: appointment %s", ErrAppointmentInvalidState, appointmentID)
	case errors.Is(err, ErrInvalidBookingPrice):
		return fmt.Errorf("%w: appointment price is invalid", ErrInvalidBookingData)
	default:
		return fmt.Errorf("failed to verify payment eligibility: %w", err)
	}
}

func timeFromMillis(value int64) time.Time {
	return time.UnixMilli(value)
}
