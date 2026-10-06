package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"profile-service/internal/infrastructure/http/middleware"
	"profile-service/internal/infrastructure/persistence/models"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreateProfileInternal(t *testing.T) {
	t.Run("TC_PROF_MNG_01 - Tạo profile nội bộ thành công", func(t *testing.T) {
		_, mock := SetupTestDB(t)

		authUUID := uuid.New()
		reqBody := map[string]string{
			"auth_id":   authUUID.String(),
			"full_name": "Patient A",
			"role":      "PATIENT",
			"email":     "patient@example.com",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		// 1. Mock query check duplicate (First) -> trả về record not found (empty row)
		mock.ExpectQuery(`SELECT \* FROM "profiles" WHERE auth_id = \$1`).
			WithArgs(authUUID, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		// 2. Mock GORM INSERT Transaction
		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "profiles"`).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
				AddRow(uuid.New(), time.Now(), time.Now()))

		mock.ExpectExec(`INSERT INTO "patient_profiles"`).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/internal/api/v1/profiles/create", bytes.NewBuffer(bodyBytes))

		CreateProfileInternal(c)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assertBaseResponse(t, w, resp)
		assert.Equal(t, "Tạo profile thành công", resp["message"])

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("TC_PROF_MNG_02 - Tạo profile nội bộ thất bại do trùng lặp (Conflict)", func(t *testing.T) {
		_, mock := SetupTestDB(t)

		authUUID := uuid.New()
		reqBody := map[string]string{
			"auth_id":   authUUID.String(),
			"full_name": "Patient A",
			"role":      "PATIENT",
			"email":     "patient@example.com",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		// Mock query check duplicate (First) -> trả về 1 record đã tồn tại
		mock.ExpectQuery(`SELECT \* FROM "profiles" WHERE auth_id = \$1`).
			WithArgs(authUUID, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "auth_id"}).AddRow(uuid.New(), authUUID))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/internal/api/v1/profiles/create", bytes.NewBuffer(bodyBytes))

		CreateProfileInternal(c)

		assert.Equal(t, http.StatusConflict, w.Code)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assertBaseResponse(t, w, resp)
		assert.Equal(t, "Profile cho tài khoản này đã tồn tại", resp["message"])

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGetMe(t *testing.T) {
	t.Run("TC_PROF_MNG_03 - Xem profile bản thân thành công", func(t *testing.T) {
		_, mock := SetupTestDB(t)

		authUUID := uuid.New()
		profileUUID := uuid.New()

		// Mock query find profile (First)
		mock.ExpectQuery(`SELECT \* FROM "profiles" WHERE auth_id = \$1`).
			WithArgs(authUUID, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "full_name", "role", "auth_id"}).
				AddRow(profileUUID, "patient-a", "Patient A", string(models.RolePatient), authUUID))

		// Mock preload AdminProfile
		mock.ExpectQuery(`SELECT \* FROM "admin_profiles" WHERE "admin_profiles"."profile_id" = \$1`).
			WithArgs(profileUUID).
			WillReturnRows(sqlmock.NewRows([]string{"profile_id"}))

		// Mock preload ExpertProfile
		mock.ExpectQuery(`SELECT \* FROM "expert_profiles" WHERE "expert_profiles"."profile_id" = \$1`).
			WithArgs(profileUUID).
			WillReturnRows(sqlmock.NewRows([]string{"profile_id"}))

		// Mock preload PatientProfile
		mock.ExpectQuery(`SELECT \* FROM "patient_profiles" WHERE "patient_profiles"."profile_id" = \$1`).
			WithArgs(profileUUID).
			WillReturnRows(sqlmock.NewRows([]string{"profile_id", "email", "phone_number"}).
				AddRow(profileUUID, "patient@example.com", "123456"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/profiles/me", nil)

		// Set auth_id context giả lập từ JWT middleware
		c.Set(middleware.CtxAuthID, authUUID.String())

		GetMe(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assertBaseResponse(t, w, resp)

		data := resp["result"].(map[string]interface{})
		assert.Equal(t, "Patient A", data["user_information"].(map[string]interface{})["full_name"])
		assert.Equal(t, string(models.RolePatient), data["role"])

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUpdateMe(t *testing.T) {
	t.Run("TC_PROF_MNG_04 - Cập nhật hồ sơ thành công (PATIENT)", func(t *testing.T) {
		_, mock := SetupTestDB(t)

		authUUID := uuid.New()
		profileUUID := uuid.New()

		reqBody := map[string]interface{}{
			"user_information": map[string]string{
				"full_name":     "New Patient Name",
				"phone_number":  "0911222333",
				"date_of_birth": "1995-05-15",
				"gender":        "MALE",
				"country":       "Vietnam",
			},
			"email":   "patient-new@example.com",
			"address": "123 Street",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		// 1. Mock query tìm profile cũ để update
		mock.ExpectQuery(`SELECT \* FROM "profiles" WHERE auth_id = \$1`).
			WithArgs(authUUID, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "full_name", "role", "auth_id"}).
				AddRow(profileUUID, "patient-a", "Old Patient Name", string(models.RolePatient), authUUID))

		// 2. Mock preload các sub-profiles
		mock.ExpectQuery(`SELECT \* FROM "admin_profiles"`).
			WithArgs(profileUUID).
			WillReturnRows(sqlmock.NewRows([]string{"profile_id"}))

		mock.ExpectQuery(`SELECT \* FROM "expert_profiles"`).
			WithArgs(profileUUID).
			WillReturnRows(sqlmock.NewRows([]string{"profile_id"}))

		mock.ExpectQuery(`SELECT \* FROM "patient_profiles"`).
			WithArgs(profileUUID).
			WillReturnRows(sqlmock.NewRows([]string{"profile_id", "email", "phone_number"}).
				AddRow(profileUUID, "patient-old@example.com", "000000"))

		// 3. Mock saveProfileAndSub transaction
		mock.ExpectBegin()
		// Mock Update Profile Name
		mock.ExpectExec(`UPDATE "profiles"`).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Mock Save PatientProfile (GORM uses upsert for Save)
		mock.ExpectExec(`(INSERT INTO "patient_profiles"|UPDATE "patient_profiles")`).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("PUT", "/api/v1/profiles/me", bytes.NewBuffer(bodyBytes))
		c.Set(middleware.CtxAuthID, authUUID.String())

		UpdateMe(nil)(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assertBaseResponse(t, w, resp)
		assert.Equal(t, "Cập nhật hồ sơ thành công", resp["message"])

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("TC_PROF_MNG_05 - Cập nhật hồ sơ thất bại do validation lỗi", func(t *testing.T) {
		_, mock := SetupTestDB(t)

		authUUID := uuid.New()
		profileUUID := uuid.New()

		// Email sai định dạng
		reqBody := map[string]interface{}{
			"user_information": map[string]string{"full_name": "New Patient Name"},
			"email":            "wrong-email-format",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		// 1. Mock query tìm profile cũ
		mock.ExpectQuery(`SELECT \* FROM "profiles" WHERE auth_id = \$1`).
			WithArgs(authUUID, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "full_name", "role", "auth_id"}).
				AddRow(profileUUID, "patient-a", "Old Patient Name", string(models.RolePatient), authUUID))

		// 2. Mock preload các sub-profiles
		mock.ExpectQuery(`SELECT \* FROM "admin_profiles"`).
			WithArgs(profileUUID).
			WillReturnRows(sqlmock.NewRows([]string{"profile_id"}))

		mock.ExpectQuery(`SELECT \* FROM "expert_profiles"`).
			WithArgs(profileUUID).
			WillReturnRows(sqlmock.NewRows([]string{"profile_id"}))

		mock.ExpectQuery(`SELECT \* FROM "patient_profiles"`).
			WithArgs(profileUUID).
			WillReturnRows(sqlmock.NewRows([]string{"profile_id"}))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("PUT", "/api/v1/profiles/me", bytes.NewBuffer(bodyBytes))
		c.Set(middleware.CtxAuthID, authUUID.String())

		UpdateMe(nil)(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assertBaseResponse(t, w, resp)
		assert.Equal(t, "Dữ liệu không hợp lệ", resp["message"])

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
