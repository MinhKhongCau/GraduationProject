package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"profile-service/internal/application/expertprofile"
	"profile-service/internal/domain/expert"
	"profile-service/internal/domain/profile"
	"profile-service/internal/domain/specialization"
	"profile-service/internal/infrastructure/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type stubExpertRepo struct{ snapshot *expert.Snapshot }

func (r stubExpertRepo) FindByAuthID(_ context.Context, authID uuid.UUID) (*expert.Expert, error) {
	if r.snapshot == nil || r.snapshot.AuthID != authID {
		return nil, expert.ErrExpertNotFound
	}
	return expert.Reconstitute(*r.snapshot), nil
}

func (stubExpertRepo) Save(context.Context, *expert.Expert) error { return nil }

type stubSpecRepo struct{}

func (stubSpecRepo) FindByIDs(context.Context, []string) ([]specialization.Specialization, error) {
	return nil, nil
}

type stubPublisher struct{}

func (stubPublisher) Publish(context.Context, []expert.Event) {}

func newTestExpertHandler(snapshot *expert.Snapshot) *ExpertHandler {
	repo := stubExpertRepo{snapshot: snapshot}
	return NewExpertHandler(
		expertprofile.NewReplaceExpertProfile(repo, stubSpecRepo{}, stubPublisher{}),
		expertprofile.NewPatchExpertProfile(repo, stubSpecRepo{}, stubPublisher{}),
	)
}

func serveExpert(h *ExpertHandler, method, id, role string, body any) (*httptest.ResponseRecorder, map[string]any) {
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set(middleware.CtxRole, role) })
	r.PUT("/experts/:id", h.Update)
	r.PATCH("/experts/:id", h.Patch)

	raw, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(method, "/experts/"+id, bytes.NewBuffer(raw)))

	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	return w, resp
}

func TestExpertHandler(t *testing.T) {
	authID := uuid.New()
	snapshot := &expert.Snapshot{ProfileID: uuid.New(), AuthID: authID, UserInformation: profile.UserInformation{FullName: "Dr. A"}, VerificationStatus: "PENDING"}

	t.Run("PUT thành công trả về hồ sơ, giữ trạng thái xác minh", func(t *testing.T) {
		w, resp := serveExpert(newTestExpertHandler(snapshot), http.MethodPut, authID.String(), "ADMIN",
			map[string]any{
				"user_information":    map[string]any{"full_name": "Dr. B", "country": "Vietnam"},
				"verification_status": "VERIFIED",
			})

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "Cập nhật hồ sơ thành công", resp["message"])
		data := resp["result"].(map[string]any)
		info := data["user_information"].(map[string]any)
		assert.Equal(t, "Dr. B", info["full_name"])
		assert.Equal(t, "Vietnam", info["country"])
		assert.Equal(t, "EXPERT", data["role"])
		assert.Equal(t, "PENDING", data["expert_profile"].(map[string]any)["verification_status"])
	})

	t.Run("PUT thiếu full_name -> 400", func(t *testing.T) {
		w, resp := serveExpert(newTestExpertHandler(snapshot), http.MethodPut, authID.String(), "ADMIN", map[string]any{})
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "Dữ liệu không hợp lệ", resp["message"])
	})

	t.Run("id không phải UUID -> 404", func(t *testing.T) {
		w, resp := serveExpert(newTestExpertHandler(snapshot), http.MethodPut, "abc", "ADMIN",
			map[string]any{"user_information": map[string]any{"full_name": "X"}})
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, "Không tìm thấy hồ sơ chuyên gia", resp["message"])
	})

	t.Run("không tìm thấy chuyên gia -> 404 kèm chi tiết record not found", func(t *testing.T) {
		w, resp := serveExpert(newTestExpertHandler(nil), http.MethodPatch, uuid.NewString(), "ADMIN", map[string]any{})
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, "record not found", resp["result"].(map[string]any)["error"])
	})

	t.Run("PATCH verification_status bởi không phải Admin -> 403", func(t *testing.T) {
		w, resp := serveExpert(newTestExpertHandler(snapshot), http.MethodPatch, authID.String(), "EXPERT",
			map[string]any{"verification_status": "VERIFIED"})
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Equal(t, "Chỉ Admin mới có quyền thay đổi trạng thái xác minh", resp["message"])
	})

	t.Run("PATCH verification_status sai giá trị -> 400", func(t *testing.T) {
		w, _ := serveExpert(newTestExpertHandler(snapshot), http.MethodPatch, authID.String(), "ADMIN",
			map[string]any{"verification_status": "APPROVED"})
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("PATCH bởi Admin đổi trạng thái", func(t *testing.T) {
		w, resp := serveExpert(newTestExpertHandler(snapshot), http.MethodPatch, authID.String(), "ADMIN",
			map[string]any{"verification_status": "VERIFIED"})
		assert.Equal(t, http.StatusOK, w.Code)
		data := resp["result"].(map[string]any)
		assert.Equal(t, "VERIFIED", data["expert_profile"].(map[string]any)["verification_status"])
	})
}
