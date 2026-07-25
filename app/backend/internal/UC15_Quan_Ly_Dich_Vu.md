# Tài Liệu Bàn Giao Thiết Kế Unit Test - UC-15: Quản Lý Dịch Vụ Tư Vấn

Tài liệu này chứa Ma trận dò vết và Bảng kế hoạch Unit Test chi tiết bằng Golang cho chức năng Quản lý danh mục dịch vụ tư vấn của hệ thống (Thêm mới, chỉnh sửa, vô hiệu hóa dịch vụ tư vấn, kiểm tra định dạng giá tiền và tên dịch vụ).

---

## 1. MA TRẬN DÒ VẾT (TRACEABILITY MATRIX) - UC-15

| Mã Yêu Cầu (Req ID) | Nhóm Chức Năng | Use Case (UC) | Use Case Name | Test Case Liên Kết (Test Case ID) | Trạng thái |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **REQ-SVC-01** | Tạo & cập nhật dịch vụ tư vấn | UC-15 | Quản lý dịch vụ | TC-SVC-01<br>TC-SVC-02<br>TC-SVC-03 | **READY TO RUN** |
| **REQ-SVC-02** | Xóa & vô hiệu hóa dịch vụ | UC-15 | Quản lý dịch vụ | TC-SVC-04<br>TC-SVC-05 | **READY TO RUN** |

---

## 2. BẢNG KẾ HOẠCH UNIT TEST CHI TIẾT - UC-15

Dưới đây là bảng kế hoạch chi tiết kiểm thử tầng Service Management Manager & Handler trong Golang:

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
      <td><b>TC-SVC-01</b></td>
      <td>Tạo dịch vụ tư vấn mới thành công (Happy Case - Create Service)</td>
      <td>
        <ul>
          <li>Header Admin: <code>X-User-Role = "ADMIN"</code>.</li>
          <li>Payload JSON: <code>name = "Tư vấn tâm lý trực tuyến 1:1"</code>, <code>price = 300000</code> (VNĐ), <code>duration = 60</code> (phút), <code>description = "Gặp gỡ chuyên gia qua video call"</code>.</li>
          <li>Mock Repository: Tạo mới dịch vụ gán ID = 1.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/services</code> với dữ liệu hợp lệ.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>201 Created</code></li>
          <li>Response JSON chứa <code>success = true</code> và thông tin dịch vụ vừa tạo.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-SVC-02</b></td>
      <td>Tạo dịch vụ thất bại do tên dịch vụ bị trống hoặc giá tiền âm (Validation Error)</td>
      <td>
        <ul>
          <li>Header Admin: <code>X-User-Role = "ADMIN"</code>.</li>
          <li>Payload JSON: <code>name = ""</code> (trống) hoặc <code>price = -50000</code> (giá âm).</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/services</code> với dữ liệu lỗi.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>400 Bad Request</code></li>
          <li>Response JSON trả về lỗi validation: <code>"Tên dịch vụ không được để trống và giá tiền phải lớn hơn 0"</code>.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-SVC-03</b></td>
      <td>Cập nhật giá và mô tả dịch vụ tư vấn thành công (Update Service)</td>
      <td>
        <ul>
          <li>CSDL Mock: Đã tồn tại dịch vụ ID = 1 (giá cũ 300.000 VNĐ).</li>
          <li>Header Admin: <code>X-User-Role = "ADMIN"</code>.</li>
          <li>Payload JSON: <code>price = 350000</code>, <code>description = "Mô tả mới đã cập nhật"</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>PUT /api/v1/services/1</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code></li>
          <li>Response JSON chứa thông tin dịch vụ với mức giá 350.000 VNĐ.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-SVC-04</b></td>
      <td>Vô hiệu hóa (Soft Delete) dịch vụ tư vấn thành công (Disable Service)</td>
      <td>
        <ul>
          <li>CSDL Mock: Dịch vụ ID = 1 đang hoạt động (<code>IsActive = true</code>).</li>
          <li>Header Admin: <code>X-User-Role = "ADMIN"</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>DELETE /api/v1/services/1</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>204 No Content</code></li>
          <li>Trạng thái dịch vụ chuyển thành <code>IsActive = false</code>.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-SVC-05</b></td>
      <td>Cập nhật dịch vụ thất bại do dịch vụ không tồn tại (Service Not Found)</td>
      <td>
        <ul>
          <li>Header Admin: <code>X-User-Role = "ADMIN"</code>.</li>
          <li>Path parameter: <code>id = 99999</code> (không tồn tại trong CSDL).</li>
        </ul>
      </td>
      <td>Gửi request <code>PUT /api/v1/services/99999</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>404 Not Found</code></li>
          <li>Response JSON thông báo: <code>"Dịch vụ tư vấn không tồn tại!"</code>.</li>
        </ul>
      </td>
    </tr>
  </tbody>
</table>
