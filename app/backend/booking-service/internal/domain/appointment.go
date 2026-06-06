package domain

import "time"

// Appointment - Cuộc hẹn đã đặt
type Appointment struct {
	AppointmentID      string    `json:"appointment_id" gorm:"column:appointment_id;primaryKey;type:uuid"`
	SlotID             string    `json:"slot_id" gorm:"column:slot_id;type:uuid;unique"`
	PatientID          string    `json:"patient_id" gorm:"column:patient_id;type:uuid;not null"`
	ExpertID           string    `json:"expert_id" gorm:"column:expert_id;type:uuid;not null"`
	CancellationReason string    `json:"cancellation_reason" gorm:"column:cancellation_reason"`
	Status             string    `json:"status" gorm:"column:status"` // PENDING_PAYMENT, CONFIRMED...
	MeetingLink        string    `json:"meeting_link" gorm:"column:meeting_link"`
	CreatedAt          time.Time `json:"created_at" gorm:"column:created_at"`
}

func (Appointment) TableName() string { return "Booking_Appointments" }
