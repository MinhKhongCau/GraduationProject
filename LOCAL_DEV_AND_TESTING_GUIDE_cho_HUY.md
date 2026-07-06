# 📘 CẨM NANG HƯỚNG DẪN: CHẠY & TEST DỰ ÁN BOOKING SERVICE
> **MindCare Project** - Hệ thống đặt lịch khám chuyên gia tâm lý tích hợp API Gateway (Kong) và Go Microservice.

Tài liệu này hướng dẫn chi tiết cách thiết lập môi trường phát triển local, khởi chạy các service và thực hiện kiểm thử (test) 7 API cốt lõi qua cổng API Gateway.

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

## 2. Hướng Dẫn Khởi Chạy Hệ Thống

### Bước 1: Khởi động Database Postgres
1. Mở Terminal tại thư mục gốc của dự án (`GraduationProject`).
2. Chạy lệnh dựng database:
   ```bash
   docker-compose -f docker-compose.dev.yml up -d postgres-db
   ```
   > [!TIP]
   > Cổng kết nối vào Database từ ngoài máy thật lúc này sẽ là `5433`.

### Bước 2: Build Gateway & Khởi chạy Kong
1. Đứng ở thư mục gốc `GraduationProject`.
2. Build image gateway chứa Go plugin (chỉ cần chạy 1 lần đầu tiên):
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

### Bước 3: Tạo Dữ Liệu Mẫu (Seed Data)
Để sinh được các slot khám, bạn cần chèn dữ liệu mẫu về ca làm việc (`Booking_Config_Time_Templates`) và lịch rảnh (`Booking_Config_Availability`) của chuyên gia vào database.

Kết nối vào Postgres thông qua **DBeaver / pgAdmin** (hoặc dùng terminal) với thông tin:
- **Host**: `127.0.0.1`
- **Port**: `5433`
- **User**: `admin`
- **Password**: `admin`
- **Database**: `booking_db`

Chạy đoạn mã SQL sau để chèn dữ liệu mẫu:
```sql
-- 1. Chèn Ca mẫu dùng chung hệ thống
INSERT INTO "Booking_Config_Time_Templates" (template_id, shift_name, start_time, end_time, slot_duration_minutes, is_active)
VALUES 
('t1-uuid-ca-sang-001', 'Ca Sáng (08h-12h)', '08:00', '12:00', 60, true),
('t2-uuid-ca-chieu-002', 'Ca Chiều (13h-17h)', '13:00', '17:00', 60, true);

-- 2. Cài lịch rảnh cho chuyên gia (Expert ID: "expert-uuid-111")
-- Thứ 2 (day_of_week = 1) làm Ca Sáng
-- Thứ 4 (day_of_week = 3) làm Ca Chiều
INSERT INTO "Booking_Config_Availability" (availability_id, expert_id, template_id, day_of_week, is_enabled, effective_from, effective_until)
VALUES 
('a1-avail-mon', 'expert-uuid-111', 't1-uuid-ca-sang-001', 1, true, 1719680400000, NULL), -- 1719680400000 = timestamp ms
('a2-avail-wed', 'expert-uuid-111', 't2-uuid-ca-chieu-002', 3, true, 1719680400000, NULL);
```

### Bước 4: Chạy Booking Service ở máy thật
1. Mở Terminal mới, đi vào thư mục `booking-service`:
   ```bash
   cd app/backend/booking-service
   ```
2. Khởi chạy ứng dụng:
   ```bash
   go run ./cmd/api
   ```
   > [!NOTE]
   > Chương trình sẽ kết nối thành công vào port `5433`, tự động migrate dữ liệu và chạy ở cổng `8083`.

---

## 3. Quy Trình Kiểm Thử (Test) Chi Tiết 7 API qua Gateway (Port 8000)

> [!IMPORTANT]
> Toàn bộ các API được test dưới đây đều gọi qua Gateway tại cổng **`8000`** thay vì gọi trực tiếp cổng `8083`.

---

### GIAI ĐOẠN 1: CẤU HÌNH & SINH LỊCH

### API 1: Sinh Lịch Khám Tự Động (Generate Slots)
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
- **Kết quả mong đợi**: Trả về thông báo thành công và số lượng slot được sinh ra (Ví dụ: `slots_created: 8`). Hệ thống tự chia nhỏ ca 4h thành các slot 60 phút và bỏ qua các ngày trùng lịch nghỉ.

---

### GIAI ĐOẠN 2: BỆNH NHÂN TRA CỨU & GIỮ CHỖ

### API 2: Xem Các Ngày Có Lịch Trống (Get Available Dates)
Bệnh nhân tìm xem những ngày nào chuyên gia có lịch rảnh.

