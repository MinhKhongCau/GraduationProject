package appointment

import (
	appointmentdomain "booking-service/internal/domain/appointment"
	slotdomain "booking-service/internal/domain/slot"
	"context"
	"errors"
	"testing"
	"time"
)

type fakeProfileGateway struct {
	err     error
	queries []BookingInfoQuery
}

func (g *fakeProfileGateway) GetBookingInfo(_ context.Context, query BookingInfoQuery) (*BookingInfo, error) {
	g.queries = append(g.queries, query)
	if g.err != nil {
		return nil, g.err
	}
	info := &BookingInfo{
		Expert: ExpertInfo{ExpertID: query.ExpertID, FullName: "Dr. A"},
		PatientRecord: PatientRecordInfo{
			RecordID:     query.PatientRecordID,
			FullName:     "Nguyen Van B",
			DateOfBirth:  "1990-01-02",
			Gender:       "MALE",
			PhoneNumber:  "0900000000",
			Relationship: "PARENT",
		},
	}
	if query.SpecializationID != "" {
		info.Specialization = &SpecializationInfo{SpecID: query.SpecializationID, Name: "Tâm lý"}
	}
	return info, nil
}

type fakeSlotReaderRepo struct {
	*fakeUOWRepository
	slot *slotdomain.ExpertSlot
}

func (r *fakeSlotReaderRepo) GetSlotByID(context.Context, string) (*slotdomain.ExpertSlot, error) {
	if r.slot == nil {
		return nil, ErrBookingSlotNotHeld
	}
	slot := *r.slot
	return &slot, nil
}

func lockedSlot(slotID, expertID, patientID string, expiresAt int64) slotdomain.ExpertSlot {
	return slotdomain.ExpertSlot{
		SlotID:          slotID,
		ExpertID:        expertID,
		Status:          slotdomain.SlotStatusLocked,
		LockedBy:        &patientID,
		LockedExpiresAt: &expiresAt,
		StartTime:       1_800_000_000_000,
		EndTime:         1_800_003_600_000,
		Price:           300000,
	}
}

func TestCreateAppointmentSnapshotsPatientRecordAndSpecialization(t *testing.T) {
	patientID, expertID, slotID := "patient-1", "expert-1", "slot-1"
	repo := newUOWTestRepo(appointmentdomain.Appointment{}, lockedSlot(slotID, expertID, patientID, time.Now().Add(time.Minute).UnixMilli()))
	profiles := &fakeProfileGateway{}

	appt, err := NewUsecaseWithProfiles(repo, profiles).CreateAppointment(context.Background(), CreateAppointmentCommand{
		PatientID:        patientID,
		ExpertID:         expertID,
		SlotID:           slotID,
		PatientRecordID:  "record-1",
		SpecializationID: "spec-1",
	})
	if err != nil {
		t.Fatalf("expected create success, got %v", err)
	}

	if len(profiles.queries) != 1 || profiles.queries[0].OwnerID != patientID || profiles.queries[0].PatientRecordID != "record-1" {
		t.Fatalf("expected one profile lookup owned by patient, got %#v", profiles.queries)
	}
	created := repo.tx.createdAppointment
	if created == nil {
		t.Fatal("expected appointment to be created")
	}
	if created.Patient.RecordID == nil || *created.Patient.RecordID != "record-1" || created.Patient.FullName != "Nguyen Van B" || created.Patient.Relationship != "PARENT" {
		t.Fatalf("expected patient snapshot, got %#v", created.Patient)
	}
	if created.SpecializationID == nil || *created.SpecializationID != "spec-1" || created.SpecializationName != "Tâm lý" {
		t.Fatalf("expected specialization snapshot, got %v %q", created.SpecializationID, created.SpecializationName)
	}
	if appt.PatientID != patientID {
		t.Fatalf("expected booking account as patient_id, got %q", appt.PatientID)
	}
}

func TestCreateAppointmentProfileErrorSkipsTransaction(t *testing.T) {
	patientID, expertID, slotID := "patient-1", "expert-1", "slot-1"
	repo := newUOWTestRepo(appointmentdomain.Appointment{}, lockedSlot(slotID, expertID, patientID, time.Now().Add(time.Minute).UnixMilli()))

	_, err := NewUsecaseWithProfiles(repo, &fakeProfileGateway{err: ErrBookingProfileNotFound}).CreateAppointment(context.Background(), CreateAppointmentCommand{
		PatientID: patientID, ExpertID: expertID, SlotID: slotID, PatientRecordID: "someone-elses-record",
	})
	if !errors.Is(err, ErrBookingProfileNotFound) {
		t.Fatalf("expected ErrBookingProfileNotFound, got %v", err)
	}
	if repo.withinTxCalls != 0 {
		t.Fatal("expected no transaction when patient record lookup fails")
	}
}

