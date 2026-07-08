// File: internal/handlers/expert_handler.go
package handlers

import (
	"net/http"
	"strings"

	"profile-service/config"
	"profile-service/internal/middleware"
	"profile-service/internal/models"
	"profile-service/internal/schemas"
	"profile-service/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// saveExpertWithSpecializations lưu tên (Profile.Name) + hồ sơ Expert, và đồng bộ lại danh sách
// chuyên khoa nếu specIDs khác nil (nil = không gửi trường này lên, giữ nguyên; []string{} = xoá hết).
func saveExpertWithSpecializations(profileID uuid.UUID, name string, expert *models.ExpertProfile, specIDs []string) error {
	return config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Profile{}).Where("id = ?", profileID).Update("name", name).Error; err != nil {
			return err
		}
		if err := tx.Save(expert).Error; err != nil {
			return err
		}
		if specIDs != nil {
			var specs []models.Specialization
			if len(specIDs) > 0 {
				if err := tx.Where("spec_id IN ?", specIDs).Find(&specs).Error; err != nil {
					return err
				}
			}
			if err := tx.Model(expert).Association("Specializations").Replace(specs); err != nil {
				return err
			}
			expert.Specializations = specs
		}
		return nil
	})
}

func keepVerificationStatus(existing *models.ExpertProfile) string {
	if existing != nil && existing.VerificationStatus != "" {
		return existing.VerificationStatus
	}
	return "UNVERIFIED"
}

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
	profile, err := findRoleProfileByAuthID(c.Param("id"), models.RoleExpert, "ExpertProfile.Specializations")
	if err != nil {
		response.Error(c, http.StatusNotFound, "Không tìm thấy hồ sơ chuyên gia", err.Error())
		return
	}
	response.Success(c, "Lấy thông tin thành công", profile)
}

// UpdateExpert thay thế toàn bộ (PUT) hồ sơ chuyên gia - dành cho Admin.
// @Summary      [Admin] Cập nhật toàn bộ hồ sơ chuyên gia
// @Tags         experts
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id      path string                      true "Auth Account ID"
// @Param        request body schemas.UpsertExpertRequest true "Hồ sơ chuyên gia"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      404 {object} response.Response
// @Router       /api/v1/profiles/experts/{id} [put]
func UpdateExpert(c *gin.Context) {
	profile, err := findRoleProfileByAuthID(c.Param("id"), models.RoleExpert, "ExpertProfile")
	if err != nil {
		response.Error(c, http.StatusNotFound, "Không tìm thấy hồ sơ chuyên gia", err.Error())
		return
	}
	applyExpertUpsert(c, profile)
}

// PatchExpert cập nhật một phần (PATCH) hồ sơ chuyên gia - dành cho Admin.
// Chỉ Admin mới có quyền thay đổi verification_status.
// @Summary      [Admin] Cập nhật một phần hồ sơ chuyên gia
// @Tags         experts
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id      path string                     true "Auth Account ID"
// @Param        request body schemas.PatchExpertRequest true "Các trường cần cập nhật"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      403 {object} response.Response
// @Failure      404 {object} response.Response
// @Router       /api/v1/profiles/experts/{id} [patch]
func PatchExpert(c *gin.Context) {
	profile, err := findRoleProfileByAuthID(c.Param("id"), models.RoleExpert, "ExpertProfile")
	if err != nil {
		response.Error(c, http.StatusNotFound, "Không tìm thấy hồ sơ chuyên gia", err.Error())
		return
	}
	applyExpertPatch(c, profile)
}

func applyExpertUpsert(c *gin.Context, profile *models.Profile) {
	var req schemas.UpsertExpertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", err.Error())
		return
	}

	expert := models.ExpertProfile{
		ProfileID:            profile.ID,
		PhoneNumber:          req.PhoneNumber,
		Email:                req.Email,
		AvatarURL:            req.AvatarURL,
		IntroductionVideoURL: req.IntroductionVideoURL,
		Bio:                  req.Bio,
		VerificationStatus:   keepVerificationStatus(profile.ExpertProfile),
	}

	if err := saveExpertWithSpecializations(profile.ID, req.Name, &expert, req.SpecializationIDs); err != nil {
		response.Error(c, http.StatusInternalServerError, "Không thể cập nhật hồ sơ", err.Error())
		return
	}

	profile.Name = req.Name
	profile.ExpertProfile = &expert
	response.Success(c, "Cập nhật hồ sơ thành công", profile)
}

func applyExpertPatch(c *gin.Context, profile *models.Profile) {
	var req schemas.PatchExpertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", err.Error())
		return
	}

	expert := profile.ExpertProfile
	if expert == nil {
		expert = &models.ExpertProfile{ProfileID: profile.ID, VerificationStatus: "UNVERIFIED"}
	}

	newName := profile.Name
	if req.Name != nil {
		newName = *req.Name
	}
	if req.PhoneNumber != nil {
		expert.PhoneNumber = *req.PhoneNumber
	}
	if req.Email != nil {
		expert.Email = *req.Email
	}
	if req.AvatarURL != nil {
		expert.AvatarURL = *req.AvatarURL
	}
	if req.IntroductionVideoURL != nil {
		expert.IntroductionVideoURL = *req.IntroductionVideoURL
	}
	if req.Bio != nil {
		expert.Bio = *req.Bio
	}
	if req.VerificationStatus != nil {
		if c.GetString(middleware.CtxRole) != string(models.RoleAdmin) {
			response.Error(c, http.StatusForbidden, "Chỉ Admin mới có quyền thay đổi trạng thái xác minh", "Forbidden")
			return
		}
		expert.VerificationStatus = *req.VerificationStatus
	}

	if err := saveExpertWithSpecializations(profile.ID, newName, expert, req.SpecializationIDs); err != nil {
		response.Error(c, http.StatusInternalServerError, "Không thể cập nhật hồ sơ", err.Error())
		return
	}

	profile.Name = newName
	profile.ExpertProfile = expert
	response.Success(c, "Cập nhật hồ sơ thành công", profile)
}
