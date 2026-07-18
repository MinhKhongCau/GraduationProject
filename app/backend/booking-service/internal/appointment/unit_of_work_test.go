package appointment

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"booking-service/internal/domain"
)

func TestCreateAppointmentUsesUnitOfWorkPath(t *testing.T) {
	patientID := "patient-1"
	expertID := "expert-1"
	slotID := "slot-1"
	expiresAt := time.Now().UnixMilli() + 900_000
	repo := newUOWTestRepo(domain.Appointment{}, domain.ExpertSlot{
		SlotID:          slotID,
		ExpertID:        expertID,
		Status:          domain.SlotStatusLocked,
		LockedBy:        &patientID,
		LockedExpiresAt: &expiresAt,
	})

	appt, err := NewUsecase(repo).CreateAppointment(patientID, expertID, slotID)
	if err != nil {
		t.Fatalf("expected create success, got %v", err)
	}

	assertUOWCalled(t, repo)
	assertCalls(t, repo.tx, []string{"load_slot:" + slotID, "create_appointment"})
	if repo.tx.createdAppointment == nil {
		t.Fatal("expected appointment to be created")
	}
	if repo.tx.createdAppointment.Status != domain.AppointmentStatusPendingPayment {
		t.Fatalf("expected PENDING_PAYMENT, got %s", repo.tx.createdAppointment.Status.String())
	}
	if repo.tx.createdAppointment.PatientID != patientID || repo.tx.createdAppointment.ExpertID != expertID || repo.tx.createdAppointment.SlotID != slotID {
		t.Fatalf("created appointment has wrong identity data: %#v", repo.tx.createdAppointment)
	}
	if repo.tx.createdAppointment.CreatedAt == 0 || repo.tx.createdAppointment.UpdatedAt == 0 {
		t.Fatalf("expected timestamps to be set, got created_at=%d updated_at=%d", repo.tx.createdAppointment.CreatedAt, repo.tx.createdAppointment.UpdatedAt)
	}
	if appt.AppointmentID != repo.tx.createdAppointment.AppointmentID {
		t.Fatalf("expected returned appointment to be created appointment")
	}
}

func TestCreateAppointmentUnitOfWorkValidationPreventsPersistence(t *testing.T) {
	patientID := "patient-1"
	expertID := "expert-1"
	slotID := "slot-1"
	otherPatientID := "patient-2"
	expiresAt := time.Now().UnixMilli() + 900_000
	repo := newUOWTestRepo(domain.Appointment{}, domain.ExpertSlot{
		SlotID:          slotID,
		ExpertID:        expertID,
		Status:          domain.SlotStatusLocked,
		LockedBy:        &otherPatientID,
		LockedExpiresAt: &expiresAt,
	})

	_, err := NewUsecase(repo).CreateAppointment(patientID, expertID, slotID)
	if err == nil {
		t.Fatal("expected invalid slot lock error")
	}
	const want = "slot khÃ´ng Ä‘Æ°á»£c giá»¯ bá»Ÿi báº¡n, vui lÃ²ng thá»±c hiá»‡n láº¡i tá»« Ä‘áº§u"
	if err.Error() != want {
		t.Fatalf("expected current error message %q, got %q", want, err.Error())
	}
	assertUOWCalled(t, repo)
	assertCalls(t, repo.tx, []string{"load_slot:" + slotID})
	if repo.tx.createdAppointment != nil {
		t.Fatalf("expected no appointment creation after validation failure, got %#v", repo.tx.createdAppointment)
	}
}

