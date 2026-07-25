# Tài Liệu Bàn Giao Thiết Kế Unit Test - UC-14: Trả/Giải Ngân Thù Lao

Tài liệu này chứa Ma trận dò vết và Bảng kế hoạch Unit Test chi tiết bằng Golang cho chức năng Trả/Giải ngân thù lao (Liên kết ngân hàng, Yêu cầu rút tiền & Admin duyệt/từ chối giải ngân) thuộc dịch vụ `payment-service`.

---

## 1. MA TRẬN DÒ VẾT (TRACEABILITY MATRIX) - UC-14

| Mã Yêu Cầu (Req ID) | Nhóm Chức Năng | Use Case (UC) | Use Case Name | Test Case Liên Kết (Test Case ID) | Trạng thái |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **REQ-PAY-04** | Tài khoản ngân hàng | UC-14 | Trả/Giải ngân thù lao | TC-PAY-WTD-01<br>TC-PAY-WTD-02 | **READY TO RUN** |
| **REQ-PAY-05** | Rút tiền & Duyệt giải ngân | UC-14 | Trả/Giải ngân thù lao | TC-PAY-WTD-03<br>TC-PAY-WTD-04<br>TC-PAY-WTD-05<br>TC-PAY-WTD-06<br>TC-PAY-WTD-07 | **READY TO RUN** |

---

## 2. BẢNG KẾ HOẠCH UNIT TEST CHI TIẾT - UC-14

Dưới đây là bảng kế hoạch chi tiết tích hợp các kiểm thử tầng HTTP Handler (httptest & Gin context) và tầng Service/Usecase cho chức năng Trả/Giải ngân thù lao:

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
      <td><b>TC-PAY-WTD-01</b></td>
      <td>Liên kết tài khoản ngân hàng thành công (Happy Case - Link Bank Account)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Role = "EXPERT"</code>, <code>X-User-Id = "22222222-2222-2222-2222-222222222222"</code>.</li>
          <li>Body request JSON: <code>bank_code = "MB"</code>, <code>account_number = "999999999"</code>, <code>account_holder_name = "NGUYEN VAN A"</code>.</li>
          <li>Mock Withdrawal Usecase: Hàm <code>LinkBankAccount</code> lưu và trả về <code>BankAccount</code> với <code>verified = true</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/payments/bank-accounts</code> với thông tin tài khoản hợp lệ.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = true</code></li>
              <li>Thông báo: <code>"Bank account linked successfully"</code></li>
              <li>Data chứa thông tin tài khoản ngân hàng đã được xác minh.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-PAY-WTD-02</b></td>
      <td>Liên kết tài khoản ngân hàng thất bại do tên chủ thẻ không khớp thông tin ngân hàng</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Role = "EXPERT"</code>, <code>X-User-Id = "22222222-2222-2222-2222-222222222222"</code>.</li>
          <li>Body request JSON: <code>bank_code = "MB"</code>, <code>account_number = "1011223344"</code>, <code>account_holder_name = "SAI TEN CHU THE"</code>.</li>
          <li>Mock Withdrawal Usecase: Hàm <code>LinkBankAccount</code> trả về lỗi tên chủ tài khoản không khớp.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/payments/bank-accounts</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>400 Bad Request</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = false</code></li>
              <li>Thông báo lỗi xác minh tên chủ tài khoản từ ngân hàng.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-PAY-WTD-03</b></td>
      <td>Tạo phiếu yêu cầu rút tiền thù lao thành công (Happy Case - Create Withdrawal)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Role = "EXPERT"</code>, <code>X-User-Id = "22222222-2222-2222-2222-222222222222"</code>.</li>
          <li>Body request JSON: <code>bank_account_id = "bank-account-uuid-1"</code>, <code>amount = 2000000</code>.</li>
          <li>Ví chuyên gia có số dư khả dụng đủ 2,000,000 VND.</li>
          <li>Mock Withdrawal Usecase: Hàm <code>CreateWithdrawal</code> khóa 2,000,000 VND số dư khả dụng sang <code>locked_balance</code> và tạo phiếu rút tiền.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/payments/withdrawals</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = true</code></li>
              <li>Thông báo: <code>"Withdrawal request created successfully"</code></li>
              <li>Data chứa <code>amount = 2000000</code> và trạng thái phiếu rút tiền.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-PAY-WTD-04</b></td>
      <td>Tạo phiếu yêu cầu rút tiền thất bại do vượt quá số dư khả dụng (Insufficient Balance)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Role = "EXPERT"</code>, <code>X-User-Id = "22222222-2222-2222-2222-222222222222"</code>.</li>
          <li>Body request JSON: <code>amount = 50000000</code> (vượt quá số dư ví hiện có).</li>
          <li>Mock Withdrawal Usecase: Trả về lỗi <code>wallet.ErrInsufficientBalance</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/payments/withdrawals</code> với số tiền lớn hơn số dư.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>400 Bad Request</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = false</code></li>
              <li>Thông báo lỗi: <code>"Insufficient wallet balance"</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-PAY-WTD-05</b></td>
      <td>Admin phê duyệt phiếu giải ngân thù lao thành công (Happy Case - Approve Withdrawal)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Role = "ADMIN"</code>, <code>X-User-Id = "admin-uuid-1"</code>.</li>
          <li>Path parameter: <code>id = "withdrawal-req-uuid-1"</code> (phiếu đang ở trạng thái PENDING_APPROVAL).</li>
          <li>Body request JSON: <code>note = "Đã chuyển khoản qua Napas247"</code>.</li>
          <li>Mock Withdrawal Usecase: Hàm <code>ApproveWithdrawal</code> đổi trạng thái phiếu thành APPROVED và trả về <code>nil</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/payments/withdrawals/withdrawal-req-uuid-1/approve</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = true</code></li>
              <li>Thông báo: <code>"Withdrawal request approved successfully"</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-PAY-WTD-06</b></td>
      <td>Admin từ chối phiếu rút tiền và hoàn lại tiền ví thành công (Happy Case - Reject Withdrawal)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Role = "ADMIN"</code>, <code>X-User-Id = "admin-uuid-1"</code>.</li>
          <li>Path parameter: <code>id = "withdrawal-req-uuid-1"</code>.</li>
          <li>Body request JSON: <code>note = "Thông tin tài khoản không chính xác"</code>.</li>
          <li>Mock Withdrawal Usecase: Hàm <code>RejectWithdrawal</code> hoàn trả tiền từ <code>locked_balance</code> về <code>available_balance</code> và đổi trạng thái phiếu thành REJECTED.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/payments/withdrawals/withdrawal-req-uuid-1/reject</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = true</code></li>
              <li>Thông báo: <code>"Withdrawal request rejected successfully"</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-PAY-WTD-07</b></td>
      <td>Admin phê duyệt thất bại do phiếu không ở trạng thái chờ duyệt (Pending Approval)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Role = "ADMIN"</code>, <code>X-User-Id = "admin-uuid-1"</code>.</li>
          <li>Path parameter: <code>id = "withdrawal-req-already-processed"</code>.</li>
          <li>Mock Withdrawal Usecase: Trả về lỗi <code>"yêu cầu rút tiền không ở trạng thái chờ duyệt"</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/payments/withdrawals/withdrawal-req-already-processed/approve</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>400 Bad Request</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = false</code></li>
              <li>Thông báo lỗi: <code>"yêu cầu rút tiền không ở trạng thái chờ duyệt"</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
  </tbody>
</table>
