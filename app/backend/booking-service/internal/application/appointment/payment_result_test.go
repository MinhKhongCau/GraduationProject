package appointment

import (
	appointmentdomain "booking-service/internal/domain/appointment"
	slotdomain "booking-service/internal/domain/slot"
	"errors"
	"testing"
)

func TestPlanPaymentResultSuccessFromPendingConfirmsAppointmentAndOccupiesSlot(t *testing.T) {
	now := int64(12345)
	transition, err := appointmentdomain.PlanPaymentResultTransition(
		appointmentdomain.Appointment{Status: appointmentdomain.AppointmentStatusPendingPayment},
		slotdomain.ExpertSlot{Status: slotdomain.SlotStatusLocked},
		appointmentdomain.PaymentResultSuccess,
		now,
	)
	if err != nil {
		t.Fatalf("expected success transition, got %v", err)
	}
	if transition.Noop {
		t.Fatal("expected mutation, got no-op")
	}
	if transition.AppointmentUpdates["status"] != appointmentdomain.AppointmentStatusConfirmed {
		t.Fatalf("expected appointment CONFIRMED, got %#v", transition.AppointmentUpdates["status"])
	}
	if transition.AppointmentUpdates["confirmed_at"] != now {
		t.Fatalf("expected confirmed_at %d, got %#v", now, transition.AppointmentUpdates["confirmed_at"])
	}
	if transition.SlotUpdates["status"] != slotdomain.SlotStatusOccupied {
		t.Fatalf("expected slot OCCUPIED, got %#v", transition.SlotUpdates["status"])
	}
}

func TestPlanPaymentResultDuplicateSuccessFromConfirmedIsNoop(t *testing.T) {
	confirmedAt := int64(111)
	transition, err := appointmentdomain.PlanPaymentResultTransition(
		appointmentdomain.Appointment{Status: appointmentdomain.AppointmentStatusConfirmed, ConfirmedAt: &confirmedAt},
		slotdomain.ExpertSlot{Status: slotdomain.SlotStatusOccupied},
		appointmentdomain.PaymentResultSuccess,
		222,
	)
	if err != nil {
		t.Fatalf("expected no-op success, got %v", err)
	}
	if !transition.Noop {
		t.Fatal("expected duplicate success to be no-op")
	}
	if len(transition.AppointmentUpdates) != 0 {
		t.Fatalf("expected no appointment updates, got %#v", transition.AppointmentUpdates)
	}
}

