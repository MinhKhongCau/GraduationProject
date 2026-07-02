package handlers

import (
	"net/http"
	"profile-service/config"
	"profile-service/internal/models"
	"profile-service/internal/schemas"
	"profile-service/pkg/response"

	"github.com/gin-gonic/gin"
)

// 1. TẠO CHUYÊN KHOA MỚI (Dùng cho Admin)
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
		response.Error(c, http.StatusInternalServerError, "Lỗi lưu Database!", err.Error())
		return
	}

	response.JSON(c, http.StatusCreated, true, "Tạo chuyên khoa thành công", newSpec, "")
}

// 2. LẤY DANH SÁCH CHUYÊN KHOA (Dùng cho UI để Bác sĩ/Bệnh nhân chọn)
func GetAllSpecializations(c *gin.Context) {
	var specs []models.Specialization
	config.DB.Where("is_active = ?", true).Find(&specs)

	response.Success(c, "Thành công", specs)
}
