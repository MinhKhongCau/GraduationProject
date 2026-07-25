# Tài Liệu Bàn Giao Thiết Kế Unit Test - UC-18: Quản Lý Bài Trắc Nghiệm

Tài liệu này chứa Ma trận dò vết và Bảng kế hoạch Unit Test chi tiết cho chức năng Quản lý bài trắc nghiệm dành cho Admin / Chuyên gia (Tạo, sửa, xóa bài test và thêm hàng loạt câu hỏi) thuộc dịch vụ `assessment-service`.

---

## 1. MA TRẬN DÒ VẾT (TRACEABILITY MATRIX) - UC-18

| Mã Yêu Cầu (Req ID) | Nhóm Chức Năng | Use Case (UC) | Use Case Name | Test Case Liên Kết (Test Case ID) | Trạng thái |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **REQ-ASM-03** | Quản lý bộ đề bài test | UC-18 | Quản lý bài trắc nghiệm | TC-ASM-TPL-01<br>TC-ASM-TPL-02<br>TC-ASM-TPL-03<br>TC-ASM-TPL-04<br>TC-ASM-TPL-05 | **READY TO RUN** |
| **REQ-ASM-04** | Quản lý danh sách câu hỏi | UC-18 | Quản lý bài trắc nghiệm | TC-ASM-QST-01<br>TC-ASM-QST-02<br>TC-ASM-QST-03<br>TC-ASM-QST-04 | **READY TO RUN** |

---

## 2. BẢNG KẾ HOẠCH UNIT TEST CHI TIẾT - UC-18

Dưới đây là bảng kế hoạch chi tiết kiểm thử các API Endpoint quản lý mẫu bài test (`/api/v1/assessments/templates`) và quản lý bộ câu hỏi (`/api/v1/assessments/questions`):

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
      <td><b>TC-ASM-TPL-01</b></td>
      <td>Tạo bài trắc nghiệm mới thành công (Happy Case - Create Template)</td>
      <td>
        <ul>
          <li>Mã bài test `code = "DASS21_TEST"` chưa tồn tại trong CSDL.</li>
          <li>Body chứa `title`, `description`, `instruction`, `certification`.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/assessments/templates</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code></li>
          <li>Response JSON chứa `message`: <code>"Tạo bài test thành công!"</code> và đối tượng `data` kèm chuỗi `slug` duy nhất 10 ký tự.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-ASM-TPL-02</b></td>
      <td>Tạo bài trắc nghiệm thất bại do trùng lặp mã bài test (Duplicate Code)</td>
      <td>
        <ul>
          <li>Mã bài test `code = "DASS21"` đã tồn tại và đang hoạt động (`is_active = True`).</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/assessments/templates</code> với `code = "DASS21"`.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>400 Bad Request</code></li>
          <li>Response JSON chứa `detail`: <code>"Lỗi: Mã bài test này đã tồn tại!"</code>.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-ASM-TPL-03</b></td>
      <td>Xem thông tin chi tiết bài trắc nghiệm theo slug thành công (Get Template)</td>
      <td>
        <ul>
          <li>Bài test có `slug = "dass21-slug"` tồn tại trong CSDL.</li>
        </ul>
      </td>
      <td>Gửi request <code>GET /api/v1/assessments/templates/dass21-slug</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code></li>
          <li>Trả về đầy đủ thông tin tiêu đề, mô tả và hướng dẫn bài test.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-ASM-TPL-04</b></td>
      <td>Cập nhật thông tin bài trắc nghiệm thành công (Update Template)</td>
      <td>
        <ul>
          <li>Bài test tồn tại. Payload chứa tiêu đề mới `title = "Bộ trắc nghiệm DASS-21 Chuẩn"`.</li>
        </ul>
      </td>
      <td>Gửi request <code>PATCH /api/v1/assessments/templates/dass21-slug</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code></li>
          <li>Dữ liệu bài test được cập nhật tiêu đề mới trong CSDL.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-ASM-TPL-05</b></td>
      <td>Xóa bài trắc nghiệm thành công bằng phương pháp Soft Delete (Delete Template)</td>
      <td>
        <ul>
          <li>Header `X-User-Role = "ADMIN"`. Bài test có `slug = "dass21-slug"`.</li>
        </ul>
      </td>
      <td>Gửi request <code>DELETE /api/v1/assessments/templates/dass21-slug</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code></li>
          <li>Trường `is_active` của bài test chuyển thành `False` trong CSDL.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-ASM-QST-01</b></td>
      <td>Thêm hàng loạt câu hỏi vào bài trắc nghiệm thành công (Create Bulk Questions)</td>
      <td>
        <ul>
          <li>Bài test (`template_id`), Nhóm phương án (`group_id`) và Khía cạnh (`dimension_id`) hợp lệ.</li>
          <li>Payload chứa mảng 2 câu hỏi mới.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/assessments/questions/bulk</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code></li>
          <li>Response JSON thông báo: <code>"Đã thêm thành công 2 câu hỏi vào bài test!"</code>.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-ASM-QST-02</b></td>
      <td>Thêm hàng loạt câu hỏi thất bại do khía cạnh (dimension) không tồn tại</td>
      <td>
        <ul>
          <li>`dimension_id = "non-existent-dim"` không có trong CSDL.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/assessments/questions/bulk</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>404 Not Found</code></li>
          <li>Response JSON thông báo không tìm thấy khía cạnh tương ứng.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-ASM-QST-03</b></td>
      <td>Lấy danh sách câu hỏi kèm nhóm phương án lựa chọn của một bài test thành công</td>
      <td>
        <ul>
          <li>Bài test `slug = "dass21-slug"` chứa 21 câu hỏi.</li>
        </ul>
      </td>
      <td>Gửi request <code>GET /api/v1/assessments/templates/dass21-slug/questions</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code></li>
          <li>Trả về mảng câu hỏi đã sắp xếp theo `question_order` tăng dần, mỗi câu hỏi đi kèm danh sách phương án lựa chọn (`options`).</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-ASM-QST-04</b></td>
      <td>Xóa câu hỏi thành công bằng Soft Delete (Delete Question)</td>
      <td>
        <ul>
          <li>Header `X-User-Role = "ADMIN"`. Câu hỏi `slug = "question-slug-1"`.</li>
        </ul>
      </td>
      <td>Gửi request <code>DELETE /api/v1/assessments/questions/question-slug-1</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code></li>
          <li>`is_active` của câu hỏi chuyển thành `False`.</li>
        </ul>
      </td>
    </tr>
  </tbody>
</table>
