package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"profile-service/internal/infrastructure/persistence/models"

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
			WillReturnRows(sqlmock.NewRows([]string{"id", "full_name", "role"}).
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
		assertBaseResponse(t, w, resp)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetPatient - Success", func(t *testing.T) {
		_, mock := SetupTestDB(t)

		profileID := uuid.New()
		authUUID := uuid.New()

		// 1. Mock select profile (findRoleProfileByAuthID)
		mock.ExpectQuery(`SELECT \* FROM "profiles"`).
			WithArgs(authUUID, string(models.RolePatient), 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "auth_id", "role", "full_name"}).
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
		assertBaseResponse(t, w, resp)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("UpdatePatient - Success", func(t *testing.T) {
		_, mock := SetupTestDB(t)

		profileID := uuid.New()
		authUUID := uuid.New()

		// 1. Mock find profile (findRoleProfileByAuthID)
		mock.ExpectQuery(`SELECT \* FROM "profiles"`).
			WithArgs(authUUID, string(models.RolePatient), 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "auth_id", "role", "full_name"}).
				AddRow(profileID, authUUID, string(models.RolePatient), "Old Name"))

		// 2. Mock transaction begin
		mock.ExpectBegin()

		// 3. Mock update user information trên bảng profiles (cột theo thứ tự alphabet)
		mock.ExpectExec(`UPDATE "profiles" SET "country"=\$1,"date_of_birth"=\$2,"full_name"=\$3,"gender"=\$4,"phone_number"=\$5,"updated_at"=\$6 WHERE id = \$7`).
			WithArgs("Vietnam", sqlmock.AnyArg(), "New Patient Name", "FEMALE", "0987654321", sqlmock.AnyArg(), profileID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// 4. Mock save PatientProfile (insert or update)
		mock.ExpectExec(`(INSERT INTO "patient_profiles"|UPDATE "patient_profiles")`).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		reqBody := map[string]interface{}{
			"user_information": map[string]interface{}{
				"full_name":     "New Patient Name",
				"phone_number":  "0987654321",
				"date_of_birth": "1995-10-10",
				"gender":        "FEMALE",
				"country":       "Vietnam",
			},
			"email":   "patient@new.com",
			"address": "123 Street",
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
		assertBaseResponse(t, w, resp)
		info := resp["result"].(map[string]interface{})["user_information"].(map[string]interface{})
		assert.Equal(t, "New Patient Name", info["full_name"])
		assert.Equal(t, "Vietnam", info["country"])
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("UpdatePatient - Gender không hợp lệ -> 400", func(t *testing.T) {
		_, mock := SetupTestDB(t)

		profileID := uuid.New()
		authUUID := uuid.New()

		mock.ExpectQuery(`SELECT \* FROM "profiles"`).
			WithArgs(authUUID, string(models.RolePatient), 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "auth_id", "role", "full_name"}).
				AddRow(profileID, authUUID, string(models.RolePatient), "Old Name"))

		bodyBytes, _ := json.Marshal(map[string]interface{}{
			"user_information": map[string]interface{}{"full_name": "X", "gender": "UNKNOWN"},
		})

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: authUUID.String()}}
		c.Request = httptest.NewRequest("PUT", "/api/v1/profiles/patients/"+authUUID.String(), bytes.NewBuffer(bodyBytes))

		UpdatePatient(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assertBaseResponse(t, w, resp)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
