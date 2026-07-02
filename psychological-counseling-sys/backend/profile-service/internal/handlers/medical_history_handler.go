package handlers

import (
	"net/http"
	"profile-service/config"
	"profile-service/internal/models"
	"profile-service/internal/schemas"
	"profile-service/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
)

// 1. LẤY DANH SÁCH TIỀN SỬ BỆNH
func GetMedicalHistories(c *gin.Context) {
	accountID := c.Param("account_id")

	// Bước 1: Tìm ID của Bệnh nhân thông qua AccountID
	var patient models.Patient
	if err := config.DB.Where("account_id = ?", accountID).First(&patient).Error; err != nil {
		response.Error(c, http.StatusNotFound, "Không tìm thấy hồ sơ bệnh nhân!", err.Error())
		return
	}

	// Bước 2: Truy vấn danh sách Tiền sử bệnh dựa vào PatientID
	var histories []models.MedicalHistory
	// Chỉ lấy những bệnh án đang active
	config.DB.Where("patient_id = ? AND is_active = ?", patient.PatientID, true).Find(&histories)

	response.Success(c, "Lấy danh sách tiền sử bệnh thành công", histories)
}

// 2. THÊM MỚI TIỀN SỬ BỆNH
func AddMedicalHistory(c *gin.Context) {
	accountID := c.Param("account_id")

	var req schemas.CreateMedicalHistoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", err.Error())
		return
	}

	// Tìm Bệnh nhân (Giống hệt ở trên)
	var patient models.Patient
	if err := config.DB.Where("account_id = ?", accountID).First(&patient).Error; err != nil {
		response.Error(c, http.StatusNotFound, "Bạn cần tạo Hồ sơ cá nhân trước khi thêm Tiền sử bệnh!", err.Error())
		return
	}

	diagnosedDate, _ := time.Parse("2006-01-02", req.DiagnosedAt)

	// Tạo bản ghi mới
	newHistory := models.MedicalHistory{
		PatientID:     patient.PatientID, // Dùng ID nội bộ vừa tìm được
		ConditionName: req.ConditionName,
		Description:   req.Description,
		DiagnosedAt:   diagnosedDate,
		IsChronic:     req.IsChronic,
		IsActive:      true,
	}

	if err := config.DB.Create(&newHistory).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi lưu dữ liệu", err.Error())
		return
	}

	response.JSON(c, http.StatusCreated, true, "Thêm tiền sử bệnh thành công!", newHistory, "")
}
