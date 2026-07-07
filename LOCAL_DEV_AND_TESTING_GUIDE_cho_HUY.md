# 📘 CẨM NANG HƯỚNG DẪN: TEST NHANH LUỒNG ĐẶT LỊCH

Tài liệu này tập trung vào hướng dẫn bạn cách gọi các API tuần tự để test luồng hoạt động chính (từ lúc bác sĩ sinh lịch -> bệnh nhân đặt lịch -> thanh toán -> bác sĩ nghỉ đột xuất).

Toàn bộ API dưới đây đều gọi qua API Gateway tại cổng **`8000`** (bạn dùng Postman để test cho tiện nhé).

---

## GIAI ĐOẠN 1: BÁC SĨ SINH LỊCH (GENERATE SLOTS)

Bác sĩ yêu cầu hệ thống sinh tự động các khung giờ trống cho mình dựa trên cấu hình sẵn có (đã được tự động seed lúc chạy app).

- **URL**: `POST http://localhost:8000/api/v1/booking/slots/generate`
- **Headers**:
  - `X-User-Role`: `EXPERT`
  - `X-User-Id`: `ce7b23b0-6b42-4e71-a482-84a8b0839422` (ID bác sĩ mặc định)
- **Body (JSON)**:
  ```json
  {
    "expert_id": "ce7b23b0-6b42-4e71-a482-84a8b0839422",
    "days_to_generate": 14
  }
  ```
- **Kiểm tra**: Sẽ thấy trả về `slots_created: ...` thành công.

---

## GIAI ĐOẠN 2: BỆNH NHÂN TRA CỨU & GIỮ CHỖ

### 1. Xem ngày nào có lịch trống
- **URL**: `GET http://localhost:8000/api/v1/public/booking/slots/available-dates?expert_id=ce7b23b0-6b42-4e71-a482-84a8b0839422`
- **Kiểm tra**: Trả về danh sách ngày có lịch rảnh, dạng `["2026-07-07", "2026-07-08", ...]`. Lấy một ngày bất kỳ để dùng cho API dưới.

### 2. Xem khung giờ trống trong ngày
- **URL**: `GET http://localhost:8000/api/v1/public/booking/slots/available-times?expert_id=ce7b23b0-6b42-4e71-a482-84a8b0839422&date=2026-07-08` (đổi date thành ngày bạn thấy ở bước trên)
- **Kiểm tra**: Trả về danh sách các slot. Copy một `slot_id` bất kỳ.

### 3. Giữ chỗ (Lock Slot)
Giữ slot đó trong 15 phút để chuẩn bị thanh toán.
- **URL**: `POST http://localhost:8000/api/v1/booking/slots/:slot_id/lock` (thay `:slot_id` bằng ID vừa copy)
- **Headers**:
  - `X-User-Role`: `PATIENT`
  - `X-User-Id`: `patient-uuid-999`
- **Kiểm tra**: Nhận được thông báo thành công. (Thử dùng user khác gọi lại API này sẽ bị lỗi `409 Conflict`).

### 4. Tạo cuộc hẹn (Create Appointment)
- **URL**: `POST http://localhost:8000/api/v1/booking/appointments`
- **Headers**:
  - `X-User-Role`: `PATIENT`
  - `X-User-Id`: `patient-uuid-999`
- **Body (JSON)**:
  ```json
  {
    "slot_id": "<slot_id_vừa_lock>",
    "expert_id": "ce7b23b0-6b42-4e71-a482-84a8b0839422"
  }
  ```
- **Kiểm tra**: Trả về thành công kèm theo `appointment_id`. Trạng thái cuộc hẹn lúc này là `PENDING_PAYMENT (0)`. Copy `appointment_id` này lại.

---

## GIAI ĐOẠN 3: XÁC NHẬN THANH TOÁN (WEBHOOK MOCK)

Giả lập cổng thanh toán VNPay/Momo gọi webhook báo thanh toán thành công.

- **URL**: `POST http://localhost:8000/api/v1/public/booking/appointments/webhook`
- **Body (JSON)**:
  ```json
  {
    "appointment_id": "<appointment_id_vừa_copy>",
    "status": "SUCCESS"
  }
  ```
- **Kiểm tra**: Cuộc hẹn chuyển sang trạng thái `CONFIRMED (1)` và slot thành `OCCUPIED (2)`.

---

## GIAI ĐOẠN 4: BÁC SĨ NGHỈ ĐỘT XUẤT (TIME-OFF)

Bác sĩ báo bận đột xuất, đè lên chính khung giờ có người vừa đặt ở trên.

### Bước 1: Khai báo nghỉ phép
Bạn cần truyền timestamp (milliseconds) bao trùm cái slot vừa bị chiếm.
- **URL**: `POST http://localhost:8000/api/v1/booking/time-off`
- **Headers**:
  - `X-User-Role`: `EXPERT`
  - `X-User-Id`: `ce7b23b0-6b42-4e71-a482-84a8b0839422`
