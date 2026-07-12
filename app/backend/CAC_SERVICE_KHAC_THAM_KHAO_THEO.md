# Internal Service-to-Service (M2M) Authentication Guide

> **Dành cho**: Team developers khi tích hợp service mới vào hệ thống MindCare Microservices.

---

## 🏗️ Kiến trúc tổng quan

```
Internet
    │
    ▼
[Kong API Gateway :8000]  ←── Chặn /internal/* với 403
    │
    ├──▶ /api/v1/auth/*        → [auth-service :8080]  (NO host ports — internal only)
    ├──▶ /api/v1/payments/*    → [payment-service :8082]
    ├──▶ /api/v1/booking/*     → [booking-service :8083]
    ├──▶ /api/v1/profiles/*    → [profile-service :8081]
    └──▶ /api/v1/assessments/* → [assessment-service :5000]

Docker Internal Network (app-network):
    [payment-service] ──POST /internal/auth/token──▶ [auth-service]
    [booking-service] ──POST /internal/auth/token──▶ [auth-service]
    [booking-service] ──GET  /internal/payments/...─▶ [payment-service]
                           (kèm Bearer <internal-JWT>)
```

### Mô hình bảo mật "2 lớp"

| Lớp | Ai thực hiện | Chức năng |
|-----|-------------|-----------|
| **North-South** (ngoài vào) | Kong Gateway | Block `/internal/*`, enforce JWT user token |
| **East-West** (nội bộ) | auth-service | Cấp M2M JWT 15 phút bằng client credentials |

---

## 📦 Cấu trúc folder chuẩn (DDD — theo payment-service)

```
your-service/
├── cmd/
│   └── main.go                    # Entrypoint: LoadConfig → InitPublicKey → server start
├── internal/
│   ├── config/
│   │   └── config.go              # Struct Config + LoadConfig (đọc ENV)
│   ├── <domain>/                  # VD: appointment/, invoice/, profile/
│   │   ├── repository.go          # Interface + GORM implementation
│   │   ├── usecase.go             # Business logic
│   │   ├── handler/
│   │   │   └── <domain>_handler.go
│   │   └── gateway/               # (nếu cần gọi API bên ngoài như VNPay)
│   │       └── <external>_client.go
│   └── domain/                    # (optional) Shared domain types
│       └── errors.go
├── pkg/
│   ├── database/
│   │   └── postgres.go            # ConnectDB()
│   ├── internal_auth/             # ⭐ Copy nguyên folder này từ payment-service
│   │   ├── token_manager.go       # Client: xin token để GỌI service khác
│   │   ├── public_key_cache.go    # Cache RSA Public Key từ auth-service
│   │   └── middleware.go          # Server: verify inbound internal token
│   └── response/
│       └── response.go
├── routes/
│   └── routes.go                  # Khai báo tất cả routes
├── docs/                          # Auto-generated Swagger
├── .env                           # Local dev config (KHÔNG commit production secrets)
├── Dockerfile
├── go.mod
└── go.sum
```

---

## 🚀 Checklist tích hợp service mới (7 bước)

### Bước 1: Copy `pkg/internal_auth/` từ payment-service

```bash
# Chạy từ root project
Copy-Item -Recurse app\backend\payment-service\pkg\internal_auth app\backend\your-service\pkg\internal_auth
```

> Folder này chứa đủ 3 files: `token_manager.go`, `public_key_cache.go`, `middleware.go`.
> **Không cần sửa gì** — đây là library dùng chung.

---

### Bước 2: Thêm `internal_auth` fields vào `internal/config/config.go`

```go
type Config struct {
    // ... các field DB, Server port như cũ ...

    // ─── Internal M2M Auth ───────────────────────────────────────────────────
    AuthServiceInternalURL string // "http://auth-service:8080"
    InternalClientID       string // "your-service-name"
    InternalClientSecret   string // plain text (chỉ trong ENV, KHÔNG commit)
}

func LoadConfig() {
    // ... load như cũ ...
    AppConfig = &Config{
        // ... các field cũ ...

        // Internal M2M Auth
        AuthServiceInternalURL: getEnvOrDefault("AUTH_SERVICE_INTERNAL_URL", "http://auth-service:8080"),
        InternalClientID:       getEnvOrDefault("INTERNAL_CLIENT_ID", "your-service-name"),
        InternalClientSecret:   getEnvOrDefault("INTERNAL_CLIENT_SECRET", ""),
    }
}
```

