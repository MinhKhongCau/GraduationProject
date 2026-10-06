// File: internal/infrastructure/persistence/models/patient_record.go
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Relationship là quan hệ giữa chủ tài khoản và người được khám.
type Relationship string

const (
	RelationshipSelf     Relationship = "SELF"
	RelationshipParent   Relationship = "PARENT"
	RelationshipChild    Relationship = "CHILD"
	RelationshipSpouse   Relationship = "SPOUSE"
	RelationshipSibling  Relationship = "SIBLING"
	RelationshipRelative Relationship = "RELATIVE"
	RelationshipOther    Relationship = "OTHER"
)

// MaxPatientRecordsPerOwner giới hạn số hồ sơ người khám của một tài khoản.
const MaxPatientRecordsPerOwner = 10

// ==========================================
// HỒ SƠ NGƯỜI KHÁM (PATIENT RECORD)
// ==========================================
// Một tài khoản bệnh nhân (owner_auth_id) có thể có nhiều hồ sơ người khám: hồ sơ SELF
// (sinh từ Profile của chính họ) và hồ sơ người thân. booking-service đọc hồ sơ qua gRPC
// để hiển thị trang xác nhận và lưu snapshot vào cuộc hẹn.
type PatientRecord struct {
	RecordID     uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"record_id"`
	OwnerAuthID  uuid.UUID      `gorm:"type:uuid;not null;index" json:"owner_auth_id"`
	FullName     string         `gorm:"type:varchar(255);not null" json:"full_name"`
	DateOfBirth  *time.Time     `gorm:"type:date" json:"date_of_birth"`
	Gender       string         `gorm:"type:varchar(20)" json:"gender"`
	PhoneNumber  string         `gorm:"type:varchar(20)" json:"phone_number"`
	Email        string         `gorm:"type:varchar(255)" json:"email"`
	Address      string         `gorm:"type:text" json:"address"`
	Relationship Relationship   `gorm:"type:varchar(20);not null;default:'SELF'" json:"relationship"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// NewSelfPatientRecord dựng hồ sơ SELF từ Profile (và PatientProfile nếu đã preload).
func NewSelfPatientRecord(profile *Profile) PatientRecord {
	info := profile.UserInformation
	record := PatientRecord{
		OwnerAuthID:  profile.AuthID,
		FullName:     info.FullName,
		DateOfBirth:  info.DateOfBirth,
		Gender:       info.Gender,
		PhoneNumber:  info.PhoneNumber,
		Relationship: RelationshipSelf,
	}
	if profile.PatientProfile != nil {
		record.Email = profile.PatientProfile.Email
		record.Address = profile.PatientProfile.Address
	}
	return record
}
