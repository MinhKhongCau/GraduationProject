package domain

// AppointmentStatus - Trạng thái cuộc hẹn (Enum)
type AppointmentStatus string

const (
	AppointmentStatusPendingPayment AppointmentStatus = "PENDING_PAYMENT"
	AppointmentStatusConfirmed      AppointmentStatus = "CONFIRMED"
	AppointmentStatusCancelled      AppointmentStatus = "CANCELLED"
)

// Appointment - Cuộc hẹn đã đặt
type Appointment struct {
	AppointmentID      string            `json:"appointment_id" gorm:"column:appointment_id;primaryKey;type:uuid"`
	SlotID             string            `json:"slot_id" gorm:"column:slot_id;type:uuid;unique"`
	PatientID          string            `json:"patient_id" gorm:"column:patient_id;type:uuid;not null"`
	ExpertID           string            `json:"expert_id" gorm:"column:expert_id;type:uuid;not null"`
	CancellationReason string            `json:"cancellation_reason" gorm:"column:cancellation_reason"`
	Status             AppointmentStatus `json:"status" gorm:"column:status;type:varchar(20)"` // PENDING_PAYMENT, CONFIRMED, CANCELLED
	MeetingLink        string            `json:"meeting_link" gorm:"column:meeting_link"`
	CreatedAt          int64             `json:"created_at" gorm:"column:created_at"` // Unix timestamp 13 số (milliseconds)
}

func (Appointment) TableName() string { return "Booking_Appointments" }
