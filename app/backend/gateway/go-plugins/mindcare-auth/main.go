package main

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Kong/go-pdk"
	"github.com/Kong/go-pdk/server"
	"github.com/golang-jwt/jwt/v5"
)

type Config struct {
	AuthServiceUrl string `json:"auth_service_url"`
}

var (
	publicKeyCache *rsa.PublicKey
	pubKeyMutex    sync.RWMutex
)

func New() interface{} {
	return &Config{}
}

// Hàm lấy Public Key từ Auth Service (có cache)
func getPublicKey(authUrl string, kong *pdk.PDK) (*rsa.PublicKey, error) {
	pubKeyMutex.RLock()
	if publicKeyCache != nil {
		pubKeyMutex.RUnlock()
		return publicKeyCache, nil
	}
	pubKeyMutex.RUnlock()

	pubKeyMutex.Lock()
	defer pubKeyMutex.Unlock()

	// Double check
	if publicKeyCache != nil {
		return publicKeyCache, nil
	}

	url := authUrl
	if url == "" || strings.HasPrefix(url, "$") {
		if envUrl := os.Getenv("AUTH_SERVICE_URL"); envUrl != "" {
			url = envUrl
		} else {
			url = "http://auth-service:8080/api/v1/auth/public-key"
		}
	}

	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		kong.Log.Err(fmt.Sprintf("Failed to fetch public key from %s: %v", url, err))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("auth service returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data struct {
		PublicKey string `json:"publicKey"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	pubKeyBytes, err := base64.StdEncoding.DecodeString(data.PublicKey)
	if err != nil {
		return nil, err
	}

	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(pubKeyBytes)
	if err != nil {
		// Fallback thử load raw X509 (nếu spring boot không xuất PEM chuẩn)
		pubKey, err = jwt.ParseRSAPublicKeyFromPEM([]byte("-----BEGIN PUBLIC KEY-----\n" + data.PublicKey + "\n-----END PUBLIC KEY-----"))
		if err != nil {
			return nil, err
		}
	}

	publicKeyCache = pubKey
	kong.Log.Notice("✅ Successfully fetched and cached RSA Public Key from Auth Service")
	return publicKeyCache, nil
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


// Access chỉ giải mã & gắn header định danh KHI request có kèm Authorization.
// Không có Authorization -> cho qua ẩn danh (route công khai như GET /experts tự quyết
// định không cần định danh); có Authorization nhưng sai định dạng/hết hạn/sai chữ ký ->
// chặn 401. Các route bắt buộc đăng nhập tự kiểm tra sự tồn tại của X-User-Id ở tầng
// service (xem profile-service/internal/middleware.RequireAuth).
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

	// 3. Fetch Public Key
	pubKey, err := getPublicKey(conf.AuthServiceUrl, kong)
	if err != nil {
		kong.Log.Err("Could not get public key: " + err.Error())
		exitWithError(kong, 500, "Internal Server Error", "Could not fetch public key for validation")
		return
	}

	// 4. Verify JWT Token using RS256
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		// Ensure signing method is RSA
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			kong.Log.Err(fmt.Sprintf("Invalid signing method: %v", t.Header["alg"]))
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return pubKey, nil
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
