package domain

// ExpertTimeOff - Lịch nghỉ đột xuất của chuyên gia
type ExpertTimeOff struct {
	TimeOffID     string `json:"time_off_id" gorm:"column:time_off_id;primaryKey;type:uuid"`
	ExpertID      string `json:"expert_id" gorm:"column:expert_id;type:uuid"`
	StartDatetime int64  `json:"start_datetime" gorm:"column:start_datetime"` // Unix timestamp 13 số (ms)
	EndDatetime   int64  `json:"end_datetime" gorm:"column:end_datetime"`     // Unix timestamp 13 số (ms)
	Reason        string `json:"reason" gorm:"column:reason"`
}

func (ExpertTimeOff) TableName() string { return "Booking_Expert_Time_Off" }
