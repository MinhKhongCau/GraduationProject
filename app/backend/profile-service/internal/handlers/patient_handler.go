// File: internal/handlers/patient_handler.go
package handlers

import (
	"net/http"
	"profile-service/config"
	"profile-service/internal/models"
	"profile-service/internal/schemas"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Lấy hồ sơ Bệnh nhân dựa vào account_id
func GetPatientProfile(c *gin.Context) {
	accountID := c.Param("account_id") // Lấy từ URL path

	var patient models.Patient
	// Truy vấn DB qua GORM
	if err := config.DB.Where("account_id = ?", accountID).First(&patient).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy hồ sơ bệnh nhân!"})
		return
	}

	// Trả về JSON
	c.JSON(http.StatusOK, gin.H{
		"message": "Lấy hồ sơ thành công",
		"data":    patient,
	})
}

func CreatePatientProfile(c *gin.Context) {
	accountID := c.Param("account_id")

	var req schemas.UpdatePatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ: " + err.Error()})
		return
	}

	dob, _ := time.Parse("2006-01-02", req.DateOfBirth)
	parsedAccountID, _ := uuid.Parse(accountID)

	// KHÔNG CẦN TÌM TRƯỚC NỮA. CỨ THẾ INSERT!
	newPatient := models.Patient{
		AccountID:   parsedAccountID,
		FullName:    req.FullName,
		PhoneNumber: req.PhoneNumber,
		Email:       req.Email,
		AvatarURL:   req.AvatarURL,
		DateOfBirth: dob,
		Gender:      req.Gender,
		Address:     req.Address,
	}

	// GORM sẽ cố gắng lưu vào DB. Nếu trùng Bộ 3 (Account + Tên + Ngày sinh), DB sẽ chửi và văng lỗi!
	if errCreate := config.DB.Create(&newPatient).Error; errCreate != nil {
		// Kiểm tra xem có phải lỗi do trùng lặp dữ liệu không (Violate Unique Constraint)
		c.JSON(http.StatusConflict, gin.H{
			"error": "Hồ sơ cho bệnh nhân này đã tồn tại trong tài khoản của bạn!",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Tạo hồ sơ thành công!", "data": newPatient})
}
