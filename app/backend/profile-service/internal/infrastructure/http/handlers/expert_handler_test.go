package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"profile-service/internal/infrastructure/persistence/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestListExperts(t *testing.T) {
	t.Run("TC_PROF_EXP_01 - Tìm kiếm chuyên gia thành công có kết quả", func(t *testing.T) {
		_, mock := SetupTestDB(t)

		profileID := uuid.New()
		now := time.Now()

		// 1. Mock query COUNT
		mock.ExpectQuery(`SELECT count\(\*\) FROM "profiles"`).
			WithArgs(string(models.RoleExpert), "%Dr. A%").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		// 2. Mock query SELECT profiles
		mock.ExpectQuery(`SELECT \* FROM "profiles"`).
			WithArgs(string(models.RoleExpert), "%Dr. A%", 10).
			WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "name", "role", "created_at", "updated_at", "deleted_at"}).
				AddRow(profileID, "dr-a", "Dr. A", string(models.RoleExpert), now, now, nil))

		// 3. Mock query SELECT preloaded ExpertProfile
		mock.ExpectQuery(`SELECT \* FROM "expert_profiles"`).
			WithArgs(profileID).
			WillReturnRows(sqlmock.NewRows([]string{"profile_id", "email", "phone_number", "verification_status"}).
				AddRow(profileID, "dr-a@example.com", "123456789", "UNVERIFIED"))

		specUUID := uuid.New()

		// 4. Mock query SELECT preloaded Specializations (ManyToMany)
		mock.ExpectQuery(`SELECT \* FROM "expert_specializations"`).
			WithArgs(profileID).
			WillReturnRows(sqlmock.NewRows([]string{"expert_profile_id", "spec_id"}).
				AddRow(profileID, specUUID))

		// 5. Mock query SELECT Specializations details
		mock.ExpectQuery(`SELECT \* FROM "specializations"`).
			WillReturnRows(sqlmock.NewRows([]string{"spec_id", "name", "description"}).
				AddRow(specUUID, "Chuyên khoa tâm thần", "Mô tả"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/profiles/experts?search=Dr.+A&page=1&page_size=10", nil)

		ListExperts(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp["success"].(bool))
		assert.Equal(t, "Lấy danh sách chuyên gia thành công", resp["message"])

		data := resp["data"].(map[string]interface{})
		assert.Equal(t, float64(1), data["total_items"])
		assert.Equal(t, float64(1), data["total_pages"])
		items := data["items"].([]interface{})
		assert.Len(t, items, 1)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("TC_PROF_EXP_02 - Tìm kiếm chuyên gia không ra kết quả", func(t *testing.T) {
		_, mock := SetupTestDB(t)

		// Mock query COUNT
		mock.ExpectQuery(`SELECT count\(\*\) FROM "profiles"`).
			WithArgs(string(models.RoleExpert), "%Unknown%").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		// Mock query SELECT profiles (Find)
		mock.ExpectQuery(`SELECT \* FROM "profiles"`).
			WithArgs(string(models.RoleExpert), "%Unknown%", 20).
			WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "name", "role"}))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/profiles/experts?search=Unknown", nil)

		ListExperts(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.True(t, resp["success"].(bool))

		data := resp["data"].(map[string]interface{})
		assert.Equal(t, float64(0), data["total_items"])
		items := data["items"].([]interface{})
		assert.Len(t, items, 0)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("TC_PROF_EXP_03 - Lỗi truy vấn cơ sở dữ liệu", func(t *testing.T) {
		_, mock := SetupTestDB(t)

		// Mock query COUNT trả về lỗi
		mock.ExpectQuery(`SELECT count\(\*\) FROM "profiles"`).
			WillReturnError(errors.New("connection failed"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/profiles/experts", nil)

		ListExperts(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.False(t, resp["success"].(bool))
		assert.Equal(t, "Lỗi truy vấn danh sách chuyên gia", resp["message"])

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGetExpert(t *testing.T) {
	t.Run("TC_PROF_EXP_04 - Xem chi tiết chuyên gia thành công", func(t *testing.T) {
		_, mock := SetupTestDB(t)

		profileID := uuid.New()
		expertUUID := uuid.New()

		// Mock query Find profile by ID (First)
		mock.ExpectQuery(`SELECT \* FROM "profiles"`).
			WithArgs(expertUUID, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "name", "role", "auth_id"}).
				AddRow(profileID, "dr-b", "Dr. B", string(models.RoleExpert), expertUUID))

		// Mock preload AdminProfile
		mock.ExpectQuery(`SELECT \* FROM "admin_profiles"`).
			WithArgs(profileID).
			WillReturnRows(sqlmock.NewRows([]string{"profile_id"}))

		// Mock preload ExpertProfile
		mock.ExpectQuery(`SELECT \* FROM "expert_profiles"`).
			WithArgs(profileID).
			WillReturnRows(sqlmock.NewRows([]string{"profile_id", "email", "phone_number"}).
				AddRow(profileID, "dr-b@example.com", "0987654321"))

		// Mock preload ExpertProfile.Specializations
		mock.ExpectQuery(`SELECT \* FROM "expert_specializations"`).
			WithArgs(profileID).
			WillReturnRows(sqlmock.NewRows([]string{"expert_profile_id", "specialization_spec_id"}))

		// Mock preload PatientProfile
		mock.ExpectQuery(`SELECT \* FROM "patient_profiles"`).
			WithArgs(profileID).
			WillReturnRows(sqlmock.NewRows([]string{"profile_id"}))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: expertUUID.String()}}

		GetExpert(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp["success"].(bool))

		data := resp["data"].(map[string]interface{})
		assert.Equal(t, "Dr. B", data["name"])
		assert.Equal(t, string(models.RoleExpert), data["role"])

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("TC_PROF_EXP_05 - Xem chi tiết chuyên gia không tồn tại", func(t *testing.T) {
		_, mock := SetupTestDB(t)

		expertUUID := uuid.New()

		// Mock query First trả về record not found
		mock.ExpectQuery(`SELECT \* FROM "profiles"`).
			WithArgs(expertUUID, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "name", "role"})) // Row rỗng tức là Not Found trong sqlmock

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: expertUUID.String()}}

		GetExpert(c)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.False(t, resp["success"].(bool))
		assert.Equal(t, "Không tìm thấy hồ sơ", resp["message"])

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
