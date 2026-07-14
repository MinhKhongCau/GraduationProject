# 🔄 HƯỚNG DẪN GỌI INTERNAL API (SERVICE-TO-SERVICE)

Trong kiến trúc Microservices của MindCare, các service cần giao tiếp trực tiếp với nhau trong mạng nội bộ Docker — ví dụ `booking-service` gọi `payment-service` để kiểm tra số dư ví của user.

Để bảo mật, **không được gọi thẳng URL của nhau mà không xác thực**. Hệ thống dùng **Internal JWT Token (M2M Auth)** — service phải xin token từ `auth-service` trước, rồi mới được phép gọi sang service khác.

---

## ⚙️ Cơ chế hoạt động (M2M Flow)

```
Service A (Caller)                auth-service                  Service B (Receiver)
      │                                │                                │
      │── POST /internal/auth/token ──>│                                │
      │   {clientId, clientSecret}     │                                │
      │<── {access_token} ────────────│                                │
      │                                │                                │
      │── GET /internal/... ──────────────────────────────────────────>│
      │   Authorization: Bearer <token>                                 │
      │                                                                  │
      │                          [Service B verify token với RSA key]   │
      │<── Response ─────────────────────────────────────────────────── │
```

**3 bước cốt lõi:**

1. **Service A** gửi `client_id` + `client_secret` lên `auth-service` để xin Internal Token.
2. **Service A** đính token vào header `Authorization: Bearer <token>` khi gọi sang **Service B**.
3. **Service B** xác thực token bằng **RSA Public Key** (lấy từ `auth-service`). Nếu hợp lệ → xử lý request.

---

## 🔐 Thông tin Credentials (môi trường Dev)

| Service              | Client ID            | Plain Secret (trong `.env`)       | Bcrypt Hash (trong `auth-service`) |
| -------------------- | -------------------- | --------------------------------- | ---------------------------------- |
| `booking-service`    | `booking-service`    | `booking_internal_secret_2024`    | `INTERNAL_SECRET_BOOKING`          |
| `payment-service`    | `payment-service`    | `payment_internal_secret_2024`    | `INTERNAL_SECRET_PAYMENT`          |
| `profile-service`    | `profile-service`    | `profile_internal_secret_2024`    | `INTERNAL_SECRET_PROFILE`          |
| `assessment-service` | `assessment-service` | `assessment_internal_secret_2024` | `INTERNAL_SECRET_ASSESSMENT`       |

> Hash đầy đủ xem trong file `.env` ở root project.

---

## 🛠️ Phần 1: Cấu hình Service đi gọi (CALLER)

### Bước 1 — Khai báo biến môi trường

Trong `docker-compose.dev.huy.yml`, đảm bảo service của bạn có 3 biến này:

```yaml
your-service:
  environment:
    - AUTH_SERVICE_INTERNAL_URL=http://auth-service:8080
    - INTERNAL_CLIENT_ID=your-service
    - INTERNAL_CLIENT_SECRET=${INTERNAL_PLAIN_SECRET_YOUR_SERVICE:-your_secret_here}
```

### Bước 2 — Khởi tạo TokenManager lúc boot app

```go
// cmd/main.go

import "payment-service/pkg/internal_auth"

// Khởi tạo một lần duy nhất lúc boot — tự động cache và refresh token khi hết hạn
tokenManager := internal_auth.NewTokenManager(
    config.AppConfig.AuthServiceInternalURL, // "http://auth-service:8080"
    config.AppConfig.InternalClientID,       // "booking-service"
    config.AppConfig.InternalClientSecret,   // "booking_internal_secret_2024"
)
```

### Bước 3 — Gọi Internal API của service khác

```go
func (g *PaymentGateway) GetWalletBalance(ctx context.Context, userID string) (int64, error) {
    // 1. Lấy token (tự động refresh nếu sắp hết hạn, ~0ms nếu còn hạn)
    token, err := g.tokenManager.GetToken(ctx)
    if err != nil {
        return 0, fmt.Errorf("failed to get internal token: %w", err)
    }

    // 2. Tạo request với token
    url := "http://payment-service:8082/internal/wallets/" + userID
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return 0, err
    }
    req.Header.Set("Authorization", "Bearer "+token)

    // 3. Thực thi request
    resp, err := g.httpClient.Do(req)
    if err != nil {
        return 0, err
    }
    defer resp.Body.Close()

    // 4. Nếu 401 → token hết hạn đột xuất → invalidate và retry 1 lần
    if resp.StatusCode == http.StatusUnauthorized {
        g.tokenManager.InvalidateToken()
        return g.GetWalletBalance(ctx, userID) // retry once
    }

    // 5. Parse response...
    return balance, nil
}
```

