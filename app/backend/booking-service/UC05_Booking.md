# Tài Liệu Bàn Giao Thiết Kế Unit Test - UC-05: Đặt Lịch (Booking)

Tài liệu này chứa Ma trận dò vết và Bảng kế hoạch Unit Test chi tiết bằng Golang cho chức năng Đặt lịch hẹn (Lock Slot & Create Appointment) thuộc dịch vụ `booking-service`.

---

## 1. MA TRẬN DÒ VẾT (TRACEABILITY MATRIX) - UC-05

| Mã Yêu Cầu (Req ID) | Nhóm Chức Năng | Use Case (UC) | Use Case Name | Test Case Liên Kết (Test Case ID) | Trạng thái |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **REQ-BOOK-01** | Khóa giữ chỗ tạm thời | UC-05 | Đặt lịch | TC-BOOK-LCK-01<br>TC-BOOK-LCK-02<br>TC-BOOK-LCK-03<br>TC-BOOK-LCK-04 | **READY TO RUN** |
| **REQ-BOOK-02** | Khởi tạo cuộc hẹn | UC-05 | Đặt lịch | TC-BOOK-APT-01<br>TC-BOOK-APT-02<br>TC-BOOK-APT-03<br>TC-BOOK-APT-04<br>TC-BOOK-APT-05<br>TC-BOOK-APT-06 | **READY TO RUN** |

---

## 2. BẢNG KẾ HOẠCH UNIT TEST CHI TIẾT - UC-05

