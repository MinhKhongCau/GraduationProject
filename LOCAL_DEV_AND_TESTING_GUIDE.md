# 📘 CẨM NANG HƯỚNG DẪN: CHẠY & TEST DỰ ÁN BOOKING SERVICE
> **MindCare Project** - Hệ thống đặt lịch khám chuyên gia tâm lý tích hợp API Gateway (Kong) và Go Microservice.

Tài liệu này hướng dẫn chi tiết cách thiết lập môi trường phát triển local, khởi chạy các service và thực hiện kiểm thử (test) các API qua cổng API Gateway.

---

## 1. Chuẩn Bị Môi Trường Local

### 1.1. Cấu hình biến môi trường (`.env`)
Đảm bảo bạn đã có file `.env` ở **thư mục gốc** (`GraduationProject/.env`) và trong **booking-service** (`GraduationProject/app/backend/booking-service/.env`) với nội dung port đã được điều chỉnh thành **`5433`** để tránh tranh chấp với Postgres cài trên Windows.

```ini
# GraduationProject/app/backend/booking-service/.env
APP_ENV=dev
DB_HOST=127.0.0.1
DB_PORT=5433
DB_USER=admin
DB_PASSWORD=admin
BOOKING_DB_NAME=booking_db
DB_SSLMODE=disable
JWT_SECRET_KEY=404E635266556A586E3272357538782F413F4428472B4B6250645367566B5970
```

---

## 2. Cơ chế Xác thực & Phân quyền (Role Authorization)

Booking Service hoạt động đằng sau **Kong API Gateway**. 
- Kong giải mã mã token JWT và gửi thông tin danh tính về cho Booking Service qua HTTP Headers.
- Booking Service không giải mã JWT trực tiếp mà kiểm tra phân quyền bằng cách đọc các Headers:
  - `X-User-Role`: Chứa quyền hạn (`ADMIN`, `EXPERT`, `PATIENT`).
  - `X-User-Id`: Chứa UUID của tài khoản người dùng từ Auth Service.

