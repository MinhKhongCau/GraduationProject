package handler

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	appappointment "booking-service/internal/booking/application/appointment"
	"booking-service/internal/booking/domain"
	"booking-service/pkg/internal_auth"

	"github.com/gin-gonic/gin"
)

func TestInternalPaymentWebhookRejectsInvalidStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	usecase := &fakeAppointmentUsecase{}
	handler := NewHandler(usecase)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/internal/appointments/appt-1/webhook", bytes.NewBufferString(`{"appointment_id":"ignored","status":"BOGUS"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "appt-1"}}
	c.Set(internal_auth.ContextKeyCallerID, "payment-service")

	handler.InternalPaymentWebhook(c)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if len(usecase.commands) != 0 {
		t.Fatalf("expected usecase not to be called, got %d calls", len(usecase.commands))
	}
}

func TestInternalPaymentWebhookReturnsNon2xxWhenUsecaseFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	usecase := &fakeAppointmentUsecase{handleErr: errors.New("repository unavailable")}
	handler := NewHandler(usecase)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/internal/appointments/appt-1/webhook", bytes.NewBufferString(`{"appointment_id":"ignored","status":"FAILED"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "appt-1"}}
	c.Set(internal_auth.ContextKeyCallerID, "payment-service")

	handler.InternalPaymentWebhook(c)

	if recorder.Code < 400 {
		t.Fatalf("expected non-2xx response, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if len(usecase.commands) != 1 {
		t.Fatalf("expected exactly one usecase call, got %d", len(usecase.commands))
	}
	if usecase.commands[0].AppointmentID != "appt-1" {
		t.Fatalf("expected path appointment id to win, got %q", usecase.commands[0].AppointmentID)
	}
	if usecase.commands[0].Status != appappointment.PaymentResultFailed {
		t.Fatalf("expected FAILED command, got %q", usecase.commands[0].Status)
	}
}

func TestInternalPaymentEligibilityRequiresPaymentServiceCaller(t *testing.T) {
	gin.SetMode(gin.TestMode)
	usecase := &fakeAppointmentUsecase{}
	handler := NewHandler(usecase)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/internal/appointments/"+uuidString()+"/payment-eligibility", bytes.NewBufferString(`{"payer_id":"`+uuidString()+`"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: uuidString()}}

	handler.InternalPaymentEligibility(c)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if len(usecase.eligibilityCommands) != 0 {
		t.Fatalf("expected usecase not to be called, got %d calls", len(usecase.eligibilityCommands))
	}
}

func TestInternalPaymentEligibilityRejectsInvalidPayer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	usecase := &fakeAppointmentUsecase{}
	handler := NewHandler(usecase)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/internal/appointments/"+uuidString()+"/payment-eligibility", bytes.NewBufferString(`{"payer_id":"not-a-uuid"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: uuidString()}}
	c.Set(internal_auth.ContextKeyCallerID, "payment-service")

	handler.InternalPaymentEligibility(c)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if len(usecase.eligibilityCommands) != 0 {
		t.Fatalf("expected usecase not to be called, got %d calls", len(usecase.eligibilityCommands))
	}
}

func TestInternalPaymentEligibilityMapsTypedErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{name: "not found", err: appappointment.ErrNotFound, wantStatus: http.StatusNotFound},
		{name: "forbidden", err: appappointment.ErrPaymentEligibilityForbidden, wantStatus: http.StatusForbidden},
		{name: "conflict", err: appappointment.ErrPaymentEligibilityConflict, wantStatus: http.StatusConflict},
		{name: "invalid price", err: appappointment.ErrInvalidBookingPrice, wantStatus: http.StatusBadRequest},
		{name: "unexpected", err: errors.New("database unavailable"), wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := &fakeAppointmentUsecase{eligibilityErr: tt.err}
			handler := NewHandler(usecase)
			appointmentID := uuidString()
			payerID := uuidString()

			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodPost, "/internal/appointments/"+appointmentID+"/payment-eligibility", bytes.NewBufferString(`{"payer_id":"`+payerID+`"}`))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: appointmentID}}
			c.Set(internal_auth.ContextKeyCallerID, "payment-service")

			handler.InternalPaymentEligibility(c)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d: %s", tt.wantStatus, recorder.Code, recorder.Body.String())
			}
			if len(usecase.eligibilityCommands) != 1 {
				t.Fatalf("expected one usecase call, got %d", len(usecase.eligibilityCommands))
			}
			if usecase.eligibilityCommands[0].AppointmentID != appointmentID {
				t.Fatalf("expected path appointment id, got %q", usecase.eligibilityCommands[0].AppointmentID)
			}
		})
	}
}

type fakeAppointmentUsecase struct {
	commands            []appappointment.HandlePaymentResultCommand
	eligibilityCommands []appappointment.GetPaymentEligibilityCommand
	handleErr           error
	eligibilityErr      error
}

func (u *fakeAppointmentUsecase) CreateAppointment(patientID, expertID, slotID string) (*domain.Appointment, error) {
	return nil, nil
}

func (u *fakeAppointmentUsecase) GetAppointmentByID(appointmentID string) (*domain.Appointment, error) {
	return nil, nil
}

func (u *fakeAppointmentUsecase) GetPaymentEligibility(command appappointment.GetPaymentEligibilityCommand) (appappointment.PaymentEligibility, error) {
	u.eligibilityCommands = append(u.eligibilityCommands, command)
	if u.eligibilityErr != nil {
		return appappointment.PaymentEligibility{}, u.eligibilityErr
	}
	return appappointment.PaymentEligibility{
		AppointmentID: command.AppointmentID,
		ExpertID:      uuidString(),
		AmountVND:     200000,
		ExpiresAt:     9999999999999,
	}, nil
}

func (u *fakeAppointmentUsecase) CancelAppointment(appointmentID, userID, userRole, reason string) error {
	return nil
}

func (u *fakeAppointmentUsecase) ConfirmPayment(appointmentID string) error {
	return u.HandlePaymentResult(appappointment.HandlePaymentResultCommand{AppointmentID: appointmentID, Status: appappointment.PaymentResultSuccess})
}

func (u *fakeAppointmentUsecase) HandlePaymentFailure(appointmentID string) error {
	return u.HandlePaymentResult(appappointment.HandlePaymentResultCommand{AppointmentID: appointmentID, Status: appappointment.PaymentResultFailed})
}

func (u *fakeAppointmentUsecase) HandlePaymentResult(command appappointment.HandlePaymentResultCommand) error {
	u.commands = append(u.commands, command)
	return u.handleErr
}

func (u *fakeAppointmentUsecase) GetAppointmentsByPatient(patientID string) ([]domain.Appointment, error) {
	return nil, nil
}

func (u *fakeAppointmentUsecase) GetAppointmentsByExpert(expertID string, fromDate, toDate int64, status *domain.AppointmentStatus) ([]domain.Appointment, error) {
	return nil, nil
}

func uuidString() string {
	return "11111111-1111-1111-1111-111111111111"
}
