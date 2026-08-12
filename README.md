# 🌿 Psychological Counseling System (MindCare) - Microservices Architecture

Hệ thống cung cấp nền tảng tư vấn tâm lý trực tuyến (**MindCare**), giúp kết nối **Patient (Bệnh nhân/Khách hàng)** với các **Chuyên gia (Expert)** thông qua quy trình đặt lịch, thanh toán ví nội bộ/VNPay, làm bài trắc nghiệm tâm lý có đánh giá hỗ trợ bởi AI, nhắn tin và gọi thoại realtime, cùng diễn đàn thảo luận cộng đồng.

Hệ thống được thiết kế theo kiến trúc **Microservices** đa ngôn ngữ (Java, Golang, Python, Node.js) đứng sau **Kong API Gateway** nhằm đảm bảo tính sẵn sàng cao, bảo mật chặt chẽ và khả năng mở rộng linh hoạt.

---

## 🔗 Liên kết Dịch vụ & Tài liệu API (Service Links & API Docs)

Toàn bộ các yêu cầu từ Client (hoặc Postman) đều đi qua **API Gateway (Kong)** ở cổng `8000`, định tuyến đến các service backend tương ứng qua prefix `/api/v1/*`.

### 🌐 Môi trường Production (Sandbox Demo)
- **Frontend App**: [https://sandbox.qmcloud.io.vn](https://sandbox.qmcloud.io.vn)
- **RabbitMQ Management**: [https://rabbitmq.qmcloud.io.vn](https://rabbitmq.qmcloud.io.vn)
- **System Observer (Dozzle)**: [https://observer.qmcloud.io.vn](https://observer.qmcloud.io.vn)

### 📄 Swagger / OpenAPI UI (Bản chạy Local - Cổng `8000`)
- **Auth Service**: [http://localhost:8000/auth/swagger-ui/index.html](http://localhost:8000/auth/swagger-ui/index.html)
- **Profile Service**: [http://localhost:8000/profile/swagger/index.html](http://localhost:8000/profile/swagger/index.html)
- **Booking Service**: [http://localhost:8000/booking/swagger/index.html](http://localhost:8000/booking/swagger/index.html)
- **Payment Service**: [http://localhost:8000/payment/swagger/index.html](http://localhost:8000/payment/swagger/index.html)
- **Assessment Service**: [http://localhost:8000/assessment/swagger-ui](http://localhost:8000/assessment/swagger-ui)
- **Forum Service**: [http://localhost:8000/forum/swagger/index.html](http://localhost:8000/forum/swagger/index.html)
- **Chatroom Service**: [http://localhost:8000/chatroom/docs](http://localhost:8000/chatroom/docs)

---

## 🏗️ Kiến trúc Hệ thống (System Architecture)

```mermaid
graph TD
    Client[Client: Next.js / WebRTC / Socket.IO] -->|Request: Cổng 8000| Kong[Kong API Gateway]
    
    subgraph Gateway Layer
        Kong -->|Go Custom Plugin| AuthVerify[mindcare-auth: Verify JWT]
    end

    subgraph Internal Network (Docker Bridge)
        Kong -->|/api/v1/auth| Auth[Auth Service: Spring Boot]
        Kong -->|/api/v1/profiles| Profile[Profile Service: Go Gin]
        Kong -->|/api/v1/booking| Booking[Booking Service: Go Gin]
        Kong -->|/api/v1/payments| Payment[Payment Service: Go Gin]
        Kong -->|/api/v1/assessments| Assessment[Assessment Service: FastAPI]
        Kong -->|/api/v1/forum| Forum[Forum Service: Go Gin]
        Kong -->|/api/v1/chatroom| Chatroom[Chatroom Service: Node.js]
        Kong -->|/api/v1/chat| Chatbot[Chatbot Service: FastAPI]
    end

    subgraph Infrastructure
        Payment -->|Atomicity: Outbox Pattern| PG[(PostgreSQL 16)]
        Booking & Profile & Forum & Chatroom & Chatbot & Auth & Assessment --> PG
        Chatroom -->|State Management| Redis[(Redis 7)]
        Chatbot -->|Vector Embeddings / LLM| Ollama[Ollama: nomic-embed-text]
        
        %% RabbitMQ Messaging %%
        Profile -.->|Publish: profile.sync_seed_authors| Rabbit[RabbitMQ Broker]
        Rabbit -.->|Subscribe| Forum
    end
```

---

## 🛠️ Danh sách Microservices & Công nghệ

### 1. **Frontend App** (`app/frontend`)
- **Công nghệ**: Next.js 16 (App Router) + React 19 + TypeScript + Tailwind CSS v4 + TanStack Query.
- **Tính năng**: Tích hợp đầy đủ luồng nghiệp vụ đặt lịch (Booking), thanh toán nội bộ & VNPay, làm bài test tâm lý, quản lý ví, chat thời gian thực và gọi thoại WebRTC.

### 2. **API Gateway (Kong)** (`app/backend/gateway`)
- **Công nghệ**: Kong Gateway (Declarative Config) + Custom Go Plugin (`mindcare-auth`).
- **Nhiệm vụ**: Xác thực JWT tập trung, phân quyền cơ bản, chặn tuyệt đối các endpoint nội bộ (`/internal/*`) từ bên ngoài internet, cấu hình CORS và Rate Limiting.

### 3. **Auth Service** (`app/backend/auth-service`)
- **Công nghệ**: Java 21 + Spring Boot 3 + Maven.
- **Nhiệm vụ**: Quản lý tài khoản, đăng ký/đăng nhập, phân quyền Role-Based Access Control (RBAC: `ADMIN`, `EXPERT`, `PATIENT`), phát hành token JWT (ký bằng cặp khóa RSA), cung cấp API cấp mã Internal Token (M2M Auth) cho các service giao tiếp nội bộ.

### 4. **Profile Service** (`app/backend/profile-service`)
- **Công nghệ**: Go (Gin + GORM).
- **Nhiệm vụ**: Quản lý thông tin chi tiết hồ sơ bệnh nhân, chuyên gia, quản trị viên, lịch sử bệnh án và chuyên môn của Expert.

### 5. **Booking Service** (`app/backend/booking-service`)
- **Công nghệ**: Go (Gin + GORM).
- **Nhiệm vụ**: Quản lý lịch trình rảnh của Expert (Availabilities), tự động sinh lịch hẹn (Slots), cơ chế khóa slot giữ chỗ (Slot locking) trong 15 phút, quản lý ca hẹn (Appointments), lịch nghỉ phép (Time-off).

### 6. **Payment Service** (`app/backend/payment-service`)
- **Công nghệ**: Go (Gin + GORM).
- **Nhiệm vụ**: Quản lý ví điện tử nội bộ, nạp/rút tiền, lưu vết giao dịch (Ledger), tạo đơn thanh toán (Payment Order) tích hợp cổng VNPay Sandbox, đối soát giao dịch (VNPay IPN webhook), cơ chế Outbox Pattern tin cậy để đồng bộ trạng thái cuộc hẹn.

### 7. **Assessment Service** (`app/backend/assessment-service`)
- **Công nghệ**: Python 3.10+ (FastAPI + SQLAlchemy + Alembic).
- **Nhiệm vụ**: Quản lý ngân hàng câu hỏi trắc nghiệm tâm lý (PHQ-9, ASRS, MDQ...), thu thập câu trả lời từ bệnh nhân, chấm điểm tự động và gọi API Google Gemini để sinh nhận xét/lời khuyên AI cá nhân hóa.

### 8. **Forum Service** (`app/backend/forum-service`)
- **Công nghệ**: Go (Gin + GORM).
- **Nhiệm vụ**: Diễn đàn hỏi đáp cộng đồng tâm lý, tạo/quản lý bài viết, chuyên mục, bình luận dưới dạng comment tree, bookmark và yêu thích bài viết.

### 9. **Chatroom Service** (`app/backend/chatroom-service`)
- **Công nghệ**: Node.js + Socket.IO + Redis + Postgres.
- **Nhiệm vụ**:
  - Nhắn tin thời gian thực (1-1 DM và phòng chat chung).
  - Gửi tin nhắn thoại (Voice messages) định dạng Base64 kết hợp khuếch đại âm thanh (Web Audio API).
  - Gọi thoại trực tiếp peer-to-peer (**1-on-1 WebRTC Calls**) tích hợp UI đổ chuông, chấp nhận/từ chối, bật/tắt mic, và bộ đếm thời gian cuộc gọi.
  - Kết nối với Chatbot Service để đưa trợ lý ảo vào cuộc trò chuyện.

### 10. **Chatbot Service** (`app/backend/chatbot-service`)
- **Công nghệ**: Python (FastAPI + PyTorch + HuggingFace Speech Emotion Recognition) + Ollama.
- **Nhiệm vụ**: Trợ lý AI hỗ trợ tư vấn tâm lý, phân tích cảm xúc qua giọng nói/văn bản, tìm kiếm ngữ nghĩa (Semantic search) dựa trên Vector database PostgreSQL (PGVector) và mô hình `nomic-embed-text` cục bộ qua Ollama.

---

## 🔄 Cơ chế Giao tiếp & Đồng bộ (Communication Patterns)

### 1. REST M2M Authentication (Machine-to-Machine Auth)
Để bảo mật mạng nội bộ, các service khi gọi API trực tiếp của nhau qua Docker Network (bypass Kong Gateway) bắt buộc phải sử dụng **Internal JWT Token**.
- **Luồng hoạt động**:
  1. Service gọi (ví dụ: `booking-service`) gửi credentials (`clientId` + `clientSecret` khai báo trong `.env`) lên `/internal/auth/token` của `auth-service`.
  2. `auth-service` kiểm tra, trả về một token nội bộ có thời hạn 15 phút.
  3. Service gọi đính kèm token này vào header `Authorization: Bearer <token>`.
  4. Service nhận xác thực token thông qua RSA Public Key lấy từ `auth-service`.
- **An toàn tuyệt đối**: Kong Gateway đã chặn toàn bộ truy cập vào path `/internal/*` từ Internet (trả về HTTP `403 Forbidden`).

### 2. Sự kiện bất đồng bộ qua RabbitMQ (Event-Driven)
Hệ thống sử dụng RabbitMQ để xử lý các tác vụ bất đồng bộ hoặc đồng bộ hóa dữ liệu không chặn (Non-blocking):
- Sử dụng **Topic Exchange** (ví dụ: `user.exchange`, `booking.exchange`, `payment.exchange`).
- Quy tắc đặt tên Routing Key: `<entity>.<action>` (ví dụ: `user.created`, `profile.sync_seed_authors`, `payment.success`).
- Tất cả các hàng đợi (Queues) đều là **Durable** và tin nhắn là **Persistent**.
- Điển hình: `profile-service` phát sự kiện `profile.sync_seed_authors` qua RabbitMQ, `forum-service` nhận sự kiện để cập nhật ID chuyên gia đồng bộ với tác giả bài viết.

### 3. Outbox Pattern trong Thanh toán
Để tránh lỗi mất mát trạng thái giữa lúc VNPay gọi Webhook IPN đến `payment-service` và cập nhật lịch hẹn tại `booking-service`:
- Khi giao dịch thành công, `payment-service` thực hiện ghi nhận thông tin thanh toán, cập nhật ví và lưu một thông điệp vào bảng `payment_outbox` trong **cùng một transaction cơ sở dữ liệu** (ACID).
- Một worker chạy ngầm định kỳ quét bảng outbox, tiến hành gọi REST API nội bộ `/internal/appointments/:id/webhook` sang `booking-service` với cơ chế retry tối đa 10 lần (Exponential Backoff), đảm bảo cuộc hẹn chắc chắn được xác nhận hoặc xử lý bồi hoàn (Compensation case) nếu xảy ra xung đột.

---

## 🔒 Cơ chế Phân quyền (Security & RBAC)

Hệ thống chia làm 3 Role chính:
- **`ADMIN`**: Quản lý danh mục chuyên gia, duyệt rút tiền, xem log bồi hoàn hệ thống, cấu hình ca mẫu.
- **`EXPERT`**: Đăng ký ca rảnh có giá, xuất lịch hẹn, xem thu nhập và ví tiền, thực hiện tư vấn và viết bệnh án.
- **`PATIENT`** *(trong code là `PATIENT`, tài liệu nghiệp vụ ghi `CLIENT`)*: Tìm kiếm chuyên gia, đặt lịch, nạp/rút tiền, làm trắc nghiệm tâm lý hỗ trợ bởi AI, chat & gọi điện thoại trực tiếp.

---

## ⚡ Hướng dẫn Khởi chạy dưới Local (Development)

### 📋 Yêu cầu hệ thống (Prerequisites)
- Docker & Docker Compose
- Java 21 & Maven 3.9+
- Go 1.21+
- Python 3.10+ (cùng pip)
- Node.js 20+ & npm

### 🚀 Bước 1: Khởi tạo File Môi trường
Sao chép cấu hình mẫu và chỉnh sửa nếu cần thiết:
```bash
cp .env.example .env
```
*(Mặc định `.env` đã được điền đầy đủ thông tin cổng, db name và cặp khóa RSA cho dev local).*

### 🚀 Bước 2: Chạy Hạ tầng Docker (Databases, Broker, Gateway, AI)
Tại thư mục root của dự án, khởi chạy PostgreSQL, Redis, RabbitMQ, Ollama và Kong Gateway:
```bash
# 1. Khởi động DB và các Message Broker
docker compose up -d postgres-db redis rabbitmq ollama

# 2. Chờ DB khởi động và tự chạy script init-db.sh để tạo database logic

# 3. Khởi động API Gateway (Kong)
docker compose up -d api-gateway
```

### 🚀 Bước 3: Chạy các Backend Services (Hoặc chạy qua Docker)
Bạn có thể chạy toàn bộ service bằng docker-compose hoặc chạy thủ công từng tab terminal để debug code nhanh hơn.

#### **Cách A: Chạy toàn bộ bằng Docker Compose (Khuyên dùng)**
```bash
docker compose up -d --build
```
*(Lệnh này sẽ build và chạy tất cả backend + frontend).*

#### **Cách B: Chạy thủ công từng Service (Dành cho nhà phát triển)**
Mở các cửa sổ Terminal riêng biệt:
- **Auth Service**:
  ```bash
  cd app/backend/auth-service
  ./mvnw spring-boot:run
  ```
- **Booking Service**:
  ```bash
  cd app/backend/booking-service
  go run cmd/api/main.go
  ```
- **Payment Service**:
  ```bash
  cd app/backend/payment-service
  go run cmd/main.go
  ```
- **Profile Service**:
  ```bash
  cd app/backend/profile-service
  go run cmd/main.go
  ```
- **Assessment Service**:
  ```bash
  cd app/backend/assessment-service
  # Tạo venv và cài dependencies
  python -m venv .venv && source .venv/bin/activate
  pip install -r requirements.txt
  alembic upgrade head
  uvicorn app.main:app --port 5000 --reload
  ```
- **Forum Service**:
  ```bash
  cd app/backend/forum-service
  go run cmd/main.go
  ```
- **Chatroom Service**:
  ```bash
  cd app/backend/chatroom-service
  npm install
  npm run dev
  ```
- **Chatbot Service**:
  ```bash
  cd app/backend/chatbot-service
  python -m venv .venv && source .venv/bin/activate
  pip install -r requirements.txt
  python run_console.py # Hoặc uvicorn server
  ```

### 🚀 Bước 4: Khởi chạy Frontend App
```bash
cd app/frontend
npm install
cp .env.local.example .env.local
npm run dev
```
Truy cập ứng dụng tại: [http://localhost:3000](http://localhost:3000).

---

## 👥 Tài khoản Thử nghiệm Local (Local Credentials)

| Vai trò (Role) | Email tài khoản | Mật khẩu (Password) |
| :--- | :--- | :--- |
| **ADMIN** | `admin@mindcare.com` | `admin@mindcare.com` |
| **EXPERT** | `expert@mindcare.com` | `expert@mindcare.com` |
| **PATIENT** | `patient@mindcare.com` | `patient@mindcare.com` |

> ⚠️ **Chú ý bảo mật**: Tuyệt đối không sử dụng các tài khoản mặc định và mật khẩu này trên môi trường Production thực tế.

---

© 2026 Psychological Counseling System (MindCare) - Tài liệu hướng dẫn phát triển hệ thống.
