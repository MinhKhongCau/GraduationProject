package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"profile-service/internal/infrastructure/http/handlers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSetupRoutes_RegistersAvatarUpload(t *testing.T) {
	r := gin.New()
	// Chỉ cần đăng ký route, không gọi handler, nên các handler khác có thể để nil.
	SetupRoutes(r, Handlers{Avatar: handlers.NewAvatarHandler(nil)})

	found := false
	for _, route := range r.Routes() {
		if route.Method == http.MethodPost && route.Path == "/api/v1/profiles/:id/avatar" {
			found = true
		}
	}
	assert.True(t, found, "POST /api/v1/profiles/:id/avatar must be registered")
}

func TestAvatarUpload_RequiresAuthHeader(t *testing.T) {
	r := gin.New()
	SetupRoutes(r, Handlers{Avatar: handlers.NewAvatarHandler(nil)})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/profiles/00000000-0000-0000-0000-000000000003/avatar", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
