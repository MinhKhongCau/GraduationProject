package domain

import "gorm.io/gorm"

type AppointmentStatus int

const (
	AppointmentStatusPendingPayment AppointmentStatus = 0 // "PENDING_PAYMENT"
	AppointmentStatusConfirmed      AppointmentStatus = 1 // "CONFIRMED"
	AppointmentStatusCancelled      AppointmentStatus = 2 // "CANCELLED"
	AppointmentStatusCompleted      AppointmentStatus = 3 // "COMPLETED"
)

func (s AppointmentStatus) String() string {
	statuses := [...]string{"PENDING_PAYMENT", "CONFIRMED", "CANCELLED", "COMPLETED"}
	if int(s) >= 0 && int(s) < len(statuses) {
		return statuses[s]
	}
	return "UNKNOWN"
}

// Appointment - Cuộc hẹn đã đặt
type Appointment struct {
	AppointmentID      string            `json:"appointment_id"      gorm:"column:appointment_id;primaryKey;type:uuid"`
	SlotID             string            `json:"slot_id"             gorm:"column:slot_id;type:uuid;index"`
	PatientID          string            `json:"patient_id"          gorm:"column:patient_id;type:uuid;not null"`
	ExpertID           string            `json:"expert_id"           gorm:"column:expert_id;type:uuid;not null"`
	CancellationReason string            `json:"cancellation_reason" gorm:"column:cancellation_reason"`
	CancelledBy        *string           `json:"cancelled_by"        gorm:"column:cancelled_by;type:varchar(50)"` // SYSTEM / PATIENT / EXPERT
	Status             AppointmentStatus `json:"status"              gorm:"column:status;type:smallint;default:0"`
	StatusLabel        string            `json:"status_label"        gorm:"-"` // Tự động điền bởi AfterFind hook
	Price              float64           `json:"price"               gorm:"-"` // Giá tiền lấy từ bảng Slot thông qua JOIN
	StartTime          int64             `json:"start_time"          gorm:"-"`
	EndTime            int64             `json:"end_time"            gorm:"-"`
	MeetingLink        string            `json:"meeting_link"        gorm:"column:meeting_link"`
	CreatedAt          int64             `json:"created_at"          gorm:"column:created_at"`   // Unix ms
	UpdatedAt          int64             `json:"updated_at"          gorm:"column:updated_at"`   // Unix ms
	ConfirmedAt        *int64            `json:"confirmed_at"        gorm:"column:confirmed_at"` // Unix ms, nullable
}

func (Appointment) TableName() string { return "Booking_Appointments" }

// AfterFind - GORM hook: tự động chạy sau mỗi lần SELECT từ DB
func (a *Appointment) AfterFind(tx *gorm.DB) error {
	a.StatusLabel = a.Status.String()
	return nil
}
