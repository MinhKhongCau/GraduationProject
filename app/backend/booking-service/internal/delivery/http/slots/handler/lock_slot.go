package handler

import (
	"booking-service/internal/repository/postgres"
	"booking-service/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LockSlotHandler struct {
	repo *postgres.AppointmentRepository
}

func NewLockSlotHandler(repo *postgres.AppointmentRepository) *LockSlotHandler {
	return &LockSlotHandler{repo: repo}
}

// Handle - POST /api/v1/slots/:id/lock
// Phân quyền: Chỉ PATIENT mới được giữ chỗ
//
//	@Summary      Giữ chỗ slot tạm thời 15 phút
//	@Description  Bệnh nhân click chọn giờ, hệ thống khóa slot trong 15 phút để tiến hành thanh toán
//	@Tags         Slots
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id   path      string  true  "Slot ID"
//	@Success      200  {object}  map[string]interface{}
//	@Failure      400  {object}  map[string]interface{}
//	@Failure      403  {object}  map[string]interface{}
//	@Failure      409  {object}  map[string]interface{}
//	@Router       /slots/{id}/lock [post]
func (h *LockSlotHandler) Handle(c *gin.Context) {
	// Chỉ PATIENT mới được giữ chỗ
	userRole := c.GetHeader("X-User-Role")
	if userRole != "PATIENT" {
		response.Error(c, http.StatusForbidden, "Chỉ bệnh nhân mới có thể đặt lịch", "Forbidden")
		return
	}

	patientID := c.GetHeader("X-User-Id")
	if patientID == "" {
		response.Error(c, http.StatusUnauthorized, "Không thể xác định danh tính người dùng", "Missing X-User-Id header")
		return
	}

	slotID := c.Param("id")
	if slotID == "" {
		response.Error(c, http.StatusBadRequest, "Thiếu Slot ID trong đường dẫn", "Missing slot_id")
		return
	}

	// Thực hiện Atomic Lock
	if err := h.repo.LockSlot(slotID, patientID); err != nil {
		response.Error(c, http.StatusConflict, err.Error(), "Slot unavailable")
		return
	}

	response.Success(c, "Giữ chỗ thành công! Bạn có 15 phút để hoàn tất thanh toán.", gin.H{
		"slot_id":              slotID,
		"lock_duration_seconds": 900,
	})
}
