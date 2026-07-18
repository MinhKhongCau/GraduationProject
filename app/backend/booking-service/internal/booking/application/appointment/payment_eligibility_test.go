package appointment

import (
	"errors"
	"math"
	"testing"

	"booking-service/internal/domain"
)

func TestGetPaymentEligibilityValidSnapshot(t *testing.T) {
	now := int64(1000)
	expiresAt := now + 900000
	payerID := "patient-1"
	expertID := "expert-1"
	appointmentID := "appt-1"
	repo := &fakeEligibilityRepository{snapshot: &PaymentEligibilitySnapshot{
		Appointment: domain.Appointment{
			AppointmentID: appointmentID,
			SlotID:        "slot-1",
			PatientID:     payerID,
			ExpertID:      expertID,
			Status:        domain.AppointmentStatusPendingPayment,
		},
		Slot: domain.ExpertSlot{
			SlotID:          "slot-1",
			ExpertID:        expertID,
			Status:          domain.SlotStatusLocked,
			Price:           200000,
			LockedBy:        &payerID,
			LockedExpiresAt: &expiresAt,
		},
	}}

	eligibility, err := domain.BuildPaymentEligibility(repo.snapshot.Appointment, repo.snapshot.Slot, payerID, now)
	if err != nil {
		t.Fatalf("expected valid eligibility, got %v", err)
	}
	if eligibility.AppointmentID != appointmentID {
		t.Fatalf("expected appointment_id %s, got %s", appointmentID, eligibility.AppointmentID)
	}
	if eligibility.ExpertID != expertID {
		t.Fatalf("expected expert_id %s, got %s", expertID, eligibility.ExpertID)
	}
	if eligibility.AmountVND != 200000 {
		t.Fatalf("expected amount_vnd 200000, got %d", eligibility.AmountVND)
	}
	if eligibility.ExpiresAt != expiresAt {
		t.Fatalf("expected expires_at %d, got %d", expiresAt, eligibility.ExpiresAt)
	}
}

