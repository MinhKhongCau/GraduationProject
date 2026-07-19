package application

import (
	"context"
	"errors"
	"fmt"
	"log"
	"payment-service/internal/domain/entity"
	paymentdomain "payment-service/internal/payment/domain"

	"github.com/google/uuid"
)

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

	// 2. TrÃ­ch xuáº¥t thÃ´ng tin giao dá»‹ch
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

	vnpAmount, err := paymentdomain.ParseVNPayAmount(params)
	if err != nil {
		return false, err
	}

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

	// 3. Tiáº¿n hÃ nh cáº­p nháº­t Database
	var alreadyProcessed bool
	var appointmentID *uuid.UUID
	err = u.uow.WithinTx(ctx, func(tx Tx) error {
		order, appointmentOrders, err := loadOrderForIPN(ctx, tx, orderID)
		if err != nil {
			return err
		}
		if order.AppointmentID != nil {
			value := *order.AppointmentID
			appointmentID = &value
		}

		if order.Status == entity.OrderStatusSuccess || order.Status == entity.OrderStatusFailed {
			alreadyProcessed = true
			return nil
		}
		if order.Status != entity.OrderStatusPending && order.Status != entity.OrderStatusExpired {
			alreadyProcessed = true
			return nil
		}

		if order.GrossAmount.Int64() != vnpAmount {
			return errors.New("amount mismatch")
		}
		locallyExpired := order.Status == entity.OrderStatusExpired ||
			(order.Status == entity.OrderStatusPending && paymentdomain.IsOrderExpired(u.clock(), order.ExpiresAt))
		if locallyExpired && vnpResponseCode != "00" {
			if order.Status == entity.OrderStatusPending {
				order.Status = entity.OrderStatusExpired
				if err := tx.UpdateOrder(ctx, order); err != nil {
					return err
				}
			}
			alreadyProcessed = true
			return nil
		}

		if vnpResponseCode == "00" && order.AppointmentID != nil {
			for i := range appointmentOrders {
				other := &appointmentOrders[i]
				if other.ID != order.ID && other.Status == entity.OrderStatusSuccess {
					if order.Status == entity.OrderStatusPending {
						order.Status = entity.OrderStatusExpired
						if err := tx.UpdateOrder(ctx, order); err != nil {
							return err
						}
					}
					alreadyProcessed = true
					return nil
				}
			}
		}

		// Replay attack check
		if vnpTxnNo != "" {
			existing, err := tx.GetGatewayTxnRef(ctx, vnpTxnNo)
			if err != nil {
				if !errors.Is(err, ErrTxRecordNotFound) {
					return err
				}
			} else if existing.ID != order.ID {
				return errors.New("replay attack: transaction reference already processed")
			}
		}

		paidAt := u.clock().UnixMilli()
		if vnpResponseCode == "00" {
			if order.AppointmentID != nil {
				if err := tx.ExpireOtherPendingOrders(ctx, *order.AppointmentID, order.ID); err != nil {
					return err
				}
			}
			order.Status = entity.OrderStatusSuccess
			order.PaidAt = &paidAt
			order.GatewayTxnRef = vnpTxnNo

			if err := tx.UpdateOrder(ctx, order); err != nil {
				return err
			}

			creditKey := fmt.Sprintf("credit_order_%s", order.ID.String())
			err = tx.CreditWalletPending(ctx, order.ExpertID, order.GrossAmount, order.ID, creditKey)
			if err != nil {
				return fmt.Errorf("wallet credit failed: %w", err)
			}

			debitKey := fmt.Sprintf("debit_order_%s", order.ID.String())
			err = tx.DebitWalletPending(ctx, order.ExpertID, order.CommissionAmount, order.ID, debitKey)
			if err != nil {
				return fmt.Errorf("wallet commission debit failed: %w", err)
			}

			// LÆ°u Outbox event: thÃ´ng bÃ¡o Booking Service xÃ¡c nháº­n lá»‹ch háº¹n.
			// Chá»‰ emit náº¿u order nÃ y cÃ³ appointment_id.
			if order.AppointmentID != nil {
				bookingPayload, _ := paymentdomain.BookingOutboxPayload(order.AppointmentID.String(), order.ID.String(), "SUCCESS")
				bookingOutbox := &entity.OutboxEvent{
					AggregateType: "PAYMENT_ORDER",
					AggregateID:   order.ID,
					EventType:     "booking.appointment.confirm",
					Payload:       bookingPayload,
					Published:     false,
				}
				if err := tx.SaveOutboxEvent(ctx, bookingOutbox); err != nil {
					return err
				}
			}

		} else {
			order.Status = entity.OrderStatusFailed
			if err := tx.UpdateOrder(ctx, order); err != nil {
				return err
			}

			if order.AppointmentID != nil {
				bookingPayload, _ := paymentdomain.BookingOutboxPayload(order.AppointmentID.String(), order.ID.String(), "FAILED")
				bookingOutbox := &entity.OutboxEvent{
					AggregateType: "PAYMENT_ORDER",
					AggregateID:   order.ID,
					EventType:     "booking.appointment.fail",
					Payload:       bookingPayload,
					Published:     false,
				}
				if err := tx.SaveOutboxEvent(ctx, bookingOutbox); err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, ErrAppointmentAlreadyPaid) && appointmentID != nil {
			winnerExists, reloadErr := u.hasSuccessfulOrder(ctx, *appointmentID)
			if reloadErr != nil {
				return false, reloadErr
			}
			if winnerExists {
				return true, nil
			}
		}
		return false, err
	}

	return alreadyProcessed, nil
}

func (u *paymentUsecase) hasSuccessfulOrder(ctx context.Context, appointmentID uuid.UUID) (bool, error) {
	var found bool
	err := u.uow.WithinTx(ctx, func(tx Tx) error {
		orders, err := tx.ListOrdersForAppointmentForUpdate(ctx, appointmentID)
		if err != nil {
			return err
		}
		for i := range orders {
			if orders[i].Status == entity.OrderStatusSuccess {
				found = true
				break
			}
		}
		return nil
	})
	return found, err
}

func loadOrderForIPN(ctx context.Context, tx Tx, orderID uuid.UUID) (*entity.PaymentOrder, []entity.PaymentOrder, error) {
	order, err := tx.GetOrder(ctx, orderID)
	if err != nil {
		return nil, nil, err
	}
	if order.AppointmentID == nil {
		locked, err := tx.GetOrderForUpdate(ctx, orderID)
		return locked, nil, err
	}

	orders, err := tx.ListOrdersForAppointmentForUpdate(ctx, *order.AppointmentID)
	if err != nil {
		return nil, nil, err
	}
	for i := range orders {
		if orders[i].ID == orderID {
			return &orders[i], orders, nil
		}
	}
	return nil, nil, ErrTxRecordNotFound
}
