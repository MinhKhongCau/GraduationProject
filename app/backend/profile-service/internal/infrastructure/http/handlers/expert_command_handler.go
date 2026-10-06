// File: internal/infrastructure/http/handlers/expert_command_handler.go
package handlers

import (
	"errors"
	"net/http"

	"profile-service/internal/application/expertprofile"
	"profile-service/internal/domain/expert"
	"profile-service/internal/domain/profile"
	"profile-service/internal/domain/specialization"
	"profile-service/internal/infrastructure/http/middleware"
	"profile-service/internal/infrastructure/http/response"
	"profile-service/internal/infrastructure/http/schemas"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ExpertHandler là HTTP adapter cho các use case ghi hồ sơ chuyên gia.
// Handler chỉ bind/validate request, gọi use case và map lỗi domain sang HTTP status.
type ExpertHandler struct {
	replace         *expertprofile.ReplaceExpertProfile
	patch           *expertprofile.PatchExpertProfile
	specializations *expertprofile.ManageExpertSpecializations
}

func NewExpertHandler(
	replace *expertprofile.ReplaceExpertProfile,
	patch *expertprofile.PatchExpertProfile,
	specializations *expertprofile.ManageExpertSpecializations,
) *ExpertHandler {
	return &ExpertHandler{replace: replace, patch: patch, specializations: specializations}
}

// AddMySpecialization: chuyên gia tự đăng ký thêm 1 chuyên khoa (đã có thì giữ nguyên).
// @Summary      [Expert] Thêm chuyên khoa cho chính mình
// @Tags         experts
// @Security     BearerAuth
// @Produce      json
// @Param        specId path string true "Specialization ID"
// @Success      200 {object} response.BaseResponse
// @Failure      404 {object} response.BaseResponse
// @Failure      422 {object} response.BaseResponse
// @Router       /api/v1/profiles/me/specializations/{specId} [post]
func (h *ExpertHandler) AddMySpecialization(c *gin.Context) {
	authID, specID, ok := parseMySpecializationParams(c)
	if !ok {
		return
	}

	view, err := h.specializations.Add(c.Request.Context(), authID, specID)
	if err != nil {
		if errors.Is(err, specialization.ErrSpecializationNotFound) {
			response.Error(c, http.StatusNotFound, "Không tìm thấy chuyên khoa", err.Error())
			return
		}
		writeExpertError(c, err)
		return
	}
	response.Success(c, "Đã thêm chuyên khoa", view)
}

// RemoveMySpecialization: chuyên gia tự huỷ đăng ký 1 chuyên khoa.
// @Summary      [Expert] Bỏ chuyên khoa của chính mình
// @Tags         experts
// @Security     BearerAuth
// @Produce      json
// @Param        specId path string true "Specialization ID"
// @Success      200 {object} response.BaseResponse
// @Failure      404 {object} response.BaseResponse
// @Router       /api/v1/profiles/me/specializations/{specId} [delete]
func (h *ExpertHandler) RemoveMySpecialization(c *gin.Context) {
	authID, specID, ok := parseMySpecializationParams(c)
	if !ok {
		return
	}

	view, err := h.specializations.Remove(c.Request.Context(), authID, specID)
	if err != nil {
		writeExpertError(c, err)
		return
	}
	response.Success(c, "Đã bỏ chuyên khoa", view)
}

func parseMySpecializationParams(c *gin.Context) (uuid.UUID, uuid.UUID, bool) {
	authID, err := uuid.Parse(c.GetString(middleware.CtxAuthID))
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Không thể xác định danh tính người dùng", err.Error())
		return uuid.Nil, uuid.Nil, false
	}
	specID, err := uuid.Parse(c.Param("specId"))
	if err != nil {
		response.Error(c, http.StatusNotFound, "Không tìm thấy chuyên khoa", err.Error())
		return uuid.Nil, uuid.Nil, false
	}
	return authID, specID, true
}

