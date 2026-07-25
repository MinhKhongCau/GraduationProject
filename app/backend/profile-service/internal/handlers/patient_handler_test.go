package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"profile-service/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestPatientHandler(t *testing.T) {
	t.Run("ListPatients - Success", func(t *testing.T) {
		_, mock := SetupTestDB(t)

		profileID := uuid.New()

		// 1. Mock count query
		mock.ExpectQuery(`SELECT count\(\*\) FROM "profiles"`).
			WithArgs(string(models.RolePatient)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		// 2. Mock profiles select query
		mock.ExpectQuery(`SELECT \* FROM "profiles"`).
			WithArgs(string(models.RolePatient), 20).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "role"}).
				AddRow(profileID, "Patient One", string(models.RolePatient)))

		// 3. Mock preloaded PatientProfile query
		mock.ExpectQuery(`SELECT \* FROM "patient_profiles"`).
			WithArgs(profileID).
			WillReturnRows(sqlmock.NewRows([]string{"profile_id", "phone_number"}).
				AddRow(profileID, "0912345678"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/profiles/patients", nil)

		ListPatients(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.True(t, resp["success"].(bool))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetPatient - Success", func(t *testing.T) {
		_, mock := SetupTestDB(t)

		profileID := uuid.New()
		authUUID := uuid.New()

		// 1. Mock select profile (findRoleProfileByAuthID)
		mock.ExpectQuery(`SELECT \* FROM "profiles"`).
			WithArgs(authUUID, string(models.RolePatient), 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "auth_id", "role", "name"}).
				AddRow(profileID, authUUID, string(models.RolePatient), "Patient One"))

		// 2. Mock preload PatientProfile
		mock.ExpectQuery(`SELECT \* FROM "patient_profiles"`).
			WithArgs(profileID).
			WillReturnRows(sqlmock.NewRows([]string{"profile_id", "phone_number"}).
				AddRow(profileID, "0912345678"))

		// 3. Mock preload MedicalHistories
		mock.ExpectQuery(`SELECT \* FROM "medical_histories"`).
			WithArgs(profileID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "patient_profile_id", "condition_name"}))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: authUUID.String()}}
		c.Request = httptest.NewRequest("GET", "/api/v1/profiles/patients/"+authUUID.String(), nil)

		GetPatient(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.True(t, resp["success"].(bool))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("UpdatePatient - Success", func(t *testing.T) {
		_, mock := SetupTestDB(t)

		profileID := uuid.New()
		authUUID := uuid.New()

		// 1. Mock find profile (findRoleProfileByAuthID)
		mock.ExpectQuery(`SELECT \* FROM "profiles"`).
			WithArgs(authUUID, string(models.RolePatient), 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "auth_id", "role", "name"}).
				AddRow(profileID, authUUID, string(models.RolePatient), "Old Name"))

		// 2. Mock transaction begin
		mock.ExpectBegin()

		// 3. Mock update Profile name
		mock.ExpectExec(`UPDATE "profiles" SET "name"=\$1`).
			WithArgs("New Patient Name", sqlmock.AnyArg(), profileID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// 4. Mock save PatientProfile (insert or update)
		mock.ExpectExec(`(INSERT INTO "patient_profiles"|UPDATE "patient_profiles")`).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		reqBody := map[string]interface{}{
			"name":          "New Patient Name",
			"phone_number":  "0987654321",
			"email":         "patient@new.com",
			"date_of_birth": "1995-10-10",
			"gender":        "FEMALE",
			"address":       "123 Street",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: authUUID.String()}}
		c.Request = httptest.NewRequest("PUT", "/api/v1/profiles/patients/"+authUUID.String(), bytes.NewBuffer(bodyBytes))

		UpdatePatient(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.True(t, resp["success"].(bool))
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
