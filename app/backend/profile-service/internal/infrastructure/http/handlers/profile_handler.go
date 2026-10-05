// File: internal/infrastructure/http/handlers/profile_handler.go
package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"profile-service/config"
	"profile-service/internal/infrastructure/http/middleware"
	"profile-service/internal/infrastructure/http/response"
	"profile-service/internal/infrastructure/http/schemas"
	"profile-service/internal/infrastructure/messaging"
	"profile-service/internal/infrastructure/persistence/models"
	"profile-service/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ---------- Helpers dùng chung ----------

func parsePagination(c *gin.Context) schemas.PaginationQuery {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return schemas.PaginationQuery{Page: page, PageSize: pageSize}
}

func totalPages(total int64, pageSize int) int64 {
	if pageSize <= 0 {
		return 0
	}
	return (total + int64(pageSize) - 1) / int64(pageSize)
}

func findProfileByAuthID(authIDStr string) (*models.Profile, error) {
	authID, err := uuid.Parse(authIDStr)
	if err != nil {
		return nil, err
	}
	var profile models.Profile
	err = config.DB.
		Preload("PatientProfile").
		Preload("ExpertProfile.Specializations").
		Preload("AdminProfile").
		Where("auth_id = ?", authID).
		First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

// findRoleProfileByAuthID tìm 1 Profile theo auth_id (account id bên auth-service), ràng buộc
// đúng vai trò (role) mong đợi. Dùng auth_id (không phải Profile.ID nội bộ) làm định danh công khai
// vì đây chính là "expert_id"/"patient_id" mà booking-service và frontend đã dùng xuyên suốt hệ thống.
func findRoleProfileByAuthID(authIDStr string, role models.Role, preload string) (*models.Profile, error) {
	authID, err := uuid.Parse(authIDStr)
	if err != nil {
		return nil, err
	}
	var profile models.Profile
	query := config.DB.Where("auth_id = ? AND role = ?", authID, role)
	if preload != "" {
		query = query.Preload(preload)
	}
	if err := query.First(&profile).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

// CreateProfileCore builds a Profile + its role-specific sub-record and persists it via GORM.
// Exported so internal/consumer can reuse the exact same logic when reacting to the
// user.created RabbitMQ event (see RABBITMQ_CONVENTION.md), instead of duplicating it.
func CreateProfileCore(req schemas.CreateProfileRequest) (*models.Profile, error) {
	authID, err := uuid.Parse(req.AuthID)
	if err != nil {
		return nil, err
	}

	var existing models.Profile
	if err := config.DB.Where("auth_id = ?", authID).First(&existing).Error; err == nil {
		return nil, ErrProfileConflict
	}

	profile := models.Profile{
		Slug:   utils.GenerateUniqueSlug(req.Name),
		Name:   req.Name,
		AuthID: authID,
		Role:   models.Role(req.Role),
	}

	switch profile.Role {
	case models.RolePatient:
		profile.PatientProfile = &models.PatientProfile{Email: req.Email}
	case models.RoleExpert:
		profile.ExpertProfile = &models.ExpertProfile{Email: req.Email, VerificationStatus: "UNVERIFIED"}
	case models.RoleAdmin:
		profile.AdminProfile = &models.AdminProfile{Email: req.Email}
	}

	if err := config.DB.Create(&profile).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

// ErrProfileConflict is returned when a profile already exists for the given auth_id.
// The user.created consumer treats this as an idempotent no-op rather than a failure.
var ErrProfileConflict = &conflictError{"auth_id already has a profile"}

type conflictError struct{ msg string }

func (e *conflictError) Error() string { return e.msg }

// ---------- /internal/api/v1/profiles ----------

// CreateProfileInternal khởi tạo Profile + hồ sơ theo vai trò ngay khi auth-service báo có
// tài khoản mới (service-to-service, không đi qua Gateway nên không cần JWT).
// @Summary      [Internal] Tạo profile khi có tài khoản mới
// @Description  auth-service gọi endpoint này (qua mạng nội bộ, không qua Gateway) ngay sau khi tạo tài khoản.
// @Tags         internal
// @Accept       json
// @Produce      json
// @Param        request body schemas.CreateProfileRequest true "Thông tin tài khoản mới"
// @Success      201 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      409 {object} response.Response
// @Router       /internal/api/v1/profiles/create [post]
func CreateProfileInternal(c *gin.Context) {
	var req schemas.CreateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", err.Error())
		return
	}

	profile, err := CreateProfileCore(req)
	if err != nil {
		if err == ErrProfileConflict {
			response.Error(c, http.StatusConflict, "Profile cho tài khoản này đã tồn tại", err.Error())
			return
		}
		response.Error(c, http.StatusBadRequest, "Không thể tạo profile", err.Error())
		return
	}

	response.Created(c, "Tạo profile thành công", profile)
}

// SyncSeedAuthorsInternal triggers an event to update seeded post authors in forum-service
// @Summary      [Internal] Đồng bộ authorId cho các post mẫu
// @Description  Lấy ID của expert hiện tại và publish qua RabbitMQ để forum-service cập nhật lại toàn bộ authorId của những id mặc định ban đầu.
// @Tags         internal
// @Produce      json
// @Success      200 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /internal/api/v1/profiles/sync-seed-authors [post]
func SyncSeedAuthorsInternal(c *gin.Context) {
	var expertAuthID string
	err := config.DB.Table("profiles").
		Where("role = ?", "EXPERT").
		Order("created_at asc").
		Limit(1).
		Pluck("auth_id", &expertAuthID).
		Error

	if err != nil || expertAuthID == "" {
		expertAuthID = "00000000-0000-0000-0000-000000000002"
	}

	err = messaging.PublishSyncSeedAuthors(expertAuthID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Không thể publish event đồng bộ", err.Error())
		return
	}

	response.Success(c, "Đã publish event đồng bộ thành công", gin.H{"expert_auth_id": expertAuthID})
}

// ---------- /api/v1/profiles/me ----------

// GetMe trả về hồ sơ của chính người dùng đang đăng nhập.
// @Summary      Xem hồ sơ của chính mình
// @Tags         profiles
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} response.Response
// @Failure      401 {object} response.Response
// @Failure      404 {object} response.Response
// @Router       /api/v1/profiles/me [get]
func GetMe(c *gin.Context) {
	profile, err := findProfileByAuthID(c.GetString(middleware.CtxAuthID))
	if err != nil {
		response.Error(c, http.StatusNotFound, "Không tìm thấy hồ sơ của bạn", err.Error())
		return
	}
	response.Success(c, "Lấy hồ sơ thành công", profile)
}

// UpdateMe thay thế toàn bộ (PUT) hồ sơ của chính mình - nội dung tuỳ theo vai trò.
// @Summary      Cập nhật toàn bộ hồ sơ của chính mình
// @Tags         profiles
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      404 {object} response.Response
// @Router       /api/v1/profiles/me [put]
func UpdateMe(experts *ExpertHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		profile, err := findProfileByAuthID(c.GetString(middleware.CtxAuthID))
		if err != nil {
			response.Error(c, http.StatusNotFound, "Không tìm thấy hồ sơ của bạn", err.Error())
			return
		}

		switch profile.Role {
		case models.RolePatient:
			applyPatientUpsert(c, profile)
		case models.RoleExpert:
			experts.replaceFor(c, profile.AuthID)
		default:
			applyAdminUpsert(c, profile)
		}
	}
}