func TestCreateAppointmentUnitOfWorkPreservesCurrentErrorMessages(t *testing.T) {
	patientID := "patient-1"
	expertID := "expert-1"
	slotID := "slot-1"
	expiresAt := time.Now().UnixMilli() + 900_000

	tests := []struct {
		name    string
		txSetup func(tx *fakeAppointmentTx)
		wantErr string
	}{
		{
			name: "slot load failure",
			txSetup: func(tx *fakeAppointmentTx) {
				tx.loadSlotErr = errors.New("missing")
			},
			wantErr: "khÃ´ng tÃ¬m tháº¥y slot: missing",
		},
		{
			name: "expired lock",
			txSetup: func(tx *fakeAppointmentTx) {
				expiredAt := time.Now().UnixMilli() - 1
				tx.slot.LockedExpiresAt = &expiredAt
			},
			wantErr: "phiÃªn giá»¯ chá»— Ä‘Ã£ háº¿t háº¡n 15 phÃºt, vui lÃ²ng chá»n láº¡i",
		},
		{
			name: "appointment insert failure",
			txSetup: func(tx *fakeAppointmentTx) {
				tx.createAppointmentErr = errors.New("insert failed")
			},
			wantErr: "lá»—i khi táº¡o cuá»™c háº¹n: insert failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newUOWTestRepo(domain.Appointment{}, domain.ExpertSlot{
				SlotID:          slotID,
				ExpertID:        expertID,
				Status:          domain.SlotStatusLocked,
				LockedBy:        &patientID,
				LockedExpiresAt: &expiresAt,
			})
			tt.txSetup(repo.tx)

			_, err := NewUsecase(repo).CreateAppointment(patientID, expertID, slotID)
			if err == nil {
				t.Fatal("expected error")
			}
			if err.Error() != tt.wantErr {
				t.Fatalf("expected %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestGetPaymentEligibilityUsesUnitOfWorkPath(t *testing.T) {
	payerID := "patient-1"
	expertID := "expert-1"
	appointmentID := "appt-1"
	slotID := "slot-1"
	expiresAt := time.Now().UnixMilli() + 900_000
	repo := newUOWTestRepo(domain.Appointment{
		AppointmentID: appointmentID,
		SlotID:        slotID,
		PatientID:     payerID,
		ExpertID:      expertID,
		Status:        domain.AppointmentStatusPendingPayment,
	}, domain.ExpertSlot{
		SlotID:          slotID,
		ExpertID:        expertID,
		Status:          domain.SlotStatusLocked,
		Price:           250000,
		LockedBy:        &payerID,
		LockedExpiresAt: &expiresAt,
	})

	eligibility, err := NewUsecase(repo).GetPaymentEligibility(GetPaymentEligibilityCommand{
		AppointmentID: appointmentID,
		PayerID:       payerID,
	})
	if err != nil {
		t.Fatalf("expected eligibility success, got %v", err)
	}

	assertUOWCalled(t, repo)
	assertCalls(t, repo.tx, []string{"load_appointment:" + appointmentID, "load_slot:" + slotID})
	if eligibility.AppointmentID != appointmentID || eligibility.ExpertID != expertID || eligibility.AmountVND != 250000 || eligibility.ExpiresAt != expiresAt {
		t.Fatalf("unexpected eligibility: %#v", eligibility)
	}
}

func TestGetPaymentEligibilityUnitOfWorkRejectsIneligibleSnapshot(t *testing.T) {
	payerID := "patient-1"
	expertID := "expert-1"
	appointmentID := "appt-1"
	slotID := "slot-1"
	expiresAt := time.Now().UnixMilli() + 900_000
	repo := newUOWTestRepo(domain.Appointment{
		AppointmentID: appointmentID,
		SlotID:        slotID,
		PatientID:     payerID,
		ExpertID:      expertID,
		Status:        domain.AppointmentStatusConfirmed,
	}, domain.ExpertSlot{
		SlotID:          slotID,
		ExpertID:        expertID,
		Status:          domain.SlotStatusLocked,
		Price:           250000,
		LockedBy:        &payerID,
		LockedExpiresAt: &expiresAt,
	})

	_, err := NewUsecase(repo).GetPaymentEligibility(GetPaymentEligibilityCommand{
		AppointmentID: appointmentID,
		PayerID:       payerID,
	})
	if !errors.Is(err, ErrPaymentEligibilityConflict) {
		t.Fatalf("expected ErrPaymentEligibilityConflict, got %v", err)
	}
	assertUOWCalled(t, repo)
	assertCalls(t, repo.tx, []string{"load_appointment:" + appointmentID, "load_slot:" + slotID})
	if repo.tx.updateAppointmentCalls != 0 || repo.tx.updateSlotCalls != 0 || repo.tx.createAppointmentCalls != 0 {
		t.Fatalf("expected eligibility check to avoid writes, tx=%#v", repo.tx)
	}
}

func TestHandlePaymentResultUnitOfWorkSuccessAndFailureTransitions(t *testing.T) {
	tests := []struct {
		name                  string
		status                PaymentResultStatus
		wantAppointmentStatus domain.AppointmentStatus
		wantSlotStatus        domain.SlotStatus
		wantReason            interface{}
	}{
		{
			name:                  "success confirms and occupies",
			status:                PaymentResultSuccess,
			wantAppointmentStatus: domain.AppointmentStatusConfirmed,
			wantSlotStatus:        domain.SlotStatusOccupied,
		},
		{
			name:                  "failure cancels and releases",
			status:                PaymentResultFailed,
			wantAppointmentStatus: domain.AppointmentStatusCancelled,
			wantSlotStatus:        domain.SlotStatusAvailable,
			wantReason:            "Payment failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := paymentResultUOWRepo(domain.AppointmentStatusPendingPayment, domain.SlotStatusLocked)

			err := NewUsecase(repo).HandlePaymentResult(HandlePaymentResultCommand{
				AppointmentID: "appt-1",
				Status:        tt.status,
			})
			if err != nil {
				t.Fatalf("expected payment result success, got %v", err)
			}

			assertUOWCalled(t, repo)
			assertCalls(t, repo.tx, []string{"load_appointment:appt-1", "load_slot:slot-1", "update_appointment", "update_slot:slot-1"})
			if repo.tx.updateAppointmentCalls != 1 || repo.tx.updateSlotCalls != 1 {
				t.Fatalf("expected one appointment and slot update, got appointment=%d slot=%d", repo.tx.updateAppointmentCalls, repo.tx.updateSlotCalls)
			}
			if repo.tx.lastAppointmentUpdates["status"] != tt.wantAppointmentStatus {
				t.Fatalf("expected appointment status %s, got %#v", tt.wantAppointmentStatus.String(), repo.tx.lastAppointmentUpdates["status"])
			}
			if repo.tx.lastSlotUpdates["status"] != tt.wantSlotStatus {
				t.Fatalf("expected slot status %s, got %#v", tt.wantSlotStatus.String(), repo.tx.lastSlotUpdates["status"])
			}
			if repo.tx.lastExpectedSlotStatus == nil || *repo.tx.lastExpectedSlotStatus != domain.SlotStatusLocked {
				t.Fatalf("expected slot update to require LOCKED, got %#v", repo.tx.lastExpectedSlotStatus)
			}
			if tt.wantReason != nil && repo.tx.lastAppointmentUpdates["cancellation_reason"] != tt.wantReason {
				t.Fatalf("expected cancellation reason %q, got %#v", tt.wantReason, repo.tx.lastAppointmentUpdates["cancellation_reason"])
			}
		})
	}
}

func TestHandlePaymentResultUnitOfWorkDuplicateCallbacksAreNoop(t *testing.T) {
	tests := []struct {
		name       string
		apptStatus domain.AppointmentStatus
		slotStatus domain.SlotStatus
		status     PaymentResultStatus
	}{
		{name: "duplicate success", apptStatus: domain.AppointmentStatusConfirmed, slotStatus: domain.SlotStatusOccupied, status: PaymentResultSuccess},
		{name: "duplicate failure", apptStatus: domain.AppointmentStatusCancelled, slotStatus: domain.SlotStatusAvailable, status: PaymentResultFailed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := paymentResultUOWRepo(tt.apptStatus, tt.slotStatus)

			err := NewUsecase(repo).HandlePaymentResult(HandlePaymentResultCommand{
				AppointmentID: "appt-1",
				Status:        tt.status,
			})
			if err != nil {
				t.Fatalf("expected duplicate callback no-op, got %v", err)
			}

			assertUOWCalled(t, repo)
			assertCalls(t, repo.tx, []string{"load_appointment:appt-1", "load_slot:slot-1"})
			if repo.tx.updateAppointmentCalls != 0 || repo.tx.updateSlotCalls != 0 {
				t.Fatalf("expected no updates for duplicate callback, got appointment=%d slot=%d", repo.tx.updateAppointmentCalls, repo.tx.updateSlotCalls)
			}
		})
	}
}

