// File: internal/infrastructure/http/handlers/expert_handler.go
package handlers

import (
	"net/http"
	"strings"

	"profile-service/config"
	"profile-service/internal/infrastructure/http/middleware"
	"profile-service/internal/infrastructure/http/response"
	"profile-service/internal/infrastructure/persistence/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ListExperts trả về danh sách chuyên gia có phân trang (public - dùng cho màn hình đặt lịch).
// @Summary      Danh sách chuyên gia
// @Tags         experts
// @Produce      json
// @Param        page      query int    false "Trang" default(1)
// @Param        page_size query int    false "Số dòng/trang" default(20)
// @Param        search    query string false "Tìm theo tên"
// @Param        specialization_id query string false "Lọc theo chuyên khoa (spec_id)"
// @Success      200 {object} response.BaseResponse{result=response.PageResult[models.Profile]}
// @Failure      400 {object} response.BaseResponse
// @Router       /api/v1/profiles/experts [get]
func ListExperts(c *gin.Context) {
	pagination := parsePagination(c)

	query := config.DB.Model(&models.Profile{}).Where("role = ?", models.RoleExpert)
	if search := strings.TrimSpace(c.Query("search")); search != "" {
		query = query.Where("full_name ILIKE ?", "%"+search+"%")
	}
	if rawSpecID := strings.TrimSpace(c.Query("specialization_id")); rawSpecID != "" {
		specID, err := uuid.Parse(rawSpecID)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Mã chuyên khoa không hợp lệ", err.Error())
			return
		}
		query = query.Where("id IN (?)", config.DB.Table("expert_specializations").
			Select("expert_profile_id").
			Where("spec_id = ?", specID))
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
	for i := range experts {
		sanitizePublicProfile(&experts[i])
	}

	response.Success(c, "Lấy danh sách chuyên gia thành công",
		response.NewPageResult(experts, total, pagination.Page, pagination.PageSize))
}

// ListManagedExperts trả về danh sách chuyên gia do Admin đang đăng nhập quản lý (đã duyệt).
// @Summary      [Admin] Danh sách chuyên gia tôi quản lý
// @Tags         experts
// @Security     BearerAuth
// @Produce      json
// @Param        page      query int    false "Trang" default(1)
// @Param        page_size query int    false "Số dòng/trang" default(20)
// @Param        search    query string false "Tìm theo tên"
// @Success      200 {object} response.BaseResponse{result=response.PageResult[models.Profile]}
// @Failure      401 {object} response.BaseResponse
// @Failure      403 {object} response.BaseResponse
// @Router       /api/v1/profiles/experts/managed [get]
func ListManagedExperts(c *gin.Context) {
	adminID, err := uuid.Parse(c.GetString(middleware.CtxAuthID))
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Không thể xác định danh tính người dùng", err.Error())
		return
	}
	pagination := parsePagination(c)

	query := config.DB.Model(&models.Profile{}).
		Where("role = ?", models.RoleExpert).
		Where("id IN (?)", config.DB.Table("expert_profiles").
			Select("profile_id").
			Where("managed_by_admin_id = ?", adminID))
	if search := strings.TrimSpace(c.Query("search")); search != "" {
		query = query.Where("full_name ILIKE ?", "%"+search+"%")
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

	response.Success(c, "Lấy danh sách chuyên gia thành công",
		response.NewPageResult(experts, total, pagination.Page, pagination.PageSize))
}

// GetExpert trả về chi tiết một chuyên gia theo profile id (public).
// @Summary      Xem chi tiết chuyên gia
// @Tags         experts
// @Produce      json
// @Param        id path string true "Auth Account ID"
// @Success      200 {object} response.BaseResponse
// @Failure      404 {object} response.BaseResponse
// @Router       /api/v1/profiles/experts/{id} [get]
func GetExpert(c *gin.Context) {
	profile, err := findProfileByAuthID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusNotFound, "Không tìm thấy hồ sơ", err.Error())
		return
	}

	// Sanitize sensitive patient/admin info if accessed through this public endpoint
	sanitizePublicProfile(profile)

	response.Success(c, "Lấy thông tin thành công", profile)
}
