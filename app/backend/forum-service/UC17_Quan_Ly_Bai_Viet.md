# Tài Liệu Bàn Giao Thiết Kế Unit Test - UC-17: Quản Lý Bài Viết & Bình Luận

Tài liệu này chứa Ma trận dò vết và Bảng kế hoạch Unit Test chi tiết bằng Golang cho chức năng Quản lý bài viết và bình luận (Chỉnh sửa/Xóa bài viết cá nhân, Thêm/Sửa/Xóa bình luận thảo luận) thuộc dịch vụ `forum-service`.

---

## 1. MA TRẬN DÒ VẾT (TRACEABILITY MATRIX) - UC-17

| Mã Yêu Cầu (Req ID) | Nhóm Chức Năng | Use Case (UC) | Use Case Name | Test Case Liên Kết (Test Case ID) | Trạng thái |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **REQ-FORUM-03** | Cập nhật & xóa bài viết | UC-17 | Quản lý bài viết | TC-FORUM-MNG-01<br>TC-FORUM-MNG-02<br>TC-FORUM-MNG-03<br>TC-FORUM-MNG-04 | **READY TO RUN** |
| **REQ-FORUM-04** | Quản lý bình luận | UC-17 | Quản lý bài viết | TC-FORUM-MNG-05<br>TC-FORUM-MNG-06<br>TC-FORUM-MNG-07<br>TC-FORUM-MNG-08 | **READY TO RUN** |

---

## 2. BẢNG KẾ HOẠCH UNIT TEST CHI TIẾT - UC-17

Dưới đây là bảng kế hoạch chi tiết kiểm thử các API Cập nhật/Xóa bài viết (`PUT/DELETE /api/v1/forum/posts/:id`) và Bình luận (`POST/PUT/DELETE /api/v1/forum/comments`):

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
      <td><b>TC-FORUM-MNG-01</b></td>
      <td>Chủ bài viết chỉnh sửa nội dung bài viết thành công (Happy Case - Update Post)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Id = "user-uuid-1"</code> (là tác giả bài viết ID = 10).</li>
          <li>Body request JSON: <code>title = "Tiêu đề bài viết đã cập nhật"</code>, <code>content = "Nội dung mới..."</code>.</li>
          <li>Mock Service: Trả về đối tượng `Post` đã cập nhật.</li>
        </ul>
      </td>
      <td>Gửi request <code>PUT /api/v1/forum/posts/10</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường `success = true`</li>
              <li>Thông báo: <code>"Post updated"</code></li>
              <li>Data chứa tiêu đề và nội dung mới.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-FORUM-MNG-02</b></td>
      <td>Từ chối chỉnh sửa bài viết của người khác (Forbidden Update)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Id = "other-user-uuid"</code> (không phải tác giả và không phải Admin).</li>
          <li>Mock Service: Trả về lỗi `service.ErrForbidden`.</li>
        </ul>
      </td>
      <td>Gửi request <code>PUT /api/v1/forum/posts/10</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>403 Forbidden</code></li>
          <li>Response JSON chứa thông báo: <code>"You do not have permission to edit or delete this post"</code>.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-FORUM-MNG-03</b></td>
      <td>Xóa bài viết cá nhân thành công bằng phương pháp Soft Delete (Delete Post)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Id = "user-uuid-1"</code> (chủ bài viết ID = 10).</li>
          <li>Mock Service: Trả về `nil` (xóa thành công).</li>
        </ul>
      </td>
      <td>Gửi request <code>DELETE /api/v1/forum/posts/10</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>204 No Content</code></li>
          <li>Bài viết được đánh dấu xóa trong CSDL (Soft Delete).</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-FORUM-MNG-04</b></td>
      <td>Xóa bài viết thất bại do bài viết không tồn tại (Post Not Found)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Id = "user-uuid-1"</code>.</li>
          <li>Path parameter: <code>id = 99999</code>.</li>
          <li>Mock Service: Trả về `service.ErrNotFound`.</li>
        </ul>
      </td>
      <td>Gửi request <code>DELETE /api/v1/forum/posts/99999</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>404 Not Found</code></li>
          <li>Response JSON chứa thông báo: <code>"Post not found"</code>.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-FORUM-MNG-05</b></td>
      <td>Người dùng gửi bình luận vào bài viết thành công (Create Comment)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Id = "user-uuid-1"</code>.</li>
          <li>Path parameter: <code>id = 10</code> (bài viết tồn tại).</li>
          <li>Body request JSON: <code>content = "Cảm ơn tác giả về bài viết rất hữu ích!"</code>.</li>
          <li>Mock Service: Tạo thành công bình luận mới và tăng `commentCount`.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/forum/posts/10/comments</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>201 Created</code></li>
          <li>Response JSON chứa `success = true`, thông báo <code>"Comment created"</code> và dữ liệu bình luận vừa tạo.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-FORUM-MNG-06</b></td>
      <td>Gửi bình luận thất bại do nội dung bình luận rỗng (Validation Error)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Id = "user-uuid-1"</code>.</li>
          <li>Body request JSON: <code>content = ""</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/forum/posts/10/comments</code> với nội dung rỗng.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>400 Bad Request</code></li>
          <li>Response JSON thông báo lỗi binding validation.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-FORUM-MNG-07</b></td>
      <td>Chỉnh sửa nội dung bình luận cá nhân thành công (Edit Comment)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Id = "user-uuid-1"</code> (tác giả bình luận ID = 50).</li>
          <li>Body request JSON: <code>content = "Nội dung bình luận đã được chỉnh sửa."</code>.</li>
        </ul>
      </td>
      <td>Gửi request <code>PUT /api/v1/forum/comments/50</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code></li>
          <li>Response JSON chứa thông báo: <code>"Comment updated"</code> và nội dung mới.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-FORUM-MNG-08</b></td>
      <td>Xóa bình luận cá nhân thành công theo cơ chế hiển thị Placeholder (Soft Delete Comment)</td>
      <td>
        <ul>
          <li>Header request: <code>X-User-Id = "user-uuid-1"</code> (tác giả bình luận ID = 50).</li>
        </ul>
      </td>
      <td>Gửi request <code>DELETE /api/v1/forum/comments/50</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>204 No Content</code></li>
          <li>Bình luận đánh dấu `deleted = true`, nội dung được ẩn để duy trì cây phản hồi (`replies`).</li>
        </ul>
      </td>
    </tr>
  </tbody>
</table>
