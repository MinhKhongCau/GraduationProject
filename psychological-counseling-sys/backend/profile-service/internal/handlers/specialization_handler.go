package handlers

import (
	"net/http"
	"profile-service/config"
	"profile-service/internal/models"
	"profile-service/internal/schemas"

	"github.com/gin-gonic/gin"
)

// 1. TẠO CHUYÊN KHOA MỚI (Dùng cho Admin)
func CreateSpecialization(c *gin.Context) {
	var req schemas.CreateSpecializationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ: " + err.Error()})
		return
	}

	newSpec := models.Specialization{
		Name:        req.Name,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		IsActive:    true,
	}

	if err := config.DB.Create(&newSpec).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi lưu Database!"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Tạo chuyên khoa thành công", "data": newSpec})
}

// 2. LẤY DANH SÁCH CHUYÊN KHOA (Dùng cho UI để Bác sĩ/Bệnh nhân chọn)
func GetAllSpecializations(c *gin.Context) {
	var specs []models.Specialization
	config.DB.Where("is_active = ?", true).Find(&specs)

	c.JSON(http.StatusOK, gin.H{"message": "Thành công", "data": specs})
}
