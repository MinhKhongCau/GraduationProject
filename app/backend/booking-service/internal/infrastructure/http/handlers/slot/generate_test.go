package handler

import (
	"booking-service/internal/application/slot"
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeGenerationApplication struct {
	result slot.GenerationResult
	err    error
}

func (u fakeGenerationApplication) GenerateExpert(context.Context, string, int) (slot.GenerationResult, error) {
	return u.result, u.err
}
func (fakeGenerationApplication) GenerateAll(context.Context, int) slot.GenerationRunSummary {
	return slot.GenerationRunSummary{}
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
			response := performGenerate(fakeGenerationApplication{result: tt.result})
			if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), tt.want) {
				t.Fatalf("unexpected response %d: %s", response.Code, response.Body.String())
			}
		})
	}
}

func TestGenerateMapsScheduleConflictWithoutFalseSuccess(t *testing.T) {
	response := performGenerate(fakeGenerationApplication{err: errors.Join(slot.ErrSlotOverlap, errors.New("overlap"))})
	if response.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", response.Code, response.Body.String())
	}
	if strings.Contains(response.Body.String(), "Slots generated successfully") {
		t.Fatal("conflict returned false success")
	}
}

func performGenerate(generation slot.GenerationApplication) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	h := &Handler{generation: generation}
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
