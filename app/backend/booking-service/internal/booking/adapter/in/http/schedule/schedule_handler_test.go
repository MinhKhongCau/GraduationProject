package handler

import (
	"bytes"
	"encoding/json"
	bookingquery "booking-service/internal/booking/application/query"
	"booking-service/internal/booking/domain"
	"booking-service/internal/schedule"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type fakeScheduleUsecase struct {
	createTimeTemplateFn func(shiftName, startTime, endTime string, slotDuration int) (*domain.TimeTemplate, error)
	getTimeTemplatesFn   func() ([]domain.TimeTemplate, error)
	updateTemplateFn     func(templateID string, isActive bool) error
	createAvailabilityFn func(expertID, templateID string, dayOfWeek int, effectiveFrom int64, effectiveUntil *int64, price float64) (*domain.Availability, error)
	getAvailabilitiesFn   func(expertID string) ([]domain.Availability, error)
	updateAvailabilityFn func(availID, expertID string, updates map[string]interface{}) error
	listAvailabilitiesFn func(query schedule.AvailabilityListQuery) (bookingquery.Page[domain.Availability], error)
	listTimeTemplatesFn  func(query schedule.TemplateListQuery) (bookingquery.Page[domain.TimeTemplate], error)
}

func (f fakeScheduleUsecase) CreateTimeTemplate(shiftName, startTime, endTime string, slotDuration int) (*domain.TimeTemplate, error) {
	if f.createTimeTemplateFn != nil {
		return f.createTimeTemplateFn(shiftName, startTime, endTime, slotDuration)
	}
	return nil, nil
}

func (f fakeScheduleUsecase) GetTimeTemplates() ([]domain.TimeTemplate, error) {
	if f.getTimeTemplatesFn != nil {
		return f.getTimeTemplatesFn()
	}
	return nil, nil
}

func (f fakeScheduleUsecase) UpdateTemplate(templateID string, isActive bool) error {
	if f.updateTemplateFn != nil {
		return f.updateTemplateFn(templateID, isActive)
	}
	return nil
}

func (f fakeScheduleUsecase) CreateAvailability(expertID, templateID string, dayOfWeek int, effectiveFrom int64, effectiveUntil *int64, price float64) (*domain.Availability, error) {
	if f.createAvailabilityFn != nil {
		return f.createAvailabilityFn(expertID, templateID, dayOfWeek, effectiveFrom, effectiveUntil, price)
	}
	return nil, nil
}

func (f fakeScheduleUsecase) GetAvailabilities(expertID string) ([]domain.Availability, error) {
	if f.getAvailabilitiesFn != nil {
		return f.getAvailabilitiesFn(expertID)
	}
	return nil, nil
}

func (f fakeScheduleUsecase) UpdateAvailability(availID, expertID string, updates map[string]interface{}) error {
	if f.updateAvailabilityFn != nil {
		return f.updateAvailabilityFn(availID, expertID, updates)
	}
	return nil
}

func (f fakeScheduleUsecase) ListAvailabilities(query schedule.AvailabilityListQuery) (bookingquery.Page[domain.Availability], error) {
	if f.listAvailabilitiesFn != nil {
		return f.listAvailabilitiesFn(query)
	}
	return bookingquery.Page[domain.Availability]{}, nil
}

func (f fakeScheduleUsecase) ListTimeTemplates(query schedule.TemplateListQuery) (bookingquery.Page[domain.TimeTemplate], error) {
	if f.listTimeTemplatesFn != nil {
		return f.listTimeTemplatesFn(query)
	}
	return bookingquery.Page[domain.TimeTemplate]{}, nil
}

func TestCreateTemplate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("CreateTemplate - Forbidden for non-ADMIN", func(t *testing.T) {
		h := NewHandler(fakeScheduleUsecase{})
		router := gin.New()
		router.POST("/booking/templates", h.CreateTemplate)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/booking/templates", nil)
		req.Header.Set("X-User-Role", "EXPERT")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("CreateTemplate - Success", func(t *testing.T) {
		templateID := uuid.New().String()
		h := NewHandler(fakeScheduleUsecase{
			createTimeTemplateFn: func(shiftName, startTime, endTime string, slotDuration int) (*domain.TimeTemplate, error) {
				return &domain.TimeTemplate{
					TemplateID:          templateID,
					ShiftName:           shiftName,
					StartTime:           startTime,
					EndTime:             endTime,
					SlotDurationMinutes: slotDuration,
					IsActive:            true,
				}, nil
			},
		})

		router := gin.New()
		router.POST("/booking/templates", h.CreateTemplate)

		w := httptest.NewRecorder()
		reqBody := map[string]interface{}{
			"shift_name":            "Morning Shift",
			"start_time":            "08:00",
			"end_time":              "12:00",
			"slot_duration_minutes": 60,
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/booking/templates", bytes.NewBuffer(bodyBytes))
		req.Header.Set("X-User-Role", "ADMIN")
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.True(t, resp["success"].(bool))
	})
}

func TestGetTemplates(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("GetTemplates - Success", func(t *testing.T) {
		h := NewHandler(fakeScheduleUsecase{
			listTimeTemplatesFn: func(query schedule.TemplateListQuery) (bookingquery.Page[domain.TimeTemplate], error) {
				return bookingquery.Page[domain.TimeTemplate]{
					Items:      []domain.TimeTemplate{{TemplateID: "t-123", ShiftName: "Shift 1", IsActive: true}},
					TotalItems: 1,
				}, nil
			},
		})

		router := gin.New()
		router.GET("/public/booking/templates", h.GetTemplates)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/public/booking/templates", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.True(t, resp["success"].(bool))
	})
}

func TestCreateAvailability(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("CreateAvailability - Forbidden for non-EXPERT", func(t *testing.T) {
		h := NewHandler(fakeScheduleUsecase{})
		router := gin.New()
		router.POST("/booking/availabilities", h.CreateAvailability)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/booking/availabilities", nil)
		req.Header.Set("X-User-Role", "PATIENT")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("CreateAvailability - Success", func(t *testing.T) {
		h := NewHandler(fakeScheduleUsecase{
			createAvailabilityFn: func(expertID, templateID string, dayOfWeek int, effectiveFrom int64, effectiveUntil *int64, price float64) (*domain.Availability, error) {
				return &domain.Availability{
					AvailabilityID: "avail-123",
					ExpertID:       expertID,
					TemplateID:     templateID,
					DayOfWeek:      dayOfWeek,
					Price:          &price,
				}, nil
			},
		})

		router := gin.New()
		router.POST("/booking/availabilities", h.CreateAvailability)

		w := httptest.NewRecorder()
		reqBody := map[string]interface{}{
			"template_id":    "temp-abc",
			"day_of_week":    1,
			"effective_from": 1711111111,
			"price":          150000,
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/booking/availabilities", bytes.NewBuffer(bodyBytes))
		req.Header.Set("X-User-Role", "EXPERT")
		req.Header.Set("X-User-Id", "expert-456")
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.True(t, resp["success"].(bool))
	})
}

func TestUpdateAvailability(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("UpdateAvailability - Success", func(t *testing.T) {
		h := NewHandler(fakeScheduleUsecase{
			updateAvailabilityFn: func(availID, expertID string, updates map[string]interface{}) error {
				return nil
			},
		})

		router := gin.New()
		router.PATCH("/booking/availabilities/:id", h.UpdateAvailability)

		w := httptest.NewRecorder()
		reqBody := map[string]interface{}{
			"price": 200000,
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("PATCH", "/booking/availabilities/avail-123", bytes.NewBuffer(bodyBytes))
		req.Header.Set("X-User-Role", "EXPERT")
		req.Header.Set("X-User-Id", "expert-456")
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.True(t, resp["success"].(bool))
	})
}
