# Tài Liệu Bàn Giao Thiết Kế Unit Test - UC-11: Nhận Thông Báo Hệ Thống

Tài liệu này chứa Ma trận dò vết và Bảng kế hoạch Unit Test chi tiết bằng Golang cho cơ chế đẩy/nhận thông báo thời gian thực (Push Notification & Domain Event Consumers) qua hàng đợi tin nhắn Message Queue thuộc hệ thống backend.

---

## 1. MA TRẬN DÒ VẾT (TRACEABILITY MATRIX) - UC-11

| Mã Yêu Cầu (Req ID) | Nhóm Chức Năng | Use Case (UC) | Use Case Name | Test Case Liên Kết (Test Case ID) | Trạng thái |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **REQ-NOTIF-01** | Đẩy thông báo sự kiện real-time | UC-11 | Nhận thông báo | TC-NOTIF-01<br>TC-NOTIF-02<br>TC-NOTIF-03 | **READY TO RUN** |
| **REQ-NOTIF-02** | Xử lý lỗi & Retry hàng đợi | UC-11 | Nhận thông báo | TC-NOTIF-04<br>TC-NOTIF-05 | **READY TO RUN** |

---

## 2. BẢNG KẾ HOẠCH UNIT TEST CHI TIẾT - UC-11

Dưới đây là bảng kế hoạch chi tiết kiểm thử tầng Notification Event Processor & Message Publisher trong Golang:

<table>
  <thead>
    <tr>
      <th width="15%">ID</th>
      <th width="20%">Tên Test Case</th>
      <th width="30%">Điều kiện (Setup Data)</th>
      <th width="15%">Các bước thực hiện (Execution)</th>
      <th width="20%">Kết quả mong đợi (Expected Output)</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td><b>TC-NOTIF-01</b></td>
      <td>Gửi thông báo sự kiện đặt lịch hẹn thành công (Happy Case - Appointment Event)</td>
      <td>
        <ul>
          <li>Event Data: <code>AppointmentCreatedEvent</code> chứa <code>AppointmentID = 100</code>, <code>UserID = "patient-uuid-1"</code>, <code>Time = "2026-07-25 09:00"</code>.</li>
          <li>Mock Channel/MQ: Đẩy tin nhắn vào Exchange <code>"notification.events"</code>.</li>
        </ul>
      </td>
      <td>Gọi hàm <code>PublishAppointmentNotification(evt)</code>.</td>
      <td>
        <ul>
          <li>Hàm trả về <code>err == nil</code>.</li>
          <li>Payload thông báo được định dạng đúng JSON và gửi đến đúng topic của bệnh nhân.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-NOTIF-02</b></td>
      <td>Gửi thông báo sự kiện thanh toán thành công (Happy Case - Payment Event)</td>
      <td>
        <ul>
          <li>Event Data: <code>PaymentSuccessEvent</code> chứa <code>OrderID = "ORD-888"</code>, <code>Amount = 500000</code>, <code>UserID = "patient-uuid-1"</code>.</li>
          <li>Mock Channel/MQ: Đặt trạng thái kết nối hoạt động bình thường.</li>
        </ul>
      </td>
      <td>Gọi hàm <code>PublishPaymentNotification(evt)</code>.</td>
      <td>
        <ul>
          <li>Hàm trả về <code>err == nil</code>.</li>
          <li>Thông báo ghi nhận đúng số tiền thanh toán và mã đơn hàng.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-NOTIF-03</b></td>
      <td>Đánh dấu thông báo đã đọc thành công (Mark Notification as Read)</td>
      <td>
        <ul>
          <li>CSDL Mock: Đã tồn tại bản ghi Notification ID = 5, <code>IsRead = false</code>, <code>UserID = "patient-uuid-1"</code>.</li>
        </ul>
      </td>
      <td>Gọi hàm <code>MarkAsRead(notifID, userID)</code>.</td>
      <td>
        <ul>
          <li>Trạng thái bản ghi chuyển thành <code>IsRead = true</code>.</li>
          <li>Thời gian <code>ReadAt</code> được cập nhật.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-NOTIF-04</b></td>
      <td>Xử lý thất bại khi kết nối hàng đợi tin nhắn bị ngắt (MQ Connection Disconnected)</td>
      <td>
        <ul>
          <li>Mock Channel/MQ: Mô phỏng lỗi ngắt kết nối RabbitMQ / Broker <code>ErrConnectionClosed</code>.</li>
        </ul>
      </td>
      <td>Gọi hàm <code>PublishAppointmentNotification(evt)</code> khi ngắt kết nối.</td>
      <td>
        <ul>
          <li>Hàm trả về lỗi <code>ErrBrokerUnavailable</code>.</li>
          <li>Tin nhắn được đưa vào hàng đợi chờ đẩy lại (Dead Letter Queue / Retry mechanism).</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-NOTIF-05</b></td>
      <td>Từ chối gửi thông báo do payload sự kiện bị rỗng hoặc sai định dạng (Invalid Payload)</td>
      <td>
        <ul>
          <li>Event Data: Chứa <code>UserID = ""</code> (rỗng) hoặc <code>AppointmentID = 0</code>.</li>
        </ul>
      </td>
      <td>Gọi hàm <code>PublishAppointmentNotification(evt)</code> với dữ liệu không hợp lệ.</td>
      <td>
        <ul>
          <li>Hàm trả về lỗi <code>ErrInvalidNotificationPayload</code>.</li>
          <li>Hệ thống hủy sự kiện và log lại cảnh báo validation.</li>
        </ul>
      </td>
    </tr>
  </tbody>
</table>