Dưới đây là bảng kế hoạch chi tiết tích hợp các kiểm thử tầng HTTP Handler (httptest & Gin context) và tầng Service/Usecase cho chức năng Đặt lịch hẹn:

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
      <td><b>TC-BOOK-LCK-01</b></td>
      <td>Giữ chỗ tạm thời thành công (Happy Case - Lock Slot)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Role = "PATIENT"</code>, <code>X-User-Id = "patient-uuid-1"</code>.</li>
          <li>Path parameter: <code>id = "slot-uuid-1"</code>.</li>
          <li>Mock Slot Usecase: Gọi hàm <code>LockSlot("slot-uuid-1", "patient-uuid-1")</code> trả về <code>nil</code> (thành công).</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/booking/slots/slot-uuid-1/lock</code> với đầy đủ headers.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = true</code></li>
              <li>Thông báo: <code>"Slot locked successfully. You have 15 minutes to complete the booking."</code></li>
              <li>Data chứa <code>slot_id = "slot-uuid-1"</code> và thời gian hết hạn <code>expires = 900</code> (15 phút).</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-BOOK-LCK-02</b></td>
      <td>Giữ chỗ thất bại do slot đã bị người khác khóa (Conflict)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Role = "PATIENT"</code>, <code>X-User-Id = "patient-uuid-2"</code>.</li>
          <li>Path parameter: <code>id = "slot-uuid-locked"</code>.</li>
          <li>Mock Slot Usecase: Gọi hàm <code>LockSlot</code> trả về lỗi <code>slot.ErrSlotAlreadyLocked</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/booking/slots/slot-uuid-locked/lock</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>409 Conflict</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = false</code></li>
              <li>Thông báo lỗi: <code>"slot is already locked by someone else"</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-BOOK-LCK-03</b></td>
      <td>Giữ chỗ thất bại do vai trò không phải Bệnh nhân (Forbidden)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Role = "EXPERT"</code>, <code>X-User-Id = "expert-uuid-1"</code>.</li>
          <li>Path parameter: <code>id = "slot-uuid-1"</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/booking/slots/slot-uuid-1/lock</code> với role EXPERT.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>403 Forbidden</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = false</code></li>
              <li>Thông báo lỗi: <code>"Only patients are allowed to lock slots"</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-BOOK-LCK-04</b></td>
      <td>Giữ chỗ thất bại do thiếu header định danh người dùng (Unauthorized)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Role = "PATIENT"</code>, thiếu header <code>X-User-Id</code> (rỗng).</li>
          <li>Path parameter: <code>id = "slot-uuid-1"</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/booking/slots/slot-uuid-1/lock</code> không có X-User-Id.</td>
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
      <td><b>TC-BOOK-APT-01</b></td>
      <td>Tạo cuộc hẹn thành công (Happy Case - Create Appointment)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Role = "PATIENT"</code>, <code>X-User-Id = "patient-uuid-1"</code>.</li>
          <li>Body request JSON: <code>expert_id = "expert-uuid-1"</code>, <code>slot_id = "slot-uuid-1"</code>.</li>
          <li>Mock Appointment Usecase: Gọi <code>CreateAppointment("patient-uuid-1", "expert-uuid-1", "slot-uuid-1")</code> trả về đối tượng <code>Appointment</code> có <code>status = AppointmentStatusPendingPayment</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/booking/appointments</code> với body JSON hợp lệ.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = true</code></li>
              <li>Thông báo: <code>"Appointment created successfully! Please complete payment within 15 minutes."</code></li>
              <li>Trường <code>data</code> chứa <code>appointment_id</code> và trạng thái <code>status = 0</code> (PENDING_PAYMENT).</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-BOOK-APT-02</b></td>
      <td>Tạo cuộc hẹn thất bại do slot chưa được giữ chỗ bởi bệnh nhân này</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Role = "PATIENT"</code>, <code>X-User-Id = "patient-uuid-1"</code>.</li>
          <li>Body request JSON chứa <code>slot_id</code> đang ở trạng thái AVAILABLE hoặc do người khác giữ.</li>
          <li>Mock Appointment Usecase: Trả về lỗi <code>"slot không được giữ bởi bạn, vui lòng thực hiện lại từ đầu"</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/booking/appointments</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>500 Internal Server Error</code> (hoặc 400 theo quy định API)</li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = false</code></li>
              <li>Thông báo lỗi chứa nội dung không giữ slot.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-BOOK-APT-03</b></td>
      <td>Tạo cuộc hẹn thất bại do phiên giữ chỗ hết hạn (15 phút)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Role = "PATIENT"</code>, <code>X-User-Id = "patient-uuid-1"</code>.</li>
          <li>Body request chứa <code>slot_id</code> đã hết hạn khóa (<code>locked_expires_at <= nowMs</code>).</li>
          <li>Mock Appointment Usecase: Trả về lỗi <code>"phiên giữ chỗ đã hết hạn 15 phút, vui lòng chọn lại"</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/booking/appointments</code> với slot hết hạn.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>500 Internal Server Error</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = false</code></li>
              <li>Thông báo lỗi hết hạn phiên giữ chỗ.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-BOOK-APT-04</b></td>
      <td>Tạo cuộc hẹn thất bại do chuyên gia vắng mặt / nghỉ phép (Time-off)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Role = "PATIENT"</code>, <code>X-User-Id = "patient-uuid-1"</code>.</li>
          <li>Mock Appointment Usecase: Trả về lỗi <code>"slot is covered by expert time-off"</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/booking/appointments</code> vào thời gian chuyên gia nghỉ phép.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>500 Internal Server Error</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = false</code></li>
              <li>Thông báo lỗi: <code>"slot is covered by expert time-off"</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-BOOK-APT-05</b></td>
      <td>Tạo cuộc hẹn thất bại do vai trò không phải Bệnh nhân (Forbidden)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Role = "EXPERT"</code>, <code>X-User-Id = "expert-uuid-1"</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/booking/appointments</code> với header role là EXPERT.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>403 Forbidden</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = false</code></li>
              <li>Thông báo lỗi: <code>"Only patients are allowed to book appointments"</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-BOOK-APT-06</b></td>
      <td>Tạo cuộc hẹn thất bại do thiếu thông tin bắt buộc trong body (Bad Request)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Role = "PATIENT"</code>, <code>X-User-Id = "patient-uuid-1"</code>.</li>
          <li>Body request JSON thiếu <code>slot_id</code> hoặc <code>expert_id</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/booking/appointments</code> với body JSON rỗng <code>{}</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>400 Bad Request</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = false</code></li>
              <li>Thông báo lỗi: <code>"Invalid request body data"</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
  </tbody>
</table>
