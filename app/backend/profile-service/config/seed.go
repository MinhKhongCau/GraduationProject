// File: config/seed.go
package config

import (
	"log"
	"time"

	"profile-service/internal/infrastructure/persistence/models"
	"profile-service/internal/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Các auth_id cố định dùng cho dữ liệu mẫu (local/dev). Khi test thật, hãy tạo tài khoản
// tương ứng bên auth-service với cùng account_id này để /me trả về đúng hồ sơ.
var (
	SeedAdminAuthID   = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	SeedExpertAuthID  = uuid.MustParse("00000000-0000-0000-0000-000000000002")
	SeedPatientAuthID = uuid.MustParse("00000000-0000-0000-0000-000000000003")
)

// SeedInitialData chèn dữ liệu khởi tạo (idempotent) cho 3 loại hồ sơ: Admin, Expert, Patient.
// Đây là bước "pour data" thay cho migration SQL thủ công - đồng bộ với cách các service Go
// khác trong dự án dùng GORM AutoMigrate thay vì thư mục migrations/ riêng.
func SeedInitialData(db *gorm.DB) {
	seedAdmin(db)
	seedExpert(db)
	seedPatient(db)
}

func seedAdmin(db *gorm.DB) {
	var count int64
	db.Model(&models.Profile{}).Where("auth_id = ?", SeedAdminAuthID).Count(&count)
	if count > 0 {
		return
	}

	profile := models.Profile{
		Slug:   utils.GenerateUniqueSlug("System Admin"),
		Name:   "System Admin",
		AuthID: SeedAdminAuthID,
		Role:   models.RoleAdmin,
		AdminProfile: &models.AdminProfile{
			Email: "admin@mindcare.local",
			Note:  "Tài khoản quản trị mặc định được seed khi khởi tạo hệ thống",
		},
	}

	if err := db.Create(&profile).Error; err != nil {
		log.Printf("⚠️  Seed Admin profile thất bại: %v", err)
		return
	}
	log.Println("🌱 Đã seed Admin profile mặc định")
}

func seedExpert(db *gorm.DB) {
	var count int64
	db.Model(&models.Profile{}).Where("auth_id = ?", SeedExpertAuthID).Count(&count)
	if count > 0 {
		return
	}

	spec := models.Specialization{
		Name:        "Tâm lý học lâm sàng",
		Description: "Chuyên khoa mẫu được seed cùng dữ liệu khởi tạo",
		IsActive:    true,
	}
	db.Where(models.Specialization{Name: spec.Name}).FirstOrCreate(&spec)

	profile := models.Profile{
		Slug:   utils.GenerateUniqueSlug("Expert Demo"),
		Name:   "Expert Demo",
		AuthID: SeedExpertAuthID,
		Role:   models.RoleExpert,
		ExpertProfile: &models.ExpertProfile{
			Email:              "expert.demo@mindcare.local",
			Bio:                "Hồ sơ chuyên gia mẫu",
			VerificationStatus: "VERIFIED",
			Specializations:    []models.Specialization{spec},
		},
	}

	if err := db.Create(&profile).Error; err != nil {
		log.Printf("⚠️  Seed Expert profile thất bại: %v", err)
		return
	}
	log.Println("🌱 Đã seed Expert profile mẫu")
}

func seedPatient(db *gorm.DB) {
	var count int64
	db.Model(&models.Profile{}).Where("auth_id = ?", SeedPatientAuthID).Count(&count)
	if count > 0 {
		return
	}

	dob := time.Date(1999, time.January, 1, 0, 0, 0, 0, time.UTC)
	profile := models.Profile{
		Slug:   utils.GenerateUniqueSlug("Patient Demo"),
		Name:   "Patient Demo",
		AuthID: SeedPatientAuthID,
		Role:   models.RolePatient,
		PatientProfile: &models.PatientProfile{
			Email:       "patient.demo@mindcare.local",
			DateOfBirth: &dob,
			Gender:      "OTHER",
		},
	}

	if err := db.Create(&profile).Error; err != nil {
		log.Printf("⚠️  Seed Patient profile thất bại: %v", err)
		return
	}
	log.Println("🌱 Đã seed Patient profile mẫu")
}
