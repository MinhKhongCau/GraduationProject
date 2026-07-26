# Tài Liệu Bàn Giao Thiết Kế Unit Test - UC-16: Quản Lý Người Dùng

Tài liệu này chứa Ma trận dò vết và Bảng kế hoạch Unit Test chi tiết bằng Golang cho chức năng Admin quản lý danh sách tài khoản người dùng (Khóa/kích hoạt tài khoản, phân quyền vai trò người dùng, xử lý tài khoản không tồn tại).

---

## 1. MA TRẬN DÒ VẾT (TRACEABILITY MATRIX) - UC-16

| Mã Yêu Cầu (Req ID) | Nhóm Chức Năng | Use Case (UC) | Use Case Name | Test Case Liên Kết (Test Case ID) | Trạng thái |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **REQ-ADM-01** | Quản lý trạng thái tài khoản | UC-16 | Quản lý người dùng | TC-ADM-01<br>TC-ADM-02<br>TC-ADM-03 | **READY TO RUN** |
| **REQ-ADM-02** | Cập nhật phân quyền người dùng | UC-16 | Quản lý người dùng | TC-ADM-04<br>TC-ADM-05 | **READY TO RUN** |

---

## 2. BẢNG KẾ HOẠCH UNIT TEST CHI TIẾT - UC-16

Dưới đây là bảng kế hoạch chi tiết kiểm thử tầng User Management Admin Manager & Handler trong Golang:

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
      <td><b>TC-ADM-01</b></td>
      <td>Admin thực hiện khóa tài khoản vi phạm thành công (Lock Account)</td>
      <td>
        <ul>
          <li>Header Admin: <code>X-User-Role = "ADMIN"</code>.</li>
          <li>Target Account: <code>UserID = "user-violator-uuid"</code>, <code>IsBlocked = false</code>.</li>
          <li>Payload JSON: <code>reason = "Spam nội dung vi phạm tiêu chuẩn cộng đồng"</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/admin/users/user-violator-uuid/block</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code></li>
          <li>Tài khoản đổi trạng thái <code>IsBlocked = true</code>, revoked toàn bộ JWT tokens hiện tại.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-ADM-02</b></td>
      <td>Admin thực hiện mở khóa tài khoản thành công (Unlock Account)</td>
      <td>
        <ul>
          <li>Header Admin: <code>X-User-Role = "ADMIN"</code>.</li>
          <li>Target Account: <code>UserID = "user-blocked-uuid"</code>, <code>IsBlocked = true</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/admin/users/user-blocked-uuid/unblock</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code></li>
          <li>Tài khoản chuyển lại trạng thái hoạt động <code>IsBlocked = false</code>.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-ADM-03</b></td>
      <td>Thao tác khóa/mở khóa thất bại do tài khoản không tồn tại (Account Not Found)</td>
      <td>
        <ul>
          <li>Header Admin: <code>X-User-Role = "ADMIN"</code>.</li>
          <li>Target Account: <code>UserID = "non-existent-uuid"</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/admin/users/non-existent-uuid/block</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>404 Not Found</code></li>
          <li>Response JSON thông báo: <code>"Tài khoản người dùng không tồn tại!"</code>.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-ADM-04</b></td>
      <td>Admin nâng cấp phân quyền vai trò cho người dùng thành công (Update Role)</td>
      <td>
        <ul>
          <li>Header Admin: <code>X-User-Role = "ADMIN"</code>.</li>
          <li>Target Account: <code>UserID = "user-promoted-uuid"</code>, role hiện tại: <code>"PATIENT"</code>.</li>
          <li>Payload JSON: <code>role = "EXPERT"</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>PUT /api/v1/admin/users/user-promoted-uuid/role</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code></li>
          <li>Vai trò tài khoản cập nhật thành <code>"EXPERT"</code>.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-ADM-05</b></td>
      <td>Từ chối thao tác quản lý người dùng nếu người thực hiện không phải Admin (Forbidden Access)</td>
      <td>
        <ul>
          <li>Header User: <code>X-User-Role = "PATIENT"</code> (không phải Admin).</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/admin/users/some-uuid/block</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>403 Forbidden</code></li>
          <li>Response JSON thông báo: <code>"Bạn không có quyền thực hiện thao tác này!"</code>.</li>
        </ul>
      </td>
    </tr>
  </tbody>
</table>
