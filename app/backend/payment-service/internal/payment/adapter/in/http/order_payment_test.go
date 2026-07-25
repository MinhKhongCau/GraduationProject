package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"payment-service/internal/domain/entity"
	"payment-service/internal/domain/vo"
	apppayment "payment-service/internal/payment/application"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type mockPaymentOrderUsecase struct {
	createOrderFunc func(ctx context.Context, payerID uuid.UUID, appointmentID string, clientIP string) (*entity.PaymentOrder, string, error)
	processIPNFunc  func(ctx context.Context, queryParams map[string][]string) (bool, error)
}

func (m *mockPaymentOrderUsecase) CreateOrder(ctx context.Context, payerID uuid.UUID, appointmentID string, clientIP string) (*entity.PaymentOrder, string, error) {
	if m.createOrderFunc != nil {
		return m.createOrderFunc(ctx, payerID, appointmentID, clientIP)
	}
	return nil, "", nil
}

func (m *mockPaymentOrderUsecase) ProcessIPN(ctx context.Context, queryParams map[string][]string) (bool, error) {
	if m.processIPNFunc != nil {
		return m.processIPNFunc(ctx, queryParams)
	}
	return false, nil
}

func (m *mockPaymentOrderUsecase) ListCompensationCases(ctx context.Context, filter apppayment.CompensationCaseFilter) (*apppayment.CompensationCasePage, error) {
	return nil, nil
}

func (m *mockPaymentOrderUsecase) GetCompensationCase(ctx context.Context, caseID uuid.UUID) (*apppayment.CompensationCase, error) {
	return nil, nil
}

