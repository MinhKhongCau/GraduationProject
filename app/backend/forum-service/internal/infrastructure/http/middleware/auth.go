package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"forum-service/internal/infrastructure/http/response"
)

const (
	HeaderUserID   = "X-User-Id"
	HeaderUserRole = "X-User-Role"

	contextUserID = "userId"
	contextRole   = "userRole"
)

// RequireAuth enforces that the gateway-injected X-User-Id header is
// present (SPEC.md §4) — forum-service trusts this header rather than
// validating a JWT itself. X-User-Id is the account's UUID (same as every
// other MindCare service's accountId/expertId/patientId), not a numeric
// id — matches the plain c.GetHeader() convention used by booking-service
// et al. Returns 401 if missing.
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetHeader(HeaderUserID)
		if userID == "" {
			response.Error(c, http.StatusUnauthorized, "Missing X-User-Id header", "unauthorized")
			c.Abort()
			return
		}
		c.Set(contextUserID, userID)
		c.Set(contextRole, c.GetHeader(HeaderUserRole))
		c.Next()
	}
}

// RequireRole enforces that the caller's role (from X-User-Role, set by
// RequireAuth) is one of the given roles. Returns 403 otherwise. Must be
// chained after RequireAuth.
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(c *gin.Context) {
		roleStr, _ := c.Get(contextRole)
		role, _ := roleStr.(string)
		if !allowed[role] {
			response.Error(c, http.StatusForbidden, "You do not have permission to perform this action", "forbidden")
			c.Abort()
			return
		}
		c.Next()
	}
}

// UserID reads the authenticated user id (accountId, a UUID string) set by
// RequireAuth. ok is false if RequireAuth didn't run on this route (i.e. a
// public endpoint).
func UserID(c *gin.Context) (string, bool) {
	v, exists := c.Get(contextUserID)
	if !exists {
		return "", false
	}
	id, ok := v.(string)
	return id, ok
}

// IsAdmin reads whether the authenticated caller's role is ADMIN.
func IsAdmin(c *gin.Context) bool {
	v, _ := c.Get(contextRole)
	role, _ := v.(string)
	return role == "ADMIN"
}
