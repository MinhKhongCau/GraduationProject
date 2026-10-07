// File: internal/infrastructure/http/handlers/patient_record_handler.go
package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"profile-service/internal/infrastructure/http/middleware"
	"profile-service/internal/infrastructure/http/response"
	"profile-service/internal/infrastructure/http/schemas"
	"profile-service/internal/infrastructure/persistence/models"
	"profile-service/internal/infrastructure/persistence/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PatientRecordHandler quản lý hồ sơ người khám của bệnh nhân đang đăng nhập.
type PatientRecordHandler struct {
	repo *repository.PatientRecordRepository
}

func NewPatientRecordHandler(repo *repository.PatientRecordRepository) *PatientRecordHandler {
	return &PatientRecordHandler{repo: repo}
}

// ListMine trả về hồ sơ người khám của chính mình (hồ sơ SELF luôn đứng đầu).
// @Summary      Danh sách hồ sơ người khám của chính mình
// @Tags         patient-records
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} response.BaseResponse{result=[]models.PatientRecord}
// @Failure      401 {object} response.BaseResponse
// @Router       /api/v1/profiles/me/patient-records [get]
func (h *PatientRecordHandler) ListMine(c *gin.Context) {
	ownerID, err := uuid.Parse(c.GetString(middleware.CtxAuthID))
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Không thể xác định danh tính người dùng", err.Error())
		return
	}

	records, err := h.repo.ListByOwner(c.Request.Context(), ownerID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi truy vấn hồ sơ người khám", err.Error())
		return
	}
	if records == nil {
		records = []models.PatientRecord{}
	}

	response.Success(c, "Lấy danh sách hồ sơ người khám thành công", records)
}

// CreateMine thêm hồ sơ người thân để đặt lịch hộ.
// @Summary      Tạo hồ sơ người khám mới
// @Tags         patient-records
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body schemas.CreatePatientRecordRequest true "Hồ sơ người khám"
// @Success      201 {object} response.BaseResponse{result=models.PatientRecord}
// @Failure      400 {object} response.BaseResponse
// @Failure      409 {object} response.BaseResponse
// @Router       /api/v1/profiles/me/patient-records [post]
func (h *PatientRecordHandler) CreateMine(c *gin.Context) {
	ownerID, err := uuid.Parse(c.GetString(middleware.CtxAuthID))
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Không thể xác định danh tính người dùng", err.Error())
		return
	}

	var req schemas.CreatePatientRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", err.Error())
		return
	}

	dateOfBirth, err := parseOptionalDate(req.DateOfBirth)
	if err != nil || dateOfBirth == nil {
		response.Error(c, http.StatusBadRequest, "Định dạng ngày sinh không hợp lệ (YYYY-MM-DD)", "invalid date_of_birth")
		return
	}
	if dateOfBirth.After(time.Now()) {
		response.Error(c, http.StatusBadRequest, "Ngày sinh không được ở tương lai", "date_of_birth in the future")
		return
	}

	record := models.PatientRecord{
		OwnerAuthID:  ownerID,
		FullName:     strings.TrimSpace(req.FullName),
		DateOfBirth:  dateOfBirth,
		Gender:       req.Gender,
		PhoneNumber:  strings.TrimSpace(req.PhoneNumber),
		Email:        strings.TrimSpace(req.Email),
		Address:      strings.TrimSpace(req.Address),
		Relationship: models.Relationship(req.Relationship),
	}
	if record.FullName == "" {
		response.Error(c, http.StatusBadRequest, "Họ tên không được để trống", "full_name is blank")
		return
	}

	if err := h.repo.Create(c.Request.Context(), &record); err != nil {
		if errors.Is(err, repository.ErrPatientRecordLimit) {
			response.Error(c, http.StatusConflict, "Bạn đã đạt số hồ sơ người khám tối đa", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Lỗi lưu hồ sơ người khám", err.Error())
		return
	}

	response.Created(c, "Tạo hồ sơ người khám thành công", record)
}
