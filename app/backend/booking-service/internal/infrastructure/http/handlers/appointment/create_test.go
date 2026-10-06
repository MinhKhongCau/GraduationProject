package handler

import (
	appointmentdomain "booking-service/internal/domain/appointment"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	appappointment "booking-service/internal/application/appointment"
	bookingquery "booking-service/internal/application/query"

	"github.com/gin-gonic/gin"
)

type mockBookingUsecase struct {
	createFunc func(cmd appappointment.CreateAppointmentCommand) (*appointmentdomain.Appointment, error)
	cancelFunc func(appointmentID, userID, userRole, reason string) error
}

func (m *mockBookingUsecase) CreateAppointment(_ context.Context, cmd appappointment.CreateAppointmentCommand) (*appointmentdomain.Appointment, error) {
	if m.createFunc != nil {
		return m.createFunc(cmd)
	}
	return nil, nil
}

func (m *mockBookingUsecase) GetAppointmentByID(appointmentID string) (*appointmentdomain.Appointment, error) {
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

func (m *mockBookingUsecase) GetAppointmentsByPatient(patientID string) ([]appointmentdomain.Appointment, error) {
	return nil, nil
}

func (m *mockBookingUsecase) GetAppointmentsByExpert(expertID string, fromDate, toDate int64, status *appointmentdomain.AppointmentStatus) ([]appointmentdomain.Appointment, error) {
	return nil, nil
}

func (m *mockBookingUsecase) SaveMedicalRecord(actorID, actorRole, appointmentID string, cmd appappointment.SaveMedicalRecordCommand) (*appointmentdomain.MedicalRecord, error) {
	return nil, nil
}

func (m *mockBookingUsecase) GetMedicalRecordByAppointmentID(actorID, actorRole, appointmentID string) (*appointmentdomain.MedicalRecord, error) {
	return nil, nil
}

func (m *mockBookingUsecase) GetMedicalRecordByID(actorID, actorRole, recordID string) (*appointmentdomain.MedicalRecord, error) {
	return nil, nil
}

func (m *mockBookingUsecase) ListMedicalRecords(actorID, actorRole string, page bookingquery.PageRequest) (bookingquery.Page[appointmentdomain.MedicalRecord], error) {
	return bookingquery.Page[appointmentdomain.MedicalRecord]{}, nil
}

func TestCreateAppointment(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("TC-BOOK-APT-01 - Tạo cuộc hẹn thành công (Happy Case)", func(t *testing.T) {
		mockUsecase := &mockBookingUsecase{
			createFunc: func(cmd appappointment.CreateAppointmentCommand) (*appointmentdomain.Appointment, error) {
				return &appointmentdomain.Appointment{
					AppointmentID: "appt-uuid-1",
					SlotID:        cmd.SlotID,
					PatientID:     cmd.PatientID,
					ExpertID:      cmd.ExpertID,
					Status:        appointmentdomain.AppointmentStatusPendingPayment,
				}, nil
			},
		}
		h := NewHandler(mockUsecase)

		body := map[string]string{
			"expert_id": "expert-uuid-1",
			"slot_id":   "slot-uuid-1",
			"patient_record_id": "record-uuid-1",
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
			createFunc: func(cmd appappointment.CreateAppointmentCommand) (*appointmentdomain.Appointment, error) {
				return nil, errors.New("slot không được giữ bởi bạn, vui lòng thực hiện lại từ đầu")
			},
		}
		h := NewHandler(mockUsecase)

		body := map[string]string{
			"expert_id": "expert-uuid-1",
			"slot_id":   "slot-uuid-unlocked",
			"patient_record_id": "record-uuid-1",
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
			createFunc: func(cmd appappointment.CreateAppointmentCommand) (*appointmentdomain.Appointment, error) {
				return nil, errors.New("phiên giữ chỗ đã hết hạn 15 phút, vui lòng chọn lại")
			},
		}
		h := NewHandler(mockUsecase)

		body := map[string]string{
			"expert_id": "expert-uuid-1",
			"slot_id":   "slot-uuid-expired",
			"patient_record_id": "record-uuid-1",
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
			createFunc: func(cmd appappointment.CreateAppointmentCommand) (*appointmentdomain.Appointment, error) {
				return nil, errors.New("slot is covered by expert time-off")
			},
		}
		h := NewHandler(mockUsecase)

		body := map[string]string{
			"expert_id": "expert-uuid-1",
			"slot_id":   "slot-uuid-timeoff",
			"patient_record_id": "record-uuid-1",
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
			"patient_record_id": "record-uuid-1",
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
