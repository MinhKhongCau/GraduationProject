package payment

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"payment-service/internal/application/managedscope"
	apppayment "payment-service/internal/application/payment"
	"payment-service/internal/application/readquery"
	paymentdomain "payment-service/internal/domain/payment"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// fakeManageUsecase ghi lại tham số handler truyền xuống để kiểm tra phân quyền và bộ lọc.
type fakeManageUsecase struct {
	fakeCreateOrderUsecase
	actorID      uuid.UUID
	filter       apppayment.PaymentOrderFilter
	reviewAction paymentdomain.CompensationStatus
	err          error
}

func (f *fakeManageUsecase) ListExpertOrders(_ context.Context, expertID uuid.UUID, filter apppayment.PaymentOrderFilter) (*readquery.Page[apppayment.ExpertPaymentOrderView], error) {
	f.actorID, f.filter = expertID, filter
	page := readquery.NewPage([]apppayment.ExpertPaymentOrderView{}, filter.Page, 0)
	return &page, f.err
}
func (f *fakeManageUsecase) GetExpertOrder(_ context.Context, expertID, _ uuid.UUID) (*apppayment.ExpertPaymentOrderView, error) {
	f.actorID = expertID
	return &apppayment.ExpertPaymentOrderView{}, f.err
}
func (f *fakeManageUsecase) SummarizeExpertOrders(_ context.Context, expertID uuid.UUID, filter apppayment.PaymentOrderFilter) (*apppayment.ExpertOrderSummary, error) {
	f.actorID, f.filter = expertID, filter
	return &apppayment.ExpertOrderSummary{}, f.err
}
func (f *fakeManageUsecase) ListAdminOrders(_ context.Context, adminID uuid.UUID, filter apppayment.PaymentOrderFilter) (*readquery.Page[apppayment.AdminPaymentOrderView], error) {
	f.actorID, f.filter = adminID, filter
	page := readquery.NewPage([]apppayment.AdminPaymentOrderView{}, filter.Page, 0)
	return &page, f.err
}
func (f *fakeManageUsecase) GetAdminOrder(_ context.Context, adminID, _ uuid.UUID) (*apppayment.AdminPaymentOrderView, error) {
	f.actorID = adminID
	return &apppayment.AdminPaymentOrderView{}, f.err
}
func (f *fakeManageUsecase) SummarizeAdminOrders(_ context.Context, adminID uuid.UUID, filter apppayment.PaymentOrderFilter) (*apppayment.AdminOrderSummary, error) {
	f.actorID, f.filter = adminID, filter
	return &apppayment.AdminOrderSummary{}, f.err
}
func (f *fakeManageUsecase) ReviewOrder(_ context.Context, adminID, _ uuid.UUID, action paymentdomain.CompensationStatus, _ string) (*apppayment.CompensationCase, error) {
	f.actorID, f.reviewAction = adminID, action
	return &apppayment.CompensationCase{}, f.err
}
func (f *fakeManageUsecase) ResolveCompensationCase(_ context.Context, adminID, _ uuid.UUID, _ string) (*apppayment.CompensationCase, error) {
	f.actorID = adminID
	return &apppayment.CompensationCase{}, f.err
}

func serveManage(usecase *fakeManageUsecase, method, path, role, userID, body string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewHandler(usecase)
	router.GET("/expert/orders", handler.ListExpertOrders)
	router.GET("/expert/orders/summary", handler.SummarizeExpertOrders)
	router.GET("/admin/orders", handler.ListAdminOrders)
	router.GET("/admin/orders/summary", handler.SummarizeAdminOrders)
	router.GET("/admin/orders/:id", handler.GetAdminOrder)
	router.POST("/admin/orders/:id/review", handler.ReviewOrder)
	router.POST("/compensation-cases/:id/resolve", handler.ResolveCompensationCase)
	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-User-Role", role)
	request.Header.Set("X-User-Id", userID)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func TestManageEndpointsEnforceRoles(t *testing.T) {
	for _, tc := range []struct {
		method, path, role string
	}{
		{http.MethodGet, "/expert/orders", "PATIENT"},
		{http.MethodGet, "/expert/orders/summary", "ADMIN"},
		{http.MethodGet, "/admin/orders", "EXPERT"},
		{http.MethodGet, "/admin/orders/summary", "PATIENT"},
		{http.MethodGet, "/admin/orders/" + uuid.NewString(), "EXPERT"},
		{http.MethodPost, "/admin/orders/" + uuid.NewString() + "/review", "EXPERT"},
		{http.MethodPost, "/compensation-cases/" + uuid.NewString() + "/resolve", "PATIENT"},
	} {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			usecase := &fakeManageUsecase{}
			recorder := serveManage(usecase, tc.method, tc.path, tc.role, uuid.NewString(), `{"action":"MANUAL_REVIEW"}`)
			if recorder.Code != http.StatusForbidden || usecase.actorID != uuid.Nil {
				t.Fatalf("expected 403 without calling usecase, got %d", recorder.Code)
			}
		})
	}
}

func TestAdminOrdersPassesExpertAndPayerFilters(t *testing.T) {
	usecase := &fakeManageUsecase{}
	adminID, expertID, payerID := uuid.New(), uuid.New(), uuid.New()
	recorder := serveManage(usecase, http.MethodGet, "/admin/orders?expert_id="+expertID.String()+"&payer_id="+payerID.String()+"&status=SUCCESS&type=APPOINTMENT", "ADMIN", adminID.String(), "")
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	f := usecase.filter
	if usecase.actorID != adminID || f.ExpertID != expertID || f.PayerID != payerID || f.Status == nil || *f.Status != paymentdomain.OrderStatusSuccess || f.Type != apppayment.PaymentOrderTypeAppointment {
		t.Fatalf("filter not forwarded: actor=%s filter=%+v", usecase.actorID, f)
	}

	if recorder := serveManage(usecase, http.MethodGet, "/admin/orders?expert_id=bad", "ADMIN", adminID.String(), ""); recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid expert_id, got %d", recorder.Code)
	}
	if recorder := serveManage(usecase, http.MethodGet, "/admin/orders?type=GIFT", "ADMIN", adminID.String(), ""); recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid type, got %d", recorder.Code)
	}
}

func TestAdminOrdersMapsUnmanagedExpertToForbidden(t *testing.T) {
	usecase := &fakeManageUsecase{err: managedscope.ErrNotManagedExpert}
	recorder := serveManage(usecase, http.MethodGet, "/admin/orders/"+uuid.NewString(), "ADMIN", uuid.NewString(), "")
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", recorder.Code)
	}
	usecase = &fakeManageUsecase{err: managedscope.ErrResolverUnavailable}
	recorder = serveManage(usecase, http.MethodGet, "/admin/orders", "ADMIN", uuid.NewString(), "")
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", recorder.Code)
	}
}

func TestReviewOrderNormalizesAction(t *testing.T) {
	usecase := &fakeManageUsecase{}
	recorder := serveManage(usecase, http.MethodPost, "/admin/orders/"+uuid.NewString()+"/review", "ADMIN", uuid.NewString(), `{"action":" refund_required ","note":"x"}`)
	if recorder.Code != http.StatusOK || usecase.reviewAction != paymentdomain.CompensationRefundRequired {
		t.Fatalf("unexpected review result: %d action=%s", recorder.Code, usecase.reviewAction)
	}
	if recorder := serveManage(&fakeManageUsecase{}, http.MethodPost, "/admin/orders/"+uuid.NewString()+"/review", "ADMIN", uuid.NewString(), `{}`); recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 without action, got %d", recorder.Code)
	}
}
