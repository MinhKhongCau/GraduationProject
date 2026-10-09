# Hướng Dẫn Khởi Động Dự Án - MindCare Backend (Microservices)

Chào mừng bạn gia nhập dự án! Hệ thống Backend của MindCare được xây dựng theo kiến trúc **Microservices** đa ngôn ngữ (Java, Golang, Python) kết hợp với **Kong API Gateway**.

Dưới đây là hướng dẫn từ A-Z để bạn (người mới) có thể chạy toàn bộ hệ thống lên và test thử mà không bị lỗi.

---

## 1. Yêu Cầu Cài Đặt Ban Đầu (Prerequisites)
Bạn cần đảm bảo máy tính đã cài đặt các phần mềm sau:
*   **Docker & Docker Compose** (Bắt buộc để chạy Database, Redis, RabbitMQ, Gateway).
*   **Go** (phiên bản 1.21 trở lên).
*   **Java** (phiên bản 21) & Maven.
*   *(Python 3.10+ nếu sau này bạn làm tới Assessment Service).*

---

## 2. Bước 1: Khởi Động Hạ Tầng (Databases & Gateway)

Toàn bộ Database (PostgreSQL), Redis, RabbitMQ và API Gateway đều được đóng gói bằng Docker. Bạn cần chạy chúng lên trước tiên.

Mở Terminal tại thư mục `backend/` và chạy 2 lệnh sau:

**1. Chạy các Databases & Message Brokers:**
```bash
docker compose up -d
```
*(Nếu muốn xem Database UI, bạn có thể dùng DBeaver kết nối vào `localhost`, user: `root`, pass: `rootpassword`. Port: `5434` cho Auth, `5433` cho Booking, `5437` cho Payment).*

**2. Chạy Kong API Gateway (Người gác cổng):**
```bash
docker compose -f docker-compose.gateway.yml up -d --build
```
*(Cổng chính của toàn bộ hệ thống lúc này sẽ là: `http://localhost:8000`)*

---

## 3. Bước 2: Khởi Động Các Microservices

Hệ thống có nhiều Service, nhưng hiện tại bạn chỉ cần quan tâm 3 Service lõi là: **Auth**, **Booking**, và **Payment**.
Bạn hãy mở 3 cửa sổ Terminal (hoặc 3 tab trong VS Code) để chạy 3 service này song song.

### 🔑 3.1. Auth Service (Cổng 8080)
Service này viết bằng Java Spring Boot, quản lý tài khoản và phân quyền (JWT/2FA).
*   **Thư mục:** `backend/auth-service/`
*   **Lệnh chạy:**
    ```bash
    ./mvnw spring-boot:run
    ```

### 🗓️ 3.2. Booking Service (Cổng 8083)
Service này viết bằng Golang, quản lý lịch khám và đặt lịch.
*   **Thư mục:** `backend/booking-service/`
*   **Lệnh chạy:**
    ```bash
    go run cmd/main.go
    ```

### 💰 3.3. Payment Service (Cổng 8082)
Service này viết bằng Golang, quản lý ví tiền và thanh toán.
*   **Thư mục:** `backend/payment-service/`
*   **Lệnh chạy:**
    ```bash
    go run cmd/main.go
    ```

---

## 4. Bước 3: Xem Tài Liệu API (Swagger)

Sau khi khởi chạy thành công, bạn không cần phải cắm mặt vào đọc code để biết API có những gì. Toàn bộ API đều đã được tự động hóa giao diện trực quan bằng Swagger.

Bạn có thể mở trình duyệt và click vào các link sau để test:
*   **Auth Service API:** [http://localhost:8080/swagger-ui/index.html](http://localhost:8080/swagger-ui/index.html)
*   **Booking Service API:** [http://localhost:8083/swagger-ui/index.html](http://localhost:8083/swagger-ui/index.html)
*   **Payment Service API:** [http://localhost:8082/swagger-ui/index.html](http://localhost:8082/swagger-ui/index.html)

---

## 5. Lưu Ý Khi Gọi API (Workflow)

1.  **Tuyệt đối không gọi trực tiếp vào các cổng 8080, 8083, 8082 khi code Frontend.**
2.  Mọi request từ Frontend (hoặc Postman) đều phải bắn vào **Cổng 8000 (Kong Gateway)**.
    *   *Ví dụ thay vì gọi:* `http://localhost:8080/api/v1/auth/login`
    *   *Hãy gọi:* `http://localhost:8000/api/v1/auth/login`
3.  Kong Gateway sẽ tự động kiểm tra Token (nếu có) và định tuyến request về đúng Service bên dưới. Nếu thiếu Token ở các API bảo mật (như đặt lịch), Kong sẽ tự chặn ngay lập tức.

Chúc bạn code vui vẻ! 🚀
