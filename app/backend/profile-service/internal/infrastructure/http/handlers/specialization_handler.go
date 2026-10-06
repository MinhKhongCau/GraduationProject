// File: internal/infrastructure/http/handlers/specialization_handler.go
package handlers

import (
	"net/http"
	"strings"

	"profile-service/config"
	"profile-service/internal/infrastructure/http/response"
	"profile-service/internal/infrastructure/http/schemas"
	"profile-service/internal/infrastructure/persistence/models"
	"profile-service/internal/utils"

	"github.com/gin-gonic/gin"
)

// CreateSpecialization tạo mới một chuyên khoa (dành cho Admin).
// @Summary      [Admin] Tạo chuyên khoa mới
// @Description  code và slug là duy nhất; slug bỏ trống sẽ được sinh tự động từ name.
// @Tags         specializations
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body schemas.CreateSpecializationRequest true "Chuyên khoa"
// @Success      201 {object} response.BaseResponse{result=models.Specialization}
// @Failure      400 {object} response.BaseResponse{result=response.ErrorResult}
// @Failure      409 {object} response.BaseResponse{result=response.ErrorResult}
// @Router       /api/v1/profiles/specializations [post]
func CreateSpecialization(c *gin.Context) {
	var req schemas.CreateSpecializationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", err.Error())
		return
	}

	slug := utils.Slugify(strings.TrimSpace(req.Slug))
	if strings.TrimSpace(req.Slug) == "" {
		slug = utils.Slugify(req.Name)
	}
	symptoms := req.Symptoms
	if symptoms == nil {
		symptoms = []string{}
	}

	newSpec := models.Specialization{
		Code:        strings.ToUpper(strings.TrimSpace(req.Code)),
		Name:        req.Name,
		Slug:        slug,
		Description: req.Description,
		Symptoms:    symptoms,
		Location:    req.Location,
		ImageURL:    req.ImageURL,
		IsActive:    true,
	}

	var count int64
	if err := config.DB.Model(&models.Specialization{}).
		Where("code = ? OR slug = ?", newSpec.Code, newSpec.Slug).
		Count(&count).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi truy vấn Database", err.Error())
		return
	}
	if count > 0 {
		response.Error(c, http.StatusConflict, "Mã hoặc slug chuyên khoa đã tồn tại", "specialization code or slug already exists")
		return
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
// @Success      200 {object} response.BaseResponse{result=[]models.Specialization}
// @Router       /api/v1/profiles/specializations [get]
func GetAllSpecializations(c *gin.Context) {
	specs := []models.Specialization{}
	if err := config.DB.Where("is_active = ?", true).Order("code ASC").Find(&specs).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi truy vấn danh sách chuyên khoa", err.Error())
		return
	}

	response.Success(c, "Lấy danh sách chuyên khoa thành công", specs)
}
