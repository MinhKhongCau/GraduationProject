package appointment

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	appointmentdomain "booking-service/internal/domain/appointment"
	slotdomain "booking-service/internal/domain/slot"
)

type fakeSettlementGateway struct {
	calls []string
	err   error
}

func (g *fakeSettlementGateway) SettleSession(ctx context.Context, appointmentID string) error {
	g.calls = append(g.calls, appointmentID)
	return g.err
}

func meetingUsecase(repo *fakeUOWRepository, settlement SessionSettlementGateway) MeetingUsecase {
	return NewUsecaseWithMeetings(repo, nil, MeetingOptions{
		BaseURL:    "https://mindcare.test/meet/",
		Settlement: settlement,
	}).(MeetingUsecase)
}

func startedSessionRepo(status appointmentdomain.AppointmentStatus) *fakeUOWRepository {
	repo := paymentResultUOWRepo(status, slotdomain.SlotStatusOccupied)
	repo.tx.slot.StartTime = time.Now().Add(-40 * time.Minute).UnixMilli()
	repo.tx.slot.EndTime = time.Now().Add(20 * time.Minute).UnixMilli()
	return repo
}

func TestPaymentSuccessGeneratesMeetingLink(t *testing.T) {
	repo := paymentResultUOWRepo(appointmentdomain.AppointmentStatusPendingPayment, slotdomain.SlotStatusLocked)
	usecase := NewUsecaseWithMeetings(repo, nil, MeetingOptions{BaseURL: "https://mindcare.test/meet/"})

	if err := usecase.ConfirmPayment("appt-1"); err != nil {
		t.Fatalf("ConfirmPayment returned error: %v", err)
	}
	token, _ := repo.tx.lastAppointmentUpdates["meeting_token"].(string)
	link, _ := repo.tx.lastAppointmentUpdates["meeting_link"].(string)
	if len(token) < 32 {
		t.Fatalf("expected a random meeting token, got %q", token)
	}
	if link != "https://mindcare.test/meet/"+token {
		t.Fatalf("expected meeting link built from base URL and token, got %q", link)
	}
}

func TestPaymentFailureDoesNotGenerateMeetingLink(t *testing.T) {
	repo := paymentResultUOWRepo(appointmentdomain.AppointmentStatusPendingPayment, slotdomain.SlotStatusLocked)
	if err := NewUsecase(repo).HandlePaymentFailure("appt-1"); err != nil {
		t.Fatalf("HandlePaymentFailure returned error: %v", err)
	}
	if _, ok := repo.tx.lastAppointmentUpdates["meeting_link"]; ok {
		t.Fatal("failed payment must not create a meeting link")
	}
}

func TestCompleteSessionMarksCompletedAndSettles(t *testing.T) {
	repo := startedSessionRepo(appointmentdomain.AppointmentStatusConfirmed)
	settlement := &fakeSettlementGateway{}

	result, err := meetingUsecase(repo, settlement).CompleteSession(context.Background(), CompleteSessionCommand{
		AppointmentID: "appt-1",
		ExpertID:      "expert-1",
		Presence:      31 * time.Minute,
	})
	if err != nil {
		t.Fatalf("CompleteSession returned error: %v", err)
	}
	if !result.Settled || result.AlreadyCompleted {
		t.Fatalf("unexpected result: %+v", result)
	}
	if repo.tx.lastAppointmentUpdates["status"] != appointmentdomain.AppointmentStatusCompleted {
		t.Fatalf("expected COMPLETED, got %#v", repo.tx.lastAppointmentUpdates["status"])
	}
	if repo.tx.lastAppointmentUpdates["session_completed_at"] == nil {
		t.Fatal("expected session_completed_at to be set")
	}
	if len(settlement.calls) != 1 || settlement.calls[0] != "appt-1" {
		t.Fatalf("expected one settlement call, got %v", settlement.calls)
	}
}

