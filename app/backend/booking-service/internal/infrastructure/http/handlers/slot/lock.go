package handler

import (
	"booking-service/internal/application/slot"
	"booking-service/internal/infrastructure/http/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Lock - POST /api/v1/slots/:id/lock
// PATIENT chọn một giờ trống (AVAILABLE) -> gọi API này để giữ chỗ tạm thời (LOCKED)
//
//	@Summary      [PATIENT] Lock a slot temporarily
//	@Description  [PATIENT] Locks an available slot for 15 minutes before payment
//	@Tags         Slots
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id    path      string  true  "Slot ID"
//	@Success      200   {object}  map[string]interface{}
//	@Failure      400   {object}  map[string]interface{}
//	@Failure      403   {object}  map[string]interface{}
//	@Failure      409   {object}  map[string]interface{}
//	@Failure      500   {object}  map[string]interface{}
//	@Router       /booking/slots/{id}/lock [post]
func (h *Handler) Lock(c *gin.Context) {
	// Only PATIENT is allowed to lock slots
	userRole := c.GetHeader("X-User-Role")
	if userRole != "PATIENT" {
		response.Error(c, http.StatusForbidden, "Only patients are allowed to lock slots", "Forbidden")
		return
	}

	patientID := c.GetHeader("X-User-Id")
	if patientID == "" {
		response.Error(c, http.StatusUnauthorized, "User identity could not be determined", "Missing X-User-Id header")
		return
	}

	slotID := c.Param("id")
	if slotID == "" {
		response.Error(c, http.StatusBadRequest, "Missing Slot ID in path", "Missing slot_id")
		return
	}

	// TODO ỨNG DỤNG REDIS NHA: Sử dụng Distributed Lock
	// Để chắc chắn tuyệt đối ko có 2 request vào DB cùng lúc, có thể dùng Redis SetNX lock theo `slot_id`.
	// Tuy nhiên hiện tại DB Update Atomic cũng đã chống được race condition cơ bản.

	// Lock the slot
	err := h.usecase.LockSlot(slotID, patientID)
	if err != nil {
		if err == slot.ErrSlotAlreadyLocked {
			response.Error(c, http.StatusConflict, err.Error(), "Slot locking failed - someone else might have locked it")
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error(), "Slot locking failed")
		return
	}

	response.Success(c, "Slot locked successfully. You have 15 minutes to complete the booking.", gin.H{
		"slot_id": slotID,
		"expires": 15 * 60, // 15 minutes in seconds
	})
}