// PatchMe cập nhật một phần (PATCH) hồ sơ của chính mình - nội dung tuỳ theo vai trò.
// @Summary      Cập nhật một phần hồ sơ của chính mình
// @Tags         profiles
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      404 {object} response.Response
// @Router       /api/v1/profiles/me [patch]
func PatchMe(experts *ExpertHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		profile, err := findProfileByAuthID(c.GetString(middleware.CtxAuthID))
		if err != nil {
			response.Error(c, http.StatusNotFound, "Không tìm thấy hồ sơ của bạn", err.Error())
			return
		}

		switch profile.Role {
		case models.RolePatient:
			applyPatientPatch(c, profile)
		case models.RoleExpert:
			experts.patchFor(c, profile.AuthID)
		default:
			applyAdminPatch(c, profile)
		}
	}
}

func applyAdminUpsert(c *gin.Context, profile *models.Profile) {
	var req schemas.UpsertAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", err.Error())
		return
	}

	admin := models.AdminProfile{ProfileID: profile.ID, Email: req.Email, Note: req.Note}
	if err := saveProfileAndSub(profile.ID, req.Name, &admin); err != nil {
		response.Error(c, http.StatusInternalServerError, "Không thể cập nhật hồ sơ", err.Error())
		return
	}

	profile.Name = req.Name
	profile.AdminProfile = &admin
	response.Success(c, "Cập nhật hồ sơ thành công", profile)
}