func TestPlanPaymentResultSuccessFromCancelledConflicts(t *testing.T) {
	_, err := appointmentdomain.PlanPaymentResultTransition(
		appointmentdomain.Appointment{Status: appointmentdomain.AppointmentStatusCancelled},
		slotdomain.ExpertSlot{Status: slotdomain.SlotStatusAvailable},
		appointmentdomain.PaymentResultSuccess,
		123,
	)
	if !errors.Is(err, appointmentdomain.ErrPaymentResultConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
}

func TestPlanPaymentResultFailedFromPendingCancelsAppointmentAndReleasesSlot(t *testing.T) {
	now := int64(12345)
	transition, err := appointmentdomain.PlanPaymentResultTransition(
		appointmentdomain.Appointment{Status: appointmentdomain.AppointmentStatusPendingPayment},
		slotdomain.ExpertSlot{Status: slotdomain.SlotStatusLocked},
		appointmentdomain.PaymentResultFailed,
		now,
	)
	if err != nil {
		t.Fatalf("expected failure transition, got %v", err)
	}
	if transition.AppointmentUpdates["status"] != appointmentdomain.AppointmentStatusCancelled {
		t.Fatalf("expected appointment CANCELLED, got %#v", transition.AppointmentUpdates["status"])
	}
	if transition.AppointmentUpdates["updated_at"] != now {
		t.Fatalf("expected updated_at %d, got %#v", now, transition.AppointmentUpdates["updated_at"])
	}
	if transition.AppointmentUpdates["cancellation_reason"] != "Payment failed" {
		t.Fatalf("expected cancellation reason to be set, got %#v", transition.AppointmentUpdates["cancellation_reason"])
	}
	if transition.SlotUpdates["status"] != slotdomain.SlotStatusAvailable {
		t.Fatalf("expected slot AVAILABLE, got %#v", transition.SlotUpdates["status"])
	}
	if transition.SlotUpdates["locked_expires_at"] != nil || transition.SlotUpdates["locked_by"] != nil {
		t.Fatalf("expected slot lock metadata to be cleared, got %#v", transition.SlotUpdates)
	}
}

func TestPlanPaymentResultDuplicateFailureFromCancelledIsNoop(t *testing.T) {
	transition, err := appointmentdomain.PlanPaymentResultTransition(
		appointmentdomain.Appointment{Status: appointmentdomain.AppointmentStatusCancelled},
		slotdomain.ExpertSlot{Status: slotdomain.SlotStatusAvailable},
		appointmentdomain.PaymentResultFailed,
		123,
	)
	if err != nil {
		t.Fatalf("expected no-op failure, got %v", err)
	}
	if !transition.Noop {
		t.Fatal("expected duplicate failure to be no-op")
	}
}

func TestPlanPaymentResultFailureFromConfirmedConflicts(t *testing.T) {
	_, err := appointmentdomain.PlanPaymentResultTransition(
		appointmentdomain.Appointment{Status: appointmentdomain.AppointmentStatusConfirmed},
		slotdomain.ExpertSlot{Status: slotdomain.SlotStatusOccupied},
		appointmentdomain.PaymentResultFailed,
		123,
	)
	if !errors.Is(err, appointmentdomain.ErrPaymentResultConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
}

func TestHandlePaymentResultRollsBackWhenSlotUpdateFails(t *testing.T) {
	repo := &fakePaymentResultRepository{
		appointment: appointmentdomain.Appointment{
			AppointmentID: "appt-1",
			SlotID:        "slot-1",
			Status:        appointmentdomain.AppointmentStatusPendingPayment,
		},
		slot: slotdomain.ExpertSlot{
			SlotID: "slot-1",
			Status: slotdomain.SlotStatusLocked,
		},
		failSlotUpdate: true,
	}
	usecase := NewUsecase(repo)

	err := usecase.HandlePaymentResult(HandlePaymentResultCommand{
		AppointmentID: "appt-1",
		Status:        PaymentResultSuccess,
	})
	if err == nil {
		t.Fatal("expected slot update error")
	}
	if repo.appointment.Status != appointmentdomain.AppointmentStatusPendingPayment {
		t.Fatalf("expected appointment rollback to PENDING_PAYMENT, got %s", repo.appointment.Status.String())
	}
	if repo.slot.Status != slotdomain.SlotStatusLocked {
		t.Fatalf("expected slot rollback to LOCKED, got %s", repo.slot.Status.String())
	}
}

type fakePaymentResultRepository struct {
	appointment    appointmentdomain.Appointment
	slot           slotdomain.ExpertSlot
	failSlotUpdate bool
}

func (r *fakePaymentResultRepository) HandlePaymentResult(command HandlePaymentResultCommand) error {
	apptSnapshot := r.appointment
	slotSnapshot := r.slot

	transition, err := appointmentdomain.PlanPaymentResultTransition(r.appointment, r.slot, appointmentdomain.PaymentResultStatus(command.Status), 12345)
	if err != nil {
		return err
	}
	if transition.Noop {
		return nil
	}

	if status, ok := transition.AppointmentUpdates["status"].(appointmentdomain.AppointmentStatus); ok {
		r.appointment.Status = status
	}
	if r.failSlotUpdate {
		r.appointment = apptSnapshot
		r.slot = slotSnapshot
		return errors.New("slot update failed")
	}
	if status, ok := transition.SlotUpdates["status"].(slotdomain.SlotStatus); ok {
		r.slot.Status = status
	}
	return nil
}

func (r *fakePaymentResultRepository) GetAppointmentByID(appointmentID string) (*appointmentdomain.Appointment, error) {
	return &r.appointment, nil
}

func (r *fakePaymentResultRepository) GetPaymentEligibilitySnapshot(command GetPaymentEligibilityCommand) (*PaymentEligibilitySnapshot, error) {
	return &PaymentEligibilitySnapshot{Appointment: r.appointment, Slot: r.slot}, nil
}

func (r *fakePaymentResultRepository) GetAppointmentBySlotID(slotID string) (*appointmentdomain.Appointment, error) {
	return &r.appointment, nil
}

func (r *fakePaymentResultRepository) CancelAppointmentByExpert(appointmentID string, reason string) error {
	return nil
}

func (r *fakePaymentResultRepository) CancelAppointmentByPatient(appointmentID string, patientID string, reason string) error {
	return nil
}

func (r *fakePaymentResultRepository) GetAppointmentsByPatient(patientID string) ([]appointmentdomain.Appointment, error) {
	return nil, nil
}

func (r *fakePaymentResultRepository) GetAppointmentsByExpert(expertID string, fromDate, toDate int64, status *appointmentdomain.AppointmentStatus) ([]appointmentdomain.Appointment, error) {
	return nil, nil
}

func (r *fakePaymentResultRepository) LockSlot(slotID string, patientID string) error {
	return nil
}

func (r *fakePaymentResultRepository) CreateAppointment(appointment *appointmentdomain.Appointment) error {
	return nil
}

func (r *fakePaymentResultRepository) ConfirmPayment(appointmentID string) error {
	return r.HandlePaymentResult(HandlePaymentResultCommand{AppointmentID: appointmentID, Status: PaymentResultSuccess})
}

func (r *fakePaymentResultRepository) CancelExpiredLocks() (int64, error) {
	return 0, nil
}

func (r *fakePaymentResultRepository) UpdateAppointmentStatus(appointmentID string, status appointmentdomain.AppointmentStatus) error {
	r.appointment.Status = status
	return nil
}

func (r *fakePaymentResultRepository) SaveMedicalRecord(record *appointmentdomain.MedicalRecord) error {
	return nil
}

func (r *fakePaymentResultRepository) GetMedicalRecordByAppointmentID(appointmentID string) (*appointmentdomain.MedicalRecord, error) {
	return nil, nil
}

func (r *fakePaymentResultRepository) GetMedicalRecordByID(recordID string) (*appointmentdomain.MedicalRecord, error) {
	return nil, nil
}

func (r *fakePaymentResultRepository) ListMedicalRecordsByPatient(patientID string, limit, offset int) ([]appointmentdomain.MedicalRecord, int64, error) {
	return nil, 0, nil
}

func (r *fakePaymentResultRepository) ListMedicalRecordsByExpert(expertID string, limit, offset int) ([]appointmentdomain.MedicalRecord, int64, error) {
	return nil, 0, nil
}
