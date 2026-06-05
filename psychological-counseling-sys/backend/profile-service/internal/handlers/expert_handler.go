package handlers

import (
	"net/http"
	"profile-service/config"
	"profile-service/internal/models"
	"profile-service/internal/schemas"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CreateExpertProfile(c *gin.Context) {
	accountID := c.Param("account_id")

	var req schemas.CreateExpertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ: " + err.Error()})
		return
	}

	parsedAccountID, errParse := uuid.Parse(accountID)
	if errParse != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Account ID không hợp lệ"})
		return
	}

	// BƯỚC 1: Khởi tạo struct Chuyên gia
	newExpert := models.Expert{
		AccountID:            parsedAccountID,
		FullName:             req.FullName,
		PhoneNumber:          req.PhoneNumber,
		Email:                req.Email,
		AvatarURL:            req.AvatarURL,
		IntroductionVideoURL: req.IntroductionVideoURL,
		Bio:                  req.Bio,
		VerificationStatus:   "UNVERIFIED", // Mặc định chờ Admin duyệt
	}

	// BƯỚC 2: TÌM VÀ GÁN CHUYÊN KHOA (PHÉP THUẬT N-N LÀ Ở ĐÂY)
	var specializations []models.Specialization
	// Tìm tất cả các chuyên khoa có ID nằm trong mảng Frontend gửi lên
	if len(req.SpecializationIDs) > 0 {
		config.DB.Where("spec_id IN ?", req.SpecializationIDs).Find(&specializations)
		// Gán mảng vừa tìm được vào struct Expert
		newExpert.Specializations = specializations
	}

	// BƯỚC 3: LƯU VÀO DATABASE
	// GORM sẽ tự động INSERT bảng profile_experts VÀ INSERT luôn bảng trung gian profile_expert_specs
	if err := config.DB.Create(&newExpert).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Tài khoản này đã được đăng ký làm Chuyên gia. Mỗi tài khoản chỉ có 1 hồ sơ!"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Tạo hồ sơ Chuyên gia thành công!",
	})
}

// 2. LẤY THÔNG TIN CHI TIẾT 1 CHUYÊN GIA
func GetExpertProfile(c *gin.Context) {
	accountID := c.Param("account_id")
	var expert models.Expert

	// Dùng Preload("Specializations") để GORM tự động kéo dữ liệu từ bảng trung gian
	if err := config.DB.Preload("Specializations").Where("account_id = ?", accountID).First(&expert).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy hồ sơ chuyên gia!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Lấy thông tin thành công",
		"data":    expert,
	})
}

// 3. LẤY DANH SÁCH TẤT CẢ CHUYÊN GIA (Cho màn hình Đặt lịch)
func GetAllExperts(c *gin.Context) {
	var experts []models.Expert

	// Preload để UI biết mỗi bác sĩ thuộc những chuyên khoa nào
	if err := config.DB.Preload("Specializations").Find(&experts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi truy vấn Database!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Lấy danh sách thành công",
		"data":    experts,
	})
}