func TestPaymentEligibilityRejectsInvalidBookingStates(t *testing.T) {
	now := int64(1000)
	expiresAt := now + 900000
	payerID := "patient-1"
	otherPayerID := "patient-2"
	expertID := "expert-1"

	baseAppt := domain.Appointment{
		AppointmentID: "appt-1",
		SlotID:        "slot-1",
		PatientID:     payerID,
		ExpertID:      expertID,
		Status:        domain.AppointmentStatusPendingPayment,
	}
	baseSlot := domain.ExpertSlot{
		SlotID:          "slot-1",
		ExpertID:        expertID,
		Status:          domain.SlotStatusLocked,
		Price:           200000,
		LockedBy:        &payerID,
		LockedExpiresAt: &expiresAt,
	}

	tests := []struct {
		name    string
		command GetPaymentEligibilityCommand
		appt    domain.Appointment
		slot    domain.ExpertSlot
		wantErr error
	}{
		{
			name:    "wrong payer",
			command: GetPaymentEligibilityCommand{AppointmentID: "appt-1", PayerID: otherPayerID},
			appt:    baseAppt,
			slot:    baseSlot,
			wantErr: domain.ErrPaymentEligibilityForbidden,
		},
		{
			name:    "appointment not pending payment",
			command: GetPaymentEligibilityCommand{AppointmentID: "appt-1", PayerID: payerID},
			appt:    appointmentWithStatus(baseAppt, domain.AppointmentStatusConfirmed),
			slot:    baseSlot,
			wantErr: domain.ErrPaymentEligibilityConflict,
		},
		{
			name:    "slot not locked",
			command: GetPaymentEligibilityCommand{AppointmentID: "appt-1", PayerID: payerID},
			appt:    baseAppt,
			slot:    slotWithStatus(baseSlot, domain.SlotStatusAvailable),
			wantErr: domain.ErrPaymentEligibilityConflict,
		},
		{
			name:    "slot expert mismatch",
			command: GetPaymentEligibilityCommand{AppointmentID: "appt-1", PayerID: payerID},
			appt:    baseAppt,
			slot:    slotWithExpert(baseSlot, "expert-2"),
			wantErr: domain.ErrPaymentEligibilityConflict,
		},
		{
			name:    "locked by missing",
			command: GetPaymentEligibilityCommand{AppointmentID: "appt-1", PayerID: payerID},
			appt:    baseAppt,
			slot:    slotWithLockedBy(baseSlot, nil),
			wantErr: domain.ErrPaymentEligibilityConflict,
		},
		{
			name:    "locked by wrong payer",
			command: GetPaymentEligibilityCommand{AppointmentID: "appt-1", PayerID: payerID},
			appt:    baseAppt,
			slot:    slotWithLockedBy(baseSlot, &otherPayerID),
			wantErr: domain.ErrPaymentEligibilityConflict,
		},
		{
			name:    "expiry missing",
			command: GetPaymentEligibilityCommand{AppointmentID: "appt-1", PayerID: payerID},
			appt:    baseAppt,
			slot:    slotWithExpiry(baseSlot, nil),
			wantErr: domain.ErrPaymentEligibilityConflict,
		},
		{
			name:    "expired",
			command: GetPaymentEligibilityCommand{AppointmentID: "appt-1", PayerID: payerID},
			appt:    baseAppt,
			slot:    slotWithExpiry(baseSlot, int64Ptr(now)),
			wantErr: domain.ErrPaymentEligibilityConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.BuildPaymentEligibility(tt.appt, tt.slot, tt.command.PayerID, now)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestStrictPriceToVND(t *testing.T) {
	tests := []struct {
		name    string
		price   float64
		want    int64
		wantErr bool
	}{
		{name: "integer VND accepted", price: 200000.00, want: 200000},
		{name: "fractional rejected", price: 200000.50, wantErr: true},
		{name: "zero rejected", price: 0, wantErr: true},
		{name: "negative rejected", price: -1, wantErr: true},
		{name: "NaN rejected", price: math.NaN(), wantErr: true},
		{name: "positive infinity rejected", price: math.Inf(1), wantErr: true},
		{name: "negative infinity rejected", price: math.Inf(-1), wantErr: true},
		{name: "VNPay scaled overflow rejected", price: 92233720368547760, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.NewMoneyVNDFromPrice(tt.price)
			if tt.wantErr {
				if !errors.Is(err, domain.ErrInvalidMoneyVND) {
					t.Fatalf("expected ErrInvalidMoneyVND, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
			if int64(got) != tt.want {
				t.Fatalf("expected %d, got %d", tt.want, got)
			}
		})
	}
}

func TestAppointmentCreationSlotValidationRejectsExpertMismatchAndNilExpiry(t *testing.T) {
	now := int64(1000)
	payerID := "patient-1"
	expertID := "expert-1"
	expiresAt := now + 900000
	baseAppt := &domain.Appointment{
		PatientID: payerID,
		ExpertID:  expertID,
	}
	baseSlot := domain.ExpertSlot{
		ExpertID:        expertID,
		Status:          domain.SlotStatusLocked,
		LockedBy:        &payerID,
		LockedExpiresAt: &expiresAt,
	}

	if err := domain.ValidateAppointmentCreationSlot(slotWithExpert(baseSlot, "expert-2"), baseAppt, now); err == nil {
		t.Fatal("expected expert mismatch to be rejected")
	}
	if err := domain.ValidateAppointmentCreationSlot(slotWithExpiry(baseSlot, nil), baseAppt, now); err == nil {
		t.Fatal("expected nil lock expiry to be rejected")
	}
}

type fakeEligibilityRepository struct {
	snapshot *PaymentEligibilitySnapshot
	err      error
}

func (r *fakeEligibilityRepository) GetPaymentEligibilitySnapshot(command GetPaymentEligibilityCommand) (*PaymentEligibilitySnapshot, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.snapshot, nil
}

func (r *fakeEligibilityRepository) GetAppointmentByID(appointmentID string) (*domain.Appointment, error) {
	return nil, nil
}

func (r *fakeEligibilityRepository) GetAppointmentBySlotID(slotID string) (*domain.Appointment, error) {
	return nil, nil
}

func (r *fakeEligibilityRepository) CancelAppointmentByExpert(appointmentID string, reason string) error {
	return nil
}

func (r *fakeEligibilityRepository) CancelAppointmentByPatient(appointmentID string, patientID string, reason string) error {
	return nil
}

func (r *fakeEligibilityRepository) GetAppointmentsByPatient(patientID string) ([]domain.Appointment, error) {
	return nil, nil
}

func (r *fakeEligibilityRepository) GetAppointmentsByExpert(expertID string, fromDate, toDate int64, status *domain.AppointmentStatus) ([]domain.Appointment, error) {
	return nil, nil
}

func (r *fakeEligibilityRepository) LockSlot(slotID string, patientID string) error {
	return nil
}

func (r *fakeEligibilityRepository) CreateAppointment(appointment *domain.Appointment) error {
	return nil
}

func (r *fakeEligibilityRepository) ConfirmPayment(appointmentID string) error {
	return nil
}

func (r *fakeEligibilityRepository) HandlePaymentResult(command HandlePaymentResultCommand) error {
	return nil
}

func (r *fakeEligibilityRepository) CancelExpiredLocks() (int64, error) {
	return 0, nil
}

func appointmentWithStatus(appt domain.Appointment, status domain.AppointmentStatus) domain.Appointment {
	appt.Status = status
	return appt
}

func slotWithStatus(slot domain.ExpertSlot, status domain.SlotStatus) domain.ExpertSlot {
	slot.Status = status
	return slot
}

func slotWithExpert(slot domain.ExpertSlot, expertID string) domain.ExpertSlot {
	slot.ExpertID = expertID
	return slot
}

func slotWithLockedBy(slot domain.ExpertSlot, lockedBy *string) domain.ExpertSlot {
	slot.LockedBy = lockedBy
	return slot
}

func slotWithExpiry(slot domain.ExpertSlot, expiry *int64) domain.ExpertSlot {
	slot.LockedExpiresAt = expiry
	return slot
}

func int64Ptr(v int64) *int64 {
	return &v
}