func applyAdminPatch(c *gin.Context, profile *models.Profile) {
	var req schemas.PatchAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", err.Error())
		return
	}

	admin := profile.AdminProfile
	if admin == nil {
		admin = &models.AdminProfile{ProfileID: profile.ID}
	}

	newName := profile.Name
	if req.Name != nil {
		newName = *req.Name
	}
	if req.Email != nil {
		admin.Email = *req.Email
	}
	if req.Note != nil {
		admin.Note = *req.Note
	}

	if err := saveProfileAndSub(profile.ID, newName, admin); err != nil {
		response.Error(c, http.StatusInternalServerError, "Không thể cập nhật hồ sơ", err.Error())
		return
	}

	profile.Name = newName
	profile.AdminProfile = admin
	response.Success(c, "Cập nhật hồ sơ thành công", profile)
}

// ---------- /api/v1/profiles (quản lý chung - Admin) ----------

// ListProfiles trả về danh sách mọi profile, có thể lọc theo vai trò (dành cho Admin).
// @Summary      [Admin] Danh sách toàn bộ profile
// @Tags         profiles
// @Security     BearerAuth
// @Produce      json
// @Param        page      query int    false "Trang" default(1)
// @Param        page_size query int    false "Số dòng/trang" default(20)
// @Param        role      query string false "Lọc theo vai trò (ADMIN, EXPERT, PATIENT)"
// @Success      200 {object} response.Response
// @Router       /api/v1/profiles [get]
func ListProfiles(c *gin.Context) {
	pagination := parsePagination(c)

	query := config.DB.Model(&models.Profile{})
	if role := strings.TrimSpace(c.Query("role")); role != "" {
		query = query.Where("role = ?", strings.ToUpper(role))
	}

	var total int64
	query.Count(&total)

	var profiles []models.Profile
	if err := query.
		Order("created_at DESC").
		Limit(pagination.PageSize).
		Offset((pagination.Page - 1) * pagination.PageSize).
		Find(&profiles).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi truy vấn danh sách profile", err.Error())
		return
	}

	response.Success(c, "Lấy danh sách profile thành công", schemas.PaginatedResponse{
		Items:      profiles,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalItems: total,
		TotalPages: totalPages(total, pagination.PageSize),
	})
}

// CreateProfile cho phép Admin tạo thủ công một profile mới (VD: tạo thêm tài khoản Admin khác).
// @Summary      [Admin] Tạo mới một profile
// @Tags         profiles
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body schemas.CreateProfileRequest true "Thông tin profile"
// @Success      201 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      409 {object} response.Response
// @Router       /api/v1/profiles [post]
func CreateProfile(c *gin.Context) {
	var req schemas.CreateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", err.Error())
		return
	}

	profile, err := CreateProfileCore(req)
	if err != nil {
		if err == ErrProfileConflict {
			response.Error(c, http.StatusConflict, "Profile cho tài khoản này đã tồn tại", err.Error())
			return
		}
		response.Error(c, http.StatusBadRequest, "Không thể tạo profile", err.Error())
		return
	}

	response.Created(c, "Tạo profile thành công", profile)
}

// GetProfile trả về 1 profile bất kỳ theo auth_id (dành cho Admin).
// @Summary      [Admin] Xem chi tiết 1 profile
// @Tags         profiles
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Auth Account ID"
// @Success      200 {object} response.Response
// @Failure      404 {object} response.Response
// @Router       /api/v1/profiles/{id} [get]
func GetProfile(c *gin.Context) {
	authID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Auth Account ID không hợp lệ", err.Error())
		return
	}

	var profile models.Profile
	if err := config.DB.
		Preload("PatientProfile").
		Preload("ExpertProfile.Specializations").
		Preload("AdminProfile").
		First(&profile, "auth_id = ?", authID).Error; err != nil {
		response.Error(c, http.StatusNotFound, "Không tìm thấy profile", err.Error())
		return
	}
	response.Success(c, "Lấy profile thành công", profile)
}

