# Tài Liệu Bàn Giao Thiết Kế Unit Test - UC-13: Hỏi Đáp Cộng Đồng

Tài liệu này chứa Ma trận dò vết và Bảng kế hoạch Unit Test chi tiết bằng Golang cho chức năng Hỏi đáp cộng đồng (Đăng câu hỏi, lọc danh sách bài viết theo danh mục, thẻ tag và tìm kiếm) thuộc dịch vụ `forum-service`.

---

## 1. MA TRẬN DÒ VẾT (TRACEABILITY MATRIX) - UC-13

| Mã Yêu Cầu (Req ID) | Nhóm Chức Năng | Use Case (UC) | Use Case Name | Test Case Liên Kết (Test Case ID) | Trạng thái |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **REQ-FORUM-01** | Khởi tạo bài thảo luận / câu hỏi | UC-13 | Hỏi đáp cộng đồng | TC-FORUM-QNA-01<br>TC-FORUM-QNA-02<br>TC-FORUM-QNA-03<br>TC-FORUM-QNA-04 | **READY TO RUN** |
| **REQ-FORUM-02** | Duyệt & tìm kiếm bài viết | UC-13 | Hỏi đáp cộng đồng | TC-FORUM-QNA-05<br>TC-FORUM-QNA-06 | **READY TO RUN** |

---

## 2. BẢNG KẾ HOẠCH UNIT TEST CHI TIẾT - UC-13

Dưới đây là bảng kế hoạch chi tiết kiểm thử tầng HTTP Handler (Gin context & httptest) cho các API Đăng bài và Xem danh sách bài viết (`POST /api/v1/forum/posts` & `GET /api/v1/forum/posts`):

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
      <td><b>TC-FORUM-QNA-01</b></td>
      <td>Đăng câu hỏi / bài viết mới thành công (Happy Case - Create Post)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Id = "user-uuid-1"</code>.</li>
          <li>Body request JSON: <code>categoryId = 1</code>, <code>title = "Làm sao để vượt qua căng thẳng học đường?"</code>, <code>content = "Nội dung chi tiết câu hỏi..."</code>, <code>tags = ["tam-ly", "hoc-duong"]</code>.</li>
          <li>Mock Service: Trả về đối tượng `Post` vừa tạo kèm slug tự động sinh.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/forum/posts</code> với payload hợp lệ.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>201 Created</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường `success = true`</li>
              <li>Thông báo: <code>"Post created"</code></li>
              <li>Data chứa thông tin bài viết, slug và danh sách thẻ tag.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-FORUM-QNA-02</b></td>
      <td>Đăng bài viết thất bại do thiếu tiêu đề hoặc nội dung (Validation Error)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Id = "user-uuid-1"</code>.</li>
          <li>Body request JSON rỗng: <code>{"categoryId": 1, "title": "", "content": ""}</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/forum/posts</code> với tiêu đề và nội dung rỗng.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>400 Bad Request</code></li>
          <li>Response JSON chứa `success = false` và chi tiết lỗi binding validation.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-FORUM-QNA-03</b></td>
      <td>Đăng bài viết thất bại do danh mục (category) không tồn tại (Category Not Found)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Id = "user-uuid-1"</code>.</li>
          <li>Body request JSON: <code>categoryId = 99999</code> (không tồn tại trong CSDL).</li>
          <li>Mock Service: Trả về lỗi `service.ErrNotFound`.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/forum/posts</code> với danh mục không hợp lệ.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>400 Bad Request</code></li>
          <li>Response JSON chứa thông báo lỗi: <code>"Category not found"</code>.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-FORUM-QNA-04</b></td>
      <td>Đăng bài viết thất bại do thiếu header định danh người dùng (Unauthorized)</td>
      <td>
        <ul>
          <li>Thiếu header `X-User-Id` (rỗng/chưa đăng nhập).</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/forum/posts</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>401 Unauthorized</code></li>
          <li>Response JSON thông báo người dùng chưa được xác thực.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-FORUM-QNA-05</b></td>
      <td>Truy vấn danh sách bài viết theo danh mục và từ khóa tìm kiếm thành công (List Posts)</td>
      <td>
        <ul>
          <li>Query Parameters: <code>categoryId = 1</code>, <code>search = "căng thẳng"</code>, <code>page = 1</code>, <code>pageSize = 10</code>.</li>
          <li>Mock Service: Trả về danh sách 2 bài viết thảo luận phù hợp.</li>
        </ul>
      </td>
      <td>Gửi request <code>GET /api/v1/forum/posts?categoryId=1&amp;search=c%C4%83ng%20th%E1%BA%B3ng</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code></li>
          <li>Response JSON chứa mảng `items` gồm các bài viết thảo luận công khai (`PUBLISHED`) và tổng số lượng phân trang.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-FORUM-QNA-06</b></td>
      <td>Xem chi tiết bài viết thảo luận theo slug thành công (Get Post Detail)</td>
      <td>
        <ul>
          <li>Path parameter: <code>id = "lam-sao-de-vuot-qua-cang-thang"</code>.</li>
          <li>Mock Service: Trả về đối tượng `Post` kèm nội dung đầy đủ và tự động tăng `viewCount`.</li>
        </ul>
      </td>
      <td>Gửi request <code>GET /api/v1/forum/posts/lam-sao-de-vuot-qua-cang-thang</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code></li>
          <li>Response JSON chứa đầy đủ chi tiết bài viết, tác giả và lượt xem.</li>
        </ul>
      </td>
    </tr>
  </tbody>
</table>
