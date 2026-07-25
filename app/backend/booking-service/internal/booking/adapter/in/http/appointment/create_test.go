package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	appappointment "booking-service/internal/booking/application/appointment"
	"booking-service/internal/booking/domain"

	"github.com/gin-gonic/gin"
)

type mockBookingUsecase struct {
	createFunc func(patientID, expertID, slotID string) (*domain.Appointment, error)
	cancelFunc func(appointmentID, userID, userRole, reason string) error
}

func (m *mockBookingUsecase) CreateAppointment(patientID, expertID, slotID string) (*domain.Appointment, error) {
	if m.createFunc != nil {
		return m.createFunc(patientID, expertID, slotID)
	}
	return nil, nil
}

func (m *mockBookingUsecase) GetAppointmentByID(appointmentID string) (*domain.Appointment, error) {
	return nil, nil
}

func (m *mockBookingUsecase) GetPaymentEligibility(command appappointment.GetPaymentEligibilityCommand) (appappointment.PaymentEligibility, error) {
	return appappointment.PaymentEligibility{}, nil
}

func (m *mockBookingUsecase) CancelAppointment(appointmentID, userID, userRole, reason string) error {
	if m.cancelFunc != nil {
		return m.cancelFunc(appointmentID, userID, userRole, reason)
	}
	return nil
}

func (m *mockBookingUsecase) ConfirmPayment(appointmentID string) error {
	return nil
}

func (m *mockBookingUsecase) HandlePaymentFailure(appointmentID string) error {
	return nil
}

func (m *mockBookingUsecase) HandlePaymentResult(command appappointment.HandlePaymentResultCommand) error {
	return nil
}

func (m *mockBookingUsecase) GetAppointmentsByPatient(patientID string) ([]domain.Appointment, error) {
	return nil, nil
}

func (m *mockBookingUsecase) GetAppointmentsByExpert(expertID string, fromDate, toDate int64, status *domain.AppointmentStatus) ([]domain.Appointment, error) {
	return nil, nil
}

func TestCreateAppointment(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("TC-BOOK-APT-01 - Tạo cuộc hẹn thành công (Happy Case)", func(t *testing.T) {
		mockUsecase := &mockBookingUsecase{
			createFunc: func(patientID, expertID, slotID string) (*domain.Appointment, error) {
				return &domain.Appointment{
					AppointmentID: "appt-uuid-1",
					SlotID:        slotID,
					PatientID:     patientID,
					ExpertID:      expertID,
					Status:        domain.AppointmentStatusPendingPayment,
				}, nil
			},
		}
		h := NewHandler(mockUsecase)

		body := map[string]string{
			"expert_id": "expert-uuid-1",
			"slot_id":   "slot-uuid-1",
		}
		jsonBytes, _ := json.Marshal(body)

		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest(http.MethodPost, "/booking/appointments", bytes.NewBuffer(jsonBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Request.Header.Set("X-User-Role", "PATIENT")
		c.Request.Header.Set("X-User-Id", "patient-uuid-1")

		h.Create(c)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", recorder.Code, recorder.Body.String())
		}

		var resp map[string]interface{}
		json.Unmarshal(recorder.Body.Bytes(), &resp)
		if resp["success"] != true {
			t.Fatalf("expected success = true, got %v", resp["success"])
		}

		data := resp["data"].(map[string]interface{})
		if data["appointment_id"] != "appt-uuid-1" {
			t.Fatalf("expected appt-uuid-1, got %v", data["appointment_id"])
		}
	})

	t.Run("TC-BOOK-APT-02 - Tạo cuộc hẹn thất bại do slot chưa được giữ chỗ", func(t *testing.T) {
		mockUsecase := &mockBookingUsecase{
			createFunc: func(patientID, expertID, slotID string) (*domain.Appointment, error) {
				return nil, errors.New("slot không được giữ bởi bạn, vui lòng thực hiện lại từ đầu")
			},
		}
		h := NewHandler(mockUsecase)

		body := map[string]string{
			"expert_id": "expert-uuid-1",
			"slot_id":   "slot-uuid-unlocked",
		}
		jsonBytes, _ := json.Marshal(body)

		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest(http.MethodPost, "/booking/appointments", bytes.NewBuffer(jsonBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Request.Header.Set("X-User-Role", "PATIENT")
		c.Request.Header.Set("X-User-Id", "patient-uuid-1")

		h.Create(c)

		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 Internal Server Error, got %d: %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("TC-BOOK-APT-03 - Tạo cuộc hẹn thất bại do phiên giữ chỗ hết hạn (15 phút)", func(t *testing.T) {
		mockUsecase := &mockBookingUsecase{
			createFunc: func(patientID, expertID, slotID string) (*domain.Appointment, error) {
				return nil, errors.New("phiên giữ chỗ đã hết hạn 15 phút, vui lòng chọn lại")
			},
		}
		h := NewHandler(mockUsecase)

		body := map[string]string{
			"expert_id": "expert-uuid-1",
			"slot_id":   "slot-uuid-expired",
		}
		jsonBytes, _ := json.Marshal(body)

		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest(http.MethodPost, "/booking/appointments", bytes.NewBuffer(jsonBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Request.Header.Set("X-User-Role", "PATIENT")
		c.Request.Header.Set("X-User-Id", "patient-uuid-1")

		h.Create(c)

		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 Internal Server Error, got %d: %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("TC-BOOK-APT-04 - Tạo cuộc hẹn thất bại do chuyên gia nghỉ phép", func(t *testing.T) {
		mockUsecase := &mockBookingUsecase{
			createFunc: func(patientID, expertID, slotID string) (*domain.Appointment, error) {
				return nil, errors.New("slot is covered by expert time-off")
			},
		}
		h := NewHandler(mockUsecase)

		body := map[string]string{
			"expert_id": "expert-uuid-1",
			"slot_id":   "slot-uuid-timeoff",
		}
		jsonBytes, _ := json.Marshal(body)

		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest(http.MethodPost, "/booking/appointments", bytes.NewBuffer(jsonBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Request.Header.Set("X-User-Role", "PATIENT")
		c.Request.Header.Set("X-User-Id", "patient-uuid-1")

		h.Create(c)

		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 Internal Server Error, got %d: %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("TC-BOOK-APT-05 - Tạo cuộc hẹn thất bại do vai trò không phải Bệnh nhân", func(t *testing.T) {
		mockUsecase := &mockBookingUsecase{}
		h := NewHandler(mockUsecase)

		body := map[string]string{
			"expert_id": "expert-uuid-1",
			"slot_id":   "slot-uuid-1",
		}
		jsonBytes, _ := json.Marshal(body)

		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest(http.MethodPost, "/booking/appointments", bytes.NewBuffer(jsonBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Request.Header.Set("X-User-Role", "EXPERT")
		c.Request.Header.Set("X-User-Id", "expert-uuid-1")

		h.Create(c)

		if recorder.Code != http.StatusForbidden {
			t.Fatalf("expected 403 Forbidden, got %d: %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("TC-BOOK-APT-06 - Tạo cuộc hẹn thất bại do thiếu request body bắt buộc", func(t *testing.T) {
		mockUsecase := &mockBookingUsecase{}
		h := NewHandler(mockUsecase)

		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest(http.MethodPost, "/booking/appointments", bytes.NewBufferString("{}"))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Request.Header.Set("X-User-Role", "PATIENT")
		c.Request.Header.Set("X-User-Id", "patient-uuid-1")

		h.Create(c)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %d: %s", recorder.Code, recorder.Body.String())
		}
	})
}
