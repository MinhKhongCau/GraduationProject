package application

import (
	"context"
	"errors"
	"payment-service/internal/domain/entity"
	"payment-service/internal/payment/application/readquery"
	paymentdomain "payment-service/internal/payment/domain"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
)

func (r *lifecycleRepo) ListPaymentOrders(_ context.Context, filter PaymentOrderFilter) ([]entity.PaymentOrder, int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	items := make([]entity.PaymentOrder, 0)
	for _, order := range r.orders {
		if order.PayerID != filter.PayerID || (filter.AppointmentID != nil && (order.AppointmentID == nil || *order.AppointmentID != *filter.AppointmentID)) || (filter.Status != nil && order.Status != *filter.Status) || (filter.FulfillmentStatus != "" && order.FulfillmentStatus != filter.FulfillmentStatus) || order.CreatedAt < filter.FromMs || order.CreatedAt >= filter.ToMs {
			continue
		}
		items = append(items, order)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].CreatedAt == items[j].CreatedAt {
			return items[i].ID.String() > items[j].ID.String()
		}
		return items[i].CreatedAt > items[j].CreatedAt
	})
	total := int64(len(items))
	start := filter.Page.Offset()
	if start >= len(items) {
		return []entity.PaymentOrder{}, total, nil
	}
	end := start + filter.Page.Size
	if end > len(items) {
		end = len(items)
	}
	return items[start:end], total, nil
}

func (r *lifecycleRepo) GetPaymentOrder(_ context.Context, orderID uuid.UUID) (*entity.PaymentOrder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	order, ok := r.orders[orderID]
	if !ok {
		return nil, ErrPaymentOrderNotFound
	}
	return &order, nil
}

func TestPaymentOrderReadsAreOwnedPaginatedAndExposeStatus(t *testing.T) {
	now := time.Now().UnixMilli()
	payer, other, appointment := uuid.New(), uuid.New(), uuid.New()
	first := lifecycleOrder(appointment, payer, uuid.New(), entity.OrderStatusSuccess, now+1000)
	first.CreatedAt = now
	first.FulfillmentStatus = paymentdomain.FulfillmentBookingConfirmed
	first.GatewayCaptureStatus = paymentdomain.GatewayCaptureSucceeded
	second := lifecycleOrder(appointment, payer, uuid.New(), entity.OrderStatusFailed, now+2000)
	second.CreatedAt = now - 1
	hidden := lifecycleOrder(appointment, other, uuid.New(), entity.OrderStatusSuccess, now+3000)
	hidden.CreatedAt = now
	usecase := lifecycleUsecase(newLifecycleRepo(first, second, hidden), &fakePaymentGateway{}, &fakeBookingClient{}, time.Now(), time.Minute, time.Second).(ReadUsecase)
	page, err := usecase.ListPaymentOrders(context.Background(), PaymentOrderFilter{PayerID: payer, FromMs: now - 100, ToMs: now + 100, Page: readquery.PageRequest{Page: 0, Size: 1}})
	if err != nil || page.TotalItems != 2 || len(page.Items) != 1 || !page.HasNext || page.Items[0].Status != "SUCCESS" || page.Items[0].FulfillmentStatus != paymentdomain.FulfillmentBookingConfirmed {
		t.Fatalf("unexpected page: %+v err=%v", page, err)
	}
	if _, err := usecase.GetPaymentOrder(context.Background(), other, first.ID, false); !errors.Is(err, ErrPaymentOrderForbidden) {
		t.Fatalf("expected ownership error, got %v", err)
	}
	if got, err := usecase.GetPaymentOrder(context.Background(), payer, first.ID, false); err != nil || got.AmountVND != first.GrossAmount.Int64() {
		t.Fatalf("unexpected detail: %+v %v", got, err)
	}
}

func TestPaymentOrderParsersRejectUnsupportedStatuses(t *testing.T) {
	if _, err := ParsePaymentOrderStatus("MOCK"); !errors.Is(err, ErrInvalidPaymentOrderFilter) {
		t.Fatalf("expected status error, got %v", err)
	}
	if _, err := ParseFulfillmentStatus("UNKNOWN"); !errors.Is(err, ErrInvalidPaymentOrderFilter) {
		t.Fatalf("expected fulfillment error, got %v", err)
	}
}
