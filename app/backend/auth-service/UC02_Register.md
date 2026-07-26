# Tài Liệu Bàn Giao Thiết Kế Unit Test - UC-02: Register

Tài liệu này chứa Ma trận dò vết và Bảng kế hoạch Unit Test chi tiết cho chức năng Đăng ký tài khoản (Register) thuộc dịch vụ `auth-service`.

---

## 1. MA TRẬN DÒ VẾT (TRACEABILITY MATRIX) - UC-02

| Mã Yêu Cầu (Req ID) | Nhóm Chức Năng | Use Case (UC) | Use Case Name | Test Case Liên Kết (Test Case ID) | Trạng thái |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **REQ-AUTH-03** | Xác thực người dùng | UC-02 | Register (Đăng ký) | TC-AUTH-REG-CON-01<br>TC-AUTH-REG-CON-02<br>TC-AUTH-REG-CON-03<br>TC-AUTH-REG-CON-04<br>TC-AUTH-REG-CON-05<br>TC-AUTH-REG-CON-06<br>TC-AUTH-REG-CON-07<br>TC-AUTH-REG-SVC-01<br>TC-AUTH-REG-SVC-02<br>TC-AUTH-REG-SVC-03<br>TC-AUTH-REG-SVC-04 | **READY TO RUN** |

---

## 2. BẢNG KẾ HOẠCH UNIT TEST CHI TIẾT - REGISTER

