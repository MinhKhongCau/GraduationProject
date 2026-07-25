package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"booking-service/internal/booking/domain"
	"booking-service/internal/slot"

	"github.com/gin-gonic/gin"
)

type fakeSlotUsecase struct {
	lockErr error
	locked  map[string]string
}

func (f *fakeSlotUsecase) GenerateSlotsForNextDays(expertID string, daysToGenerate int, availabilities []domain.Availability, timeTemplates []domain.TimeTemplate, timeOffs []domain.ExpertTimeOff) (slot.GenerationResult, error) {
	return slot.GenerationResult{}, nil
}

func (f *fakeSlotUsecase) LockSlot(slotID, patientID string) error {
	if f.lockErr != nil {
		return f.lockErr
	}
	if f.locked == nil {
		f.locked = make(map[string]string)
	}
	f.locked[slotID] = patientID
	return nil
}

func (f *fakeSlotUsecase) GetDates(expertID string, startDate, endDate time.Time) ([]string, error) {
	return nil, nil
}

func (f *fakeSlotUsecase) GetTimes(expertID string, date string) ([]slot.SlotTimeResult, error) {
	return nil, nil
}

func TestSlotLock(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("TC-BOOK-LCK-01 - Giữ chỗ tạm thời thành công (Happy Case)", func(t *testing.T) {
		fakeUsecase := &fakeSlotUsecase{}
		h := NewHandler(nil, nil, fakeUsecase, nil)

		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/booking/slots/slot-uuid-1/lock", nil)
		c.Request.Header.Set("X-User-Role", "PATIENT")
		c.Request.Header.Set("X-User-Id", "patient-uuid-1")
		c.Params = gin.Params{{Key: "id", Value: "slot-uuid-1"}}

		h.Lock(c)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", recorder.Code, recorder.Body.String())
		}

		var resp map[string]interface{}
		if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if resp["success"] != true {
			t.Fatalf("expected success = true, got %v", resp["success"])
		}

		data := resp["data"].(map[string]interface{})
		if data["slot_id"] != "slot-uuid-1" {
			t.Fatalf("expected slot_id slot-uuid-1, got %v", data["slot_id"])
		}
	})

	t.Run("TC-BOOK-LCK-02 - Giữ chỗ thất bại do slot đã bị người khác khóa (Conflict)", func(t *testing.T) {
		fakeUsecase := &fakeSlotUsecase{lockErr: slot.ErrSlotAlreadyLocked}
		h := NewHandler(nil, nil, fakeUsecase, nil)

		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/booking/slots/slot-uuid-locked/lock", nil)
		c.Request.Header.Set("X-User-Role", "PATIENT")
		c.Request.Header.Set("X-User-Id", "patient-uuid-2")
		c.Params = gin.Params{{Key: "id", Value: "slot-uuid-locked"}}

		h.Lock(c)

		if recorder.Code != http.StatusConflict {
			t.Fatalf("expected status 409, got %d: %s", recorder.Code, recorder.Body.String())
		}

		var resp map[string]interface{}
		json.Unmarshal(recorder.Body.Bytes(), &resp)
		if resp["success"] != false {
			t.Fatalf("expected success = false, got %v", resp["success"])
		}
	})

	t.Run("TC-BOOK-LCK-03 - Giữ chỗ thất bại do vai trò không phải Bệnh nhân (Forbidden)", func(t *testing.T) {
		fakeUsecase := &fakeSlotUsecase{}
		h := NewHandler(nil, nil, fakeUsecase, nil)

		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/booking/slots/slot-uuid-1/lock", nil)
		c.Request.Header.Set("X-User-Role", "EXPERT")
		c.Request.Header.Set("X-User-Id", "expert-uuid-1")
		c.Params = gin.Params{{Key: "id", Value: "slot-uuid-1"}}

		h.Lock(c)

		if recorder.Code != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d: %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("TC-BOOK-LCK-04 - Giữ chỗ thất bại do thiếu header định danh người dùng (Unauthorized)", func(t *testing.T) {
		fakeUsecase := &fakeSlotUsecase{}
		h := NewHandler(nil, nil, fakeUsecase, nil)

		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/booking/slots/slot-uuid-1/lock", nil)
		c.Request.Header.Set("X-User-Role", "PATIENT")
		// X-User-Id rỗng
		c.Params = gin.Params{{Key: "id", Value: "slot-uuid-1"}}

		h.Lock(c)

		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d: %s", recorder.Code, recorder.Body.String())
		}
	})
}
