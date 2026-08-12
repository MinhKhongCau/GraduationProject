package appointment

import (
	"errors"
	"testing"

	bookingquery "booking-service/internal/booking/application/query"
	"booking-service/internal/booking/domain"
)

type fakeMedicalRecordRepo struct {
	appointment domain.Appointment
	record      *domain.MedicalRecord
	saveCalled  bool
	statusUpdated domain.AppointmentStatus
}

func (f *fakeMedicalRecordRepo) GetAppointmentByID(appointmentID string) (*domain.Appointment, error) {
	if f.appointment.AppointmentID == "" {
		return nil, ErrNotFound
	}
	return &f.appointment, nil
}
func (f *fakeMedicalRecordRepo) GetPaymentEligibilitySnapshot(command GetPaymentEligibilityCommand) (*PaymentEligibilitySnapshot, error) {
	return nil, nil
}
func (f *fakeMedicalRecordRepo) GetAppointmentBySlotID(slotID string) (*domain.Appointment, error) {
	return nil, nil
}
func (f *fakeMedicalRecordRepo) CancelAppointmentByExpert(appointmentID string, reason string) error {
	return nil
}
func (f *fakeMedicalRecordRepo) CancelAppointmentByPatient(appointmentID string, patientID string, reason string) error {
	return nil
}
func (f *fakeMedicalRecordRepo) GetAppointmentsByPatient(patientID string) ([]domain.Appointment, error) {
	return nil, nil
}
func (f *fakeMedicalRecordRepo) GetAppointmentsByExpert(expertID string, fromDate, toDate int64, status *domain.AppointmentStatus) ([]domain.Appointment, error) {
	return nil, nil
}
func (f *fakeMedicalRecordRepo) LockSlot(slotID string, patientID string) error {
	return nil
}
func (f *fakeMedicalRecordRepo) CreateAppointment(appointment *domain.Appointment) error {
	return nil
}
func (f *fakeMedicalRecordRepo) ConfirmPayment(appointmentID string) error {
	return nil
}
func (f *fakeMedicalRecordRepo) HandlePaymentResult(command HandlePaymentResultCommand) error {
	return nil
}
func (f *fakeMedicalRecordRepo) CancelExpiredLocks() (int64, error) {
	return 0, nil
}
func (f *fakeMedicalRecordRepo) UpdateAppointmentStatus(appointmentID string, status domain.AppointmentStatus) error {
	f.statusUpdated = status
	return nil
}
func (f *fakeMedicalRecordRepo) SaveMedicalRecord(record *domain.MedicalRecord) error {
	f.saveCalled = true
	f.record = record
	return nil
}
func (f *fakeMedicalRecordRepo) GetMedicalRecordByAppointmentID(appointmentID string) (*domain.MedicalRecord, error) {
	return f.record, nil
}
func (f *fakeMedicalRecordRepo) GetMedicalRecordByID(recordID string) (*domain.MedicalRecord, error) {
	return f.record, nil
}
func (f *fakeMedicalRecordRepo) ListMedicalRecordsByPatient(patientID string, limit, offset int) ([]domain.MedicalRecord, int64, error) {
	if f.record != nil {
		return []domain.MedicalRecord{*f.record}, 1, nil
	}
	return []domain.MedicalRecord{}, 0, nil
}
func (f *fakeMedicalRecordRepo) ListMedicalRecordsByExpert(expertID string, limit, offset int) ([]domain.MedicalRecord, int64, error) {
	if f.record != nil {
		return []domain.MedicalRecord{*f.record}, 1, nil
	}
	return []domain.MedicalRecord{}, 0, nil
}

func TestSaveMedicalRecord(t *testing.T) {
	repo := &fakeMedicalRecordRepo{
		appointment: domain.Appointment{
			AppointmentID: "appt-1",
			PatientID:     "patient-1",
			ExpertID:      "expert-1",
			Status:        domain.AppointmentStatusConfirmed,
		},
	}
	usecase := NewUsecase(repo)

	t.Run("Unauthorized if actor is patient", func(t *testing.T) {
		_, err := usecase.SaveMedicalRecord("patient-1", "PATIENT", "appt-1", SaveMedicalRecordCommand{})
		if !errors.Is(err, ErrUnauthorized) {
			t.Fatalf("expected ErrUnauthorized, got %v", err)
		}
	})

	t.Run("Unauthorized if expert does not own appointment", func(t *testing.T) {
		_, err := usecase.SaveMedicalRecord("expert-2", "EXPERT", "appt-1", SaveMedicalRecordCommand{})
		if !errors.Is(err, ErrUnauthorized) {
			t.Fatalf("expected ErrUnauthorized, got %v", err)
		}
	})

	t.Run("Success saves medical record and completes appointment", func(t *testing.T) {
		nextDate := int64(1700000000000)
		rec, err := usecase.SaveMedicalRecord("expert-1", "EXPERT", "appt-1", SaveMedicalRecordCommand{
			Diagnosis:           "Cảm cúm nhẹ",
			Symptoms:            "Sốt, ho, đau họng",
			ActionsToAvoid:      "Tránh uống nước đá, thức khuya",
			ActionsToTake:       "Uống đủ 2 lít nước ấm, nghỉ ngơi",
			NextAppointmentDate: &nextDate,
			NextAppointmentNote: "Tái khám sau 1 tuần nếu chưa dứt ho",
			ExpertNotes:         "Theo dõi thêm nhiệt độ",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec == nil {
			t.Fatal("expected record, got nil")
		}
		if rec.Diagnosis != "Cảm cúm nhẹ" || rec.Symptoms != "Sốt, ho, đau họng" {
			t.Fatalf("unexpected content in saved record: %+v", rec)
		}
		if repo.statusUpdated != domain.AppointmentStatusCompleted {
			t.Fatalf("expected appointment status to be completed, got %v", repo.statusUpdated)
		}
	})
}

func TestGetMedicalRecord(t *testing.T) {
	record := &domain.MedicalRecord{
		RecordID:      "rec-1",
		AppointmentID: "appt-1",
		PatientID:     "patient-1",
		ExpertID:      "expert-1",
		Diagnosis:     "Chẩn đoán mẫu",
	}
	repo := &fakeMedicalRecordRepo{
		appointment: domain.Appointment{
			AppointmentID: "appt-1",
			PatientID:     "patient-1",
			ExpertID:      "expert-1",
			Status:        domain.AppointmentStatusCompleted,
		},
		record: record,
	}
	usecase := NewUsecase(repo)

	t.Run("Patient can view own medical record", func(t *testing.T) {
		rec, err := usecase.GetMedicalRecordByAppointmentID("patient-1", "PATIENT", "appt-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.RecordID != "rec-1" {
			t.Fatalf("expected rec-1, got %v", rec.RecordID)
		}
	})

	t.Run("Other patient cannot view medical record", func(t *testing.T) {
		_, err := usecase.GetMedicalRecordByAppointmentID("patient-2", "PATIENT", "appt-1")
		if !errors.Is(err, ErrUnauthorized) {
			t.Fatalf("expected ErrUnauthorized, got %v", err)
		}
	})

	t.Run("List medical records for patient", func(t *testing.T) {
		page, err := usecase.ListMedicalRecords("patient-1", "PATIENT", bookingquery.PageRequest{Page: 0, Size: 10})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(page.Items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(page.Items))
		}
	})
}