func TestHandlePaymentResultUnitOfWorkRejectsInvalidTransition(t *testing.T) {
	repo := paymentResultUOWRepo(domain.AppointmentStatusCancelled, domain.SlotStatusAvailable)

	err := NewUsecase(repo).HandlePaymentResult(HandlePaymentResultCommand{
		AppointmentID: "appt-1",
		Status:        PaymentResultSuccess,
	})
	if !errors.Is(err, ErrPaymentResultConflict) {
		t.Fatalf("expected ErrPaymentResultConflict, got %v", err)
	}

	assertUOWCalled(t, repo)
	assertCalls(t, repo.tx, []string{"load_appointment:appt-1", "load_slot:slot-1"})
	if repo.tx.updateAppointmentCalls != 0 || repo.tx.updateSlotCalls != 0 {
		t.Fatalf("expected invalid transition to avoid writes, got appointment=%d slot=%d", repo.tx.updateAppointmentCalls, repo.tx.updateSlotCalls)
	}
}

type fakeUOWRepository struct {
	tx            *fakeAppointmentTx
	withinTxCalls int
}

func newUOWTestRepo(appt domain.Appointment, slot domain.ExpertSlot) *fakeUOWRepository {
	return &fakeUOWRepository{tx: &fakeAppointmentTx{
		appointment: appt,
		slot:        slot,
		slotRows:    1,
	}}
}

