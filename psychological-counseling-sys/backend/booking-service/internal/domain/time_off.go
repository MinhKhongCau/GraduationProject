package domain

import "time"

// ExpertTimeOff - Lịch nghỉ đột xuất
type ExpertTimeOff struct {
	TimeOffID     string    `json:"time_off_id" gorm:"column:time_off_id;primaryKey;type:uuid"`
	ExpertID      string    `json:"expert_id" gorm:"column:expert_id;type:uuid"`
	StartDatetime time.Time `json:"start_datetime" gorm:"column:start_datetime"`
	EndDatetime   time.Time `json:"end_datetime" gorm:"column:end_datetime"`
	Reason        string    `json:"reason" gorm:"column:reason"`
}

func (ExpertTimeOff) TableName() string { return "Booking_Expert_Time_Off" }
