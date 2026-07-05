package domain

type AppointmentStatus int

const (
	AppointmentStatusPendingPayment AppointmentStatus = 0 // "PENDING_PAYMENT"
	AppointmentStatusConfirmed      AppointmentStatus = 1 // "CONFIRMED"
	AppointmentStatusCancelled      AppointmentStatus = 2 // "CANCELLED"
)

func (s AppointmentStatus) String() string {
	return [...]string{"PENDING_PAYMENT", "CONFIRMED", "CANCELLED"}[s]
}

// Appointment - Cuộc hẹn đã đặt
type Appointment struct {
	AppointmentID      string            `json:"appointment_id"      gorm:"column:appointment_id;primaryKey;type:uuid"`
	SlotID             string            `json:"slot_id"             gorm:"column:slot_id;type:uuid;unique"`
	PatientID          string            `json:"patient_id"          gorm:"column:patient_id;type:uuid;not null"`
	ExpertID           string            `json:"expert_id"           gorm:"column:expert_id;type:uuid;not null"`
	CancellationReason string            `json:"cancellation_reason" gorm:"column:cancellation_reason"`
	Status             AppointmentStatus `json:"status"              gorm:"column:status;type:smallint;default:0"`
	StatusLabel        string            `json:"status_label"        gorm:"-"` // Tự động điền bởi AfterFind hook
	MeetingLink        string            `json:"meeting_link"        gorm:"column:meeting_link"`
	CreatedAt          int64             `json:"created_at"          gorm:"column:created_at"` // Unix ms
}

func (Appointment) TableName() string { return "Booking_Appointments" }

// AfterFind - GORM hook: tự động chạy sau mỗi lần SELECT từ DB
func (a *Appointment) AfterFind(tx interface{}) error {
	a.StatusLabel = a.Status.String()
	return nil
}