- **URL**: `GET http://localhost:8000/api/v1/public/booking/slots/available-dates`
- **Params**:
  - `expert_id`: `expert-uuid-111`
  - `start_date`: `2026-07-06`
  - `end_date`: `2026-07-20`
- **Kết quả mong đợi**: Trả về mảng các chuỗi ngày dạng `["2026-07-06", "2026-07-08", ...]` tương ứng với những ngày có Slot rảnh.

---

### API 3: Xem Khung Giờ Trống Trong Ngày (Get Available Times)
Bệnh nhân chọn một ngày cụ thể để xem chi tiết các khung giờ.

- **URL**: `GET http://localhost:8000/api/v1/public/booking/slots/available-times`
- **Params**:
  - `expert_id`: `expert-uuid-111`
  - `date`: `2026-07-06`
- **Kết quả mong đợi**: Trả về danh sách các slot trống:
  ```json
  [
    {
      "slot_id": "cfa84bb6-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
      "start_time": 1783324800000,
      "end_time": 1783328400000
    }
  ]
  ```

---

### API 4: Giữ Chỗ Tạm Thời (Lock Slot)
Bệnh nhân bấm chọn khung giờ và giữ chỗ trong vòng 15 phút.

- **URL**: `POST http://localhost:8000/api/v1/booking/slots/:id/lock` (thay `:id` bằng `slot_id` thực tế ở API 3)
- **Headers**:
  - `X-User-Role`: `PATIENT`
  - `X-User-Id`: `patient-uuid-999`
- **Kết quả mong đợi**: Trả về `200 OK`. Slot chuyển sang trạng thái `LOCKED (1)` trên DB.
- **Kịch bản test phụ**:
  - Thử lấy User khác gọi API Lock slot đó tiếp -> Hệ thống phải trả về lỗi `409 Conflict` báo slot đã bị giữ chỗ.

---

### API 5: Tạo Lịch Hẹn (Create Appointment)
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
- **Kết quả mong đợi**: Trả về thông tin cuộc hẹn được tạo với trạng thái ban đầu là `PENDING_PAYMENT (0)`.

---

### GIAI ĐOẠN 3: XÁC NHẬN THANH TOÁN (MOCK)

### API 6: Mock Payment Webhook
Giả lập phản hồi từ cổng thanh toán báo về hệ thống khi người dùng thanh toán xong.

- **URL**: `POST http://localhost:8000/api/v1/public/booking/appointments/webhook`
- **Body (JSON)**:
  ```json
  {
    "appointment_id": "uuid-appointment-cua-buoc-5",
    "status": "success"
  }
  ```
- **Kết quả mong đợi**: Trả về xác nhận thành công. Trạng thái `Appointment` chuyển sang `CONFIRMED (1)` và trạng thái của `ExpertSlot` chuyển sang `OCCUPIED (2)`.

---

### GIAI ĐOẠN 4: ĐĂNG KÝ NGHỈ PHÉP ĐỘT XUẤT

### API 7: Bác Sĩ Đăng Ký Nghỉ Đột Xuất (Time Off)
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
  - Giả sử khung giờ này chứa Slot ở bước 5 đã chuyển sang `OCCUPIED`.
  - Gọi API với `"force": false`. Hệ thống từ chối và trả về HTTP `409 Conflict` kèm mảng `affected_appointments` chứa ID lịch hẹn bị đụng để bác sĩ tự cân nhắc.
- **Trường hợp C (Đè trúng lịch hẹn - Cưỡng chế):**
  - Gửi lại request trên với `"force": true`.
  - Hệ thống tạo `TimeOff` thành công. 
  - Quá 30 giây sau, Background Worker sẽ quét qua, tự động chuyển trạng thái lịch hẹn bị đè sang `CANCELLED (2)` (ghi rõ `cancelled_by = EXPERT`) và xoá các slot trống để không cho ai đặt nữa. Check log để thấy thông báo TODO của dịch vụ notification.

---

## 4. Xử Lý Các Sự Cố Hay Gặp (Troubleshooting)

1. **Lỗi `cannot execute: required file not found` khi khởi động database:**
   - Do file `init-db.sh` bị lưu bằng định dạng dòng Windows (CRLF). 
   - Khắc phục bằng cách chạy PowerShell chuyển về LF như đã hướng dẫn ở trên, sau đó xoá sạch volume cũ bằng `docker-compose -f docker-compose.dev.yml down -v` và khởi chạy lại.
2. **Lỗi `Only one usage of each socket address is normally permitted`:**
   - Cổng `8083` (Booking service) hoặc `5433` (Postgres) đang bị chiếm bởi một tiến trình chạy ngầm trước đó.
   - Kiểm tra bằng netstat: `netstat -ano | findstr 8083` và kill tiến trình đó trước khi chạy lại Go.
