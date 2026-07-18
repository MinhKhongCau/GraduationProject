package handler

import (
	"booking-service/internal/booking/domain"
	"booking-service/internal/slot"
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type fakeGenerationUsecase struct {
	result slot.GenerationResult
	err    error
}

func (u fakeGenerationUsecase) GenerateSlotsForNextDays(string, int, []domain.Availability, []domain.TimeTemplate, []domain.ExpertTimeOff) (slot.GenerationResult, error) {
	return u.result, u.err
}
func (fakeGenerationUsecase) LockSlot(string, string) error { return nil }
func (fakeGenerationUsecase) GetDates(string, time.Time, time.Time) ([]string, error) {
	return nil, nil
}
func (fakeGenerationUsecase) GetTimes(string, string) ([]slot.SlotTimeResult, error) { return nil, nil }

type fakeScheduleProvider struct{}

func (fakeScheduleProvider) GetAvailabilities(string) ([]domain.Availability, error) { return nil, nil }
func (fakeScheduleProvider) GetAllTimeTemplates() ([]domain.TimeTemplate, error)     { return nil, nil }

type fakeTimeOffProvider struct{}

func (fakeTimeOffProvider) GetTimeOffs(string, time.Time) ([]domain.ExpertTimeOff, error) {
	return nil, nil
}

func TestGenerateReportsActualInsertedCount(t *testing.T) {
	tests := []struct {
		name   string
		result slot.GenerationResult
		want   string
	}{
		{"first generation", slot.GenerationResult{Candidates: 3, Inserted: 2}, `"slots_created":2`},
		{"idempotent generation", slot.GenerationResult{Candidates: 3, Inserted: 0}, `"slots_created":0`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := performGenerate(fakeGenerationUsecase{result: tt.result})
			if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), tt.want) {
				t.Fatalf("unexpected response %d: %s", response.Code, response.Body.String())
			}
		})
	}
}

func TestGenerateMapsScheduleConflictWithoutFalseSuccess(t *testing.T) {
	response := performGenerate(fakeGenerationUsecase{err: errors.Join(slot.ErrSlotOverlap, errors.New("overlap"))})
	if response.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", response.Code, response.Body.String())
	}
	if strings.Contains(response.Body.String(), "Slots generated successfully") {
		t.Fatal("conflict returned false success")
	}
}

func performGenerate(usecase slot.Usecase) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	h := &Handler{usecase: usecase, scheduleRepo: fakeScheduleProvider{}, timeoffRepo: fakeTimeOffProvider{}}
	router := gin.New()
	router.POST("/api/v1/booking/slots/generate", h.Generate)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/booking/slots/generate", bytes.NewBufferString(`{"days_to_generate":1}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-User-Role", "EXPERT")
	request.Header.Set("X-User-Id", "expert")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}