- **Body (JSON)**:
  ```json
  {
    "start_datetime": 1783324800000, 
    "end_datetime": 1783342800000,
    "reason": "Bận họp chuyên môn đột xuất"
  }
  ```
- **Kiểm tra**: Do khoảng thời gian này đã có cuộc hẹn `CONFIRMED` bên trên, hệ thống sẽ từ chối và trả về HTTP `409 Conflict` kèm danh sách các `affected_appointments`.

### Bước 2: Ép buộc huỷ lịch (Confirm Time-Off)
Bác sĩ xác nhận bắt buộc phải nghỉ, hệ thống sẽ lưu TimeOff và huỷ cuộc hẹn.
- **URL**: `POST http://localhost:8000/api/v1/booking/time-off/confirm`
- **Headers**:
  - `X-User-Role`: `EXPERT`
  - `X-User-Id`: `ce7b23b0-6b42-4e71-a482-84a8b0839422`
- **Body (JSON)**: Dùng lại y chang Body của Bước 1.
- **Kiểm tra**: Trả về `200 OK`. 
- **Worker chạy ngầm**: Bạn đợi khoảng 1-2 phút, nhìn vào log terminal của ứng dụng sẽ thấy Worker quét qua và tự động đổi trạng thái cuộc hẹn kia thành `CANCELLED`.

### Bước 3: Xem lại lịch nghỉ đã tạo
- **URL**: `GET http://localhost:8000/api/v1/booking/time-off`
- **Headers**:
  - `X-User-Role`: `EXPERT`
  - `X-User-Id`: `ce7b23b0-6b42-4e71-a482-84a8b0839422`
- **Kiểm tra**: Thấy TimeOff đã được lưu.

### Bước 4: Xoá lịch nghỉ (nếu lỡ tạo nhầm)
- **URL**: `DELETE http://localhost:8000/api/v1/booking/time-off/:time_off_id`
- **Headers**: (giống Bước 3)
- **Kiểm tra**: Lịch nghỉ bị xoá khỏi DB. *(Lưu ý: Các cuộc hẹn đã bị huỷ bởi worker sẽ KHÔNG tự động phục hồi).*

---

## GIAI ĐOẠN 5: CÁC API QUẢN LÝ BỔ SUNG ĐỂ TEST ĐỦ LUỒNG

Nếu bạn muốn test toàn diện hệ thống (chủ yếu là các luồng Create, Read, Update cho cấu hình), hãy gọi tiếp các API sau:

### Nhóm 1: Quản lý Template (Ca Khám)
1. **Tạo Ca Khám (Admin)**
   - `POST http://localhost:8000/api/v1/booking/templates`
   - Role: `ADMIN`
   - Body: `{"shift_name": "Ca Tối", "start_time": "18:00", "end_time": "21:00", "slot_duration_minutes": 60, "is_active": true}`
2. **Xem Danh Sách Ca Khám (Public)**
   - `GET http://localhost:8000/api/v1/public/booking/templates`

### Nhóm 2: Quản lý Availability (Lịch Rảnh Cố Định)
3. **Đăng Ký Lịch Rảnh (Expert)**
   - `POST http://localhost:8000/api/v1/booking/availabilities`
   - Role: `EXPERT`
   - Body: `{"template_id": "...", "day_of_week": 4, "effective_from": 1719680400000}`
4. **Xem Lịch Rảnh Của Mình (Expert)**
   - `GET http://localhost:8000/api/v1/booking/availabilities`
   - Role: `EXPERT`
5. **Chỉnh Sửa Lịch Rảnh (Expert)**
   - `PATCH http://localhost:8000/api/v1/booking/availabilities/:id`
   - Role: `EXPERT`
   - Body: `{"is_enabled": false}` (Ví dụ tắt lịch rảnh này đi)

### Nhóm 3: Quản lý Cuộc Hẹn & Slot (Appt & Slot)
6. **Xem Tổng Hợp Slot (Expert)**
   - `GET http://localhost:8000/api/v1/booking/slots/expert`
   - Role: `EXPERT`
   - Trả về tất cả slot (AVAILABLE, LOCKED, OCCUPIED) để vẽ lên lịch.
7. **Xem Lịch Hẹn Bệnh Nhân (Patient)**
   - `GET http://localhost:8000/api/v1/booking/appointments`
   - Role: `PATIENT`
8. **Xem Lịch Hẹn Bác Sĩ (Expert)**
   - `GET http://localhost:8000/api/v1/booking/appointments/expert`
   - Role: `EXPERT`
9. **Huỷ Cuộc Hẹn (Patient / Expert)**
   - `PATCH http://localhost:8000/api/v1/booking/appointments/:appointment_id/cancel`
   - Role: `PATIENT` hoặc `EXPERT`
   - Body: `{"reason": "Bận việc gia đình"}`
