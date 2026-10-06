// File: internal/infrastructure/persistence/models/profile.go
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Role định danh vai trò của một Profile, đồng bộ với enum Role bên auth-service.
type Role string

const (
	RoleAdmin   Role = "ADMIN"
	RoleExpert  Role = "EXPERT"
	RolePatient Role = "PATIENT"
)

// ==========================================
// 0. THỰC THỂ GỐC (PROFILE)
// ==========================================
// Profile là thực thể chung cho mọi tài khoản đã đăng ký bên auth-service.
// auth_id trỏ về Account.accountId bên auth-service (KHÔNG có FK vật lý vì khác DB/service).
type Profile struct {
	ID              uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Slug            string          `gorm:"type:varchar(255);not null;uniqueIndex" json:"slug"`
	UserInformation UserInformation `gorm:"embedded" json:"user_information"`
	AuthID          uuid.UUID       `gorm:"type:uuid;not null;uniqueIndex" json:"auth_id"`
	Role            Role            `gorm:"type:varchar(20);not null;index" json:"role"`
	CreatedAt       time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt  `gorm:"index" json:"deleted_at,omitempty" swaggertype:"string"`

	PatientProfile *PatientProfile `gorm:"foreignKey:ProfileID;references:ID" json:"patient_profile,omitempty"`
	ExpertProfile  *ExpertProfile  `gorm:"foreignKey:ProfileID;references:ID" json:"expert_profile,omitempty"`
	AdminProfile   *AdminProfile   `gorm:"foreignKey:ProfileID;references:ID" json:"admin_profile,omitempty"`
}

// UserInformation là thông tin cá nhân dùng chung mọi vai trò, nhúng (embedded) thẳng vào bảng
// profiles nhưng trả về JSON dạng object lồng "user_information".
type UserInformation struct {
	FullName    string     `gorm:"column:full_name;type:varchar(255);not null" json:"full_name"`
	DateOfBirth *time.Time `gorm:"column:date_of_birth;type:date" json:"date_of_birth"`
	Gender      string     `gorm:"column:gender;type:varchar(20)" json:"gender"`
	PhoneNumber string     `gorm:"column:phone_number;type:varchar(20)" json:"phone_number"`
	Country     string     `gorm:"column:country;type:varchar(100)" json:"country"`
}

// ==========================================
// 1. HỒ SƠ BỆNH NHÂN (PATIENT PROFILE)
// ==========================================
// Dùng chung khóa chính với Profile (quan hệ 1-1 kiểu joined-table).
// Họ tên, ngày sinh, giới tính, SĐT nằm ở Profile.UserInformation.
type PatientProfile struct {
	ProfileID uuid.UUID `gorm:"type:uuid;primary_key" json:"profile_id"`
	Email     string    `gorm:"type:varchar(255)" json:"email"`
	AvatarURL string    `gorm:"type:text" json:"avatar_url"`
	Address   string    `gorm:"type:text" json:"address"`

	MedicalHistories []MedicalHistory `gorm:"foreignKey:PatientProfileID" json:"medical_histories,omitempty"`
}

// ==========================================
// 2. TIỀN SỬ BỆNH (MEDICAL HISTORY)
// ==========================================
type MedicalHistory struct {
	HistoryID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"history_id"`
	PatientProfileID uuid.UUID `gorm:"type:uuid;not null;index" json:"patient_profile_id"`
	ConditionName    string    `gorm:"type:varchar(255);not null" json:"condition_name"`
	Description      string    `gorm:"type:text" json:"description"`
	DiagnosedAt      time.Time `gorm:"type:date" json:"diagnosed_at"`
	IsChronic        bool      `gorm:"default:false" json:"is_chronic"`
	IsActive         bool      `gorm:"default:true" json:"is_active"`
}

// ==========================================
// 3. HỒ SƠ CHUYÊN GIA (EXPERT PROFILE)
// ==========================================
type ExpertProfile struct {
	ProfileID            uuid.UUID `gorm:"type:uuid;primary_key" json:"profile_id"`
	Email                string    `gorm:"type:varchar(255)" json:"email"`
	AvatarURL            string    `gorm:"type:text" json:"avatar_url"`
	IntroductionVideoURL string    `gorm:"type:text" json:"introduction_video_url"`
	Bio                  string    `gorm:"type:text" json:"bio"`
	VerificationStatus   string    `gorm:"type:varchar(50);default:'UNVERIFIED'" json:"verification_status"`

	// Quan hệ N-N: Chuyên gia và Chuyên khoa (GORM tự tạo bảng trung gian)
	Specializations []Specialization `gorm:"many2many:expert_specializations;joinForeignKey:ExpertProfileID;joinReferences:SpecID" json:"specializations,omitempty"`
}

// ==========================================
// 4. CHUYÊN KHOA (SPECIALIZATION)
// ==========================================
type Specialization struct {
	SpecID      uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"spec_id"`
	Code        string    `gorm:"type:varchar(50);not null;uniqueIndex" json:"code"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Slug        string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"slug"`
	Description string    `gorm:"type:text" json:"description"`
	Symptoms    []string  `gorm:"type:jsonb;not null;default:'[]';serializer:json" json:"symptoms"`
	Location    string    `gorm:"type:varchar(255)" json:"location"`
	ImageURL    string    `gorm:"type:text" json:"image_url"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
}

// ==========================================
// 5. HỒ SƠ QUẢN TRỊ VIÊN (ADMIN PROFILE)
// ==========================================
type AdminProfile struct {
	ProfileID uuid.UUID `gorm:"type:uuid;primary_key" json:"profile_id"`
	Email     string    `gorm:"type:varchar(255)" json:"email"`
	Note      string    `gorm:"type:text" json:"note"`
}
