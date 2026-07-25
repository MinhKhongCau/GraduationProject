package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	appappointment "booking-service/internal/booking/application/appointment"
	bookingquery "booking-service/internal/booking/application/query"
	"booking-service/internal/booking/domain"

	"github.com/gin-gonic/gin"
)

type mockReadUsecase struct {
	listFunc   func(query appappointment.AppointmentListQuery) (bookingquery.Page[domain.Appointment], error)
	detailFunc func(actorID, actorRole, appointmentID string) (*domain.Appointment, error)
}

func (m *mockReadUsecase) ListAppointments(query appappointment.AppointmentListQuery) (bookingquery.Page[domain.Appointment], error) {
	if m.listFunc != nil {
		return m.listFunc(query)
	}
	return bookingquery.Page[domain.Appointment]{}, nil
}

func (m *mockReadUsecase) GetAppointmentDetail(actorID, actorRole, appointmentID string) (*domain.Appointment, error) {
	if m.detailFunc != nil {
		return m.detailFunc(actorID, actorRole, appointmentID)
	}
	return nil, nil
}

type fullMockHandlerUsecase struct {
	mockBookingUsecase
	mockReadUsecase
}

func TestAppointmentManagement(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("TC-BOOK-MNG-01 - Bệnh nhân xem danh sách lịch hẹn thành công", func(t *testing.T) {
		fullMock := &fullMockHandlerUsecase{}
		fullMock.listFunc = func(query appappointment.AppointmentListQuery) (bookingquery.Page[domain.Appointment], error) {
			return bookingquery.Page[domain.Appointment]{
				Items: []domain.Appointment{
					{
						AppointmentID: "appt-uuid-1",
						PatientID:     "patient-uuid-1",
						ExpertID:      "expert-uuid-1",
						Status:        domain.AppointmentStatusConfirmed,
					},
				},
				Page:       0,
				Size:       10,
				TotalItems: 1,
				TotalPages: 1,
			}, nil
		}

		h := NewHandler(fullMock)

		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest(http.MethodGet, "/booking/appointments?page=0&size=10", nil)
		c.Request.Header.Set("X-User-Role", "PATIENT")
		c.Request.Header.Set("X-User-Id", "patient-uuid-1")

		h.GetPatient(c)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", recorder.Code, recorder.Body.String())
		}

		var resp map[string]interface{}
		json.Unmarshal(recorder.Body.Bytes(), &resp)
		if resp["success"] != true {
			t.Fatalf("expected success = true, got %v", resp["success"])
		}
	})

	t.Run("TC-BOOK-MNG-02 - Chuyên gia xem danh sách lịch hẹn thành công", func(t *testing.T) {
		fullMock := &fullMockHandlerUsecase{}
		fullMock.listFunc = func(query appappointment.AppointmentListQuery) (bookingquery.Page[domain.Appointment], error) {
			return bookingquery.Page[domain.Appointment]{
				Items: []domain.Appointment{
					{
						AppointmentID: "appt-uuid-1",
						PatientID:     "patient-uuid-1",
						ExpertID:      "expert-uuid-1",
						Status:        domain.AppointmentStatusConfirmed,
					},
				},
				Page:       0,
				Size:       10,
				TotalItems: 1,
				TotalPages: 1,
			}, nil
		}

		h := NewHandler(fullMock)

		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest(http.MethodGet, "/booking/appointments/expert?page=0&size=10", nil)
		c.Request.Header.Set("X-User-Role", "EXPERT")
		c.Request.Header.Set("X-User-Id", "expert-uuid-1")

		h.GetExpert(c)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", recorder.Code, recorder.Body.String())
		}

		var resp map[string]interface{}
		json.Unmarshal(recorder.Body.Bytes(), &resp)
		if resp["success"] != true {
			t.Fatalf("expected success = true, got %v", resp["success"])
		}
	})

	t.Run("TC-BOOK-MNG-03 - Xem chi tiết lịch hẹn thành công", func(t *testing.T) {
		fullMock := &fullMockHandlerUsecase{}
		fullMock.detailFunc = func(actorID, actorRole, appointmentID string) (*domain.Appointment, error) {
			return &domain.Appointment{
				AppointmentID: "appt-uuid-1",
				PatientID:     actorID,
				ExpertID:      "expert-uuid-1",
				Status:        domain.AppointmentStatusConfirmed,
			}, nil
		}

		h := NewHandler(fullMock)

		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest(http.MethodGet, "/booking/appointments/appt-uuid-1", nil)
		c.Request.Header.Set("X-User-Role", "PATIENT")
		c.Request.Header.Set("X-User-Id", "patient-uuid-1")
		c.Params = gin.Params{{Key: "id", Value: "appt-uuid-1"}}

		h.GetDetail(c)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", recorder.Code, recorder.Body.String())
		}

		var resp map[string]interface{}
		json.Unmarshal(recorder.Body.Bytes(), &resp)
		if resp["success"] != true {
			t.Fatalf("expected success = true, got %v", resp["success"])
		}
	})

	t.Run("TC-BOOK-MNG-04 - Xem chi tiết lịch hẹn thất bại do không tìm thấy (Not Found)", func(t *testing.T) {
		fullMock := &fullMockHandlerUsecase{}
		fullMock.detailFunc = func(actorID, actorRole, appointmentID string) (*domain.Appointment, error) {
			return nil, appappointment.ErrNotFound
		}

		h := NewHandler(fullMock)

		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest(http.MethodGet, "/booking/appointments/non-existent-id", nil)
		c.Request.Header.Set("X-User-Role", "PATIENT")
		c.Request.Header.Set("X-User-Id", "patient-uuid-1")
		c.Params = gin.Params{{Key: "id", Value: "non-existent-id"}}

		h.GetDetail(c)

		if recorder.Code != http.StatusNotFound {
			t.Fatalf("expected 404 Not Found, got %d: %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("TC-BOOK-MNG-05 - Hủy lịch hẹn thành công (Happy Case)", func(t *testing.T) {
		mockUsecase := &mockBookingUsecase{
			cancelFunc: func(appointmentID, userID, userRole, reason string) error {
				return nil
			},
		}
		h := NewHandler(mockUsecase)

		body := map[string]string{
			"reason": "Bận việc đột xuất",
		}
		jsonBytes, _ := json.Marshal(body)

		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest(http.MethodPatch, "/booking/appointments/appt-uuid-1/cancel", bytes.NewBuffer(jsonBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Request.Header.Set("X-User-Role", "PATIENT")
		c.Request.Header.Set("X-User-Id", "patient-uuid-1")
		c.Params = gin.Params{{Key: "id", Value: "appt-uuid-1"}}

		h.Cancel(c)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", recorder.Code, recorder.Body.String())
		}

		var resp map[string]interface{}
		json.Unmarshal(recorder.Body.Bytes(), &resp)
		if resp["success"] != true {
			t.Fatalf("expected success = true, got %v", resp["success"])
		}
	})

	t.Run("TC-BOOK-MNG-06 - Hủy lịch hẹn thất bại do cuộc hẹn đã bị hủy trước đó", func(t *testing.T) {
		mockUsecase := &mockBookingUsecase{
			cancelFunc: func(appointmentID, userID, userRole, reason string) error {
				return domain.ErrAppointmentAlreadyCancelled
			},
		}
		h := NewHandler(mockUsecase)

		body := map[string]string{
			"reason": "Hủy lại lần nữa",
		}
		jsonBytes, _ := json.Marshal(body)

		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest(http.MethodPatch, "/booking/appointments/appt-already-cancelled/cancel", bytes.NewBuffer(jsonBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Request.Header.Set("X-User-Role", "PATIENT")
		c.Request.Header.Set("X-User-Id", "patient-uuid-1")
		c.Params = gin.Params{{Key: "id", Value: "appt-already-cancelled"}}

		h.Cancel(c)

		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 Internal Server Error, got %d: %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("TC-BOOK-MNG-07 - Hủy lịch hẹn thất bại do thiếu header định danh người dùng (Unauthorized)", func(t *testing.T) {
		mockUsecase := &mockBookingUsecase{}
		h := NewHandler(mockUsecase)

		body := map[string]string{
			"reason": "Lý do hủy",
		}
		jsonBytes, _ := json.Marshal(body)

		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest(http.MethodPatch, "/booking/appointments/appt-uuid-1/cancel", bytes.NewBuffer(jsonBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		// Thiếu header X-User-Id & X-User-Role
		c.Params = gin.Params{{Key: "id", Value: "appt-uuid-1"}}

		h.Cancel(c)

		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 Unauthorized, got %d: %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("TC-BOOK-MNG-08 - Hủy lịch hẹn thất bại do thiếu lý do hủy (Bad Request)", func(t *testing.T) {
		mockUsecase := &mockBookingUsecase{}
		h := NewHandler(mockUsecase)

		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest(http.MethodPatch, "/booking/appointments/appt-uuid-1/cancel", bytes.NewBufferString("{}"))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Request.Header.Set("X-User-Role", "PATIENT")
		c.Request.Header.Set("X-User-Id", "patient-uuid-1")
		c.Params = gin.Params{{Key: "id", Value: "appt-uuid-1"}}

		h.Cancel(c)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %d: %s", recorder.Code, recorder.Body.String())
		}
	})
}
