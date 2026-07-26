# Tài Liệu Bàn Giao Thiết Kế Unit Test - UC-06: Quản Lý Lịch Hẹn

Tài liệu này chứa Ma trận dò vết và Bảng kế hoạch Unit Test chi tiết bằng Golang cho chức năng Quản lý lịch hẹn (Xem danh sách, xem chi tiết và hủy lịch hẹn) thuộc dịch vụ `booking-service`.

---

## 1. MA TRẬN DÒ VẾT (TRACEABILITY MATRIX) - UC-06

| Mã Yêu Cầu (Req ID) | Nhóm Chức Năng | Use Case (UC) | Use Case Name | Test Case Liên Kết (Test Case ID) | Trạng thái |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **REQ-BOOK-03** | Truy vấn danh sách & chi tiết cuộc hẹn | UC-06 | Quản lý lịch hẹn | TC-BOOK-MNG-01<br>TC-BOOK-MNG-02<br>TC-BOOK-MNG-03<br>TC-BOOK-MNG-04 | **READY TO RUN** |
| **REQ-BOOK-04** | Hủy cuộc hẹn | UC-06 | Quản lý lịch hẹn | TC-BOOK-MNG-05<br>TC-BOOK-MNG-06<br>TC-BOOK-MNG-07<br>TC-BOOK-MNG-08 | **READY TO RUN** |

---

## 2. BẢNG KẾ HOẠCH UNIT TEST CHI TIẾT - UC-06

