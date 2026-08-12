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

type mockMedicalRecordUsecase struct {
	mockBookingUsecase
	savedRecord *domain.MedicalRecord
}

func (m *mockMedicalRecordUsecase) SaveMedicalRecord(actorID, actorRole, appointmentID string, cmd appappointment.SaveMedicalRecordCommand) (*domain.MedicalRecord, error) {
	if actorRole != "EXPERT" {
		return nil, appappointment.ErrUnauthorized
	}
	m.savedRecord = &domain.MedicalRecord{
		RecordID:      "rec-123",
		AppointmentID: appointmentID,
		PatientID:     "pat-123",
		ExpertID:      actorID,
		Diagnosis:     cmd.Diagnosis,
		Symptoms:      cmd.Symptoms,
	}
	return m.savedRecord, nil
}

func (m *mockMedicalRecordUsecase) GetMedicalRecordByAppointmentID(actorID, actorRole, appointmentID string) (*domain.MedicalRecord, error) {
	if m.savedRecord != nil {
		return m.savedRecord, nil
	}
	return nil, appappointment.ErrMedicalRecordNotFound
}

func (m *mockMedicalRecordUsecase) ListMedicalRecords(actorID, actorRole string, page bookingquery.PageRequest) (bookingquery.Page[domain.MedicalRecord], error) {
	var items []domain.MedicalRecord
	if m.savedRecord != nil {
		items = append(items, *m.savedRecord)
	}
	return bookingquery.Page[domain.MedicalRecord]{
		Items:      items,
		Page:       0,
		Size:       10,
		TotalItems: int64(len(items)),
		TotalPages: 1,
	}, nil
}

func TestMedicalRecordHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Expert saves medical record successfully", func(t *testing.T) {
		mockU := &mockMedicalRecordUsecase{}
		h := NewHandler(mockU)

		body := map[string]interface{}{
			"diagnosis":        "Viêm xoang cấp",
			"symptoms":         "Nghẹt mũi, nhức trán",
			"actions_to_avoid": "Khói bụi, máy lạnh nhiệt độ thấp",
			"actions_to_take":  "Rửa mũi bằng nước muối sinh lý",
		}
		jsonBytes, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/booking/appointments/appt-1/medical-record", bytes.NewBuffer(jsonBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Request.Header.Set("X-User-Role", "EXPERT")
		c.Request.Header.Set("X-User-Id", "exp-1")
		c.Params = gin.Params{{Key: "id", Value: "appt-1"}}

		h.SaveMedicalRecord(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("Patient forbidden from saving medical record", func(t *testing.T) {
		mockU := &mockMedicalRecordUsecase{}
		h := NewHandler(mockU)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/booking/appointments/appt-1/medical-record", bytes.NewBufferString("{}"))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Request.Header.Set("X-User-Role", "PATIENT")
		c.Request.Header.Set("X-User-Id", "pat-1")
		c.Params = gin.Params{{Key: "id", Value: "appt-1"}}

		h.SaveMedicalRecord(c)

		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403 Forbidden, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("Get medical record list", func(t *testing.T) {
		mockU := &mockMedicalRecordUsecase{
			savedRecord: &domain.MedicalRecord{RecordID: "rec-123", Diagnosis: "Test"},
		}
		h := NewHandler(mockU)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/booking/medical-records?page=0&size=10", nil)
		c.Request.Header.Set("X-User-Role", "PATIENT")
		c.Request.Header.Set("X-User-Id", "pat-1")

		h.ListMedicalRecords(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
		}
	})
}
