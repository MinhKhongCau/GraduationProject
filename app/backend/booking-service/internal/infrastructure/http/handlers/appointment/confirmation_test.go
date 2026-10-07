package handler

import (
	appappointment "booking-service/internal/application/appointment"
	appointmentdomain "booking-service/internal/domain/appointment"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type mockConfirmationUsecase struct {
	mockBookingUsecase
	confirmFunc func(query appappointment.BookingConfirmationQuery) (*appappointment.BookingConfirmation, error)
}

func (m *mockConfirmationUsecase) GetBookingConfirmation(_ context.Context, query appappointment.BookingConfirmationQuery) (*appappointment.BookingConfirmation, error) {
	return m.confirmFunc(query)
}

func newConfirmationContext(target string) (*httptest.ResponseRecorder, *gin.Context) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, target, nil)
	c.Request.Header.Set("X-User-Role", "PATIENT")
	c.Request.Header.Set("X-User-Id", "patient-1")
	return recorder, c
}

func TestGetConfirmation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const target = "/booking/appointments/confirmation?slot_id=slot-1&expert_id=expert-1&patient_record_id=record-1&specialization_id=spec-1"

	t.Run("returns confirmation for the requesting patient", func(t *testing.T) {
		var got appappointment.BookingConfirmationQuery
		h := NewHandler(&mockConfirmationUsecase{confirmFunc: func(query appappointment.BookingConfirmationQuery) (*appappointment.BookingConfirmation, error) {
			got = query
			return &appappointment.BookingConfirmation{
				Slot:          appappointment.ConfirmationSlot{SlotID: query.SlotID, Price: 300000},
				PatientRecord: appappointment.PatientRecordInfo{RecordID: query.PatientRecordID, FullName: "Nguyen Van B"},
			}, nil
		}})

		recorder, c := newConfirmationContext(target)
		h.GetConfirmation(c)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
		}
		want := appappointment.BookingConfirmationQuery{PatientID: "patient-1", ExpertID: "expert-1", SlotID: "slot-1", PatientRecordID: "record-1", SpecializationID: "spec-1"}
		if got != want {
			t.Fatalf("expected query %#v, got %#v", want, got)
		}
		var resp map[string]interface{}
		json.Unmarshal(recorder.Body.Bytes(), &resp)
		record := resp["data"].(map[string]interface{})["patient_record"].(map[string]interface{})
		if record["full_name"] != "Nguyen Van B" {
			t.Fatalf("expected patient record in response, got %v", record)
		}
	})

	t.Run("maps booking errors to HTTP status", func(t *testing.T) {
		tests := map[error]int{
			appappointment.ErrBookingSlotNotHeld:            http.StatusConflict,
			appappointment.ErrBookingProfileNotFound:        http.StatusNotFound,
			appappointment.ErrBookingSpecializationMismatch: http.StatusUnprocessableEntity,
			appappointment.ErrProfileServiceUnavailable:     http.StatusServiceUnavailable,
		}
		for usecaseErr, wantStatus := range tests {
			h := NewHandler(&mockConfirmationUsecase{confirmFunc: func(appappointment.BookingConfirmationQuery) (*appappointment.BookingConfirmation, error) {
				return nil, usecaseErr
			}})
			recorder, c := newConfirmationContext(target)
			h.GetConfirmation(c)
			if recorder.Code != wantStatus {
				t.Fatalf("%v: expected %d, got %d", usecaseErr, wantStatus, recorder.Code)
			}
		}
	})

	t.Run("requires patient_record_id", func(t *testing.T) {
		h := NewHandler(&mockConfirmationUsecase{})
		recorder, c := newConfirmationContext("/booking/appointments/confirmation?slot_id=slot-1&expert_id=expert-1")
		h.GetConfirmation(c)
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", recorder.Code)
		}
	})

	t.Run("rejects non-patients", func(t *testing.T) {
		h := NewHandler(&mockConfirmationUsecase{})
		recorder, c := newConfirmationContext(target)
		c.Request.Header.Set("X-User-Role", "EXPERT")
		h.GetConfirmation(c)
		if recorder.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", recorder.Code)
		}
	})
}

func TestCreateAppointmentRequiresPatientRecordAndMapsProfileErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	post := func(h *Handler, body map[string]string) *httptest.ResponseRecorder {
		jsonBytes, _ := json.Marshal(body)
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest(http.MethodPost, "/booking/appointments", bytes.NewBuffer(jsonBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Request.Header.Set("X-User-Role", "PATIENT")
		c.Request.Header.Set("X-User-Id", "patient-1")
		h.Create(c)
		return recorder
	}

	missing := post(NewHandler(&mockBookingUsecase{}), map[string]string{"slot_id": "slot-1", "expert_id": "expert-1"})
	if missing.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 without patient_record_id, got %d", missing.Code)
	}

	notOwned := post(NewHandler(&mockBookingUsecase{createFunc: func(appappointment.CreateAppointmentCommand) (*appointmentdomain.Appointment, error) {
		return nil, appappointment.ErrBookingProfileNotFound
	}}), map[string]string{"slot_id": "slot-1", "expert_id": "expert-1", "patient_record_id": "someone-elses"})
	if notOwned.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for a patient record not owned by caller, got %d", notOwned.Code)
	}
}
