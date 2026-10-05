// File: internal/infrastructure/http/middleware/auth.go
package middleware

import (
	"net/http"

	"profile-service/internal/infrastructure/http/response"

	"github.com/gin-gonic/gin"
)

// Context keys dùng để lưu thông tin định danh do Gateway (Kong + mindcare-auth plugin) chèn vào.
const (
	CtxAuthID = "authID"
	CtxRole   = "userRole"
)

// RequireAuth đọc header X-User-Id / X-User-Role do API Gateway chèn vào sau khi verify JWT.
// Service không tự parse JWT - tin tưởng Gateway đã xác thực trước khi request tới đây.
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authID := c.GetHeader("X-User-Id")
		if authID == "" {
			response.Error(c, http.StatusUnauthorized, "Không thể xác định danh tính người dùng", "Missing X-User-Id header")
			c.Abort()
			return
		}

		c.Set(CtxAuthID, authID)
		c.Set(CtxRole, c.GetHeader("X-User-Role"))
		c.Next()
	}
}

// RequireRole đảm bảo vai trò của người dùng (từ header X-User-Role) nằm trong danh sách cho phép.
// Phải đặt sau RequireAuth() trong chuỗi middleware.
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}

	return func(c *gin.Context) {
		role := c.GetString(CtxRole)
		if role == "" || !allowed[role] {
			response.Error(c, http.StatusForbidden, "Bạn không có quyền truy cập tài nguyên này", "Forbidden")
			c.Abort()
			return
		}
		c.Next()
	}
}