func TestCompleteSessionRejectsShortPresenceWrongExpertAndEarlyCalls(t *testing.T) {
	tests := []struct {
		name     string
		expertID string
		presence time.Duration
		future   bool
		want     error
	}{
		{name: "presence below 30 minutes", expertID: "expert-1", presence: 29 * time.Minute, want: ErrSessionTooShort},
		{name: "another expert", expertID: "expert-2", presence: time.Hour, want: ErrSessionForbidden},
		{name: "before the slot starts", expertID: "expert-1", presence: time.Hour, future: true, want: ErrSessionNotStarted},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := startedSessionRepo(appointmentdomain.AppointmentStatusConfirmed)
			if tt.future {
				repo.tx.slot.StartTime = time.Now().Add(time.Hour).UnixMilli()
			}
			settlement := &fakeSettlementGateway{}
			_, err := meetingUsecase(repo, settlement).CompleteSession(context.Background(), CompleteSessionCommand{
				AppointmentID: "appt-1",
				ExpertID:      tt.expertID,
				Presence:      tt.presence,
			})
			if !errors.Is(err, tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, err)
			}
			if repo.tx.updateAppointmentCalls != 0 || len(settlement.calls) != 0 {
				t.Fatal("rejected completion must not update the appointment or pay out")
			}
		})
	}
}

func TestCompleteSessionRejectsUnpaidAppointment(t *testing.T) {
	repo := startedSessionRepo(appointmentdomain.AppointmentStatusPendingPayment)
	_, err := meetingUsecase(repo, &fakeSettlementGateway{}).CompleteSession(context.Background(), CompleteSessionCommand{
		AppointmentID: "appt-1",
		ExpertID:      "expert-1",
		Presence:      time.Hour,
	})
	if !errors.Is(err, ErrSessionNotCompletable) {
		t.Fatalf("expected ErrSessionNotCompletable, got %v", err)
	}
}

func TestCompleteSessionRetryOnlySettlesAgain(t *testing.T) {
	repo := startedSessionRepo(appointmentdomain.AppointmentStatusCompleted)
	completedAt := time.Now().Add(-time.Minute).UnixMilli()
	repo.tx.appointment.SessionCompletedAt = &completedAt
	settlement := &fakeSettlementGateway{}

	result, err := meetingUsecase(repo, settlement).CompleteSession(context.Background(), CompleteSessionCommand{
		AppointmentID: "appt-1",
		ExpertID:      "expert-1",
	})
	if err != nil {
		t.Fatalf("retry returned error: %v", err)
	}
	if !result.AlreadyCompleted || result.SessionCompletedAt != completedAt || !result.Settled {
		t.Fatalf("unexpected retry result: %+v", result)
	}
	if repo.tx.updateAppointmentCalls != 0 || len(settlement.calls) != 1 {
		t.Fatalf("retry should only call settlement: updates=%d settle=%d", repo.tx.updateAppointmentCalls, len(settlement.calls))
	}
}

func TestCompleteSessionReportsPayoutFailureAfterCommit(t *testing.T) {
	repo := startedSessionRepo(appointmentdomain.AppointmentStatusConfirmed)
	settlement := &fakeSettlementGateway{err: errors.New("payment-service down")}

	result, err := meetingUsecase(repo, settlement).CompleteSession(context.Background(), CompleteSessionCommand{
		AppointmentID: "appt-1",
		ExpertID:      "expert-1",
		Presence:      30 * time.Minute,
	})
	if !errors.Is(err, ErrSessionSettlementFails) || !strings.Contains(err.Error(), "payment-service down") {
		t.Fatalf("expected settlement failure, got %v", err)
	}
	if result == nil || result.Settled || repo.tx.updateAppointmentCalls != 1 {
		t.Fatalf("session should be recorded but not settled: result=%+v updates=%d", result, repo.tx.updateAppointmentCalls)
	}
}
