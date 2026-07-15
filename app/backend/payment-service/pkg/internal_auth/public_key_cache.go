package internal_auth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// PublicKeyCache lưu RSA public key từ Auth Service trong RAM.
// Chỉ fetch 1 lần lúc khởi động, dùng lại mãi cho đến khi restart.
// Thread-safe.
type PublicKeyCache struct {
	mu         sync.RWMutex
	publicKey  *rsa.PublicKey
	httpClient *http.Client
}

var defaultPublicKeyCache = &PublicKeyCache{
	httpClient: &http.Client{Timeout: 5 * time.Second},
}

// InitPublicKey gọi Auth Service để lấy RSA Public Key và cache vào RAM.
// Gọi hàm này 1 lần khi khởi động service (trong main()).
// Nếu không lấy được sau maxRetries lần → Fatal (service không thể hoạt động).
func InitPublicKey(authServiceURL string) {
	defaultPublicKeyCache.init(authServiceURL, 10, 3*time.Second)
}

// GetPublicKey trả về RSA Public Key đã cache. Thread-safe.
func GetPublicKey() (*rsa.PublicKey, error) {
	return defaultPublicKeyCache.get()
}

func (c *PublicKeyCache) init(authServiceURL string, maxRetries int, retryDelay time.Duration) {
	url := authServiceURL + "/api/v1/auth/public-key"
	log.Printf("🔑 internal_auth: fetching RSA public key from %s", url)

	for i := 0; i < maxRetries; i++ {
		if err := c.fetch(url); err == nil {
			log.Println("✅ internal_auth: RSA public key cached successfully")
			return
		} else {
			log.Printf("⚠️  internal_auth: retry %d/%d — %v", i+1, maxRetries, err)
			time.Sleep(retryDelay)
		}
	}
	log.Fatalf("❌ internal_auth: cannot fetch RSA public key after %d retries. Shutting down.", maxRetries)
}

func (c *PublicKeyCache) fetch(url string) error {
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("GET %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth service returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}

	var data struct {
		PublicKey string `json:"publicKey"` // Base64-encoded DER
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	keyBytes, err := base64.StdEncoding.DecodeString(data.PublicKey)
	if err != nil {
		return fmt.Errorf("base64 decode: %w", err)
	}

	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(keyBytes)
	if err != nil {
		// Thử wrap thành PEM nếu Auth Service trả về raw DER bytes
		pemWrapped := "-----BEGIN PUBLIC KEY-----\n" + data.PublicKey + "\n-----END PUBLIC KEY-----"
		pubKey, err = jwt.ParseRSAPublicKeyFromPEM([]byte(pemWrapped))
		if err != nil {
			return fmt.Errorf("parse RSA key: %w", err)
		}
	}

	c.mu.Lock()
	c.publicKey = pubKey
	c.mu.Unlock()
	return nil
}

func (c *PublicKeyCache) get() (*rsa.PublicKey, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.publicKey == nil {
		return nil, fmt.Errorf("internal_auth: public key not initialized")
	}
	return c.publicKey, nil
}
