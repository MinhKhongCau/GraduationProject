package appointment

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
	SpecializationID   *string           `json:"specialization_id"   gorm:"column:specialization_id;type:uuid"`
	SpecializationName string            `json:"specialization_name" gorm:"column:specialization_name;type:varchar(255)"`
	Patient            PatientSnapshot   `json:"patient"             gorm:"embedded;embeddedPrefix:patient_"`
	// Expert / PatientAccount: hồ sơ chuyên gia và tài khoản đã đặt lịch, lấy từ profile-service qua
	// gRPC mỗi lần đọc lịch hẹn (không lưu DB). nil khi profile-service không trả về.
	Expert         *ParticipantProfile `json:"expert,omitempty"          gorm:"-"`
	PatientAccount *ParticipantProfile `json:"patient_account,omitempty" gorm:"-"`
	CreatedAt      int64               `json:"created_at"          gorm:"column:created_at"`   // Unix ms
	UpdatedAt      int64               `json:"updated_at"          gorm:"column:updated_at"`   // Unix ms
	ConfirmedAt    *int64              `json:"confirmed_at"        gorm:"column:confirmed_at"` // Unix ms, nullable
}

// PatientSnapshot - Hồ sơ người khám tại thời điểm đặt lịch, lấy từ profile-service qua gRPC.
// Lưu bản sao để cuộc hẹn không đổi khi hồ sơ gốc bị sửa sau này. Cuộc hẹn cũ để trống.
type PatientSnapshot struct {
	RecordID     *string `json:"record_id"     gorm:"column:record_id;type:uuid"`
	FullName     string  `json:"full_name"     gorm:"column:full_name;type:varchar(255)"`
	DateOfBirth  string  `json:"date_of_birth" gorm:"column:date_of_birth;type:varchar(10)"` // YYYY-MM-DD
	Gender       string  `json:"gender"        gorm:"column:gender;type:varchar(20)"`
	PhoneNumber  string  `json:"phone_number"  gorm:"column:phone_number;type:varchar(20)"`
	Email        string  `json:"email"         gorm:"column:email;type:varchar(255)"`
	Relationship string  `json:"relationship"  gorm:"column:relationship;type:varchar(20)"`
}

// ParticipantProfile - Thông tin hiển thị hiện tại của chuyên gia hoặc tài khoản đặt lịch, lấy từ
// profile-service (gRPC GetProfileSummaries). Khác PatientSnapshot: không chụp lại lúc đặt lịch mà
// luôn phản ánh hồ sơ mới nhất.
type ParticipantProfile struct {
	AuthID             string `json:"auth_id"`
	Role               string `json:"role"`
	FullName           string `json:"full_name"`
	AvatarURL          string `json:"avatar_url,omitempty"`
	Email              string `json:"email,omitempty"`
	PhoneNumber        string `json:"phone_number,omitempty"`
	VerificationStatus string `json:"verification_status,omitempty"`
}

// AttachProfiles gắn hồ sơ chuyên gia và tài khoản đặt lịch từ kết quả tra cứu theo auth id.
func (a *Appointment) AttachProfiles(profiles map[string]ParticipantProfile) {
	if profile, ok := profiles[a.ExpertID]; ok {
		a.Expert = &profile
	}
	if profile, ok := profiles[a.PatientID]; ok {
		a.PatientAccount = &profile
	}
}

func (Appointment) TableName() string { return "Booking_Appointments" }

// AfterFind - GORM hook: tự động chạy sau mỗi lần SELECT từ DB
func (a *Appointment) AfterFind(tx *gorm.DB) error {
	a.StatusLabel = a.Status.String()
	return nil
}