---

## 🛡️ Phần 2: Cấu hình Service bị gọi (RECEIVER)

> Nếu service của bạn có API **chỉ dành cho các service khác gọi** (không phải Frontend), bạn phải đặt nó vào route group `/internal` và bảo vệ bằng middleware.

### Bước 1 — Đăng ký route `/internal` có middleware xác thực

```go
// routes/routes.go

import "your-service/pkg/internal_auth"

func SetupRoutes(r *gin.Engine, h *handler.Handler) {
    // ✅ Các API public — đi qua Kong, xác thực user JWT
    api := r.Group("/api/v1/booking")
    {
        api.POST("/create", h.Create)
        api.GET("/list", h.List)
    }

    // 🔒 Các API internal — CHỈ service khác trong Docker network gọi được
    // Kong đã block toàn bộ /internal/* từ Internet → an toàn tuyệt đối
    internal := r.Group("/internal", internal_auth.Middleware())
    {
        internal.GET("/bookings/:id", h.GetBookingInternal)
        internal.POST("/bookings/batch", h.GetBookingsBatchInternal)
    }
}
```

### Bước 2 — Xử lý logic trong Handler

```go
// handler/booking_handler.go

func (h *Handler) GetBookingInternal(c *gin.Context) {
    // Lấy thông tin service đang gọi mình (audit log, rate limiting...)
    callerID, _ := internal_auth.GetCallerID(c) // "payment-service"
    log.Printf("Internal request from service: %s", callerID)

    bookingID := c.Param("id")
    // ... xử lý logic bình thường
    c.JSON(http.StatusOK, gin.H{"booking_id": bookingID, "status": "confirmed"})
}
```

---

## 🚫 Bảo mật: Tại sao `/internal` an toàn?

Kong API Gateway đã cấu hình **block hoàn toàn** mọi request từ Internet vào path `/internal`:

```yaml
# kong.yml
- name: block-internal-routes
  url: http://auth-service:8080
  routes:
    - name: block-internal
      paths:
        - /internal
      plugins:
        - name: request-termination
          config:
            status_code: 403
            body: '{"message": "Forbidden: /internal paths are not accessible from outside"}'
```

Kết quả:

- **Từ Internet / Postman:** Gọi `http://localhost:8000/internal/...` → **403 Forbidden** ngay lập tức.
- **Từ container khác trong Docker network:** Gọi `http://booking-service:8083/internal/...` trực tiếp (bypass Kong) → ✅ Được phép nếu có token hợp lệ.

---

## 🧪 Test nhanh Internal Call trong Docker

Chạy lệnh sau từ terminal để test gọi internal API từ bên trong container:

```bash
# Test lấy internal token (từ booking-service sang auth-service)
docker exec -it booking-service sh -c '
  TOKEN=$(wget -qO- -post-data "{\"clientId\":\"booking-service\",\"clientSecret\":\"booking_internal_secret_2024\"}" \
    --header="Content-Type: application/json" \
    http://auth-service:8080/internal/auth/token | grep -o "\"access_token\":\"[^\"]*\"" | cut -d\" -f4)
  echo "Token: $TOKEN"
'

# Test gọi internal API payment từ booking
docker exec -it booking-service sh -c '
  TOKEN=$(...)  # lấy token như trên
  wget -qO- --header="Authorization: Bearer $TOKEN" \
    http://payment-service:8082/internal/wallets/<user-uuid>
'
```

---

## ❓ FAQ

**Q: Token được cache ở đâu?**

> Trong **RAM** của từng service. Restart service → tự động fetch lại lần đầu tiên khi cần.

**Q: Token TTL là bao nhiêu?**

> **15 phút (900 giây)**. `TokenManager` tự động xin token mới khi còn dưới 60 giây.

**Q: Muốn thêm secret cho service mới thì làm gì?**

> 1. Thêm plain secret vào `.env`: `INTERNAL_PLAIN_SECRET_YOUR_SERVICE=your_secret`
> 2. Tạo bcrypt hash từ plain secret đó (cost 10).
> 3. Thêm hash vào `.env`: `INTERNAL_SECRET_YOUR_SERVICE=$2a$10$...hash...`
> 4. Truyền cả 2 vào docker-compose → Restart `auth-service` → Xong.

**Q: Service Python (assessment-service) thì xử lý token bằng gì?**

> Dùng `PyJWT` + `cryptography`. Lấy RSA Public Key từ `GET /api/v1/auth/public-key`, decode base64, verify với `algorithms=["RS256"]`, kiểm tra `payload["role"] == "internal"`.