func paymentResultUOWRepo(apptStatus domain.AppointmentStatus, slotStatus domain.SlotStatus) *fakeUOWRepository {
	return newUOWTestRepo(domain.Appointment{
		AppointmentID: "appt-1",
		SlotID:        "slot-1",
		PatientID:     "patient-1",
		ExpertID:      "expert-1",
		Status:        apptStatus,
	}, domain.ExpertSlot{
		SlotID:   "slot-1",
		ExpertID: "expert-1",
		Status:   slotStatus,
	})
}

func (r *fakeUOWRepository) WithinTx(ctx context.Context, fn func(tx Tx) error) error {
	r.withinTxCalls++
	return fn(r.tx)
}

func (r *fakeUOWRepository) GetAppointmentByID(appointmentID string) (*domain.Appointment, error) {
	return nil, nil
}

func (r *fakeUOWRepository) GetPaymentEligibilitySnapshot(command GetPaymentEligibilityCommand) (*PaymentEligibilitySnapshot, error) {
	return nil, errors.New("legacy eligibility path should not be used")
}

func (r *fakeUOWRepository) GetAppointmentBySlotID(slotID string) (*domain.Appointment, error) {
	return nil, nil
}

func (r *fakeUOWRepository) CancelAppointmentByExpert(appointmentID string, reason string) error {
	return nil
}

func (r *fakeUOWRepository) CancelAppointmentByPatient(appointmentID string, patientID string, reason string) error {
	return nil
}

func (r *fakeUOWRepository) GetAppointmentsByPatient(patientID string) ([]domain.Appointment, error) {
	return nil, nil
}

func (r *fakeUOWRepository) GetAppointmentsByExpert(expertID string, fromDate, toDate int64, status *domain.AppointmentStatus) ([]domain.Appointment, error) {
	return nil, nil
}

