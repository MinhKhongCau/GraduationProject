package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Kong/go-pdk"
	"github.com/Kong/go-pdk/server"
	"github.com/golang-jwt/jwt/v5"
)

type Config struct {
	JwtSecret string `json:"jwt_secret"`
}

func New() interface{} {
	return &Config{}
}

func exitWithError(kong *pdk.PDK, status int, friendlyMsg string, errDetail string) {
	resp := map[string]interface{}{
		"success": false,
		"message": friendlyMsg,
		"error":   errDetail,
	}
	respBytes, _ := json.Marshal(resp)
	kong.Response.Exit(status, respBytes, map[string][]string{"Content-Type": {"application/json"}})
}

func (conf *Config) Access(kong *pdk.PDK) {
	// 1. Lấy Authorization Header
	authHeader, err := kong.Request.GetHeader("Authorization")
	if err != nil {
		kong.Log.Err("Error getting Authorization header: " + err.Error())
		exitWithError(kong, 401, "Thiếu header Authorization", err.Error())
		return
	}

	if authHeader == "" {
		exitWithError(kong, 401, "Bạn chưa đăng nhập hoặc thiếu token xác thực", "Unauthorized")
		return
	}

	// 2. Tách Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		exitWithError(kong, 401, "Định dạng token không đúng (Phải là Bearer <token>)", "Invalid Token Format")
		return
	}
	tokenString := parts[1]

	// 3. Verify JWT Token
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		// Đảm bảo thuật toán là HS256
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(conf.JwtSecret), nil
	})

	if err != nil || !token.Valid {
		exitWithError(kong, 401, "Token không hợp lệ hoặc đã hết hạn", err.Error())
		return
	}

	// 4. Lấy Claims và inject vào Header
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		exitWithError(kong, 401, "Không thể đọc dữ liệu phân quyền trong Token", "Invalid Claims Format")
		return
	}

	userId, _ := claims["user_id"].(string)
	if userId == "" {
		// Thử lấy sub nếu user_id không có
		userId, _ = claims["sub"].(string)
	}
	role, _ := claims["role"].(string)

	if userId != "" {
		kong.ServiceRequest.SetHeader("X-User-Id", userId)
	}
	if role != "" {
		kong.ServiceRequest.SetHeader("X-User-Role", role)
	}
}

func main() {
	server.StartServer(New, "0.1", 1000)
}