Dưới đây là bảng kế hoạch chi tiết tích hợp cả Controller Layer và Service Layer cho chức năng Register:

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
      <td><b>TC-AUTH-REG-CON-01</b></td>
      <td>Đăng ký tài khoản thành công (Happy Case - Controller)</td>
      <td>
        <ul>
          <li>Mock <code>RegisterUseCase.execute(any(RegisterCommand.class))</code> trả về đối tượng <code>MessageResponse("User registered successfully!")</code>.</li>
          <li>Tạo body request đầy đủ và hợp lệ: <code>fullName = "John Doe"</code>, <code>email = "john@example.com"</code>, <code>password = "Secret123!"</code>, <code>confirmPassword = "Secret123!"</code>, <code>role = Role.PATIENT</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/auth/register</code> với body JSON chứa dữ liệu thiết lập.</td>
      <td>
        <ul>
          <li>HTTP Status: <code>200 OK</code></li>
          <li>Response Body chứa:
            <ul>
              <li><code>success = true</code></li>
              <li><code>message = "User registered successfully!"</code></li>
              <li><code>data.message = "User registered successfully!"</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-AUTH-REG-CON-02</b></td>
      <td>Lỗi validate do để trống họ tên</td>
      <td>
        <ul>
          <li>Tạo body request: <code>fullName = ""</code> hoặc <code>null</code>, các trường khác hợp lệ.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/auth/register</code> với body JSON chứa dữ liệu thiết lập.</td>
      <td>
        <ul>
          <li>HTTP Status: <code>400 Bad Request</code></li>
          <li>Response Body chứa:
            <ul>
              <li><code>success = false</code></li>
              <li><code>message = "Invalid input data"</code></li>
              <li><code>error</code> chứa: <code>"fullName": "Full name is required"</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-AUTH-REG-CON-03</b></td>
      <td>Lỗi validate do để trống email</td>
      <td>
        <ul>
          <li>Tạo body request: <code>email = ""</code> hoặc <code>null</code>, các trường khác hợp lệ.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/auth/register</code> với body JSON chứa dữ liệu thiết lập.</td>
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
      <td><b>TC-AUTH-REG-CON-04</b></td>
      <td>Lỗi validate do email sai định dạng</td>
      <td>
        <ul>
          <li>Tạo body request: <code>email = "wrong-format-email"</code>, các trường khác hợp lệ.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/auth/register</code> với body JSON chứa dữ liệu thiết lập.</td>
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
      <td><b>TC-AUTH-REG-CON-05</b></td>
      <td>Lỗi validate do để trống mật khẩu</td>
      <td>
        <ul>
          <li>Tạo body request: <code>password = ""</code> hoặc <code>null</code>, các trường khác hợp lệ.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/auth/register</code> với body JSON chứa dữ liệu thiết lập.</td>
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
      <td><b>TC-AUTH-REG-CON-06</b></td>
      <td>Lỗi validate do mật khẩu quá ngắn</td>
      <td>
        <ul>
          <li>Tạo body request: <code>password = "12345"</code> (dưới 6 ký tự), các trường khác hợp lệ.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/auth/register</code> với body JSON chứa dữ liệu thiết lập.</td>
      <td>
        <ul>
          <li>HTTP Status: <code>400 Bad Request</code></li>
          <li>Response Body chứa:
            <ul>
              <li><code>success = false</code></li>
              <li><code>message = "Invalid input data"</code></li>
              <li><code>error</code> chứa: <code>"password": "Password must be at least 6 characters"</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-AUTH-REG-CON-07</b></td>
      <td>Lỗi validate do thiếu xác nhận mật khẩu</td>
      <td>
        <ul>
          <li>Tạo body request: <code>confirmPassword = ""</code> hoặc <code>null</code>, các trường khác hợp lệ.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/auth/register</code> với body JSON chứa dữ liệu thiết lập.</td>
      <td>
        <ul>
          <li>HTTP Status: <code>400 Bad Request</code></li>
          <li>Response Body chứa:
            <ul>
              <li><code>success = false</code></li>
              <li><code>message = "Invalid input data"</code></li>
              <li><code>error</code> chứa: <code>"confirmPassword": "Confirmation password is required"</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-AUTH-REG-SVC-01</b></td>
      <td>Đăng ký thành công với vai trò tùy chọn (Service)</td>
      <td>
        <ul>
          <li>Khởi tạo <code>RegisterCommand</code> đầy đủ thông tin: <code>role = Role.EXPERT</code>.</li>
          <li>Mock <code>AccountPort.existsByEmail</code> trả về <code>false</code>.</li>
          <li>Mock <code>PasswordEncoder.encode</code>.</li>
          <li>Mock <code>AccountPort.save</code>.</li>
        </ul>
      </td>
      <td>Gọi hàm <code>registerUseCase.execute(command)</code>.</td>
      <td>
        <ul>
          <li>Không ném ra ngoại lệ.</li>
          <li>Trả về <code>MessageResponse("User registered successfully!")</code>.</li>
          <li>Xác minh thực thể <code>Account</code> truyền vào <code>save()</code> có:
            <ul>
              <li><code>fullName = "John Doe"</code></li>
              <li><code>email = "john@example.com"</code></li>
              <li><code>role = Role.EXPERT</code></li>
              <li><code>isEmailVerified = true</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-AUTH-REG-SVC-02</b></td>
      <td>Đăng ký thành công dùng vai trò mặc định</td>
      <td>
        <ul>
          <li>Khởi tạo <code>RegisterCommand</code> với <code>role = null</code>.</li>
          <li>Mock <code>AccountPort.existsByEmail</code> trả về <code>false</code>.</li>
          <li>Mock <code>PasswordEncoder.encode</code>.</li>
          <li>Mock <code>AccountPort.save</code>.</li>
        </ul>
      </td>
      <td>Gọi hàm <code>registerUseCase.execute(command)</code>.</td>
      <td>
        <ul>
          <li>Không ném ra ngoại lệ.</li>
          <li>Trả về <code>MessageResponse("User registered successfully!")</code>.</li>
          <li>Thực thể <code>Account</code> được lưu có <code>role = Role.PATIENT</code> (mặc định).</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-AUTH-REG-SVC-03</b></td>
      <td>Lỗi mật khẩu xác nhận không trùng khớp</td>
      <td>
        <ul>
          <li>Khởi tạo <code>RegisterCommand</code> có <code>password = "Secret123!"</code> và <code>confirmPassword = "DifferentPassword123!"</code>.</li>
        </ul>
      </td>
      <td>Gọi hàm <code>registerUseCase.execute(command)</code>.</td>
      <td>
        <ul>
          <li>Ném ra ngoại lệ <code>RuntimeException</code>.</li>
          <li>Thông điệp ngoại lệ là: <code>"Confirmation password does not match!"</code>.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-AUTH-REG-SVC-04</b></td>
      <td>Lỗi email đăng ký đã tồn tại trong DB</td>
      <td>
        <ul>
          <li>Khởi tạo <code>RegisterCommand</code> với email trùng lặp.</li>
          <li>Mock <code>AccountPort.existsByEmail</code> trả về <code>true</code>.</li>
        </ul>
      </td>
      <td>Gọi hàm <code>registerUseCase.execute(command)</code>.</td>
      <td>
        <ul>
          <li>Ném ra ngoại lệ <code>RuntimeException</code>.</li>
          <li>Thông điệp ngoại lệ là: <code>"Email is already registered!"</code>.</li>
        </ul>
      </td>
    </tr>
  </tbody>
</table>
