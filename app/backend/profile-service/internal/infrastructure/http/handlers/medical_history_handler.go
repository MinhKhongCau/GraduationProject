// File: internal/infrastructure/http/handlers/medical_history_handler.go
package handlers

import (
	"net/http"

	"profile-service/config"
	"profile-service/internal/infrastructure/http/middleware"
	"profile-service/internal/infrastructure/http/response"
	"profile-service/internal/infrastructure/http/schemas"
	"profile-service/internal/infrastructure/persistence/models"

	"github.com/gin-gonic/gin"
)

// ListMyMedicalHistories trả về tiền sử bệnh của chính bệnh nhân đang đăng nhập.
// @Summary      Xem tiền sử bệnh của chính mình
// @Tags         patients
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} response.Response
// @Failure      401 {object} response.Response
// @Failure      404 {object} response.Response
// @Router       /api/v1/profiles/me/medical-histories [get]
func ListMyMedicalHistories(c *gin.Context) {
	profile, err := findProfileByAuthID(c.GetString(middleware.CtxAuthID))
	if err != nil || profile.Role != models.RolePatient {
		response.Error(c, http.StatusNotFound, "Không tìm thấy hồ sơ bệnh nhân của bạn", "not a patient profile")
		return
	}

	var histories []models.MedicalHistory
	config.DB.Where("patient_profile_id = ? AND is_active = ?", profile.ID, true).
		Order("diagnosed_at DESC").
		Find(&histories)

	response.Success(c, "Lấy danh sách tiền sử bệnh thành công", histories)
}

// AddMyMedicalHistory thêm mới một tiền sử bệnh cho chính bệnh nhân đang đăng nhập.
// @Summary      Thêm tiền sử bệnh cho chính mình
// @Tags         patients
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body schemas.CreateMedicalHistoryRequest true "Tiền sử bệnh"
// @Success      201 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      404 {object} response.Response
// @Router       /api/v1/profiles/me/medical-histories [post]
func AddMyMedicalHistory(c *gin.Context) {
	profile, err := findProfileByAuthID(c.GetString(middleware.CtxAuthID))
	if err != nil || profile.Role != models.RolePatient {
		response.Error(c, http.StatusNotFound, "Không tìm thấy hồ sơ bệnh nhân của bạn", "not a patient profile")
		return
	}

	var req schemas.CreateMedicalHistoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", err.Error())
		return
	}

	diagnosedAt, err := parseOptionalDate(req.DiagnosedAt)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Định dạng ngày chẩn đoán không hợp lệ (YYYY-MM-DD)", err.Error())
		return
	}

	history := models.MedicalHistory{
		PatientProfileID: profile.ID,
		ConditionName:    req.ConditionName,
		Description:      req.Description,
		IsChronic:        req.IsChronic,
		IsActive:         true,
	}
	if diagnosedAt != nil {
		history.DiagnosedAt = *diagnosedAt
	}

	if err := config.DB.Create(&history).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi lưu dữ liệu", err.Error())
		return
	}

	response.Created(c, "Thêm tiền sử bệnh thành công", history)
}

// ListPatientMedicalHistories trả về tiền sử bệnh của 1 bệnh nhân bất kỳ (chỉ xem - dành cho Admin).
// @Summary      [Admin] Xem tiền sử bệnh của một bệnh nhân
// @Tags         patients
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Auth Account ID"
// @Success      200 {object} response.Response
// @Failure      404 {object} response.Response
// @Router       /api/v1/profiles/patients/{id}/medical-histories [get]
func ListPatientMedicalHistories(c *gin.Context) {
	profile, err := findRoleProfileByAuthID(c.Param("id"), models.RolePatient, "")
	if err != nil {
		response.Error(c, http.StatusNotFound, "Không tìm thấy hồ sơ bệnh nhân", err.Error())
		return
	}

	var histories []models.MedicalHistory
	config.DB.Where("patient_profile_id = ? AND is_active = ?", profile.ID, true).
		Order("diagnosed_at DESC").
		Find(&histories)

	response.Success(c, "Lấy danh sách tiền sử bệnh thành công", histories)
}
