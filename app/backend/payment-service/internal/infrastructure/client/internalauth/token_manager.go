// Package internalauth cung cấp cơ chế xác thực service-to-service (M2M)
// dựa trên JWT được ký bởi Auth Service (RS256).
//
// Cách dùng:
//
//	// Khởi tạo 1 lần khi start service
//	manager := internalauth.NewTokenManager(
//	    os.Getenv("AUTH_SERVICE_INTERNAL_URL"),
//	    os.Getenv("INTERNAL_CLIENT_ID"),
//	    os.Getenv("INTERNAL_CLIENT_SECRET"),
//	)
//
//	// Lấy token khi cần gọi internal API (thread-safe, auto-refresh)
//	token, err := manager.GetToken(ctx)
//
//	// Gắn vào HTTP request
//	req.Header.Set("Authorization", "Bearer " + token)
package internalauth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// TokenManager quản lý JWT nội bộ trong RAM.
// Thread-safe. Tự động xin token mới khi sắp hết hạn (còn < 1 phút).
type TokenManager struct {
	authServiceURL string // http://auth-service:8080
	clientID       string // "booking-service"
	clientSecret   string // "booking_internal_secret_2024"

	mu          sync.RWMutex
	token       string
	tokenExpiry time.Time

	httpClient *http.Client
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"` // giây
	TokenType   string `json:"token_type"`
}

// NewTokenManager tạo một TokenManager mới.
// Không gọi Auth Service ngay — token sẽ được lấy lazy (lần đầu gọi GetToken).
func NewTokenManager(authServiceURL, clientID, clientSecret string) *TokenManager {
	return &TokenManager{
		authServiceURL: authServiceURL,
		clientID:       clientID,
		clientSecret:   clientSecret,
		httpClient:     &http.Client{Timeout: 5 * time.Second},
	}
}

// GetToken trả về token nội bộ hợp lệ. Thread-safe.
//
//   - Nếu token còn hạn > 1 phút: trả về ngay từ RAM (0.01ms).
//   - Nếu token hết hạn hoặc sắp hết (< 1 phút): gọi Auth Service để lấy mới.
//
// ctx được dùng để cancel request nếu service đang shutdown.
func (m *TokenManager) GetToken(ctx context.Context) (string, error) {
	// Fast path: read lock
	m.mu.RLock()
	if m.isTokenValid() {
		token := m.token
		m.mu.RUnlock()
		return token, nil
	}
	m.mu.RUnlock()

	// Slow path: write lock + double-check
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isTokenValid() {
		return m.token, nil
	}

	return m.fetchNewToken(ctx)
}

// InvalidateToken xóa token hiện tại, buộc lần GetToken tiếp theo phải fetch mới.
// Gọi hàm này khi nhận 401 từ internal API.
func (m *TokenManager) InvalidateToken() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.token = ""
	m.tokenExpiry = time.Time{}
	log.Println("🔄 Internal token invalidated, will fetch new on next request")
}

// isTokenValid kiểm tra token có tồn tại và còn hạn > 1 phút không.
// Phải gọi trong khi đang giữ lock.
func (m *TokenManager) isTokenValid() bool {
	return m.token != "" && time.Now().Add(time.Minute).Before(m.tokenExpiry)
}

// fetchNewToken gọi POST /internal/auth/token và cập nhật cache.
// Phải gọi trong khi đang giữ write lock.
func (m *TokenManager) fetchNewToken(ctx context.Context) (string, error) {
	url := m.authServiceURL + "/internal/auth/token"

	body, err := json.Marshal(map[string]string{
		"clientId":     m.clientID,
		"clientSecret": m.clientSecret,
	})
	if err != nil {
		return "", fmt.Errorf("internal_auth: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("internal_auth: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("internal_auth: call auth service at %s: %w", url, err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("internal_auth: auth service returned %d: %s", resp.StatusCode, string(respBody))
	}

	var tokenResp tokenResponse
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return "", fmt.Errorf("internal_auth: parse response: %w", err)
	}

	m.token = tokenResp.AccessToken
	m.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	log.Printf("✅ internal_auth: fetched new token for [%s], expires at %s",
		m.clientID, m.tokenExpiry.Format(time.RFC3339))

	return m.token, nil
}