---

### Bước 3: Khởi tạo trong `cmd/main.go`

```go
import (
    "your-service/pkg/internal_auth"
    "your-service/internal/config"
    // ...
)

func main() {
    config.LoadConfig()

    // 1. Cache RSA Public Key từ Auth Service vào RAM
    //    Dùng để: verify inbound internal JWT + verify user JWT
    internal_auth.InitPublicKey(config.AppConfig.AuthServiceInternalURL)

    // 2. Khởi tạo TokenManager để GỌI internal APIs của service khác
    internalTokenManager := internal_auth.NewTokenManager(
        config.AppConfig.AuthServiceInternalURL,
        config.AppConfig.InternalClientID,
        config.AppConfig.InternalClientSecret,
    )

    // 3. Kết nối DB, khởi tạo layers...
    // (inject internalTokenManager vào usecase/gateway nếu cần gọi service khác)
}
```

---

### Bước 4: Đăng ký credentials với Auth Service

Báo cho người maintain `auth-service` thêm service vào `application.yaml`:

```yaml
internal-clients:
  your-service: ${INTERNAL_SECRET_YOUR_SERVICE:$2a$10$<bcrypt-hash-cost10>}
```

**Bạn sẽ nhận về:**
- `client_id`: tên service (vd: `your-service`)
- `client_secret`: plain text (vd: `your_service_internal_secret_2024`)

**Tạo bcrypt hash:**
```bash
# Tạo hash bằng Go (recommend)
go run app/backend/gen_bcrypt.go

# Hoặc online: https://bcrypt-generator.com (cost=10)
```

---

### Bước 5: Thêm ENV vars

**`.env` (ở root project):**
```env
# BCrypt hash cho auth-service
INTERNAL_SECRET_YOUR_SERVICE=$2a$10$<hash>

# Plain text cho service container của bạn
INTERNAL_PLAIN_SECRET_YOUR_SERVICE=your_service_internal_secret_2024
```

**`docker-compose.dev.yml`** — thêm vào service của bạn:
```yaml
your-service:
  expose:
    - "808X"  # Chỉ expose internal, KHÔNG dùng ports: - "808X:808X"
  environment:
    - AUTH_SERVICE_INTERNAL_URL=http://auth-service:8080
    - INTERNAL_CLIENT_ID=your-service
    - INTERNAL_CLIENT_SECRET=${INTERNAL_PLAIN_SECRET_YOUR_SERVICE:-your_service_internal_secret_2024}
```

---

### Bước 6: Bảo vệ route nhận inbound internal call

```go
// routes/routes.go

func SetupRoutes(r *gin.Engine, h *handler.Handler) {
    // Public routes (qua Kong, verify user JWT)
    api := r.Group("/api/v1/your-domain")
    api.GET("/list", h.List)

    // ⭐ Internal routes — chỉ service khác trong Docker network gọi được
    // Kong block /internal/* từ Internet → an toàn
    internal := r.Group("/internal", internal_auth.Middleware())
    {
        internal.GET("/your-domain/:id", h.GetByIDInternal)
        internal.POST("/your-domain/batch", h.BatchQueryInternal)
    }
}
```

**Trong handler:**
```go
func (h *Handler) GetByIDInternal(c *gin.Context) {
    callerID, _ := internal_auth.GetCallerID(c) // "booking-service"
    log.Printf("Internal request from: %s", callerID)
    // Xử lý bình thường...
}
```

---

### Bước 7: Gọi internal API của service khác (Outbound)

