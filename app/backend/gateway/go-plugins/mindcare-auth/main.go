package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
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
	// 1. Get Authorization Header
	authHeader, err := kong.Request.GetHeader("Authorization")
	if err != nil {
		kong.Log.Err("Error getting Authorization header: " + err.Error())
		exitWithError(kong, 401, "Missing Authorization header", err.Error())
		return
	}

	if authHeader == "" {
		exitWithError(kong, 401, "You are not logged in or authentication token is missing", "Unauthorized")
		return
	}

	// 2. Extract Bearer token (support both "Bearer <token>" and raw "<token>")
	tokenString := authHeader
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		tokenString = authHeader[7:]
	}
	tokenString = strings.TrimSpace(tokenString)

	// Resolve the JWT Secret Key (support environment variable fallback if config is a placeholder)
	jwtSecret := conf.JwtSecret
	if jwtSecret == "" || strings.HasPrefix(jwtSecret, "$") {
		if envSecret := os.Getenv("JWT_SECRET_KEY"); envSecret != "" {
			jwtSecret = envSecret
		}
	}

	// Decode secret from Base64 to match Java auth-service's sign key logic
	kong.Log.Err(fmt.Sprintf("Resolved JWT Secret: %s", jwtSecret))
	secretBytes, err := base64.StdEncoding.DecodeString(jwtSecret)
	if err != nil {
		// Fallback to raw bytes if it is not valid base64
		kong.Log.Err(fmt.Sprintf("Base64 decode failed, using raw secret. Error: %v", err))
		secretBytes = []byte(jwtSecret)
	}

	// 3. Verify JWT Token
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		// Ensure signing method is HMAC
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			kong.Log.Err(fmt.Sprintf("Invalid signing method: %v", t.Header["alg"]))
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secretBytes, nil
	})

	if err != nil {
		kong.Log.Err("Token validation failed: " + err.Error())
		if strings.Contains(err.Error(), "token is expired") {
			exitWithError(kong, 401, "Token has expired, please login again", err.Error())
		} else if strings.Contains(err.Error(), "signature is invalid") {
			exitWithError(kong, 401, "Token signature is invalid or mismatched secret key", err.Error())
		} else {
			exitWithError(kong, 401, "Invalid token", err.Error())
		}
		return
	}

	if !token.Valid {
		kong.Log.Err("Token is invalid but no error returned from parser")
		exitWithError(kong, 401, "Invalid token", "Token is not valid")
		return
	}

	// 4. Extract Claims and inject into headers
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		kong.Log.Err("Failed to cast token claims to jwt.MapClaims")
		exitWithError(kong, 401, "Failed to read authorization claims from token", "Invalid Claims Format")
		return
	}

	userId, ok := claims["accountId"].(string)
	if !ok || userId == "" {
		kong.Log.Err("Token does not contain accountId claim")
		exitWithError(kong, 403, "Access denied: missing user ID in token", "Missing accountId")
		return
	}

	userRole, ok := claims["role"].(string)
	if !ok || userRole == "" {
		kong.Log.Err("Token does not contain role claim")
		exitWithError(kong, 403, "Access denied: missing role in token", "Missing role")
		return
	}

	if userId != "" {
		kong.ServiceRequest.SetHeader("X-User-Id", userId)
	}
	if userRole != "" {
		kong.ServiceRequest.SetHeader("X-User-Role", userRole)
	}
}

func main() {
	server.StartServer(New, "0.1", 1000)
}
