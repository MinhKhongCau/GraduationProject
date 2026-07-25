# Tài Liệu Bàn Giao Thiết Kế Unit Test - UC-07: Thanh Toán (Payment)

Tài liệu này chứa Ma trận dò vết và Bảng kế hoạch Unit Test chi tiết bằng Golang cho chức năng Thanh toán đơn hàng và Xử lý Webhook VNPay (Create Payment Order & VNPay IPN) thuộc dịch vụ `payment-service`.

---

## 1. MA TRẬN DÒ VẾT (TRACEABILITY MATRIX) - UC-07

| Mã Yêu Cầu (Req ID) | Nhóm Chức Năng | Use Case (UC) | Use Case Name | Test Case Liên Kết (Test Case ID) | Trạng thái |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **REQ-PAY-01** | Tạo đơn thanh toán | UC-07 | Thanh toán | TC-PAY-ORD-01<br>TC-PAY-ORD-02<br>TC-PAY-ORD-03<br>TC-PAY-ORD-04<br>TC-PAY-ORD-05<br>TC-PAY-ORD-06 | **READY TO RUN** |
| **REQ-PAY-02** | Xử lý Webhook VNPay | UC-07 | Thanh toán | TC-PAY-IPN-01<br>TC-PAY-IPN-02<br>TC-PAY-IPN-03<br>TC-PAY-IPN-04 | **READY TO RUN** |

---

## 2. BẢNG KẾ HOẠCH UNIT TEST CHI TIẾT - UC-07

