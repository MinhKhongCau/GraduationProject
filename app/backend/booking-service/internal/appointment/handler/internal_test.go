package handler

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"booking-service/internal/appointment"
	"booking-service/internal/domain"
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
	if usecase.commands[0].Status != appointment.PaymentResultFailed {
		t.Fatalf("expected FAILED command, got %q", usecase.commands[0].Status)
	}
}

type fakeAppointmentUsecase struct {
	commands  []appointment.HandlePaymentResultCommand
	handleErr error
}

func (u *fakeAppointmentUsecase) CreateAppointment(patientID, expertID, slotID string) (*domain.Appointment, error) {
	return nil, nil
}

func (u *fakeAppointmentUsecase) GetAppointmentByID(appointmentID string) (*domain.Appointment, error) {
	return nil, nil
}

func (u *fakeAppointmentUsecase) CancelAppointment(appointmentID, userID, userRole, reason string) error {
	return nil
}

func (u *fakeAppointmentUsecase) ConfirmPayment(appointmentID string) error {
	return u.HandlePaymentResult(appointment.HandlePaymentResultCommand{AppointmentID: appointmentID, Status: appointment.PaymentResultSuccess})
}

func (u *fakeAppointmentUsecase) HandlePaymentFailure(appointmentID string) error {
	return u.HandlePaymentResult(appointment.HandlePaymentResultCommand{AppointmentID: appointmentID, Status: appointment.PaymentResultFailed})
}

func (u *fakeAppointmentUsecase) HandlePaymentResult(command appointment.HandlePaymentResultCommand) error {
	u.commands = append(u.commands, command)
	return u.handleErr
}

func (u *fakeAppointmentUsecase) GetAppointmentsByPatient(patientID string) ([]domain.Appointment, error) {
	return nil, nil
}

func (u *fakeAppointmentUsecase) GetAppointmentsByExpert(expertID string, fromDate, toDate int64, status *domain.AppointmentStatus) ([]domain.Appointment, error) {
	return nil, nil
}