Dưới đây là bảng kế hoạch chi tiết tích hợp các kiểm thử tầng HTTP Handler (httptest & Gin context) và tầng Service/Usecase cho chức năng Quản lý lịch hẹn:

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
      <td><b>TC-BOOK-MNG-01</b></td>
      <td>Bệnh nhân xem danh sách lịch hẹn thành công (Get Patient Appointments)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Role = "PATIENT"</code>, <code>X-User-Id = "patient-uuid-1"</code>.</li>
          <li>Query parameters: <code>page = 0</code>, <code>size = 10</code>.</li>
          <li>Mock ReadUsecase: Hàm <code>ListAppointments</code> trả về danh sách phân trang chứa 1 cuộc hẹn của Bệnh nhân.</li>
        </ul>
      </td>
      <td>Gửi request <code>GET /api/v1/booking/appointments?page=0&amp;size=10</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = true</code></li>
              <li>Thông báo: <code>"Get appointments successfully"</code></li>
              <li>Trường <code>data.items</code> chứa danh sách cuộc hẹn khớp với dữ liệu giả lập.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-BOOK-MNG-02</b></td>
      <td>Chuyên gia xem danh sách lịch hẹn thành công (Get Expert Appointments)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Role = "EXPERT"</code>, <code>X-User-Id = "expert-uuid-1"</code>.</li>
          <li>Query parameters: <code>page = 0</code>, <code>size = 10</code>.</li>
          <li>Mock ReadUsecase: Hàm <code>ListAppointments</code> trả về danh sách cuộc hẹn của Chuyên gia.</li>
        </ul>
      </td>
      <td>Gửi request <code>GET /api/v1/booking/appointments/expert?page=0&amp;size=10</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = true</code></li>
              <li>Thông báo: <code>"Get appointments successfully"</code></li>
              <li>Data chứa <code>items</code> tương ứng của chuyên gia.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-BOOK-MNG-03</b></td>
      <td>Xem chi tiết lịch hẹn thành công (Get Appointment Detail)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Role = "PATIENT"</code>, <code>X-User-Id = "patient-uuid-1"</code>.</li>
          <li>Path parameter: <code>id = "appt-uuid-1"</code>.</li>
          <li>Mock ReadUsecase: Hàm <code>GetAppointmentDetail</code> trả về chi tiết đối tượng <code>Appointment</code> có <code>appointment_id = "appt-uuid-1"</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>GET /api/v1/booking/appointments/appt-uuid-1</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = true</code></li>
              <li>Chi tiết cuộc hẹn với <code>appointment_id = "appt-uuid-1"</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-BOOK-MNG-04</b></td>
      <td>Xem chi tiết lịch hẹn thất bại do không tìm thấy (Not Found)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Role = "PATIENT"</code>, <code>X-User-Id = "patient-uuid-1"</code>.</li>
          <li>Path parameter: <code>id = "non-existent-id"</code>.</li>
          <li>Mock ReadUsecase: Hàm <code>GetAppointmentDetail</code> trả về <code>appointment.ErrNotFound</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>GET /api/v1/booking/appointments/non-existent-id</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>404 Not Found</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = false</code></li>
              <li>Thông báo lỗi: <code>"Appointment not found"</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-BOOK-MNG-05</b></td>
      <td>Hủy lịch hẹn thành công (Happy Case - Cancel Appointment)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Role = "PATIENT"</code>, <code>X-User-Id = "patient-uuid-1"</code>.</li>
          <li>Path parameter: <code>id = "appt-uuid-1"</code>.</li>
          <li>Body request JSON: <code>reason = "Bận việc đột xuất"</code>.</li>
          <li>Mock Appointment Usecase: Hàm <code>CancelAppointment("appt-uuid-1", "patient-uuid-1", "PATIENT", "Bận việc đột xuất")</code> trả về <code>nil</code> (thành công).</li>
        </ul>
      </td>
      <td>Gửi request <code>PATCH /api/v1/booking/appointments/appt-uuid-1/cancel</code> với lý do hủy hợp lệ.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = true</code></li>
              <li>Thông báo: <code>"Appointment cancelled successfully"</code></li>
              <li>Data chứa <code>appointment_id = "appt-uuid-1"</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-BOOK-MNG-06</b></td>
      <td>Hủy lịch hẹn thất bại do cuộc hẹn đã bị hủy trước đó (Conflict/BadRequest)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Role = "PATIENT"</code>, <code>X-User-Id = "patient-uuid-1"</code>.</li>
          <li>Path parameter: <code>id = "appt-already-cancelled"</code>.</li>
          <li>Body request JSON: <code>reason = "Hủy lại"</code>.</li>
          <li>Mock Appointment Usecase: Trả về <code>domain.ErrAppointmentAlreadyCancelled</code> ("appointment already cancelled").</li>
        </ul>
      </td>
      <td>Gửi request <code>PATCH /api/v1/booking/appointments/appt-already-cancelled/cancel</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>500 Internal Server Error</code> (hoặc 400 Bad Request)</li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = false</code></li>
              <li>Thông báo lỗi: <code>"Failed to cancel appointment"</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-BOOK-MNG-07</b></td>
      <td>Hủy lịch hẹn thất bại do thiếu header định danh người dùng (Unauthorized)</td>
      <td>
        <ul>
          <li>Header request: Thiếu <code>X-User-Id</code> hoặc <code>X-User-Role</code>.</li>
          <li>Path parameter: <code>id = "appt-uuid-1"</code>.</li>
          <li>Body request JSON: <code>reason = "Lý do hủy"</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>PATCH /api/v1/booking/appointments/appt-uuid-1/cancel</code> thiếu headers.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>401 Unauthorized</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = false</code></li>
              <li>Thông báo lỗi: <code>"User identity could not be determined"</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-BOOK-MNG-08</b></td>
      <td>Hủy lịch hẹn thất bại do thiếu lý do hủy (Bad Request)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Role = "PATIENT"</code>, <code>X-User-Id = "patient-uuid-1"</code>.</li>
          <li>Path parameter: <code>id = "appt-uuid-1"</code>.</li>
          <li>Body request JSON rỗng: <code>{}</code> (thiếu trường <code>reason</code>).</li>
        </ul>
      </td>
      <td>Gửi request <code>PATCH /api/v1/booking/appointments/appt-uuid-1/cancel</code> với body không có lý do.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>400 Bad Request</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = false</code></li>
              <li>Thông báo lỗi: <code>"Invalid request data"</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
  </tbody>
</table>
