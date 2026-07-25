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

		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "specializations"`).
			WillReturnRows(sqlmock.NewRows([]string{"spec_id"}).AddRow(uuid.New()))
		mock.ExpectCommit()

		reqBody := map[string]string{
			"name":        "Psychiatry",
			"description": "Mental health care",
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
		assert.True(t, resp["success"].(bool))
		assert.NoError(t, mock.ExpectationsWereMet())
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
		assert.True(t, resp["success"].(bool))
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
