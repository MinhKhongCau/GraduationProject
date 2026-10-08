package payment

import (
	"context"
	"errors"
	"payment-service/internal/application/managedscope"
	"payment-service/internal/application/readquery"
	paymentdomain "payment-service/internal/domain/payment"
	"testing"
	"time"

	"github.com/google/uuid"
)

type manageFixture struct {
	repo                      *lifecycleRepo
	usecase                   ManageUsecase
	reader                    ReadUsecase
	adminA, adminB            uuid.UUID
	expertA, expertB, patient uuid.UUID
	paidA, pendingA, paidB    paymentdomain.PaymentOrder
	filter                    PaymentOrderFilter
}

// newManageFixture: adminA quản lý expertA (2 đơn), adminB quản lý expertB (1 đơn).
func newManageFixture(t *testing.T) manageFixture {
	t.Helper()
	now := time.Now()
	f := manageFixture{adminA: uuid.New(), adminB: uuid.New(), expertA: uuid.New(), expertB: uuid.New(), patient: uuid.New()}
	f.paidA = lifecycleOrder(uuid.New(), f.patient, f.expertA, paymentdomain.OrderStatusSuccess, now.UnixMilli()+1000)
	f.pendingA = lifecycleOrder(uuid.New(), f.patient, f.expertA, paymentdomain.OrderStatusPending, now.UnixMilli()+2000)
	f.paidB = lifecycleOrder(uuid.New(), f.patient, f.expertB, paymentdomain.OrderStatusSuccess, now.UnixMilli()+3000)
	f.paidB.GrossAmount, f.paidB.CommissionAmount, f.paidB.NetAmount = 2000, 300, 1700
	f.repo = newLifecycleRepo(f.paidA, f.pendingA, f.paidB)
	f.repo.managed = map[uuid.UUID][]uuid.UUID{f.adminA: {f.expertA}, f.adminB: {f.expertB}}
	usecase := lifecycleUsecase(f.repo, &fakePaymentGateway{}, &fakeBookingClient{}, now, time.Minute, time.Second)
	f.usecase = usecase.(ManageUsecase)
	f.reader = usecase.(ReadUsecase)
	f.filter = PaymentOrderFilter{FromMs: now.Add(-time.Hour).UnixMilli(), ToMs: now.Add(time.Hour).UnixMilli(), Page: readquery.PageRequest{Size: 20}}
	return f
}

func TestPatientSummaryCountsOnlySuccessfulPayments(t *testing.T) {
	f := newManageFixture(t)
	filter := f.filter
	filter.PayerID = f.patient
	summary, err := f.reader.SummarizePatientOrders(context.Background(), filter)
	if err != nil {
		t.Fatal(err)
	}
	if summary.TotalOrders != 3 || summary.SuccessOrders != 2 || summary.TotalPaid != 3000 {
		t.Fatalf("unexpected patient summary: %+v", summary)
	}
	if _, err := f.reader.SummarizePatientOrders(context.Background(), f.filter); !errors.Is(err, ErrPaymentOrderForbidden) {
		t.Fatalf("expected forbidden without payer, got %v", err)
	}
}

func TestExpertSeesOnlyOwnOrdersWithCommissionAndNet(t *testing.T) {
	f := newManageFixture(t)
	filter := f.filter
	filter.PayerID = uuid.New() // phạm vi từ request bị bỏ qua
	page, err := f.usecase.ListExpertOrders(context.Background(), f.expertA, filter)
	if err != nil {
		t.Fatal(err)
	}
	if page.TotalItems != 2 {
		t.Fatalf("expected 2 orders for expert A, got %+v", page)
	}
	summary, err := f.usecase.SummarizeExpertOrders(context.Background(), f.expertA, f.filter)
	if err != nil {
		t.Fatal(err)
	}
	if summary.GrossTotal != 1000 || summary.CommissionTotal != 150 || summary.NetTotal != 850 || summary.TotalOrders != 2 || summary.SuccessOrders != 1 {
		t.Fatalf("unexpected expert summary: %+v", summary)
	}
	if _, err := f.usecase.GetExpertOrder(context.Background(), f.expertA, f.paidB.ID); !errors.Is(err, ErrPaymentOrderForbidden) {
		t.Fatalf("expert must not read another expert's order, got %v", err)
	}
	if view, err := f.usecase.GetExpertOrder(context.Background(), f.expertA, f.paidA.ID); err != nil || view.NetAmount != 850 || view.Type != PaymentOrderTypeAppointment {
		t.Fatalf("unexpected expert order: %+v err=%v", view, err)
	}
}