func (r *fakeUOWRepository) LockSlot(slotID string, patientID string) error {
	return nil
}

func (r *fakeUOWRepository) CreateAppointment(appointment *domain.Appointment) error {
	return errors.New("legacy create path should not be used")
}

func (r *fakeUOWRepository) ConfirmPayment(appointmentID string) error {
	return errors.New("legacy confirm path should not be used")
}

func (r *fakeUOWRepository) HandlePaymentResult(command HandlePaymentResultCommand) error {
	return errors.New("legacy payment result path should not be used")
}

func (r *fakeUOWRepository) CancelExpiredLocks() (int64, error) {
	return 0, nil
}

type fakeAppointmentTx struct {
	appointment domain.Appointment
	slot        domain.ExpertSlot

	loadAppointmentErr   error
	loadSlotErr          error
	createAppointmentErr error
	updateAppointmentErr error
	updateSlotErr        error
	slotRows             int64

	calls                  []string
	createdAppointment     *domain.Appointment
	createAppointmentCalls int
	updateAppointmentCalls int
	updateSlotCalls        int
	lastAppointmentUpdates map[string]interface{}
	lastSlotUpdates        map[string]interface{}
	lastExpectedSlotStatus *domain.SlotStatus
}

func (tx *fakeAppointmentTx) LoadAppointmentForUpdate(ctx context.Context, appointmentID string) (*domain.Appointment, error) {
	tx.calls = append(tx.calls, "load_appointment:"+appointmentID)
	if tx.loadAppointmentErr != nil {
		return nil, tx.loadAppointmentErr
	}
	appt := tx.appointment
	return &appt, nil
}

func (tx *fakeAppointmentTx) LoadSlotForUpdate(ctx context.Context, slotID string) (*domain.ExpertSlot, error) {
	tx.calls = append(tx.calls, "load_slot:"+slotID)
	if tx.loadSlotErr != nil {
		return nil, tx.loadSlotErr
	}
	slot := tx.slot
	return &slot, nil
}

func (tx *fakeAppointmentTx) CreateAppointment(ctx context.Context, appointment *domain.Appointment) error {
	tx.calls = append(tx.calls, "create_appointment")
	tx.createAppointmentCalls++
	if tx.createAppointmentErr != nil {
		return tx.createAppointmentErr
	}
	appt := *appointment
	tx.createdAppointment = &appt
	return nil
}

func (tx *fakeAppointmentTx) UpdateAppointment(ctx context.Context, appointment *domain.Appointment, updates map[string]interface{}) error {
	tx.calls = append(tx.calls, "update_appointment")
	tx.updateAppointmentCalls++
	tx.lastAppointmentUpdates = cloneMap(updates)
	return tx.updateAppointmentErr
}

func (tx *fakeAppointmentTx) UpdateSlot(ctx context.Context, slotID string, expectedStatus *domain.SlotStatus, updates map[string]interface{}) (int64, error) {
	tx.calls = append(tx.calls, "update_slot:"+slotID)
	tx.updateSlotCalls++
	tx.lastSlotUpdates = cloneMap(updates)
	if expectedStatus != nil {
		status := *expectedStatus
		tx.lastExpectedSlotStatus = &status
	}
	return tx.slotRows, tx.updateSlotErr
}

func assertUOWCalled(t *testing.T, repo *fakeUOWRepository) {
	t.Helper()
	if repo.withinTxCalls != 1 {
		t.Fatalf("expected WithinTx to be called once, got %d", repo.withinTxCalls)
	}
}

func assertCalls(t *testing.T, tx *fakeAppointmentTx, want []string) {
	t.Helper()
	if !reflect.DeepEqual(tx.calls, want) {
		t.Fatalf("expected calls %v, got %v", want, tx.calls)
	}
}

func cloneMap(values map[string]interface{}) map[string]interface{} {
	cloned := make(map[string]interface{}, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}
