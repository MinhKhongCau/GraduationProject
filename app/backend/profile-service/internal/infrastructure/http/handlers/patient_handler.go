// File: internal/infrastructure/http/handlers/patient_handler.go
package handlers

import (
	"net/http"
	"strings"
	"time"

	"profile-service/config"
	"profile-service/internal/infrastructure/http/response"
	"profile-service/internal/infrastructure/http/schemas"
	"profile-service/internal/infrastructure/persistence/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// saveProfileAndSub cập nhật thông tin người dùng (Profile.UserInformation) và lưu hồ sơ con
// (Patient/Admin) trong 1 transaction.
func saveProfileAndSub(profileID uuid.UUID, info models.UserInformation, sub interface{}) error {
	return config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Profile{}).Where("id = ?", profileID).Updates(info.Columns()).Error; err != nil {
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
// @Success      200 {object} response.BaseResponse{result=response.PageResult[models.Profile]}
// @Router       /api/v1/profiles/patients [get]
func ListPatients(c *gin.Context) {
	pagination := parsePagination(c)

	query := config.DB.Model(&models.Profile{}).Where("role = ?", models.RolePatient)
	if search := strings.TrimSpace(c.Query("search")); search != "" {
		query = query.Where("full_name ILIKE ?", "%"+search+"%")
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

	response.Success(c, "Lấy danh sách bệnh nhân thành công",
		response.NewPageResult(patients, total, pagination.Page, pagination.PageSize))
}

// GetPatient trả về chi tiết một hồ sơ bệnh nhân theo profile id (dành cho Admin).
// @Summary      [Admin] Xem chi tiết bệnh nhân
// @Tags         patients
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Auth Account ID"
// @Success      200 {object} response.BaseResponse
// @Failure      404 {object} response.BaseResponse
// @Router       /api/v1/profiles/patients/{id} [get]
func GetPatient(c *gin.Context) {
	profile, err := findRoleProfileByAuthID(c.Param("id"), models.RolePatient, "PatientProfile.MedicalHistories")
	if err != nil {
		response.Error(c, http.StatusNotFound, "Không tìm thấy hồ sơ bệnh nhân", err.Error())
		return
	}
	response.Success(c, "Lấy hồ sơ thành công", profile)
}

// GetPatientPublic trả về thông tin công khai của bệnh nhân (tên, ảnh) cho chuyên gia/người dùng đã đăng nhập.
func GetPatientPublic(c *gin.Context) {
	profile, err := findRoleProfileByAuthID(c.Param("id"), models.RolePatient, "PatientProfile")
	if err != nil {
		response.Error(c, http.StatusNotFound, "Không tìm thấy hồ sơ bệnh nhân", err.Error())
		return
	}

	sanitizePublicProfile(profile)

	response.Success(c, "Lấy hồ sơ công khai thành công", profile)
}

// UpdatePatient thay thế toàn bộ (PUT) hồ sơ bệnh nhân - dành cho Admin.
// @Summary      [Admin] Cập nhật toàn bộ hồ sơ bệnh nhân
// @Tags         patients
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id      path string                       true "Auth Account ID"
// @Param        request body schemas.UpsertPatientRequest true "Hồ sơ bệnh nhân"
// @Success      200 {object} response.BaseResponse
// @Failure      400 {object} response.BaseResponse
// @Failure      404 {object} response.BaseResponse
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
// @Success      200 {object} response.BaseResponse
// @Failure      400 {object} response.BaseResponse
// @Failure      404 {object} response.BaseResponse
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

	info, err := parseUserInformation(req.UserInformation)
	if err != nil {
		writeInvalidDateOfBirth(c, err)
		return
	}

	patient := models.PatientProfile{
		ProfileID: profile.ID,
		Email:     req.Email,
		AvatarURL: req.AvatarURL,
		Address:   req.Address,
	}

	userInfo := models.NewUserInformation(info)
	if err := saveProfileAndSub(profile.ID, userInfo, &patient); err != nil {
		response.Error(c, http.StatusInternalServerError, "Không thể cập nhật hồ sơ", err.Error())
		return
	}

	profile.UserInformation = userInfo
	profile.PatientProfile = &patient
	response.Success(c, "Cập nhật hồ sơ thành công", profile)
}

func applyPatientPatch(c *gin.Context, profile *models.Profile) {
	var req schemas.PatchPatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", err.Error())
		return
	}

	patch, err := parseUserInformationPatch(req.UserInformation)
	if err != nil {
		writeInvalidDateOfBirth(c, err)
		return
	}

	patient := profile.PatientProfile
	if patient == nil {
		patient = &models.PatientProfile{ProfileID: profile.ID}
	}

	userInfo := models.NewUserInformation(profile.UserInformation.ToDomain().Apply(patch))
	if req.Email != nil {
		patient.Email = *req.Email
	}
	if req.AvatarURL != nil {
		patient.AvatarURL = *req.AvatarURL
	}
	if req.Address != nil {
		patient.Address = *req.Address
	}

	if err := saveProfileAndSub(profile.ID, userInfo, patient); err != nil {
		response.Error(c, http.StatusInternalServerError, "Không thể cập nhật hồ sơ", err.Error())
		return
	}

	profile.UserInformation = userInfo
	profile.PatientProfile = patient
	response.Success(c, "Cập nhật hồ sơ thành công", profile)
}
