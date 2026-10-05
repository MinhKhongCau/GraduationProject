package appointment

// MedicalRecord - Hồ sơ khám bệnh sau khi kết thúc tư vấn
type MedicalRecord struct {
	RecordID            string `json:"record_id" gorm:"column:record_id;primaryKey;type:uuid"`
	AppointmentID       string `json:"appointment_id" gorm:"column:appointment_id;type:uuid;unique;index"`
	PatientID           string `json:"patient_id" gorm:"column:patient_id;type:uuid;index"`
	ExpertID            string `json:"expert_id" gorm:"column:expert_id;type:uuid;index"`
	Diagnosis           string `json:"diagnosis" gorm:"column:diagnosis"`                                   // Tình trạng bệnh / Chẩn đoán
	Symptoms            string `json:"symptoms" gorm:"column:symptoms"`                                     // Triệu chứng
	ActionsToAvoid      string `json:"actions_to_avoid" gorm:"column:actions_to_avoid"`                     // Các hành động cần tránh
	ActionsToTake       string `json:"actions_to_take" gorm:"column:actions_to_take"`                       // Các hành động cần làm
	TreatmentPlan       string `json:"treatment_plan" gorm:"column:treatment_plan"`                         // Phác đồ / Kế hoạch điều trị
	NextAppointmentDate *int64 `json:"next_appointment_date,omitempty" gorm:"column:next_appointment_date"` // Buổi gặp tiếp theo (Unix ms, nullable)
	NextAppointmentNote string `json:"next_appointment_note" gorm:"column:next_appointment_note"`           // Ghi chú buổi gặp tiếp theo
	ExpertNotes         string `json:"expert_notes" gorm:"column:expert_notes"`                             // Ghi chú thêm của expert
	CreatedAt           int64  `json:"created_at" gorm:"column:created_at"`                                 // Unix timestamp 13 số (ms)
	UpdatedAt           int64  `json:"updated_at" gorm:"column:updated_at"`                                 // Unix timestamp 13 số (ms)

	// Extra fields joined from appointment / slot for response convenience
	AppointmentDate   int64   `json:"appointment_date,omitempty" gorm:"-"`
	AppointmentPrice  float64 `json:"appointment_price,omitempty" gorm:"-"`
	MeetingLink       string  `json:"meeting_link,omitempty" gorm:"-"`
	AppointmentStatus string  `json:"appointment_status,omitempty" gorm:"-"`
}

func (MedicalRecord) TableName() string { return "Booking_Medical_Records" }

// Review - Đánh giá của bệnh nhân sau buổi tư vấn
type Review struct {
	ReviewID      string `json:"review_id" gorm:"column:review_id;primaryKey;type:uuid"`
	AppointmentID string `json:"appointment_id" gorm:"column:appointment_id;type:uuid;unique"`
	Rating        int    `json:"rating" gorm:"column:rating"`
	Comment       string `json:"comment" gorm:"column:comment"`
	IsVisible     bool   `json:"is_visible" gorm:"column:is_visible"`
	CreatedAt     int64  `json:"created_at" gorm:"column:created_at"` // Unix timestamp 13 số (ms)
}

func (Review) TableName() string { return "Booking_Reviews" }
