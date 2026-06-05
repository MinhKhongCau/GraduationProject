package models

import (
	"time"

	"github.com/google/uuid"
)

// ==========================================
// 1. THỰC THỂ BỆNH NHÂN (PATIENT)
// ==========================================
type Patient struct {
	PatientID uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	AccountID uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_account_name_dob"`

	FullName string `gorm:"type:varchar(255);not null;uniqueIndex:idx_account_name_dob"`

	PhoneNumber string `gorm:"type:varchar(20)"`
	Email       string `gorm:"type:varchar(255)"`
	AvatarURL   string `gorm:"type:text"`

	DateOfBirth time.Time `gorm:"type:date;uniqueIndex:idx_account_name_dob"`

	Gender    string    `gorm:"type:varchar(20)"`
	Address   string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"autoCreateTime"`

	MedicalHistories []MedicalHistory `gorm:"foreignKey:PatientID"`
}

// ==========================================
// 2. TIỀN SỬ BỆNH (MEDICAL HISTORY)
// ==========================================
type MedicalHistory struct {
	HistoryID     uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	PatientID     uuid.UUID `gorm:"type:uuid;not null;index"`
	ConditionName string    `gorm:"type:varchar(255);not null"`
	Description   string    `gorm:"type:text"`
	DiagnosedAt   time.Time `gorm:"type:date"`
	IsChronic     bool      `gorm:"default:false"`
	IsActive      bool      `gorm:"default:true"`
}

// ==========================================
// 3. THỰC THỂ CHUYÊN GIA (EXPERT)
// ==========================================
type Expert struct {
	ExpertID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	AccountID            uuid.UUID `gorm:"type:uuid;uniqueIndex;not null"`
	FullName             string    `gorm:"type:varchar(255);not null"`
	PhoneNumber          string    `gorm:"type:varchar(20)"`
	Email                string    `gorm:"type:varchar(255)"`
	AvatarURL            string    `gorm:"type:text"`
	IntroductionVideoURL string    `gorm:"type:text"`
	Bio                  string    `gorm:"type:text"`
	VerificationStatus   string    `gorm:"type:varchar(50);default:'UNVERIFIED'"`
	CreatedAt            time.Time `gorm:"autoCreateTime"`

	// Quan hệ N-N: Chuyên gia và Chuyên khoa (GORM tự tạo bảng trung gian)
	Specializations []Specialization `gorm:"many2many:expert_specs;joinForeignKey:ExpertID;joinReferences:SpecID"`
}

// ==========================================
// 4. CHUYÊN KHOA (SPECIALIZATION)
// ==========================================
type Specialization struct {
	SpecID      uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string    `gorm:"type:varchar(255);not null"`
	Description string    `gorm:"type:text"`
	ImageURL    string    `gorm:"type:text"`
	IsActive    bool      `gorm:"default:true"`
}
