// File: internal/infrastructure/http/handlers/specialization_handler.go
package handlers

import (
	"net/http"

	"profile-service/config"
	"profile-service/internal/infrastructure/http/response"
	"profile-service/internal/infrastructure/http/schemas"
	"profile-service/internal/infrastructure/persistence/models"

	"github.com/gin-gonic/gin"
)

// CreateSpecialization tạo mới một chuyên khoa (dành cho Admin).
// @Summary      [Admin] Tạo chuyên khoa mới
// @Tags         specializations
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body schemas.CreateSpecializationRequest true "Chuyên khoa"
// @Success      201 {object} response.Response
// @Failure      400 {object} response.Response
// @Router       /api/v1/profiles/specializations [post]
func CreateSpecialization(c *gin.Context) {
	var req schemas.CreateSpecializationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", err.Error())
		return
	}

	newSpec := models.Specialization{
		Name:        req.Name,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		IsActive:    true,
	}

	if err := config.DB.Create(&newSpec).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi lưu Database", err.Error())
		return
	}

	response.Created(c, "Tạo chuyên khoa thành công", newSpec)
}

// GetAllSpecializations trả về danh sách chuyên khoa đang hoạt động (public).
// @Summary      Danh sách chuyên khoa
// @Tags         specializations
// @Produce      json
// @Success      200 {object} response.Response
// @Router       /api/v1/profiles/specializations [get]
func GetAllSpecializations(c *gin.Context) {
	var specs []models.Specialization
	config.DB.Where("is_active = ?", true).Find(&specs)

	response.Success(c, "Lấy danh sách chuyên khoa thành công", specs)
}
