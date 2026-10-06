package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSpecializationHandler(t *testing.T) {
	t.Run("CreateSpecialization - Success", func(t *testing.T) {
		_, mock := SetupTestDB(t)

		mock.ExpectQuery(`SELECT count\(\*\) FROM "specializations" WHERE code = \$1 OR slug = \$2`).
			WithArgs("SPEC-010", "tam-than-hoc").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "specializations"`).
			WithArgs("SPEC-010", "Tâm thần học", "tam-than-hoc", "Mental health care",
				"Tầng 3", "http://example.com/image.png", true, `["Mất ngủ","Lo âu"]`).
			WillReturnRows(sqlmock.NewRows([]string{"spec_id", "symptoms"}).
				AddRow(uuid.New(), []byte(`["Mất ngủ","Lo âu"]`)))
		mock.ExpectCommit()

		reqBody := map[string]interface{}{
			"code":        "spec-010",
			"name":        "Tâm thần học",
			"description": "Mental health care",
			"symptoms":    []string{"Mất ngủ", "Lo âu"},
			"location":    "Tầng 3",
			"image_url":   "http://example.com/image.png",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/profiles/specializations", bytes.NewBuffer(bodyBytes))

		CreateSpecialization(c)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assertBaseResponse(t, w, resp)
		result := resp["result"].(map[string]interface{})
		assert.Equal(t, "SPEC-010", result["code"])
		assert.Equal(t, "tam-than-hoc", result["slug"])
		assert.Equal(t, []interface{}{"Mất ngủ", "Lo âu"}, result["symptoms"])
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("CreateSpecialization - Trùng code/slug -> 409", func(t *testing.T) {
		_, mock := SetupTestDB(t)

		mock.ExpectQuery(`SELECT count\(\*\) FROM "specializations"`).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		bodyBytes, _ := json.Marshal(map[string]interface{}{"code": "SPEC-001", "name": "Tâm lý học lâm sàng"})

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/profiles/specializations", bytes.NewBuffer(bodyBytes))

		CreateSpecialization(c)

		assert.Equal(t, http.StatusConflict, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assertBaseResponse(t, w, resp)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("CreateSpecialization - Thiếu code -> 400", func(t *testing.T) {
		SetupTestDB(t)

		bodyBytes, _ := json.Marshal(map[string]interface{}{"name": "Psychiatry"})

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/profiles/specializations", bytes.NewBuffer(bodyBytes))

		CreateSpecialization(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("GetAllSpecializations - Success", func(t *testing.T) {
		_, mock := SetupTestDB(t)

		mock.ExpectQuery(`SELECT \* FROM "specializations"`).
			WithArgs(true).
			WillReturnRows(sqlmock.NewRows([]string{"spec_id", "name", "is_active"}).
				AddRow(uuid.New(), "Psychiatry", true))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/profiles/specializations", nil)

		GetAllSpecializations(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assertBaseResponse(t, w, resp)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
