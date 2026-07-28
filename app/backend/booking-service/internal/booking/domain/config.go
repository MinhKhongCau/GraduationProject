package domain

// TimeTemplate - Cấu hình khung giờ (Ca sáng, Ca chiều)
// StartTime và EndTime lưu dưới dạng chuỗi "HH:MM" (ví dụ: "08:00", "12:00")
type TimeTemplate struct {
	TemplateID          string `json:"template_id"           gorm:"column:template_id;primaryKey;type:uuid"`
	ShiftName           string `json:"shift_name"             gorm:"column:shift_name"`
	StartTime           string `json:"start_time"             gorm:"column:start_time;type:varchar(5)"` // "HH:MM", ví dụ: "08:00"
	EndTime             string `json:"end_time"               gorm:"column:end_time;type:varchar(5)"`   // "HH:MM", ví dụ: "12:00"
	SlotDurationMinutes int    `json:"slot_duration_minutes" gorm:"column:slot_duration_minutes"`      // Thời lượng mỗi slot (phút)
	IsActive            bool   `json:"is_active"             gorm:"column:is_active"`
}

func (TimeTemplate) TableName() string { return "Booking_Config_Time_Templates" }

// Availability - Cấu hình lịch rảnh (Thứ 2,4,6 làm Ca sáng)
type Availability struct {
	AvailabilityID string   `json:"availability_id" gorm:"column:availability_id;primaryKey;type:uuid"`
	ExpertID       string   `json:"expert_id" gorm:"column:expert_id;type:uuid;not null"`
	TemplateID     string   `json:"template_id" gorm:"column:template_id;type:uuid;not null"`
	DayOfWeek      int      `json:"day_of_week" gorm:"column:day_of_week"` // 1=Mon...7=Sun
	IsEnabled      bool     `json:"is_enabled" gorm:"column:is_enabled"`
	EffectiveFrom  int64    `json:"effective_from" gorm:"column:effective_from"`   // Unix ms
	EffectiveUntil *int64   `json:"effective_until" gorm:"column:effective_until"` // Unix ms, nullable
	Price          *float64 `json:"price" gorm:"column:price;type:decimal(12,2)"`
}

func (Availability) TableName() string { return "Booking_Config_Availability" }