func TestAdminSeesOnlyManagedExperts(t *testing.T) {
	f := newManageFixture(t)
	page, err := f.usecase.ListAdminOrders(context.Background(), f.adminA, f.filter)
	if err != nil {
		t.Fatal(err)
	}
	if page.TotalItems != 2 {
		t.Fatalf("admin A should see only expert A orders, got %d", page.TotalItems)
	}
	for _, item := range page.Items {
		if item.ExpertID != f.expertA {
			t.Fatalf("leaked order of unmanaged expert: %+v", item)
		}
	}

	summary, err := f.usecase.SummarizeAdminOrders(context.Background(), f.adminB, f.filter)
	if err != nil {
		t.Fatal(err)
	}
	if summary.TotalOrders != 1 || summary.GrossTotal != 2000 || summary.CommissionTotal != 300 || summary.NetTotal != 1700 || summary.ManagedExperts != 1 {
		t.Fatalf("unexpected admin summary: %+v", summary)
	}

	filter := f.filter
	filter.ExpertID = f.expertB
	if _, err := f.usecase.ListAdminOrders(context.Background(), f.adminA, filter); !errors.Is(err, managedscope.ErrNotManagedExpert) {
		t.Fatalf("expected ErrNotManagedExpert for unmanaged expert filter, got %v", err)
	}
	if _, err := f.usecase.GetAdminOrder(context.Background(), f.adminA, f.paidB.ID); !errors.Is(err, managedscope.ErrNotManagedExpert) {
		t.Fatalf("expected ErrNotManagedExpert for unmanaged order, got %v", err)
	}
	if view, err := f.usecase.GetAdminOrder(context.Background(), f.adminA, f.paidA.ID); err != nil || view.CommissionAmount != 150 || view.Gateway != "VNPAY" {
		t.Fatalf("unexpected admin order view: %+v err=%v", view, err)
	}
}

func TestAdminWithoutManagedExpertsGetsEmptyResults(t *testing.T) {
	f := newManageFixture(t)
	stranger := uuid.New()
	page, err := f.usecase.ListAdminOrders(context.Background(), stranger, f.filter)
	if err != nil || page.TotalItems != 0 || len(page.Items) != 0 {
		t.Fatalf("expected empty page, got %+v err=%v", page, err)
	}
	summary, err := f.usecase.SummarizeAdminOrders(context.Background(), stranger, f.filter)
	if err != nil || summary.TotalOrders != 0 {
		t.Fatalf("expected empty summary, got %+v err=%v", summary, err)
	}
}

func TestAdminReviewCreatesCaseAndResolveClosesIt(t *testing.T) {
	f := newManageFixture(t)
	ctx := context.Background()

	if _, err := f.usecase.ReviewOrder(ctx, f.adminB, f.paidA.ID, paymentdomain.CompensationRefundRequired, ""); !errors.Is(err, managedscope.ErrNotManagedExpert) {
		t.Fatalf("unmanaged admin must not review, got %v", err)
	}
	if _, err := f.usecase.ReviewOrder(ctx, f.adminA, f.pendingA.ID, paymentdomain.CompensationRefundRequired, ""); !errors.Is(err, paymentdomain.ErrOrderNotReviewable) {
		t.Fatalf("pending order must not be reviewable, got %v", err)
	}
	if _, err := f.usecase.ReviewOrder(ctx, f.adminA, f.paidA.ID, "APPROVE", ""); !errors.Is(err, paymentdomain.ErrInvalidAdminReviewAction) {
		t.Fatalf("expected invalid action, got %v", err)
	}

	created, err := f.usecase.ReviewOrder(ctx, f.adminA, f.paidA.ID, paymentdomain.CompensationRefundRequired, "patient complaint")
	if err != nil {
		t.Fatal(err)
	}
	if created.Status != paymentdomain.CompensationRefundRequired || created.Type != paymentdomain.CompensationAdminReview || created.SafeReason != "patient complaint" || created.ExpertID != f.expertA {
		t.Fatalf("unexpected case: %+v", created)
	}
	if f.repo.orders[f.paidA.ID].FulfillmentStatus != paymentdomain.FulfillmentRefundRequired {
		t.Fatalf("order fulfillment not updated: %s", f.repo.orders[f.paidA.ID].FulfillmentStatus)
	}
	if _, err := f.usecase.ReviewOrder(ctx, f.adminA, f.paidA.ID, paymentdomain.CompensationRefundRequired, ""); !errors.Is(err, ErrCompensationCaseExists) {
		t.Fatalf("duplicate review must conflict, got %v", err)
	}

	if _, err := f.usecase.ResolveCompensationCase(ctx, f.adminB, created.ID, "done"); !errors.Is(err, managedscope.ErrNotManagedExpert) {
		t.Fatalf("unmanaged admin must not resolve, got %v", err)
	}
	resolved, err := f.usecase.ResolveCompensationCase(ctx, f.adminA, created.ID, "refunded manually")
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Status != paymentdomain.CompensationResolved || resolved.ResolvedAt == nil || resolved.ResolutionNote == nil || *resolved.ResolutionNote != "refunded manually" {
		t.Fatalf("unexpected resolved case: %+v", resolved)
	}
	if _, err := f.usecase.ResolveCompensationCase(ctx, f.adminA, created.ID, ""); !errors.Is(err, paymentdomain.ErrCompensationAlreadyClosed) {
		t.Fatalf("resolving twice must fail, got %v", err)
	}
}
