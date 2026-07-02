// File: internal/handlers/patient_handler.go
package handlers

import (
	"net/http"
	"profile-service/config"
	"profile-service/internal/models"
	"profile-service/internal/schemas"
	"profile-service/pkg/response"
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
		response.Error(c, http.StatusNotFound, "Không tìm thấy hồ sơ bệnh nhân!", err.Error())
		return
	}

	response.Success(c, "Lấy hồ sơ thành công", patient)
}

func CreatePatientProfile(c *gin.Context) {
	accountID := c.Param("account_id")

	var req schemas.UpdatePatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", err.Error())
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
		response.Error(c, http.StatusConflict, "Hồ sơ cho bệnh nhân này đã tồn tại trong tài khoản của bạn!", errCreate.Error())
		return
	}

	response.JSON(c, http.StatusCreated, true, "Tạo hồ sơ thành công!", newPatient, "")
}
