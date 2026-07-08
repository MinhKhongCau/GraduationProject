// File: internal/handlers/patient_handler.go
package handlers

import (
	"net/http"
	"strings"
	"time"

	"profile-service/config"
	"profile-service/internal/models"
	"profile-service/internal/schemas"
	"profile-service/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// saveProfileAndSub cập nhật tên (Profile.Name) và lưu hồ sơ con (Patient/Expert/Admin) trong 1 transaction.
func saveProfileAndSub(profileID uuid.UUID, name string, sub interface{}) error {
	return config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Profile{}).Where("id = ?", profileID).Update("name", name).Error; err != nil {
			return err
		}
		return tx.Save(sub).Error
	})
}

func parseOptionalDate(value string) (*time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

// ListPatients trả về danh sách bệnh nhân có phân trang (dành cho Admin).
// @Summary      [Admin] Danh sách bệnh nhân
// @Tags         patients
// @Security     BearerAuth
// @Produce      json
// @Param        page      query int    false "Trang" default(1)
// @Param        page_size query int    false "Số dòng/trang" default(20)
// @Param        search    query string false "Tìm theo tên"
// @Success      200 {object} response.Response
// @Router       /api/v1/profiles/patients [get]
func ListPatients(c *gin.Context) {
	pagination := parsePagination(c)

	query := config.DB.Model(&models.Profile{}).Where("role = ?", models.RolePatient)
	if search := strings.TrimSpace(c.Query("search")); search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}

	var total int64
	query.Count(&total)

	var patients []models.Profile
	if err := query.Preload("PatientProfile").
		Order("created_at DESC").
		Limit(pagination.PageSize).
		Offset((pagination.Page - 1) * pagination.PageSize).
		Find(&patients).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi truy vấn danh sách bệnh nhân", err.Error())
		return
	}

	response.Success(c, "Lấy danh sách bệnh nhân thành công", schemas.PaginatedResponse{
		Items:      patients,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalItems: total,
		TotalPages: totalPages(total, pagination.PageSize),
	})
}

// GetPatient trả về chi tiết một hồ sơ bệnh nhân theo profile id (dành cho Admin).
// @Summary      [Admin] Xem chi tiết bệnh nhân
// @Tags         patients
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Auth Account ID"
// @Success      200 {object} response.Response
// @Failure      404 {object} response.Response
// @Router       /api/v1/profiles/patients/{id} [get]
func GetPatient(c *gin.Context) {
	profile, err := findRoleProfileByAuthID(c.Param("id"), models.RolePatient, "PatientProfile.MedicalHistories")
	if err != nil {
		response.Error(c, http.StatusNotFound, "Không tìm thấy hồ sơ bệnh nhân", err.Error())
		return
	}
	response.Success(c, "Lấy hồ sơ thành công", profile)
}

// UpdatePatient thay thế toàn bộ (PUT) hồ sơ bệnh nhân - dành cho Admin.
// @Summary      [Admin] Cập nhật toàn bộ hồ sơ bệnh nhân
// @Tags         patients
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id      path string                       true "Auth Account ID"
// @Param        request body schemas.UpsertPatientRequest true "Hồ sơ bệnh nhân"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      404 {object} response.Response
// @Router       /api/v1/profiles/patients/{id} [put]
func UpdatePatient(c *gin.Context) {
	profile, err := findRoleProfileByAuthID(c.Param("id"), models.RolePatient, "")
	if err != nil {
		response.Error(c, http.StatusNotFound, "Không tìm thấy hồ sơ bệnh nhân", err.Error())
		return
	}
	applyPatientUpsert(c, profile)
}

// PatchPatient cập nhật một phần (PATCH) hồ sơ bệnh nhân - dành cho Admin.
// @Summary      [Admin] Cập nhật một phần hồ sơ bệnh nhân
// @Tags         patients
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id      path string                     true "Auth Account ID"
// @Param        request body schemas.PatchPatientRequest true "Các trường cần cập nhật"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      404 {object} response.Response
// @Router       /api/v1/profiles/patients/{id} [patch]
func PatchPatient(c *gin.Context) {
	profile, err := findRoleProfileByAuthID(c.Param("id"), models.RolePatient, "PatientProfile")
	if err != nil {
		response.Error(c, http.StatusNotFound, "Không tìm thấy hồ sơ bệnh nhân", err.Error())
		return
	}
	applyPatientPatch(c, profile)
}

func applyPatientUpsert(c *gin.Context, profile *models.Profile) {
	var req schemas.UpsertPatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", err.Error())
		return
	}

	dob, err := parseOptionalDate(req.DateOfBirth)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Định dạng ngày sinh không hợp lệ (YYYY-MM-DD)", err.Error())
		return
	}

	patient := models.PatientProfile{
		ProfileID:   profile.ID,
		PhoneNumber: req.PhoneNumber,
		Email:       req.Email,
		AvatarURL:   req.AvatarURL,
		DateOfBirth: dob,
		Gender:      req.Gender,
		Address:     req.Address,
	}

	if err := saveProfileAndSub(profile.ID, req.Name, &patient); err != nil {
		response.Error(c, http.StatusInternalServerError, "Không thể cập nhật hồ sơ", err.Error())
		return
	}

	profile.Name = req.Name
	profile.PatientProfile = &patient
	response.Success(c, "Cập nhật hồ sơ thành công", profile)
}

func applyPatientPatch(c *gin.Context, profile *models.Profile) {
	var req schemas.PatchPatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", err.Error())
		return
	}

	patient := profile.PatientProfile
	if patient == nil {
		patient = &models.PatientProfile{ProfileID: profile.ID}
	}

	newName := profile.Name
	if req.Name != nil {
		newName = *req.Name
	}
	if req.PhoneNumber != nil {
		patient.PhoneNumber = *req.PhoneNumber
	}
	if req.Email != nil {
		patient.Email = *req.Email
	}
	if req.AvatarURL != nil {
		patient.AvatarURL = *req.AvatarURL
	}
	if req.DateOfBirth != nil {
		dob, err := parseOptionalDate(*req.DateOfBirth)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Định dạng ngày sinh không hợp lệ (YYYY-MM-DD)", err.Error())
			return
		}
		patient.DateOfBirth = dob
	}
	if req.Gender != nil {
		patient.Gender = *req.Gender
	}
	if req.Address != nil {
		patient.Address = *req.Address
	}

	if err := saveProfileAndSub(profile.ID, newName, patient); err != nil {
		response.Error(c, http.StatusInternalServerError, "Không thể cập nhật hồ sơ", err.Error())
		return
	}

	profile.Name = newName
	profile.PatientProfile = patient
	response.Success(c, "Cập nhật hồ sơ thành công", profile)
}
