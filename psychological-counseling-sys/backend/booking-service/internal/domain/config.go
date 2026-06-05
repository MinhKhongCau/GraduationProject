package domain

import "time"

// TimeTemplate - Cấu hình khung giờ (Ca sáng, Ca chiều)
type TimeTemplate struct {
	TemplateID string    `gorm:"column:template_id;primaryKey;type:uuid"`
	ShiftName  string    `gorm:"column:shift_name"`
	StartTime  time.Time `gorm:"column:start_time"` // Sửa thành time.Time
	EndTime    time.Time `gorm:"column:end_time"`   // Sửa thành time.Time
	IsActive   bool      `gorm:"column:is_active"`
}

func (TimeTemplate) TableName() string { return "Booking_Config_Time_Templates" }

// Availability - Cấu hình lịch rảnh (Thứ 2,4,6 làm Ca sáng)
type Availability struct {
	AvailabilityID string `json:"availability_id" gorm:"column:availability_id;primaryKey;type:uuid"`
	ExpertID       string `json:"expert_id" gorm:"column:expert_id;type:uuid;not null"`
	TemplateID     string `json:"template_id" gorm:"column:template_id;type:uuid;not null"`
	DayOfWeek      int    `json:"day_of_week" gorm:"column:day_of_week"` // 1=Mon...7=Sun
	IsEnabled      bool   `json:"is_enabled" gorm:"column:is_enabled"`
}

func (Availability) TableName() string { return "Booking_Config_Availability" }
