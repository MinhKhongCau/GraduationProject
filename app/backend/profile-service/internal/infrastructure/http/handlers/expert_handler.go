// File: internal/infrastructure/http/handlers/expert_handler.go
package handlers

import (
	"net/http"
	"strings"

	"profile-service/config"
	"profile-service/internal/infrastructure/http/response"
	"profile-service/internal/infrastructure/http/schemas"
	"profile-service/internal/infrastructure/persistence/models"

	"github.com/gin-gonic/gin"
)

// ListExperts trả về danh sách chuyên gia có phân trang (public - dùng cho màn hình đặt lịch).
// @Summary      Danh sách chuyên gia
// @Tags         experts
// @Produce      json
// @Param        page      query int    false "Trang" default(1)
// @Param        page_size query int    false "Số dòng/trang" default(20)
// @Param        search    query string false "Tìm theo tên"
// @Success      200 {object} response.Response
// @Router       /api/v1/profiles/experts [get]
func ListExperts(c *gin.Context) {
	pagination := parsePagination(c)

	query := config.DB.Model(&models.Profile{}).Where("role = ?", models.RoleExpert)
	if search := strings.TrimSpace(c.Query("search")); search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}

	var total int64
	query.Count(&total)

	var experts []models.Profile
	if err := query.Preload("ExpertProfile.Specializations").
		Order("created_at DESC").
		Limit(pagination.PageSize).
		Offset((pagination.Page - 1) * pagination.PageSize).
		Find(&experts).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi truy vấn danh sách chuyên gia", err.Error())
		return
	}

	response.Success(c, "Lấy danh sách chuyên gia thành công", schemas.PaginatedResponse{
		Items:      experts,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalItems: total,
		TotalPages: totalPages(total, pagination.PageSize),
	})
}

// GetExpert trả về chi tiết một chuyên gia theo profile id (public).
// @Summary      Xem chi tiết chuyên gia
// @Tags         experts
// @Produce      json
// @Param        id path string true "Auth Account ID"
// @Success      200 {object} response.Response
// @Failure      404 {object} response.Response
// @Router       /api/v1/profiles/experts/{id} [get]
func GetExpert(c *gin.Context) {
	profile, err := findProfileByAuthID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusNotFound, "Không tìm thấy hồ sơ", err.Error())
		return
	}

	// Sanitize sensitive patient/admin info if accessed through this public endpoint
	if profile.Role == models.RolePatient && profile.PatientProfile != nil {
		profile.PatientProfile.PhoneNumber = ""
		profile.PatientProfile.Email = ""
		profile.PatientProfile.Address = ""
		profile.PatientProfile.DateOfBirth = nil
		profile.PatientProfile.Gender = ""
		profile.PatientProfile.MedicalHistories = nil
	} else if profile.Role == models.RoleAdmin && profile.AdminProfile != nil {
		profile.AdminProfile.Email = ""
	}

	response.Success(c, "Lấy thông tin thành công", profile)
}
