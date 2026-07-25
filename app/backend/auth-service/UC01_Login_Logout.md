# Tài Liệu Bàn Giao Thiết Kế Unit Test - UC-01: Login & Logout

Tài liệu này chứa Ma trận dò vết và Bảng kế hoạch Unit Test chi tiết cho chức năng Đăng nhập (Login) và Đăng xuất (Logout) thuộc dịch vụ `auth-service`.

---

## 1. MA TRẬN DÒ VẾT (TRACEABILITY MATRIX) - UC-01

| Mã Yêu Cầu (Req ID) | Nhóm Chức Năng | Use Case (UC) | Use Case Name | Test Case Liên Kết (Test Case ID) | Trạng thái |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **REQ-AUTH-01** | Xác thực người dùng | UC-01 | Login (Đăng nhập) | TC-AUTH-LGN-CON-01<br>TC-AUTH-LGN-CON-02<br>TC-AUTH-LGN-CON-03<br>TC-AUTH-LGN-CON-04<br>TC-AUTH-LGN-CON-05<br>TC-AUTH-LGN-CON-06<br>TC-AUTH-LGN-SVC-01<br>TC-AUTH-LGN-SVC-02<br>TC-AUTH-LGN-SVC-03 | **READY TO RUN** |
| **REQ-AUTH-02** | Xác thực người dùng | UC-01 | Logout (Đăng xuất) | TC-AUTH-LGT-CON-01<br>TC-AUTH-LGT-CON-02<br>TC-AUTH-LGT-CON-03<br>TC-AUTH-LGT-SVC-01<br>TC-AUTH-LGT-SVC-02 | **READY TO RUN** |

---

## 2. BẢNG KẾ HOẠCH UNIT TEST CHI TIẾT - LOGIN

