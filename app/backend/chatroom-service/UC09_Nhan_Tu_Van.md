# Tài Liệu Bàn Giao Thiết Kế Unit Test - UC-09: Nhận Tư Vấn

Tài liệu này chứa Ma trận dò vết và Bảng kế hoạch Unit Test chi tiết bằng NodeJS cho chức năng Nhận tư vấn (Xác thực Socket JWT, Kết nối Socket.io và Nhận tin nhắn real-time) thuộc dịch vụ `chatroom-service`.

---

## 1. MA TRẬN DÒ VẾT (TRACEABILITY MATRIX) - UC-09

| Mã Yêu Cầu (Req ID) | Nhóm Chức Năng | Use Case (UC) | Use Case Name | Test Case Liên Kết (Test Case ID) | Trạng thái |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **REQ-CHAT-01** | Xác thực kết nối Socket | UC-09 | Nhận tư vấn | TC-CHAT-AUTH-01<br>TC-CHAT-AUTH-02<br>TC-CHAT-AUTH-03<br>TC-CHAT-AUTH-04 | **READY TO RUN** |
| **REQ-CHAT-02** | Nhận tin nhắn real-time | UC-09 | Nhận tư vấn | TC-CHAT-REC-01<br>TC-CHAT-REC-02 | **READY TO RUN** |

---

## 2. BẢNG KẾ HOẠCH UNIT TEST CHI TIẾT - UC-09

Dưới đây là bảng kế hoạch chi tiết kiểm thử middleware xác thực Socket.IO (`socketAuthMiddleware`) và sự kiện nhận tin nhắn thời gian thực:

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
      <td><b>TC-CHAT-AUTH-01</b></td>
      <td>Xác thực JWT Token hợp lệ khi kết nối Socket (Happy Case)</td>
      <td>
        <ul>
          <li>Cấu hình `JWT_PUBLIC_KEY` hợp lệ trong môi trường Server.</li>
          <li>Giả lập `socket.handshake.auth.token` chứa JWT RS256 có `accountId = "user-uuid-1"` và `role = "PATIENT"`.</li>
        </ul>
      </td>
      <td>Gọi hàm middleware `socketAuthMiddleware(socket, next)`.</td>
      <td>
        <ul>
          <li>Hàm `next()` được gọi không kèm lỗi.</li>
          <li>Gán dữ liệu người dùng vào `socket.data.user = { id: "user-uuid-1", email: "user@example.com", role: "PATIENT" }`.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-CHAT-AUTH-02</b></td>
      <td>Từ chối kết nối Socket do thiếu JWT Token (Missing Token)</td>
      <td>
        <ul>
          <li>Giả lập `socket.handshake.auth` không truyền thuộc tính `token` (null/undefined).</li>
        </ul>
      </td>
      <td>Gọi hàm middleware `socketAuthMiddleware(socket, next)`.</td>
      <td>
        <ul>
          <li>Hàm `next(err)` nhận lỗi với thông điệp: <code>"unauthorized: missing token"</code>.</li>
          <li>Không gán thuộc tính `socket.data.user`.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-CHAT-AUTH-03</b></td>
      <td>Từ chối kết nối Socket do JWT Token bị hết hạn (Expired Token)</td>
      <td>
        <ul>
          <li>Giả lập `socket.handshake.auth.token` bị hết hạn (TokenExpiredError).</li>
        </ul>
      </td>
      <td>Gọi hàm middleware `socketAuthMiddleware(socket, next)`.</td>
      <td>
        <ul>
          <li>Hàm `next(err)` nhận lỗi với thông điệp: <code>"unauthorized: token expired"</code>.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-CHAT-AUTH-04</b></td>
      <td>Từ chối kết nối Socket do Server chưa cấu hình Public Key</td>
      <td>
        <ul>
          <li>Giả lập môi trường thiếu `JWT_PUBLIC_KEY` (hoặc rỗng).</li>
        </ul>
      </td>
      <td>Gọi hàm middleware `socketAuthMiddleware(socket, next)`.</td>
      <td>
        <ul>
          <li>Hàm `next(err)` nhận lỗi với thông điệp: <code>"unauthorized: server misconfigured (JWT_PUBLIC_KEY not set)"</code>.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-CHAT-REC-01</b></td>
      <td>Khách hàng/Chuyên gia nhận tin nhắn trò chuyện trực tiếp (DM Message Event)</td>
      <td>
        <ul>
          <li>Socket của đối phương đang online và được theo dõi bởi `getUserSocket(toUser)`.</li>
          <li>Chuyên gia/Khách hàng phát sự kiện `dm:send` với nội dung hợp lệ.</li>
        </ul>
      </td>
      <td>Khách hàng đăng ký lắng nghe sự kiện <code>socket.on("dm:message", callback)</code>.</td>
      <td>
        <ul>
          <li>Sự kiện `dm:message` được phát tới đúng Socket ID của người nhận.</li>
          <li>Payload chứa thông tin tin nhắn bao gồm `dmId`, `text`, `fromId`, `toId` và `createdAt`.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-CHAT-REC-02</b></td>
      <td>Nhận tín hiệu báo đang gõ bàn phím từ đối phương (DM Typing Indicator)</td>
      <td>
        <ul>
          <li>Socket đối phương phát sự kiện <code>dm:typing</code> với `toUser` và `isTyping = true`.</li>
        </ul>
      </td>
      <td>Đăng ký lắng nghe sự kiện <code>socket.on("dm:typing:status", callback)</code>.</td>
      <td>
        <ul>
          <li>Callback nhận đúng thông tin <code>fromId</code> và trạng thái <code>isTyping = true</code>.</li>
        </ul>
      </td>
    </tr>
  </tbody>
</table>
