package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	appappointment "booking-service/internal/application/appointment"
	"booking-service/internal/infrastructure/http/middleware"

	"github.com/gin-gonic/gin"
)

type fakeMeetingUsecase struct {
	fakeAppointmentUsecase
	completeCommands []appappointment.CompleteSessionCommand
	completeErr      error
}

func (u *fakeMeetingUsecase) GetMeetingByToken(ctx context.Context, token string) (*appappointment.MeetingInfo, error) {
	return &appappointment.MeetingInfo{AppointmentID: "appt-1"}, nil
}

func (u *fakeMeetingUsecase) CompleteSession(ctx context.Context, command appappointment.CompleteSessionCommand) (*appappointment.SessionCompletion, error) {
	u.completeCommands = append(u.completeCommands, command)
	if u.completeErr != nil {
		return nil, u.completeErr
	}
	return &appappointment.SessionCompletion{AppointmentID: command.AppointmentID, Settled: true}, nil
}

const testAppointmentID = "6f1c2b8e-4c1a-4d55-9a4e-2f0d7b1e9c11"

func completeSessionContext(callerID, body string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/internal/appointments/"+testAppointmentID+"/complete-session", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: testAppointmentID}}
	if callerID != "" {
		c.Set(middleware.ContextKeyCallerID, callerID)
	}
	return c, recorder
}

func TestInternalCompleteSessionOnlyAcceptsChatroomService(t *testing.T) {
	usecase := &fakeMeetingUsecase{}
	c, recorder := completeSessionContext("payment-service", `{"expert_id":"expert-1","presence_seconds":1800}`)

	NewHandler(usecase).InternalCompleteSession(c)

	if recorder.Code != http.StatusForbidden || len(usecase.completeCommands) != 0 {
		t.Fatalf("expected 403 without calling usecase, got %d calls=%d", recorder.Code, len(usecase.completeCommands))
	}
}

func TestInternalCompleteSessionPassesPresence(t *testing.T) {
	usecase := &fakeMeetingUsecase{}
	c, recorder := completeSessionContext(MeetingCallerID, `{"expert_id":"expert-1","presence_seconds":1830}`)

	NewHandler(usecase).InternalCompleteSession(c)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	got := usecase.completeCommands[0]
	if got.AppointmentID != testAppointmentID || got.ExpertID != "expert-1" || got.Presence != 1830*time.Second {
		t.Fatalf("unexpected command: %+v", got)
	}
}

func TestInternalCompleteSessionMapsErrors(t *testing.T) {
	tests := []struct {
		err  error
		want int
	}{
		{appappointment.ErrSessionTooShort, http.StatusConflict},
		{appappointment.ErrSessionForbidden, http.StatusForbidden},
		{appappointment.ErrSessionSettlementFails, http.StatusBadGateway},
		{appappointment.ErrNotFound, http.StatusNotFound},
	}
	for _, tt := range tests {
		usecase := &fakeMeetingUsecase{completeErr: tt.err}
		c, recorder := completeSessionContext(MeetingCallerID, `{"expert_id":"expert-1","presence_seconds":60}`)
		NewHandler(usecase).InternalCompleteSession(c)
		if recorder.Code != tt.want {
			t.Fatalf("%v: expected %d, got %d", tt.err, tt.want, recorder.Code)
		}
	}
}