Dưới đây là bảng kế hoạch chi tiết tích hợp cả Controller Layer và Service Layer cho chức năng Login:

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
      <td><b>TC-AUTH-LGN-CON-01</b></td>
      <td>Đăng nhập thành công (Happy Case - Controller)</td>
      <td>
        <ul>
          <li>Mock <code>LoginUseCase.execute(any(LoginCommand.class))</code> trả về một đối tượng <code>LoginResponse</code> hợp lệ.</li>
          <li>Tạo body request: <code>email = "user@example.com"</code>, <code>password = "Secret123!"</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/auth/login</code> với body JSON chứa dữ liệu thiết lập.</td>
      <td>
        <ul>
          <li>HTTP Status: <code>200 OK</code></li>
          <li>Response Body chứa:
            <ul>
              <li><code>success = true</code></li>
              <li><code>message = "Login successful!"</code></li>
              <li><code>data</code> khớp với <code>LoginResponse</code> đã mock.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-AUTH-LGN-CON-02</b></td>
      <td>Lỗi validate do để trống email</td>
      <td>
        <ul>
          <li>Tạo body request: <code>email = ""</code> hoặc <code>null</code>, <code>password = "Secret123!"</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/auth/login</code> với body JSON chứa dữ liệu thiết lập.</td>
      <td>
        <ul>
          <li>HTTP Status: <code>400 Bad Request</code></li>
          <li>Response Body chứa:
            <ul>
              <li><code>success = false</code></li>
              <li><code>message = "Invalid input data"</code></li>
              <li><code>error</code> chứa: <code>"email": "Email is required"</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-AUTH-LGN-CON-03</b></td>
      <td>Lỗi validate do email sai định dạng</td>
      <td>
        <ul>
          <li>Tạo body request: <code>email = "invalid-email-format"</code>, <code>password = "Secret123!"</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/auth/login</code> với body JSON chứa dữ liệu thiết lập.</td>
      <td>
        <ul>
          <li>HTTP Status: <code>400 Bad Request</code></li>
          <li>Response Body chứa:
            <ul>
              <li><code>success = false</code></li>
              <li><code>message = "Invalid input data"</code></li>
              <li><code>error</code> chứa: <code>"email": "Invalid email format"</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-AUTH-LGN-CON-04</b></td>
      <td>Lỗi validate do để trống password</td>
      <td>
        <ul>
          <li>Tạo body request: <code>email = "user@example.com"</code>, <code>password = ""</code> hoặc <code>null</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/auth/login</code> với body JSON chứa dữ liệu thiết lập.</td>
      <td>
        <ul>
          <li>HTTP Status: <code>400 Bad Request</code></li>
          <li>Response Body chứa:
            <ul>
              <li><code>success = false</code></li>
              <li><code>message = "Invalid input data"</code></li>
              <li><code>error</code> chứa: <code>"password": "Password is required"</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-AUTH-LGN-CON-05</b></td>
      <td>Xử lý lỗi nghiệp vụ từ Service</td>
      <td>
        <ul>
          <li>Mock <code>LoginUseCase.execute(any(LoginCommand.class))</code> ném ra <code>RuntimeException("Incorrect password!")</code>.</li>
          <li>Tạo body request hợp lệ.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/auth/login</code> với body JSON chứa dữ liệu thiết lập.</td>
      <td>
        <ul>
          <li>HTTP Status: <code>400 Bad Request</code></li>
          <li>Response Body chứa:
            <ul>
              <li><code>success = false</code></li>
              <li><code>message = "Business validation failed"</code></li>
              <li><code>error = "Incorrect password!"</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-AUTH-LGN-CON-06</b></td>
      <td>Xử lý lỗi hệ thống không xác định</td>
      <td>
        <ul>
          <li>Mock <code>LoginUseCase.execute(any(LoginCommand.class))</code> ném ra checked <code>Exception("Checked exception simulation")</code> thông qua Mockito answer.</li>
          <li>Tạo body request hợp lệ.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/auth/login</code> với body JSON chứa dữ liệu thiết lập.</td>
      <td>
        <ul>
          <li>HTTP Status: <code>500 Internal Server Error</code></li>
          <li>Response Body chứa:
            <ul>
              <li><code>success = false</code></li>
              <li><code>message = "System error"</code></li>
              <li><code>error = "INTERNAL_SERVER_ERROR"</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-AUTH-LGN-SVC-01</b></td>
      <td>Xác thực thành công (Happy Case - Service)</td>
      <td>
        <ul>
          <li>Khởi tạo <code>LoginCommand</code> hợp lệ.</li>
          <li>Mock <code>AccountPort.findByEmail</code> trả về <code>Optional.of(Account)</code>.</li>
          <li>Mock <code>PasswordEncoder.matches</code> trả về <code>true</code>.</li>
          <li>Mock <code>JwtUtils</code> tạo tokens và hash.</li>
          <li>Mock <code>RefreshTokenPort.save</code> để lưu token thành công.</li>
        </ul>
      </td>
      <td>Gọi hàm <code>loginUseCase.execute(command)</code>.</td>
      <td>
        <ul>
          <li>Không ném ra ngoại lệ.</li>
          <li>Trả về đối tượng <code>LoginResponse</code> chứa tokens chính xác.</li>
          <li>Xác minh <code>refreshTokenPort.save</code> được gọi 1 lần.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-AUTH-LGN-SVC-02</b></td>
      <td>Lỗi tài khoản không tồn tại trong DB</td>
      <td>
        <ul>
          <li>Khởi tạo <code>LoginCommand</code> với email không có trong DB.</li>
          <li>Mock <code>AccountPort.findByEmail</code> trả về <code>Optional.empty()</code>.</li>
        </ul>
      </td>
      <td>Gọi hàm <code>loginUseCase.execute(command)</code>.</td>
      <td>
        <ul>
          <li>Ném ra ngoại lệ <code>RuntimeException</code>.</li>
          <li>Thông điệp ngoại lệ là: <code>"Account not found!"</code>.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-AUTH-LGN-SVC-03</b></td>
      <td>Lỗi sai mật khẩu đăng nhập</td>
      <td>
        <ul>
          <li>Khởi tạo <code>LoginCommand</code> hợp lệ.</li>
          <li>Mock <code>AccountPort.findByEmail</code> trả về <code>Optional.of(Account)</code>.</li>
          <li>Mock <code>PasswordEncoder.matches</code> trả về <code>false</code>.</li>
        </ul>
      </td>
      <td>Gọi hàm <code>loginUseCase.execute(command)</code>.</td>
      <td>
        <ul>
          <li>Ném ra ngoại lệ <code>RuntimeException</code>.</li>
          <li>Thông điệp ngoại lệ là: <code>"Incorrect password!"</code>.</li>
        </ul>
      </td>
    </tr>
  </tbody>
</table>

---