```go
// Trong usecase hoặc gateway client

type PaymentGateway struct {
    tokenManager *internal_auth.TokenManager
    httpClient   *http.Client
}

func (g *PaymentGateway) GetWalletBalance(ctx context.Context, userID string) (float64, error) {
    // 1. Lấy token (auto-refresh nếu sắp hết hạn — 0.01ms nếu còn hạn)
    token, err := g.tokenManager.GetToken(ctx)
    if err != nil {
        return 0, fmt.Errorf("get internal token: %w", err)
    }

    // 2. Gọi internal API
    req, _ := http.NewRequestWithContext(ctx, "GET",
        "http://payment-service:8082/internal/wallets/"+userID, nil)
    req.Header.Set("Authorization", "Bearer "+token)

    resp, err := g.httpClient.Do(req)
    if err != nil {
        return 0, err
    }

    // 3. Nếu 401 → token hết hạn đột xuất → invalidate và retry 1 lần
    if resp.StatusCode == 401 {
        g.tokenManager.InvalidateToken()
        return g.GetWalletBalance(ctx, userID) // retry once
    }

    // 4. Parse response...
}
```

---

## 🔐 Credentials hiện tại (dev environment)

| Service | Client ID | Plain Secret | Bcrypt Hash (auth-service) |
|---------|-----------|--------------|---------------------------|
| Booking | `booking-service` | `booking_internal_secret_2024` | `$2a$10$LGepNLvn/DTB9Yepl5SVlO...` |
| Payment | `payment-service` | `payment_internal_secret_2024` | `$2a$10$RMrOxu3QFcg.vJNCykUwMe...` |
| Profile | `profile-service` | `profile_internal_secret_2024` | `$2a$10$mVZu8ru34TDvKpqwZdFm5e...` |
| Assessment | `assessment-service` | `assessment_internal_secret_2024` | `$2a$10$El5khvnrOleBYhoSpzy.Ne...` |

> **Hash đầy đủ xem trong `.env` tại root project.**

---

## ⏱️ Token TTL & Refresh Policy

| Tham số | Giá trị |
|---------|---------|
| Token TTL | 15 phút (900s) |
| Refresh trigger | Khi còn < 1 phút |
| Public key retry | 10 lần × 3 giây khi startup |

---

## 🧪 Test commands

```bash
# 1. Test M2M token (từ trong container)
docker exec -it booking-service sh -c '
  curl -s -X POST http://auth-service:8080/internal/auth/token \
    -H "Content-Type: application/json" \
    -d "{\"clientId\":\"booking-service\",\"clientSecret\":\"booking_internal_secret_2024\"}" | jq .
'

# 2. Xác nhận Kong block từ Internet
curl -X POST http://localhost:8000/internal/auth/token
# → 403 Forbidden

# 3. Test gọi internal API payment từ booking
docker exec -it booking-service sh -c '
  TOKEN=$(curl -s -X POST http://auth-service:8080/internal/auth/token \
    -H "Content-Type: application/json" \
    -d "{\"clientId\":\"booking-service\",\"clientSecret\":\"booking_internal_secret_2024\"}" \
    | jq -r .access_token)
  curl -s http://payment-service:8082/internal/wallets/some-user-id \
    -H "Authorization: Bearer $TOKEN" | jq .
'
```

---

## ❓ FAQ

**Q: Service của tôi cần thêm Go dependency gì không?**
> Chỉ `github.com/golang-jwt/jwt/v5` — thường đã có sẵn.

**Q: Không có Redis, token cache ở đâu?**
> RAM của từng service. Restart service → fetch lại tự động lần đầu gọi.

**Q: Muốn rotate secret thì sao?**
> 1. Đổi plain secret trong `.env` → 2. Tạo bcrypt hash mới → 3. Update `INTERNAL_SECRET_*` → 4. Restart auth-service. Tối đa 15 phút sau, mọi token cũ tự vô hiệu.

**Q: Service Python (assessment) thì dùng gì?**
> Dùng `PyJWT` + `cryptography`. Lấy public key từ `GET /api/v1/auth/public-key`, decode base64, verify với `algorithms=["RS256"]`, kiểm tra `payload["role"] == "internal"`.