// Update thay thế toàn bộ (PUT) hồ sơ chuyên gia - dành cho Admin.
// @Summary      [Admin] Cập nhật toàn bộ hồ sơ chuyên gia
// @Tags         experts
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id      path string                      true "Auth Account ID"
// @Param        request body schemas.UpsertExpertRequest true "Hồ sơ chuyên gia"
// @Success      200 {object} response.BaseResponse
// @Failure      400 {object} response.BaseResponse
// @Failure      404 {object} response.BaseResponse
// @Router       /api/v1/profiles/experts/{id} [put]
func (h *ExpertHandler) Update(c *gin.Context) {
	authID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusNotFound, "Không tìm thấy hồ sơ chuyên gia", err.Error())
		return
	}
	h.replaceFor(c, authID)
}

// Patch cập nhật một phần (PATCH) hồ sơ chuyên gia - dành cho Admin.
// Chỉ Admin mới có quyền thay đổi verification_status.
// @Summary      [Admin] Cập nhật một phần hồ sơ chuyên gia
// @Tags         experts
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id      path string                     true "Auth Account ID"
// @Param        request body schemas.PatchExpertRequest true "Các trường cần cập nhật"
// @Success      200 {object} response.BaseResponse
// @Failure      400 {object} response.BaseResponse
// @Failure      403 {object} response.BaseResponse
// @Failure      404 {object} response.BaseResponse
// @Router       /api/v1/profiles/experts/{id} [patch]
func (h *ExpertHandler) Patch(c *gin.Context) {
	authID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusNotFound, "Không tìm thấy hồ sơ chuyên gia", err.Error())
		return
	}
	h.patchFor(c, authID)
}

func (h *ExpertHandler) replaceFor(c *gin.Context, authID uuid.UUID) {
	var req schemas.UpsertExpertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", err.Error())
		return
	}

	info, err := parseUserInformation(req.UserInformation)
	if err != nil {
		writeInvalidDateOfBirth(c, err)
		return
	}

	view, err := h.replace.Execute(c.Request.Context(), expertprofile.ReplaceCommand{
		AuthID:               authID,
		UserInformation:      info,
		Email:                req.Email,
		AvatarURL:            req.AvatarURL,
		IntroductionVideoURL: req.IntroductionVideoURL,
		Bio:                  req.Bio,
		SpecializationIDs:    req.SpecializationIDs,
	})
	if err != nil {
		writeExpertError(c, err)
		return
	}
	response.Success(c, "Cập nhật hồ sơ thành công", view)
}

func (h *ExpertHandler) patchFor(c *gin.Context, authID uuid.UUID) {
	var req schemas.PatchExpertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", err.Error())
		return
	}

	patch, err := parseUserInformationPatch(req.UserInformation)
	if err != nil {
		writeInvalidDateOfBirth(c, err)
		return
	}

	view, err := h.patch.Execute(c.Request.Context(), expertprofile.PatchCommand{
		AuthID:               authID,
		ActorRole:            profile.Role(c.GetString(middleware.CtxRole)),
		UserInformation:      patch,
		Email:                req.Email,
		AvatarURL:            req.AvatarURL,
		IntroductionVideoURL: req.IntroductionVideoURL,
		Bio:                  req.Bio,
		VerificationStatus:   req.VerificationStatus,
		SpecializationIDs:    req.SpecializationIDs,
	})
	if err != nil {
		writeExpertError(c, err)
		return
	}
	response.Success(c, "Cập nhật hồ sơ thành công", view)
}

func writeExpertError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, expert.ErrExpertNotFound):
		response.Error(c, http.StatusNotFound, "Không tìm thấy hồ sơ chuyên gia", err.Error())
	case errors.Is(err, expert.ErrVerificationRequiresAdmin):
		response.Error(c, http.StatusForbidden, "Chỉ Admin mới có quyền thay đổi trạng thái xác minh", "Forbidden")
	case errors.Is(err, expert.ErrInvalidVerificationStatus):
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", err.Error())
	case errors.Is(err, specialization.ErrSpecializationNotFound):
		response.Error(c, http.StatusBadRequest, "Có chuyên khoa không tồn tại", err.Error())
	case errors.Is(err, specialization.ErrSpecializationInactive):
		response.Error(c, http.StatusUnprocessableEntity, "Chuyên khoa đã ngừng hoạt động", err.Error())
	case errors.Is(err, expert.ErrSpecializationNotAssigned):
		response.Error(c, http.StatusNotFound, "Bạn chưa đăng ký chuyên khoa này", err.Error())
	default:
		response.Error(c, http.StatusInternalServerError, "Không thể cập nhật hồ sơ", err.Error())
	}
}