// GetPublicProfile trả về 1 profile công khai theo auth_id cho tất cả mọi người.
// @Summary      Xem chi tiết 1 profile công khai
// @Tags         profiles
// @Produce      json
// @Param        id path string true "Auth Account ID"
// @Success      200 {object} response.Response
// @Failure      404 {object} response.Response
// @Router       /api/v1/profiles/{id} [get]
func GetPublicProfile(c *gin.Context) {
	profile, err := findProfileByAuthID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusNotFound, "Không tìm thấy hồ sơ", err.Error())
		return
	}

	// Tránh lộ lọt thông tin nhạy cảm của Patient/Admin nếu truy cập qua endpoint này
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

// UpdateProfile thay thế (PUT) trường chung (name) của 1 profile bất kỳ (dành cho Admin).
// @Summary      [Admin] Cập nhật tên profile
// @Tags         profiles
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id      path string                      true "Auth Account ID"
// @Param        request body schemas.UpdateProfileRequest true "Tên mới"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      404 {object} response.Response
// @Router       /api/v1/profiles/{id} [put]
func UpdateProfile(c *gin.Context) {
	authID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Auth Account ID không hợp lệ", err.Error())
		return
	}

	var req schemas.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", err.Error())
		return
	}

	result := config.DB.Model(&models.Profile{}).Where("auth_id = ?", authID).Update("name", req.Name)
	if result.Error != nil {
		response.Error(c, http.StatusInternalServerError, "Không thể cập nhật profile", result.Error.Error())
		return
	}
	if result.RowsAffected == 0 {
		response.Error(c, http.StatusNotFound, "Không tìm thấy profile", "profile not found")
		return
	}

	response.Success(c, "Cập nhật profile thành công", gin.H{"auth_id": authID, "name": req.Name})
}

// PatchProfile cập nhật một phần (PATCH) trường chung (name) của 1 profile bất kỳ (dành cho Admin).
// @Summary      [Admin] Cập nhật một phần thông tin profile
// @Tags         profiles
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id      path string                     true "Auth Account ID"
// @Param        request body schemas.PatchProfileRequest true "Trường cần cập nhật"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      404 {object} response.Response
// @Router       /api/v1/profiles/{id} [patch]
func PatchProfile(c *gin.Context) {
	authID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Auth Account ID không hợp lệ", err.Error())
		return
	}

	var req schemas.PatchProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", err.Error())
		return
	}
	if req.Name == nil {
		response.Success(c, "Không có gì để cập nhật", nil)
		return
	}

	result := config.DB.Model(&models.Profile{}).Where("auth_id = ?", authID).Update("name", *req.Name)
	if result.Error != nil {
		response.Error(c, http.StatusInternalServerError, "Không thể cập nhật profile", result.Error.Error())
		return
	}
	if result.RowsAffected == 0 {
		response.Error(c, http.StatusNotFound, "Không tìm thấy profile", "profile not found")
		return
	}

	response.Success(c, "Cập nhật profile thành công", gin.H{"auth_id": authID, "name": *req.Name})
}

// DeleteProfile xoá mềm (soft delete) 1 profile (dành cho Admin).
// @Summary      [Admin] Xoá (mềm) một profile
// @Tags         profiles
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Auth Account ID"
// @Success      200 {object} response.Response
// @Failure      404 {object} response.Response
// @Router       /api/v1/profiles/{id} [delete]
func DeleteProfile(c *gin.Context) {
	authID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Auth Account ID không hợp lệ", err.Error())
		return
	}

	result := config.DB.Delete(&models.Profile{}, "auth_id = ?", authID)
	if result.Error != nil {
		response.Error(c, http.StatusInternalServerError, "Không thể xoá profile", result.Error.Error())
		return
	}
	if result.RowsAffected == 0 {
		response.Error(c, http.StatusNotFound, "Không tìm thấy profile", "profile not found")
		return
	}

	response.Success(c, "Xoá profile thành công", nil)
}