Dưới đây là bảng kế hoạch chi tiết tích hợp các kiểm thử tầng HTTP Handler (httptest & Gin context) và tầng Service/Usecase cho chức năng Thanh toán:

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
      <td><b>TC-PAY-ORD-01</b></td>
      <td>Tạo đơn hàng thanh toán VNPay thành công (Happy Case)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Id = "11111111-1111-1111-1111-111111111111"</code>.</li>
          <li>Body request JSON: <code>appointment_id = "appt-uuid-1"</code>.</li>
          <li>Mock Payment Usecase: Gọi <code>CreateOrder</code> thành công, trả về <code>PaymentOrder</code> có trạng thái <code>PENDING</code> và đường link <code>payment_url = "https://sandbox.vnpayment.vn/..."</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/payments/orders</code> với body JSON hợp lệ.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = true</code></li>
              <li>Thông báo: <code>"Payment order created successfully"</code></li>
              <li>Data chứa <code>order_id</code>, <code>payment_url</code>, <code>gross_amount</code> và trạng thái <code>PENDING</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-PAY-ORD-02</b></td>
      <td>Tạo đơn hàng thất bại do không tìm thấy cuộc hẹn (Not Found)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Id = "11111111-1111-1111-1111-111111111111"</code>.</li>
          <li>Body request JSON: <code>appointment_id = "non-existent-appt"</code>.</li>
          <li>Mock Payment Usecase: Trả về lỗi <code>ErrBookingAppointmentNotFound</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/payments/orders</code> với mã cuộc hẹn không tồn tại.</td>
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
      <td><b>TC-PAY-ORD-03</b></td>
      <td>Tạo đơn hàng thất bại do cuộc hẹn không thuộc về người thanh toán (Forbidden)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Id = "wrong-user-uuid"</code>.</li>
          <li>Body request JSON: <code>appointment_id = "appt-uuid-1"</code>.</li>
          <li>Mock Payment Usecase: Trả về lỗi <code>ErrAppointmentOwnership</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/payments/orders</code> với tài khoản không sở hữu cuộc hẹn.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>403 Forbidden</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = false</code></li>
              <li>Thông báo lỗi: <code>"Appointment does not belong to payer"</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-PAY-ORD-04</b></td>
      <td>Tạo đơn hàng thất bại do cuộc hẹn đã được thanh toán trước đó (Conflict)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Id = "11111111-1111-1111-1111-111111111111"</code>.</li>
          <li>Body request JSON: <code>appointment_id = "appt-already-paid"</code>.</li>
          <li>Mock Payment Usecase: Trả về lỗi <code>ErrAppointmentAlreadyPaid</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/payments/orders</code> với cuộc hẹn đã thanh toán.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>409 Conflict</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = false</code></li>
              <li>Thông báo lỗi: <code>"Appointment is already paid"</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-PAY-ORD-05</b></td>
      <td>Tạo đơn hàng thất bại do thiếu header người dùng (Unauthorized)</td>
      <td>
        <ul>
          <li>Thiếu header <code>X-User-Id</code> (rỗng).</li>
          <li>Body request JSON: <code>appointment_id = "appt-uuid-1"</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/payments/orders</code> không kèm header định danh.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>401 Unauthorized</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = false</code></li>
              <li>Thông báo lỗi: <code>"Missing authenticated payer"</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-PAY-ORD-06</b></td>
      <td>Tạo đơn hàng thất bại do thiếu thông tin body (Bad Request)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Id = "11111111-1111-1111-1111-111111111111"</code>.</li>
          <li>Body request JSON rỗng: <code>{}</code> (thiếu <code>appointment_id</code>).</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/payments/orders</code> với body rỗng.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>400 Bad Request</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = false</code></li>
              <li>Thông báo lỗi: <code>"Invalid request body"</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-PAY-IPN-01</b></td>
      <td>Xử lý VNPay Webhook (IPN) thanh toán thành công (Happy Case)</td>
      <td>
        <ul>
          <li>URL Query params chứa chữ ký hợp lệ: <code>vnp_TxnRef = "order-uuid-1"</code>, <code>vnp_ResponseCode = "00"</code>, <code>vnp_SecureHash = "valid-hash"</code>.</li>
          <li>Mock Payment Usecase: Hàm <code>ProcessIPN</code> xác minh thành công và trả về <code>alreadyProcessed = false, err = nil</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>GET /api/v1/payments/vnpay-ipn?...</code> từ cổng VNPay.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code></li>
          <li>Response JSON chuẩn VNPay:
            <ul>
              <li><code>RspCode = "00"</code></li>
              <li><code>Message = "Confirm Success"</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-PAY-IPN-02</b></td>
      <td>Xử lý VNPay Webhook thất bại do sai chữ ký bảo mật (Invalid Checksum)</td>
      <td>
        <ul>
          <li>URL Query params chứa chữ ký sai hoặc bị can thiệp.</li>
          <li>Mock Payment Usecase: Hàm <code>ProcessIPN</code> ném lỗi chứa chuỗi <code>"checksum mismatch"</code> hoặc <code>"invalid signature"</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>GET /api/v1/payments/vnpay-ipn?...</code> với vnp_SecureHash sai.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code> (VNPay quy định luôn trả 200 kèm RspCode lỗi)</li>
          <li>Response JSON chuẩn VNPay:
            <ul>
              <li><code>RspCode = "97"</code></li>
              <li><code>Message = "Invalid Signature"</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-PAY-IPN-03</b></td>
      <td>Xử lý VNPay Webhook thất bại do đơn hàng không tồn tại (Order Not Found)</td>
      <td>
        <ul>
          <li>URL Query params chứa <code>vnp_TxnRef = "unknown-order"</code>.</li>
          <li>Mock Payment Usecase: Hàm <code>ProcessIPN</code> ném lỗi chứa chuỗi <code>"order not found"</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>GET /api/v1/payments/vnpay-ipn?...</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code></li>
          <li>Response JSON chuẩn VNPay:
            <ul>
              <li><code>RspCode = "01"</code></li>
              <li><code>Message = "Order not found"</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-PAY-IPN-04</b></td>
      <td>Xử lý VNPay Webhook khi đơn hàng đã được xác nhận trước đó (Already Confirmed)</td>
      <td>
        <ul>
          <li>URL Query params hợp lệ gửi trùng lặp (duplication IPN retry).</li>
          <li>Mock Payment Usecase: Hàm <code>ProcessIPN</code> trả về <code>alreadyProcessed = true, err = nil</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>GET /api/v1/payments/vnpay-ipn?...</code> lần 2.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code></li>
          <li>Response JSON chuẩn VNPay:
            <ul>
              <li><code>RspCode = "02"</code></li>
              <li><code>Message = "Order already confirmed"</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
  </tbody>
</table>