## 3. BẢNG KẾ HOẠCH UNIT TEST CHI TIẾT - LOGOUT

Dưới đây là bảng kế hoạch chi tiết tích hợp cả Controller Layer và Service Layer cho chức năng Logout:

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
      <td><b>TC-AUTH-LGT-CON-01</b></td>
      <td>Đăng xuất thành công (Happy Case - Controller)</td>
      <td>
        <ul>
          <li>Mock <code>LogoutUseCase.execute(any(LogoutCommand.class))</code> trả về <code>MessageResponse("Logged out successfully!")</code>.</li>
          <li>Tạo body request: <code>refreshToken = "valid-refresh-token"</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/auth/logout</code> với body JSON chứa dữ liệu thiết lập.</td>
      <td>
        <ul>
          <li>HTTP Status: <code>200 OK</code></li>
          <li>Response Body chứa:
            <ul>
              <li><code>success = true</code></li>
              <li><code>message = "Logout successful!"</code></li>
              <li><code>data.message = "Logged out successfully!"</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-AUTH-LGT-CON-02</b></td>
      <td>Lỗi validate do thiếu Refresh Token</td>
      <td>
        <ul>
          <li>Tạo body request: <code>refreshToken = ""</code> hoặc <code>null</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/auth/logout</code> với body JSON chứa dữ liệu thiết lập.</td>
      <td>
        <ul>
          <li>HTTP Status: <code>400 Bad Request</code></li>
          <li>Response Body chứa:
            <ul>
              <li><code>success = false</code></li>
              <li><code>message = "Invalid input data"</code></li>
              <li><code>error</code> chứa: <code>"refreshToken": "Refresh token is required"</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-AUTH-LGT-CON-03</b></td>
      <td>Xử lý lỗi nghiệp vụ từ Service</td>
      <td>
        <ul>
          <li>Mock <code>LogoutUseCase.execute(any(LogoutCommand.class))</code> ném ra <code>RuntimeException("Session not found!")</code>.</li>
          <li>Tạo body request hợp lệ.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/auth/logout</code> với body JSON chứa dữ liệu thiết lập.</td>
      <td>
        <ul>
          <li>HTTP Status: <code>400 Bad Request</code></li>
          <li>Response Body chứa:
            <ul>
              <li><code>success = false</code></li>
              <li><code>message = "Business validation failed"</code></li>
              <li><code>error = "Session not found!"</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-AUTH-LGT-SVC-01</b></td>
      <td>Đăng xuất thành công (Happy Case - Service)</td>
      <td>
        <ul>
          <li>Khởi tạo <code>LogoutCommand</code> với token hợp lệ.</li>
          <li>Mock <code>JwtUtils.hashToken</code> trả về chuỗi hash.</li>
          <li>Khởi tạo thực thể <code>RefreshToken</code> với <code>isRevoked = false</code>.</li>
          <li>Mock <code>RefreshTokenPort.findByTokenHash</code> trả về thực thể trên.</li>
          <li>Mock <code>RefreshTokenPort.save</code>.</li>
        </ul>
      </td>
      <td>Gọi hàm <code>logoutUseCase.execute(command)</code>.</td>
      <td>
        <ul>
          <li>Trả về đối tượng <code>MessageResponse("Logged out successfully!")</code>.</li>
          <li>Trạng thái thực thể <code>RefreshToken</code> đổi thành <code>isRevoked = true</code>.</li>
          <li>Xác minh <code>refreshTokenPort.save</code> được gọi để cập nhật DB.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-AUTH-LGT-SVC-02</b></td>
      <td>Lỗi không tìm thấy Token trong DB</td>
      <td>
        <ul>
          <li>Khởi tạo <code>LogoutCommand</code> với token không có trong DB.</li>
          <li>Mock <code>JwtUtils.hashToken</code>.</li>
          <li>Mock <code>RefreshTokenPort.findByTokenHash</code> trả về <code>Optional.empty()</code>.</li>
        </ul>
      </td>
      <td>Gọi hàm <code>logoutUseCase.execute(command)</code>.</td>
      <td>
        <ul>
          <li>Ném ra ngoại lệ <code>RuntimeException</code>.</li>
          <li>Thông điệp ngoại lệ là: <code>"Session not found!"</code>.</li>
        </ul>
      </td>
    </tr>
  </tbody>
</table>
