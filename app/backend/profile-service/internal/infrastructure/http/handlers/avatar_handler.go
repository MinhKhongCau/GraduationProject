// File: internal/infrastructure/http/handlers/avatar_handler.go
package handlers

import (
	"context"
	"errors"
	"io"
	"net/http"

	"profile-service/config"
	"profile-service/internal/infrastructure/http/middleware"
	"profile-service/internal/infrastructure/http/response"
	"profile-service/internal/infrastructure/persistence/models"
	s3infra "profile-service/internal/infrastructure/s3"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// maxAvatarSize giới hạn dung lượng ảnh đại diện tải lên (5 MB).
const maxAvatarSize = 5 << 20

// AvatarUploader là phần lưu trữ ảnh mà handler cần (S3 client thật hoặc fake khi test).
type AvatarUploader interface {
	UploadProfileImage(ctx context.Context, profileID string, body io.ReadSeeker) (string, error)
}

// AvatarHandler xử lý upload ảnh đại diện cho bệnh nhân và chuyên gia.
type AvatarHandler struct {
	uploader AvatarUploader
}

func NewAvatarHandler(uploader AvatarUploader) *AvatarHandler {
	return &AvatarHandler{uploader: uploader}
}

// UploadAvatar nhận file ảnh (multipart/form-data, field "file") và lưu URL vào avatar_url.
// Chỉ chính chủ hoặc Admin mới được cập nhật; Admin không có ảnh đại diện nên bị từ chối.
// @Summary      Tải lên ảnh đại diện
// @Tags         profiles
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        id   path     string true "Auth Account ID"
// @Param        file formData file   true "Ảnh JPEG, PNG, GIF hoặc WebP, tối đa 5 MB"
// @Success      200 {object} response.BaseResponse
// @Failure      400 {object} response.BaseResponse
// @Failure      403 {object} response.BaseResponse
// @Failure      404 {object} response.BaseResponse
// @Failure      413 {object} response.BaseResponse
// @Failure      503 {object} response.BaseResponse
// @Router       /api/v1/profiles/{id}/avatar [post]
func (h *AvatarHandler) UploadAvatar(c *gin.Context) {
	authID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Auth Account ID không hợp lệ", err.Error())
		return
	}

	callerID, err := uuid.Parse(c.GetString(middleware.CtxAuthID))
	isOwner := err == nil && callerID == authID
	if !isOwner && c.GetString(middleware.CtxRole) != string(models.RoleAdmin) {
		response.Error(c, http.StatusForbidden, "Bạn không có quyền cập nhật ảnh đại diện này", "Forbidden")
		return
	}

	if h.uploader == nil {
		response.Error(c, http.StatusServiceUnavailable, "Chức năng lưu trữ ảnh đại diện chưa được cấu hình", "S3 is not configured")
		return
	}

	var profile models.Profile
	if err := config.DB.First(&profile, "auth_id = ?", authID).Error; err != nil {
		response.Error(c, http.StatusNotFound, "Không tìm thấy profile", err.Error())
		return
	}

	// avatar_url nằm ở bảng hồ sơ theo vai trò, không nằm ở profiles.
	var avatarRecord any
	switch profile.Role {
	case models.RolePatient:
		avatarRecord = &models.PatientProfile{}
	case models.RoleExpert:
		avatarRecord = &models.ExpertProfile{}
	default:
		response.Error(c, http.StatusBadRequest, "Vai trò này không có ảnh đại diện", "role has no avatar")
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxAvatarSize)
	fileHeader, err := c.FormFile("file")
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			response.Error(c, http.StatusRequestEntityTooLarge, "Ảnh vượt quá dung lượng cho phép (tối đa 5 MB)", err.Error())
			return
		}
		response.Error(c, http.StatusBadRequest, "Thiếu file ảnh (trường 'file')", err.Error())
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Không thể đọc file ảnh", err.Error())
		return
	}
	defer file.Close()

	url, err := h.uploader.UploadProfileImage(c.Request.Context(), profile.ID.String(), file)
	if err != nil {
		if errors.Is(err, s3infra.ErrUnsupportedImageType) {
			response.Error(c, http.StatusBadRequest, "Định dạng ảnh không được hỗ trợ (JPEG, PNG, GIF, WebP)", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Không thể tải ảnh lên", err.Error())
		return
	}

	result := config.DB.Model(avatarRecord).Where("profile_id = ?", profile.ID).Update("avatar_url", url)
	if result.Error != nil {
		response.Error(c, http.StatusInternalServerError, "Không thể cập nhật ảnh đại diện", result.Error.Error())
		return
	}
	if result.RowsAffected == 0 {
		response.Error(c, http.StatusNotFound, "Không tìm thấy hồ sơ chi tiết của profile", "profile detail not found")
		return
	}

	response.Success(c, "Cập nhật ảnh đại diện thành công", gin.H{"auth_id": authID, "avatar_url": url})
}
