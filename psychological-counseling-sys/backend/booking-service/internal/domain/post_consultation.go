package domain

import "time"

// MedicalRecord - Hồ sơ khám bệnh sau khi kết thúc
type MedicalRecord struct {
	RecordID      string    `json:"record_id" gorm:"column:record_id;primaryKey;type:uuid"`
	AppointmentID string    `json:"appointment_id" gorm:"column:appointment_id;type:uuid;unique"`
	Symptoms      string    `json:"symptoms" gorm:"column:symptoms"`
	Diagnosis     string    `json:"diagnosis" gorm:"column:diagnosis"`
	TreatmentPlan string    `json:"treatment_plan" gorm:"column:treatment_plan"`
	ExpertNotes   string    `json:"expert_notes" gorm:"column:expert_notes"`
	CreatedAt     time.Time `json:"created_at" gorm:"column:created_at"`
}

func (MedicalRecord) TableName() string { return "Booking_Medical_Records" }

// Review - Đánh giá của bệnh nhân
type Review struct {
	ReviewID      string    `json:"review_id" gorm:"column:review_id;primaryKey;type:uuid"`
	AppointmentID string    `json:"appointment_id" gorm:"column:appointment_id;type:uuid;unique"`
	Rating        int       `json:"rating" gorm:"column:rating"`
	Comment       string    `json:"comment" gorm:"column:comment"`
	IsVisible     bool      `json:"is_visible" gorm:"column:is_visible"`
	CreatedAt     time.Time `json:"created_at" gorm:"column:created_at"`
}

func (Review) TableName() string { return "Booking_Reviews" }
