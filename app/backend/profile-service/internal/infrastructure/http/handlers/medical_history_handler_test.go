package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"profile-service/internal/infrastructure/http/middleware"
	"profile-service/internal/infrastructure/persistence/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMedicalHistoryHandler(t *testing.T) {
	t.Run("ListMyMedicalHistories - Success", func(t *testing.T) {
		_, mock := SetupTestDB(t)

		profileID := uuid.New()
		authID := uuid.New()

		// 1. Mock find profile (calls findProfileByAuthID which preloads patient, expert, admin profiles)
		mock.ExpectQuery(`SELECT \* FROM "profiles"`).
			WithArgs(authID, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "auth_id", "role"}).
				AddRow(profileID, authID, string(models.RolePatient)))

		mock.ExpectQuery(`SELECT \* FROM "admin_profiles"`).
			WithArgs(profileID).
			WillReturnRows(sqlmock.NewRows([]string{"profile_id"}))

		mock.ExpectQuery(`SELECT \* FROM "expert_profiles"`).
			WithArgs(profileID).
			WillReturnRows(sqlmock.NewRows([]string{"profile_id"}))

		mock.ExpectQuery(`SELECT \* FROM "patient_profiles"`).
			WithArgs(profileID).
			WillReturnRows(sqlmock.NewRows([]string{"profile_id"}).AddRow(profileID))

		// 2. Mock list medical histories
		mock.ExpectQuery(`SELECT \* FROM "medical_histories"`).
			WithArgs(profileID, true).
			WillReturnRows(sqlmock.NewRows([]string{"id", "patient_profile_id", "condition_name", "is_active"}).
				AddRow(1, profileID, "Flu", true))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(middleware.CtxAuthID, authID.String())
		c.Request = httptest.NewRequest("GET", "/api/v1/profiles/me/medical-histories", nil)

		ListMyMedicalHistories(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assertBaseResponse(t, w, resp)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("AddMyMedicalHistory - Success", func(t *testing.T) {
		_, mock := SetupTestDB(t)

		profileID := uuid.New()
		authID := uuid.New()

		// 1. Mock find profile (preloaded)
		mock.ExpectQuery(`SELECT \* FROM "profiles"`).
			WithArgs(authID, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "auth_id", "role"}).
				AddRow(profileID, authID, string(models.RolePatient)))

		mock.ExpectQuery(`SELECT \* FROM "admin_profiles"`).
			WithArgs(profileID).
			WillReturnRows(sqlmock.NewRows([]string{"profile_id"}))

		mock.ExpectQuery(`SELECT \* FROM "expert_profiles"`).
			WithArgs(profileID).
			WillReturnRows(sqlmock.NewRows([]string{"profile_id"}))

		mock.ExpectQuery(`SELECT \* FROM "patient_profiles"`).
			WithArgs(profileID).
			WillReturnRows(sqlmock.NewRows([]string{"profile_id"}).AddRow(profileID))

		// 2. Mock create medical history
		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "medical_histories"`).
			WillReturnRows(sqlmock.NewRows([]string{"history_id"}).AddRow(uuid.New()))
		mock.ExpectCommit()

		reqBody := map[string]interface{}{
			"condition_name": "Asthma",
			"description":    "Chronic condition",
			"is_chronic":     true,
			"diagnosed_at":   "2026-01-01",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(middleware.CtxAuthID, authID.String())
		c.Request = httptest.NewRequest("POST", "/api/v1/profiles/me/medical-histories", bytes.NewBuffer(bodyBytes))

		AddMyMedicalHistory(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assertBaseResponse(t, w, resp)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ListPatientMedicalHistories - Success (Admin)", func(t *testing.T) {
		_, mock := SetupTestDB(t)

		profileID := uuid.New()
		patientAuthUUID := uuid.New()

		// 1. Mock find profile by param ID (calls findRoleProfileByAuthID)
		mock.ExpectQuery(`SELECT \* FROM "profiles"`).
			WithArgs(patientAuthUUID, string(models.RolePatient), 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "auth_id", "role"}).
				AddRow(profileID, patientAuthUUID, string(models.RolePatient)))

		// 2. Mock list medical histories
		mock.ExpectQuery(`SELECT \* FROM "medical_histories"`).
			WithArgs(profileID, true).
			WillReturnRows(sqlmock.NewRows([]string{"id", "patient_profile_id", "condition_name"}).
				AddRow(1, profileID, "COVID-19"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: patientAuthUUID.String()}}
		c.Request = httptest.NewRequest("GET", "/api/v1/profiles/patients/"+patientAuthUUID.String()+"/medical-histories", nil)

		ListPatientMedicalHistories(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assertBaseResponse(t, w, resp)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