func TestCreateAppointmentWithoutProfileGatewayIsUnavailable(t *testing.T) {
	repo := newUOWTestRepo(appointmentdomain.Appointment{}, slotdomain.ExpertSlot{})

	_, err := NewUsecase(repo).CreateAppointment(context.Background(), CreateAppointmentCommand{PatientRecordID: "record-1"})
	if !errors.Is(err, ErrProfileServiceUnavailable) {
		t.Fatalf("expected ErrProfileServiceUnavailable, got %v", err)
	}
}

func TestCreateAppointmentRejectsSlotWithActiveAppointment(t *testing.T) {
	patientID, expertID, slotID := "patient-1", "expert-1", "slot-1"
	repo := newUOWTestRepo(appointmentdomain.Appointment{}, lockedSlot(slotID, expertID, patientID, time.Now().Add(time.Minute).UnixMilli()))
	repo.tx.hasActiveAppointment = true

	_, err := createWithFakeProfiles(repo, patientID, expertID, slotID)
	if !errors.Is(err, ErrSlotAlreadyBooked) {
		t.Fatalf("expected ErrSlotAlreadyBooked, got %v", err)
	}
	if repo.tx.createAppointmentCalls != 0 {
		t.Fatal("expected no duplicate appointment")
	}
}

func TestGetBookingConfirmationReturnsSlotAndProfileInfo(t *testing.T) {
	patientID, expertID, slotID := "patient-1", "expert-1", "slot-1"
	expiresAt := time.Now().Add(10 * time.Minute).UnixMilli()
	slot := lockedSlot(slotID, expertID, patientID, expiresAt)
	repo := &fakeSlotReaderRepo{fakeUOWRepository: newUOWTestRepo(appointmentdomain.Appointment{}, slot), slot: &slot}

	usecase := NewUsecaseWithProfiles(repo, &fakeProfileGateway{}).(BookingConfirmationUsecase)
	confirmation, err := usecase.GetBookingConfirmation(context.Background(), BookingConfirmationQuery{
		PatientID: patientID, ExpertID: expertID, SlotID: slotID, PatientRecordID: "record-1", SpecializationID: "spec-1",
	})
	if err != nil {
		t.Fatalf("expected confirmation, got %v", err)
	}
	if confirmation.Slot.SlotID != slotID || confirmation.Slot.Price != 300000 || confirmation.Slot.LockedExpiresAt != expiresAt {
		t.Fatalf("unexpected slot info %#v", confirmation.Slot)
	}
	if confirmation.PatientRecord.RecordID != "record-1" || confirmation.Expert.FullName != "Dr. A" || confirmation.Specialization == nil {
		t.Fatalf("unexpected profile info %#v", confirmation)
	}
}

func TestGetBookingConfirmationRequiresSlotHeldByPatient(t *testing.T) {
	expertID, slotID := "expert-1", "slot-1"
	tests := map[string]slotdomain.ExpertSlot{
		"held by another patient": lockedSlot(slotID, expertID, "patient-2", time.Now().Add(time.Minute).UnixMilli()),
		"hold expired":            lockedSlot(slotID, expertID, "patient-1", time.Now().Add(-time.Minute).UnixMilli()),
		"not locked":              {SlotID: slotID, ExpertID: expertID, Status: slotdomain.SlotStatusAvailable},
	}

	for name, slot := range tests {
		t.Run(name, func(t *testing.T) {
			profiles := &fakeProfileGateway{}
			repo := &fakeSlotReaderRepo{fakeUOWRepository: newUOWTestRepo(appointmentdomain.Appointment{}, slot), slot: &slot}

			_, err := NewUsecaseWithProfiles(repo, profiles).(BookingConfirmationUsecase).GetBookingConfirmation(context.Background(), BookingConfirmationQuery{
				PatientID: "patient-1", ExpertID: expertID, SlotID: slotID, PatientRecordID: "record-1",
			})
			if !errors.Is(err, ErrBookingSlotNotHeld) {
				t.Fatalf("expected ErrBookingSlotNotHeld, got %v", err)
			}
			if len(profiles.queries) != 0 {
				t.Fatal("expected no profile lookup for a slot the patient does not hold")
			}
		})
	}
}
