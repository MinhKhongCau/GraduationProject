# Tài Liệu Bàn Giao Thiết Kế Unit Test - UC-04: Quản Lý Hồ Sơ

Tài liệu này chứa Ma trận dò vết và Bảng kế hoạch Unit Test chi tiết bằng Golang cho chức năng Quản lý hồ sơ (Get/Update Profile) thuộc dịch vụ `profile-service`.

---

## 1. MA TRẬN DÒ VẾT (TRACEABILITY MATRIX) - UC-04

| Mã Yêu Cầu (Req ID) | Nhóm Chức Năng | Use Case (UC) | Use Case Name | Test Case Liên Kết (Test Case ID) | Trạng thái |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **REQ-PROF-02** | Quản lý hồ sơ cá nhân | UC-04 | Quản lý hồ sơ | TC-PROF-MNG-01<br>TC-PROF-MNG-02<br>TC-PROF-MNG-03<br>TC-PROF-MNG-04<br>TC-PROF-MNG-05 | **READY TO RUN** |

---

## 2. BẢNG KẾ HOẠCH UNIT TEST CHI TIẾT - UC-04

Dưới đây là bảng kế hoạch chi tiết tích hợp cả tầng Handler (httptest & Gin context setup) và tầng Database (sqlmock) cho chức năng Quản lý hồ sơ cá nhân:

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
      <td><b>TC-PROF-MNG-01</b></td>
      <td>Tạo profile nội bộ thành công (Happy Case)</td>
      <td>
        <ul>
          <li>Body request hợp lệ chứa: <code>auth_id</code> (UUID mới), <code>name = "Patient A"</code>, <code>role = "PATIENT"</code>, <code>email = "patient@example.com"</code>.</li>
          <li>Giả lập sqlmock:
            <ul>
              <li>Query kiểm tra trùng lặp (First): Trả về <code>gorm.ErrRecordNotFound</code> (chưa tồn tại profile).</li>
              <li>Transaction (Insert Profile + Insert PatientProfile): Thực hiện INSERT thành công và lưu vào DB.</li>
            </ul>
          </li>
        </ul>
      </td>
      <td>Khởi tạo request <code>POST /internal/api/v1/profiles/create</code> với body JSON hợp lệ, gọi hàm <code>handlers.CreateProfileInternal</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>201 Created</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = true</code></li>
              <li>Thông tin profile mới được tạo bao gồm <code>auth_id</code>, <code>name</code>, và thực thể con <code>PatientProfile</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-PROF-MNG-02</b></td>
      <td>Tạo profile nội bộ thất bại do tài khoản đã có profile (Conflict)</td>
      <td>
        <ul>
          <li>Body request chứa <code>auth_id</code> trùng lặp.</li>
          <li>Giả lập sqlmock:
            <ul>
              <li>Query kiểm tra trùng lặp (First): Tìm thấy profile đang tồn tại (không trả về lỗi).</li>
            </ul>
          </li>
        </ul>
      </td>
      <td>Khởi tạo request <code>POST /internal/api/v1/profiles/create</code> với body JSON, gọi hàm <code>handlers.CreateProfileInternal</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>409 Conflict</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = false</code></li>
              <li>Thông báo lỗi <code>message = "Profile cho tài khoản này đã tồn tại"</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-PROF-MNG-03</b></td>
      <td>Xem profile cá nhân thành công (Happy Case)</td>
      <td>
        <ul>
          <li>Gán giá trị auth account ID hợp lệ vào Gin Context (giả lập từ JWT middleware): <code>CtxAuthID = "valid-user-uuid"</code>.</li>
          <li>Giả lập sqlmock:
            <ul>
              <li>Query tìm profile (First): Trả về đúng thông tin Profile tương ứng (kèm theo preload sub-profiles).</li>
            </ul>
          </li>
        </ul>
      </td>
      <td>Khởi tạo request <code>GET /api/v1/profiles/me</code>, thiết lập context <code>auth_id</code> và gọi hàm <code>handlers.GetMe</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = true</code></li>
              <li>Trường <code>data</code> chứa thông tin chi tiết hồ sơ cá nhân của tài khoản đang đăng nhập.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-PROF-MNG-04</b></td>
      <td>Cập nhật hồ sơ chuyên gia thành công (PUT /me)</td>
      <td>
        <ul>
          <li>Gán <code>CtxAuthID = "valid-expert-uuid"</code> vào Gin Context.</li>
          <li>Body request hợp lệ chứa <code>UpsertExpertRequest</code>: <code>name = "Dr. New Name"</code>, <code>email = "expert-new@example.com"</code>, <code>specialization_ids = ["spec-1", "spec-2"]</code>.</li>
          <li>Giả lập sqlmock:
            <ul>
              <li>Query tìm profile cũ (First): Trả về thực thể <code>Profile</code> với <code>role = EXPERT</code>.</li>
              <li>Transaction cập nhật:
                <ul>
                  <li>Update cột <code>name</code> ở bảng <code>profiles</code>.</li>
                  <li>Save thực thể mới vào bảng <code>expert_profiles</code>.</li>
                  <li>Đồng bộ mối quan hệ nhiều-nhiều (Replace association) ở bảng liên kết chuyên khoa.</li>
                </ul>
              </li>
            </ul>
          </li>
        </ul>
      </td>
      <td>Khởi tạo request <code>PUT /api/v1/profiles/me</code> với body JSON hợp lệ, gọi hàm <code>handlers.UpdateMe</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = true</code></li>
              <li>Hồ sơ chuyên gia sau khi cập nhật thành công (các trường <code>name</code>, <code>email</code>, và danh sách <code>specializations</code> mới).</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-PROF-MNG-05</b></td>
      <td>Cập nhật hồ sơ thất bại do dữ liệu validation không hợp lệ (Email sai định dạng)</td>
      <td>
        <ul>
          <li>Gán <code>CtxAuthID = "valid-expert-uuid"</code> vào Gin Context.</li>
          <li>Body request chứa email sai cấu trúc: <code>email = "wrong-format"</code>.</li>
        </ul>
      </td>
      <td>Khởi tạo request <code>PUT /api/v1/profiles/me</code> với body JSON chứa dữ liệu thiết lập, gọi hàm <code>handlers.UpdateMe</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>400 Bad Request</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = false</code></li>
              <li>Thông báo lỗi validate <code>message = "Dữ liệu không hợp lệ"</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
  </tbody>
</table>
