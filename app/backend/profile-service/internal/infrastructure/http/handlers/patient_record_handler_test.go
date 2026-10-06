package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"profile-service/internal/infrastructure/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreateMinePatientRecordValidation(t *testing.T) {
	valid := map[string]string{
		"full_name":     "Nguyen Van B",
		"date_of_birth": "1960-05-20",
		"gender":        "MALE",
		"phone_number":  "0900000000",
		"relationship":  "PARENT",
	}
	with := func(key, value string) map[string]string {
		body := map[string]string{}
		for k, v := range valid {
			body[k] = v
		}
		body[key] = value
		return body
	}

	tests := map[string]map[string]string{
		"SELF không được tạo thủ công": with("relationship", "SELF"),
		"thiếu ngày sinh":              with("date_of_birth", ""),
		"ngày sinh sai định dạng":      with("date_of_birth", "20/05/1960"),
		"ngày sinh ở tương lai":        with("date_of_birth", time.Now().AddDate(1, 0, 0).Format("2006-01-02")),
		"giới tính không hợp lệ":       with("gender", "UNKNOWN"),
		"họ tên chỉ có khoảng trắng":   with("full_name", "   "),
		"email không hợp lệ":           with("email", "not-an-email"),
	}

	// Mọi trường hợp đều bị chặn trước khi chạm tới repository/DB.
	h := NewPatientRecordHandler(nil)
	for name, body := range tests {
		t.Run(name, func(t *testing.T) {
			payload, _ := json.Marshal(body)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Set(middleware.CtxAuthID, uuid.NewString())
			c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/profiles/me/patient-records", bytes.NewBuffer(payload))
			c.Request.Header.Set("Content-Type", "application/json")

			h.CreateMine(c)

			assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
			var resp map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &resp)
			assertBaseResponse(t, w, resp)
		})
	}
}
