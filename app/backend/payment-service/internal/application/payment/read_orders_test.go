package payment

import (
	"context"
	"errors"
	"payment-service/internal/application/readquery"
	paymentdomain "payment-service/internal/domain/payment"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
)

func (r *lifecycleRepo) ListPaymentOrders(_ context.Context, filter PaymentOrderFilter) ([]paymentdomain.PaymentOrder, int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	items := make([]paymentdomain.PaymentOrder, 0)
	for _, order := range r.orders {
		if !matchesOrderFilter(order, filter) {
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
		return []paymentdomain.PaymentOrder{}, total, nil
	}
	end := start + filter.Page.Size
	if end > len(items) {
		end = len(items)
	}
	return items[start:end], total, nil
}

func matchesOrderFilter(order paymentdomain.PaymentOrder, filter PaymentOrderFilter) bool {
	switch {
	case filter.PayerID != uuid.Nil && order.PayerID != filter.PayerID,
		filter.ExpertID != uuid.Nil && order.ExpertID != filter.ExpertID,
		filter.ScopeExperts && !containsUUID(filter.ExpertIDs, order.ExpertID),
		filter.AppointmentID != nil && (order.AppointmentID == nil || *order.AppointmentID != *filter.AppointmentID),
		filter.Type == PaymentOrderTypeAppointment && order.AppointmentID == nil,
		filter.Type == PaymentOrderTypeTopUp && order.AppointmentID != nil,
		filter.Status != nil && order.Status != *filter.Status,
		filter.FulfillmentStatus != "" && order.FulfillmentStatus != filter.FulfillmentStatus,
		order.CreatedAt < filter.FromMs || order.CreatedAt >= filter.ToMs:
		return false
	}
	return true
}

func containsUUID(ids []uuid.UUID, id uuid.UUID) bool {
	for _, candidate := range ids {
		if candidate == id {
			return true
		}
	}
	return false
}

func (r *lifecycleRepo) SummarizePaymentOrders(_ context.Context, filter PaymentOrderFilter) (PaymentOrderTotals, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var totals PaymentOrderTotals
	for _, order := range r.orders {
		if !matchesOrderFilter(order, filter) {
			continue
		}
		totals.TotalOrders++
		switch order.Status {
		case paymentdomain.OrderStatusPending:
			totals.PendingOrders++
		case paymentdomain.OrderStatusSuccess:
			totals.SuccessOrders++
			totals.GrossAmount += order.GrossAmount.Int64()
			totals.CommissionAmount += order.CommissionAmount.Int64()
			totals.NetAmount += order.NetAmount.Int64()
		case paymentdomain.OrderStatusFailed:
			totals.FailedOrders++
		case paymentdomain.OrderStatusExpired:
			totals.ExpiredOrders++
		}
	}
	return totals, nil
}

func (r *lifecycleRepo) ListManagedExpertIDs(_ context.Context, adminID uuid.UUID) ([]uuid.UUID, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.managed[adminID], nil
}

func (r *lifecycleRepo) ApplyOrderReview(_ context.Context, orderID uuid.UUID, apply func(order *paymentdomain.PaymentOrder) (*paymentdomain.PaymentCompensationCase, error)) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	order, ok := r.orders[orderID]
	if !ok {
		return ErrPaymentOrderNotFound
	}
	compensationCase, err := apply(&order)
	if err != nil {
		return err
	}
	for _, existing := range r.compensationCases {
		if existing.PaymentOrderID == compensationCase.PaymentOrderID && existing.ReasonCode == compensationCase.ReasonCode {
			return ErrCompensationCaseExists
		}
	}
	r.orders[orderID] = order
	r.compensationCases = append(r.compensationCases, *compensationCase)
	return nil
}

func (r *lifecycleRepo) UpdateCompensationCase(_ context.Context, caseID uuid.UUID, apply func(compensationCase *paymentdomain.PaymentCompensationCase) error) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.compensationCases {
		if r.compensationCases[i].ID == caseID {
			updated := r.compensationCases[i]
			if err := apply(&updated); err != nil {
				return err
			}
			r.compensationCases[i] = updated
			return nil
		}
	}
	return ErrCompensationCaseNotFound
}

func (r *lifecycleRepo) GetPaymentOrder(_ context.Context, orderID uuid.UUID) (*paymentdomain.PaymentOrder, error) {
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
	first := lifecycleOrder(appointment, payer, uuid.New(), paymentdomain.OrderStatusSuccess, now+1000)
	first.CreatedAt = now
	first.FulfillmentStatus = paymentdomain.FulfillmentBookingConfirmed
	first.GatewayCaptureStatus = paymentdomain.GatewayCaptureSucceeded
	second := lifecycleOrder(appointment, payer, uuid.New(), paymentdomain.OrderStatusFailed, now+2000)
	second.CreatedAt = now - 1
	hidden := lifecycleOrder(appointment, other, uuid.New(), paymentdomain.OrderStatusSuccess, now+3000)
	hidden.CreatedAt = now
	usecase := lifecycleUsecase(newLifecycleRepo(first, second, hidden), &fakePaymentGateway{}, &fakeBookingClient{}, time.Now(), time.Minute, time.Second).(ReadUsecase)
	page, err := usecase.ListPaymentOrders(context.Background(), PaymentOrderFilter{PayerID: payer, FromMs: now - 100, ToMs: now + 100, Page: readquery.PageRequest{Page: 0, Size: 1}})
	if err != nil || page.TotalItems != 2 || len(page.Items) != 1 || !page.HasNext || page.Items[0].Status != "SUCCESS" || page.Items[0].FulfillmentStatus != paymentdomain.FulfillmentBookingConfirmed {
		t.Fatalf("unexpected page: %+v err=%v", page, err)
	}
	if _, err := usecase.GetPaymentOrder(context.Background(), other, first.ID); !errors.Is(err, ErrPaymentOrderForbidden) {
		t.Fatalf("expected ownership error, got %v", err)
	}
	if got, err := usecase.GetPaymentOrder(context.Background(), payer, first.ID); err != nil || got.AmountVND != first.GrossAmount.Int64() {
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