func TestPaymentOrderAndVNPayIPN(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("TC-PAY-ORD-01 - Tạo đơn hàng thanh toán VNPay thành công (Happy Case)", func(t *testing.T) {
		payerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
		expertID := uuid.New()
		appointmentID := "appt-uuid-1"

		mockUsecase := &mockPaymentOrderUsecase{
			createOrderFunc: func(ctx context.Context, pID uuid.UUID, apptID string, clientIP string) (*entity.PaymentOrder, string, error) {
				order := &entity.PaymentOrder{
					ID:               uuid.New(),
					PayerID:          pID,
					ExpertID:         expertID,
					GrossAmount:      vo.Money(250000),
					NetAmount:        vo.Money(200000),
					CommissionAmount: vo.Money(50000),
					Status:           entity.OrderStatusPending,
					ExpiresAt:        time.Now().Add(15 * time.Minute).UnixMilli(),
				}
				return order, "https://sandbox.vnpayment.vn/paymentv2/vpcpay.html?vnp_TxnRef=123", nil
			},
		}

		h := NewHandler(mockUsecase)
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)

		body := []byte(`{"appointment_id":"` + appointmentID + `"}`)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/payments/orders", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Request.Header.Set("X-User-Id", payerID.String())

		h.CreateOrder(c)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", recorder.Code, recorder.Body.String())
		}

		var resp map[string]interface{}
		json.Unmarshal(recorder.Body.Bytes(), &resp)
		if resp["success"] != true {
			t.Fatalf("expected success = true, got %v", resp["success"])
		}

		data := resp["data"].(map[string]interface{})
		if data["payment_url"] == "" {
			t.Fatal("expected non-empty payment_url")
		}
	})

	t.Run("TC-PAY-ORD-02 - Tạo đơn hàng thất bại do không tìm thấy cuộc hẹn", func(t *testing.T) {
		mockUsecase := &mockPaymentOrderUsecase{
			createOrderFunc: func(ctx context.Context, pID uuid.UUID, apptID string, clientIP string) (*entity.PaymentOrder, string, error) {
				return nil, "", apppayment.ErrBookingAppointmentNotFound
			},
		}

		h := NewHandler(mockUsecase)
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)

		body := []byte(`{"appointment_id":"non-existent-appt"}`)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/payments/orders", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Request.Header.Set("X-User-Id", uuid.New().String())

		h.CreateOrder(c)

		if recorder.Code != http.StatusNotFound {
			t.Fatalf("expected status 404 Not Found, got %d: %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("TC-PAY-ORD-03 - Tạo đơn hàng thất bại do cuộc hẹn không thuộc về người thanh toán", func(t *testing.T) {
		mockUsecase := &mockPaymentOrderUsecase{
			createOrderFunc: func(ctx context.Context, pID uuid.UUID, apptID string, clientIP string) (*entity.PaymentOrder, string, error) {
				return nil, "", apppayment.ErrAppointmentOwnership
			},
		}

		h := NewHandler(mockUsecase)
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)

		body := []byte(`{"appointment_id":"appt-uuid-1"}`)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/payments/orders", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Request.Header.Set("X-User-Id", uuid.New().String())

		h.CreateOrder(c)

		if recorder.Code != http.StatusForbidden {
			t.Fatalf("expected status 403 Forbidden, got %d: %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("TC-PAY-ORD-04 - Tạo đơn hàng thất bại do cuộc hẹn đã được thanh toán", func(t *testing.T) {
		mockUsecase := &mockPaymentOrderUsecase{
			createOrderFunc: func(ctx context.Context, pID uuid.UUID, apptID string, clientIP string) (*entity.PaymentOrder, string, error) {
				return nil, "", apppayment.ErrAppointmentAlreadyPaid
			},
		}

		h := NewHandler(mockUsecase)
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)

		body := []byte(`{"appointment_id":"appt-already-paid"}`)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/payments/orders", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Request.Header.Set("X-User-Id", uuid.New().String())

		h.CreateOrder(c)

		if recorder.Code != http.StatusConflict {
			t.Fatalf("expected status 409 Conflict, got %d: %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("TC-PAY-ORD-05 - Tạo đơn hàng thất bại do thiếu header người dùng", func(t *testing.T) {
		mockUsecase := &mockPaymentOrderUsecase{}
		h := NewHandler(mockUsecase)
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)

		body := []byte(`{"appointment_id":"appt-uuid-1"}`)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/payments/orders", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		// X-User-Id rỗng

		h.CreateOrder(c)

		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401 Unauthorized, got %d: %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("TC-PAY-ORD-06 - Tạo đơn hàng thất bại do thiếu body", func(t *testing.T) {
		mockUsecase := &mockPaymentOrderUsecase{}
		h := NewHandler(mockUsecase)
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)

		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/payments/orders", bytes.NewBufferString("{}"))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Request.Header.Set("X-User-Id", uuid.New().String())

		h.CreateOrder(c)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400 Bad Request, got %d: %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("TC-PAY-IPN-01 - Xử lý VNPay Webhook thanh toán thành công", func(t *testing.T) {
		mockUsecase := &mockPaymentOrderUsecase{
			processIPNFunc: func(ctx context.Context, queryParams map[string][]string) (bool, error) {
				return false, nil
			},
		}

		h := NewHandler(mockUsecase)
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)

		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/payments/vnpay-ipn?vnp_TxnRef=123&vnp_ResponseCode=00&vnp_SecureHash=valid", nil)

		h.HandleVNPayIPN(c)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", recorder.Code, recorder.Body.String())
		}

		var resp map[string]string
		json.Unmarshal(recorder.Body.Bytes(), &resp)
		if resp["RspCode"] != "00" {
			t.Fatalf("expected RspCode 00, got %s", resp["RspCode"])
		}
	})

	t.Run("TC-PAY-IPN-02 - Xử lý VNPay Webhook thất bại do sai chữ ký bảo mật", func(t *testing.T) {
		mockUsecase := &mockPaymentOrderUsecase{
			processIPNFunc: func(ctx context.Context, queryParams map[string][]string) (bool, error) {
				return false, errors.New("checksum mismatch invalid signature")
			},
		}

		h := NewHandler(mockUsecase)
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)

		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/payments/vnpay-ipn?vnp_SecureHash=invalid", nil)

		h.HandleVNPayIPN(c)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", recorder.Code, recorder.Body.String())
		}

		var resp map[string]string
		json.Unmarshal(recorder.Body.Bytes(), &resp)
		if resp["RspCode"] != "97" {
			t.Fatalf("expected RspCode 97 (Invalid Signature), got %s", resp["RspCode"])
		}
	})

	t.Run("TC-PAY-IPN-03 - Xử lý VNPay Webhook thất bại do đơn hàng không tồn tại", func(t *testing.T) {
		mockUsecase := &mockPaymentOrderUsecase{
			processIPNFunc: func(ctx context.Context, queryParams map[string][]string) (bool, error) {
				return false, errors.New("order not found in database")
			},
		}

		h := NewHandler(mockUsecase)
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)

		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/payments/vnpay-ipn?vnp_TxnRef=unknown", nil)

		h.HandleVNPayIPN(c)

		var resp map[string]string
		json.Unmarshal(recorder.Body.Bytes(), &resp)
		if resp["RspCode"] != "01" {
			t.Fatalf("expected RspCode 01 (Order not found), got %s", resp["RspCode"])
		}
	})

	t.Run("TC-PAY-IPN-04 - Xử lý VNPay Webhook khi đơn hàng đã được xác nhận trước đó", func(t *testing.T) {
		mockUsecase := &mockPaymentOrderUsecase{
			processIPNFunc: func(ctx context.Context, queryParams map[string][]string) (bool, error) {
				return true, nil // alreadyProcessed = true
			},
		}

		h := NewHandler(mockUsecase)
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)

		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/payments/vnpay-ipn?vnp_TxnRef=123", nil)

		h.HandleVNPayIPN(c)

		var resp map[string]string
		json.Unmarshal(recorder.Body.Bytes(), &resp)
		if resp["RspCode"] != "02" {
			t.Fatalf("expected RspCode 02 (Order already confirmed), got %s", resp["RspCode"])
		}
	})
}