> [!TIP]
> **Nơi kiểm tra Role trong code:**
> Bạn có thể mở trực tiếp các file Handler dưới đây để xem cách hệ thống kiểm tra Role:
> - **ADMIN**: Đầu hàm `CreateTemplate` trong file [schedule_handler.go](file:///e:/doAnThucTapTotNghiep/GraduationProject/app/backend/booking-service/internal/delivery/http/schedules/handler/schedule_handler.go#L31)
> - **EXPERT**: Đầu hàm `Handle` trong [generate_slots.go](file:///e:/doAnThucTapTotNghiep/GraduationProject/app/backend/booking-service/internal/delivery/http/slots/handler/generate_slots.go#L44) hoặc [timeoff_handler.go](file:///e:/doAnThucTapTotNghiep/GraduationProject/app/backend/booking-service/internal/delivery/http/timeoff/handler/timeoff_handler.go#L37)
> - **PATIENT**: Đầu hàm `Handle` trong [lock_slot.go](file:///e:/doAnThucTapTotNghiep/GraduationProject/app/backend/booking-service/internal/delivery/http/slots/handler/lock_slot.go#L35) hoặc [create_appointment.go](file:///e:/doAnThucTapTotNghiep/GraduationProject/app/backend/booking-service/internal/delivery/http/appointments/handler/create_appointment.go#L43)

---

## 3. Hướng Dẫn Khởi Chạy Hệ Thống

### Bước 1: Khởi động Database Postgres
1. Mở Terminal tại thư mục gốc của dự án (`GraduationProject`).
2. Chạy lệnh dựng database:
   ```bash
   docker-compose -f docker-compose.dev.yml up -d postgres-db
   ```

### Bước 2: Build Gateway & Khởi chạy Kong
1. Đứng ở thư mục gốc `GraduationProject`.
2. Build image gateway chứa Go plugin:
   ```bash
   docker build -t mindcare/api-gateway-dev:latest ./app/backend/gateway
   ```
3. Thiết lập biến môi trường `JWT_SECRET_KEY` cho Kong (PowerShell):
   ```powershell
   $env:JWT_SECRET_KEY="404E635266556A586E3272357538782F413F4428472B4B6250645367566B5970"
   ```
4. Khởi chạy Gateway:
   ```bash
   docker-compose -f docker-compose.dev.api-gateway.local.yml up -d
   ```

### Bước 3: Chạy Booking Service ở máy thật
1. Mở Terminal tại thư mục `booking-service`:
   ```bash
   cd app/backend/booking-service
   ```
2. Khởi chạy ứng dụng:
   ```bash
   go run ./cmd/api
   ```
   > [!NOTE]
   > Hệ thống **tự động seed sẵn 2 ca mẫu (Morning/Afternoon)** vào bảng `Booking_Config_Time_Templates` nếu cơ sở dữ liệu trống. Bạn không cần chạy script SQL thủ công nữa!

---

## 4. Quy Trình Kiểm Thử (Test) Chi Tiết Các API qua Gateway (Port 8000)

---

### GIAI ĐOẠN 1: CẤU HÌNH & SINH LỊCH (ADMIN & EXPERT)

### API 1: Admin tạo ca làm việc mẫu (Create Time Template)
Nếu muốn tạo thêm ca mẫu mới ngoài các ca mặc định đã được hệ thống tự động seed.

- **URL**: `POST http://localhost:8000/api/v1/booking/templates`
- **Headers**:
  - `X-User-Role`: `ADMIN`
- **Body (JSON)**:
  ```json
  {
    "shift_name": "Ca Tối (18h-21h)",
    "start_time": "18:00",
    "end_time": "21:00",
    "slot_duration_minutes": 60
  }
  ```

---

### API 2: Chuyên gia đăng ký lịch rảnh (Create Availability)
Bác sĩ đăng ký thời gian mình có thể nhận khám.

- **URL**: `POST http://localhost:8000/api/v1/booking/availabilities`
- **Headers**:
  - `X-User-Role`: `EXPERT`
  - `X-User-Id`: `expert-uuid-111`
- **Body (JSON)**:
  ```json
  {
    "template_id": "t1-uuid-ca-sang-001", -- ID ca mẫu đã có trong DB
    "day_of_week": 1, -- Thứ 2 (1=Mon ... 7=Sun)
    "effective_from": 1719680400000 -- Ngày bắt đầu có hiệu lực (Unix ms)
  }
  ```

---

### API 3: Sinh Lịch Khám Tự Động (Generate Slots)
Bác sĩ yêu cầu hệ thống sinh tự động các slot trống cho mình dựa trên cấu hình Availability.

- **URL**: `POST http://localhost:8000/api/v1/booking/slots/generate`
- **Headers**:
  - `X-User-Role`: `EXPERT`
  - `X-User-Id`: `expert-uuid-111`
- **Body (JSON)**:
  ```json
  {
    "expert_id": "expert-uuid-111",
    "days_to_generate": 14
  }
  ```

---

### GIAI ĐOẠN 2: BỆNH NHÂN TRA CỨU & GIỮ CHỖ (PATIENT)

### API 4: Xem Các Ngày Có Lịch Trống (Get Available Dates)
Bệnh nhân tìm xem những ngày nào chuyên gia có lịch rảnh.

- **URL**: `GET http://localhost:8000/api/v1/public/booking/slots/available-dates`
- **Params**:
  - `expert_id`: `expert-uuid-111`
  - `start_date`: `2026-07-06`
  - `end_date`: `2026-07-20`

---

### API 5: Xem Khung Giờ Trống Trong Ngày (Get Available Times)
Bệnh nhân chọn một ngày cụ thể để xem chi tiết các khung giờ.

- **URL**: `GET http://localhost:8000/api/v1/public/booking/slots/available-times`
- **Params**:
  - `expert_id`: `expert-uuid-111`
  - `date`: `2026-07-06`

---

### API 6: Giữ Chỗ Tạm Thời (Lock Slot)
Bệnh nhân bấm chọn khung giờ và giữ chỗ trong vòng 15 phút.

- **URL**: `POST http://localhost:8000/api/v1/booking/slots/:id/lock` (thay `:id` bằng `slot_id` thực tế ở API 5)
- **Headers**:
  - `X-User-Role`: `PATIENT`
  - `X-User-Id`: `patient-uuid-999`

---

### API 7: Tạo Lịch Hẹn (Create Appointment)
Trong vòng 15 phút giữ chỗ, bệnh nhân điền thông tin và xác nhận đặt lịch.

- **URL**: `POST http://localhost:8000/api/v1/booking/appointments`
- **Headers**:
  - `X-User-Role`: `PATIENT`
  - `X-User-Id`: `patient-uuid-999`
- **Body (JSON)**:
  ```json
  {
    "slot_id": "uuid-cua-slot-da-lock",
    "patient_id": "patient-uuid-999",
    "expert_id": "expert-uuid-111"
  }
  ```

---

### GIAI ĐOẠN 3: XÁC NHẬN THANH TOÁN (MOCK)

### API 8: Mock Payment Webhook
Giả lập phản hồi từ cổng thanh toán báo về hệ thống khi người dùng thanh toán xong.

- **URL**: `POST http://localhost:8000/api/v1/public/booking/appointments/webhook`
- **Body (JSON)**:
  ```json
  {
    "appointment_id": "uuid-appointment-cua-buoc-7",
    "status": "success"
  }
  ```

---

### GIAI ĐOẠN 4: ĐĂNG KÝ NGHỈ PHÉP ĐỘT XUẤT (EXPERT)

### API 9: Bác Sĩ Đăng Ký Nghỉ Đột Xuất (Time Off)
Bác sĩ đăng ký nghỉ phép đột xuất chèn qua các slot đã sinh.

- **URL**: `POST http://localhost:8000/api/v1/booking/time-off`
- **Headers**:
  - `X-User-Role`: `EXPERT`
  - `X-User-Id`: `expert-uuid-111`
- **Body (JSON)**:
  ```json
  {
    "start_datetime": 1783324800000, 
    "end_datetime": 1783342800000,
    "reason": "Bận họp chuyên môn đột xuất",
    "force": false
  }
  ```

#### Kịch Bản Test Khác Biệt:
- **Trường hợp A (Chỉ đè lên slot rảnh):** 
  - Gọi API với `"force": false`. Nếu trong khoảng giờ đó không có lịch nào đã được đặt (`OCCUPIED`), hệ thống tạo `TimeOff` thành công.
- **Trường hợp B (Đè trúng lịch hẹn của bệnh nhân - Cảnh báo):**
  - Giả sử khung giờ này chứa Slot ở bước 7 đã chuyển sang `OCCUPIED`.
  - Gọi API với `"force": false`. Hệ thống từ chối và trả về HTTP `409 Conflict` kèm mảng `affected_appointments` chứa ID lịch hẹn bị đụng để bác sĩ tự cân nhắc.
- **Trường hợp C (Đè trúng lịch hẹn - Cưỡng chế):**
  - Gửi lại request trên với `"force": true`.
  - Hệ thống tạo `TimeOff` thành công. 
  - Quá 30 giây sau, Background Worker sẽ quét qua, tự động chuyển trạng thái lịch hẹn bị đè sang `CANCELLED (2)` (ghi rõ `cancelled_by = EXPERT`) và xoá các slot trống để không cho ai đặt nữa.
