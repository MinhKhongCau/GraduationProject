package middleware

import (
	"booking-service/internal/infrastructure/client"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// InternalClaims là payload của JWT nội bộ do Auth Service ký.
type InternalClaims struct {
	ClientID string `json:"client_id"`
	Role     string `json:"role"` // luôn là "internal"
	jwt.RegisteredClaims
}

// ContextKey là các key dùng để lưu vào gin.Context sau khi verify thành công.
const (
	ContextKeyCallerID = "internal_caller_id" // "payment-service"
)

// InternalAuth là Gin middleware dùng để bảo vệ các route nội bộ (/internal/...).
//
// Logic:
//  1. Bóc Authorization: Bearer <token>
//  2. Verify JWT offline bằng RSA Public Key đã cache trong RAM
//  3. Kiểm tra role == "internal"
//  4. Inject caller_id vào gin.Context để handler biết ai đang gọi
//
// Nếu bất kỳ bước nào thất bại → trả 401 và abort request.
//
// Ví dụ đăng ký route:
//
//	internal := r.Group("/internal", middleware.InternalAuth())
//	internal.GET("/appointments/:id", handler.GetAppointmentInternal)
func InternalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := extractBearerToken(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "missing_token",
				"message": err.Error(),
			})
			return
		}

		claims, err := verifyInternalToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "invalid_token",
				"message": err.Error(),
			})
			return
		}

		// Inject caller identity vào context để handler có thể đọc
		c.Set(ContextKeyCallerID, claims.ClientID)
		c.Next()
	}
}

// GetCallerID lấy caller identity đã được inject bởi InternalAuth.
// Trả về ("", false) nếu request không đi qua internal middleware.
func GetCallerID(c *gin.Context) (string, bool) {
	val, exists := c.Get(ContextKeyCallerID)
	if !exists {
		return "", false
	}
	id, ok := val.(string)
	return id, ok
}

// ─── Private helpers ─────────────────────────────────────────────────────────

func extractBearerToken(c *gin.Context) (string, error) {
	header := c.GetHeader("Authorization")
	if header == "" {
		return "", fmt.Errorf("Authorization header is required for internal endpoints")
	}
	if !strings.HasPrefix(strings.ToLower(header), "bearer ") {
		return "", fmt.Errorf("Authorization header must use Bearer scheme")
	}
	token := strings.TrimSpace(header[7:])
	if token == "" {
		return "", fmt.Errorf("Bearer token is empty")
	}
	return token, nil
}

func verifyInternalToken(tokenString string) (*InternalClaims, error) {
	pubKey, err := client.GetPublicKey()
	if err != nil {
		return nil, fmt.Errorf("public key not available: %w", err)
	}

	token, err := jwt.ParseWithClaims(tokenString, &InternalClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return pubKey, nil
	})

	if err != nil {
		if strings.Contains(err.Error(), "token is expired") {
			return nil, fmt.Errorf("internal token has expired — service should refresh before calling")
		}
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	claims, ok := token.Claims.(*InternalClaims)
	if !ok {
		return nil, fmt.Errorf("cannot parse token claims")
	}

	if claims.Role != "internal" {
		return nil, fmt.Errorf("token role must be 'internal', got '%s'", claims.Role)
	}

	return claims, nil
}

// VerifyInternalBearer verify header "Bearer <internal JWT>" (dùng chung cho gRPC interceptor)
// và trả về client_id của service gọi.
func VerifyInternalBearer(header string) (string, error) {
	if !strings.HasPrefix(strings.ToLower(header), "bearer ") {
		return "", fmt.Errorf("authorization must use Bearer scheme")
	}
	token := strings.TrimSpace(header[7:])
	if token == "" {
		return "", fmt.Errorf("Bearer token is empty")
	}
	claims, err := verifyInternalToken(token)
	if err != nil {
		return "", err
	}
	return claims.ClientID, nil
}
