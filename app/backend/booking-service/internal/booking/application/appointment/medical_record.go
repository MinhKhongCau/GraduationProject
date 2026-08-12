package appointment

import (
	bookingquery "booking-service/internal/booking/application/query"
	"booking-service/internal/booking/domain"
	"strings"
	"time"

	"github.com/google/uuid"
)

type SaveMedicalRecordCommand struct {
	Diagnosis           string `json:"diagnosis"`
	Symptoms            string `json:"symptoms"`
	ActionsToAvoid      string `json:"actions_to_avoid"`
	ActionsToTake       string `json:"actions_to_take"`
	TreatmentPlan       string `json:"treatment_plan"`
	NextAppointmentDate *int64 `json:"next_appointment_date"`
	NextAppointmentNote string `json:"next_appointment_note"`
	ExpertNotes         string `json:"expert_notes"`
}

func (u *appointmentUsecase) SaveMedicalRecord(actorID, actorRole, appointmentID string, cmd SaveMedicalRecordCommand) (*domain.MedicalRecord, error) {
	if actorID == "" || actorRole != "EXPERT" {
		return nil, ErrUnauthorized
	}
	if strings.TrimSpace(appointmentID) == "" {
		return nil, ErrNotFound
	}
	appt, err := u.repo.GetAppointmentByID(appointmentID)
	if err != nil {
		return nil, err
	}
	if appt.ExpertID != actorID {
		return nil, ErrUnauthorized
	}
	if appt.Status == domain.AppointmentStatusCancelled {
		return nil, ErrCannotCancel // or invalid status
	}

	now := time.Now().UnixMilli()

	// Check if a record already exists
	existingRecord, err := u.repo.GetMedicalRecordByAppointmentID(appointmentID)
	var record domain.MedicalRecord
	if err == nil && existingRecord != nil {
		record = *existingRecord
		record.Diagnosis = cmd.Diagnosis
		record.Symptoms = cmd.Symptoms
		record.ActionsToAvoid = cmd.ActionsToAvoid
		record.ActionsToTake = cmd.ActionsToTake
		record.TreatmentPlan = cmd.TreatmentPlan
		record.NextAppointmentDate = cmd.NextAppointmentDate
		record.NextAppointmentNote = cmd.NextAppointmentNote
		record.ExpertNotes = cmd.ExpertNotes
		record.UpdatedAt = now
	} else {
		record = domain.MedicalRecord{
			RecordID:            uuid.New().String(),
			AppointmentID:       appointmentID,
			PatientID:           appt.PatientID,
			ExpertID:            appt.ExpertID,
			Diagnosis:           cmd.Diagnosis,
			Symptoms:            cmd.Symptoms,
			ActionsToAvoid:      cmd.ActionsToAvoid,
			ActionsToTake:       cmd.ActionsToTake,
			TreatmentPlan:       cmd.TreatmentPlan,
			NextAppointmentDate: cmd.NextAppointmentDate,
			NextAppointmentNote: cmd.NextAppointmentNote,
			ExpertNotes:         cmd.ExpertNotes,
			CreatedAt:           now,
			UpdatedAt:           now,
		}
	}

	if err := u.repo.SaveMedicalRecord(&record); err != nil {
		return nil, err
	}

	// Update appointment status to COMPLETED if it was CONFIRMED
	if appt.Status == domain.AppointmentStatusConfirmed {
		_ = u.repo.UpdateAppointmentStatus(appointmentID, domain.AppointmentStatusCompleted)
	}

	record.AppointmentDate = appt.StartTime
	record.AppointmentPrice = appt.Price
	record.MeetingLink = appt.MeetingLink
	record.AppointmentStatus = domain.AppointmentStatusCompleted.String()

	return &record, nil
}

func (u *appointmentUsecase) GetMedicalRecordByAppointmentID(actorID, actorRole, appointmentID string) (*domain.MedicalRecord, error) {
	if actorID == "" {
		return nil, ErrUnauthorized
	}
	appt, err := u.repo.GetAppointmentByID(appointmentID)
	if err != nil {
		return nil, err
	}
	if actorRole == "PATIENT" && appt.PatientID != actorID {
		return nil, ErrUnauthorized
	}
	if actorRole == "EXPERT" && appt.ExpertID != actorID {
		return nil, ErrUnauthorized
	}

	record, err := u.repo.GetMedicalRecordByAppointmentID(appointmentID)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, ErrMedicalRecordNotFound
	}

	record.AppointmentDate = appt.StartTime
	record.AppointmentPrice = appt.Price
	record.MeetingLink = appt.MeetingLink
	record.AppointmentStatus = appt.Status.String()

	return record, nil
}

func (u *appointmentUsecase) GetMedicalRecordByID(actorID, actorRole, recordID string) (*domain.MedicalRecord, error) {
	if actorID == "" {
		return nil, ErrUnauthorized
	}
	record, err := u.repo.GetMedicalRecordByID(recordID)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, ErrMedicalRecordNotFound
	}
	if actorRole == "PATIENT" && record.PatientID != actorID {
		return nil, ErrUnauthorized
	}
	if actorRole == "EXPERT" && record.ExpertID != actorID {
		return nil, ErrUnauthorized
	}

	appt, err := u.repo.GetAppointmentByID(record.AppointmentID)
	if err == nil && appt != nil {
		record.AppointmentDate = appt.StartTime
		record.AppointmentPrice = appt.Price
		record.MeetingLink = appt.MeetingLink
		record.AppointmentStatus = appt.Status.String()
	}

	return record, nil
}

func (u *appointmentUsecase) ListMedicalRecords(actorID, actorRole string, page bookingquery.PageRequest) (bookingquery.Page[domain.MedicalRecord], error) {
	if actorID == "" || (actorRole != "PATIENT" && actorRole != "EXPERT" && actorRole != "ADMIN") {
		return bookingquery.Page[domain.MedicalRecord]{}, ErrUnauthorized
	}

	var items []domain.MedicalRecord
	var total int64
	var err error

	if actorRole == "PATIENT" {
		items, total, err = u.repo.ListMedicalRecordsByPatient(actorID, page.Size, page.Offset())
	} else if actorRole == "EXPERT" {
		items, total, err = u.repo.ListMedicalRecordsByExpert(actorID, page.Size, page.Offset())
	} else {
		// ADMIN
		items, total, err = u.repo.ListMedicalRecordsByPatient("", page.Size, page.Offset())
	}

	if err != nil {
		return bookingquery.Page[domain.MedicalRecord]{}, err
	}

	return bookingquery.NewPage(items, page, total), nil
}
