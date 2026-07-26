# Tài Liệu Bàn Giao Thiết Kế Unit Test - UC-08: Ghi Nhận Thù Lao & Quản Lý Ví

Tài liệu này chứa Ma trận dò vết và Bảng kế hoạch Unit Test chi tiết bằng Golang cho logic Ghi nhận thù lao chuyên gia và Quản lý ví điện tử (Wallet Credit, Balance Query & Transaction History) thuộc dịch vụ `payment-service`.

---

## 1. MA TRẬN DÒ VẾT (TRACEABILITY MATRIX) - UC-08

| Mã Yêu Cầu (Req ID) | Nhóm Chức Năng | Use Case (UC) | Use Case Name | Test Case Liên Kết (Test Case ID) | Trạng thái |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **REQ-PAY-03** | Quản lý ví & Ghi nhận thù lao | UC-08 | Ghi nhận thù lao | TC-PAY-WLT-01<br>TC-PAY-WLT-02<br>TC-PAY-WLT-03<br>TC-PAY-WLT-04<br>TC-PAY-WLT-05 | **READY TO RUN** |

---

## 2. BẢNG KẾ HOẠCH UNIT TEST CHI TIẾT - UC-08

Dưới đây là bảng kế hoạch chi tiết tích hợp các kiểm thử tầng HTTP Handler (httptest & Gin context) và tầng Service/Usecase cho chức năng Ghi nhận thù lao & Quản lý ví:

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
      <td><b>TC-PAY-WLT-01</b></td>
      <td>Nạp tiền vào ví điện tử thành công (Happy Case - TopUp Wallet)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Role = "PATIENT"</code>, <code>X-User-Id = "22222222-2222-2222-2222-222222222222"</code>.</li>
          <li>Body request JSON: <code>amount = 500000</code>.</li>
          <li>Mock Wallet Usecase: Gọi hàm <code>CreditAvailable</code> với <code>userID = 22222222-2222-2222-2222-222222222222</code>, <code>amount = 500000</code> trả về <code>nil</code> (thành công).</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/payments/wallets/top-up</code> với số tiền 500,000 VND.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = true</code></li>
              <li>Thông báo: <code>"Wallet topped up successfully"</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-PAY-WLT-02</b></td>
      <td>Ghi nhận thù lao vào ví chuyên gia thành công (Credit Expert Revenue)</td>
      <td>
        <ul>
          <li>Chuyên gia có `userID` hợp lệ. Đơn hàng hoàn tất có tổng số tiền 500,000 VND.</li>
          <li>Hệ thống tính chiết khấu sàn 20% (100,000 VND) và thù lao chuyên gia 80% (400,000 VND).</li>
          <li>Mock Wallet Usecase: Gọi <code>CreditAvailable</code> ghi nhận 400,000 VND với <code>refType = "EXPERT_EARNING"</code> trả về <code>nil</code>.</li>
        </ul>
      </td>
      <td>Gọi hàm Usecase <code>CreditAvailable(ctx, expertID, netAmount, "EXPERT_EARNING", orderID, idempotencyKey)</code>.</td>
      <td>
        <ul>
          <li>Không ném ra ngoại lệ.</li>
          <li>Số dư khả dụng của chuyên gia tăng đúng 400,000 VND.</li>
          <li>Giao dịch lưu loại <code>EXPERT_EARNING</code> và ghi nhận Idempotency Key chính xác.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-PAY-WLT-03</b></td>
      <td>Xem thông tin và số dư ví điện tử thành công (Get Wallet Info)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Role = "EXPERT"</code>, <code>X-User-Id = "22222222-2222-2222-2222-222222222222"</code>.</li>
          <li>Mock Wallet Usecase: Hàm <code>GetOrCreateWallet</code> trả về đối tượng <code>Wallet</code> có <code>available_balance = 1000000</code>, <code>pending_balance = 0</code>, <code>locked_balance = 0</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>GET /api/v1/payments/wallets/me</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = true</code></li>
              <li>Data chứa <code>available_balance = 1000000</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-PAY-WLT-04</b></td>
      <td>Xem lịch sử giao dịch ví thành công (Get Wallet History)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Role = "PATIENT"</code>, <code>X-User-Id = "22222222-2222-2222-2222-222222222222"</code>.</li>
          <li>Query parameters: <code>page = 0</code>, <code>size = 10</code>.</li>
          <li>Mock ReadUsecase: Hàm <code>ListTransactionHistory</code> trả về danh sách lịch sử biến động số dư.</li>
        </ul>
      </td>
      <td>Gửi request <code>GET /api/v1/payments/wallets/history?page=0&amp;size=10</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = true</code></li>
              <li>Data chứa mảng <code>items</code> giao dịch ví và thông số phân trang.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-PAY-WLT-05</b></td>
      <td>Lỗi số dư ví không đủ khi thực hiện giao dịch khấu trừ (Insufficient Balance)</td>
      <td>
        <ul>
          <li>Người dùng có số dư khả dụng là 100,000 VND.</li>
          <li>Cố tình thực hiện giao dịch trừ số dư khả dụng với số tiền 200,000 VND.</li>
          <li>Mock Wallet Usecase: Hàm <code>DebitAvailable</code> trả về <code>wallet.ErrInsufficientBalance</code>.</li>
        </ul>
      </td>
      <td>Gọi hàm Usecase <code>DebitAvailable(ctx, userID, 200000, "WITHDRAWAL", refID, key)</code>.</td>
      <td>
        <ul>
          <li>Trả về lỗi <code>wallet.ErrInsufficientBalance</code> ("insufficient balance").</li>
          <li>Số dư ví không bị thay đổi.</li>
        </ul>
      </td>
    </tr>
  </tbody>
</table>
