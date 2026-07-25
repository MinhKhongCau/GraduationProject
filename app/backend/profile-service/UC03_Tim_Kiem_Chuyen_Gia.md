# Tài Liệu Bàn Giao Thiết Kế Unit Test - UC-03: Tìm Kiếm Chuyên Gia

Tài liệu này chứa Ma trận dò vết và Bảng kế hoạch Unit Test chi tiết bằng Golang cho chức năng Tìm kiếm chuyên gia (List & Get Experts) thuộc dịch vụ `profile-service`.

---

## 1. MA TRẬN DÒ VẾT (TRACEABILITY MATRIX) - UC-03

| Mã Yêu Cầu (Req ID) | Nhóm Chức Năng | Use Case (UC) | Use Case Name | Test Case Liên Kết (Test Case ID) | Trạng thái |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **REQ-PROF-01** | Tìm kiếm chuyên gia | UC-03 | Tìm kiếm chuyên gia | TC-PROF-EXP-01<br>TC-PROF-EXP-02<br>TC-PROF-EXP-03<br>TC-PROF-EXP-04<br>TC-PROF-EXP-05 | **READY TO RUN** |

---

## 2. BẢNG KẾ HOẠCH UNIT TEST CHI TIẾT - UC-03

Dưới đây là bảng kế hoạch chi tiết tích hợp cả tầng Handler (httptest) và tầng Database (Mock GORM với sqlmock) cho chức năng Tìm kiếm chuyên gia:

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
      <td><b>TC-PROF-EXP-01</b></td>
      <td>Tìm kiếm chuyên gia thành công có kết quả (Happy Case)</td>
      <td>
        <ul>
          <li>Gửi query parameters: <code>search = "Dr. A"</code>, <code>page = 1</code>, <code>page_size = 10</code>.</li>
          <li>Giả lập sqlmock:
            <ul>
              <li>Query đếm dòng (Count): Khớp SQL SELECT COUNT và trả về giá trị 1.</li>
              <li>Query lấy danh sách (Find): Khớp SQL SELECT với filter <code>name ILIKE %Dr. A%</code> và trả về 1 record profile có thông tin chuyên gia hợp lệ.</li>
            </ul>
          </li>
        </ul>
      </td>
      <td>Khởi tạo request <code>GET /api/v1/profiles/experts?search=Dr.+A&amp;page=1&amp;page_size=10</code>, ghi vào <code>httptest.NewRecorder()</code> và gọi hàm <code>handlers.ListExperts</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = true</code></li>
              <li>Trường <code>data.items</code> chứa danh sách chuyên gia khớp dữ liệu giả lập.</li>
              <li>Các thông số phân trang chính xác: <code>total_items = 1</code>, <code>total_pages = 1</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-PROF-EXP-02</b></td>
      <td>Tìm kiếm chuyên gia không ra kết quả</td>
      <td>
        <ul>
          <li>Gửi query parameters: <code>search = "Unknown"</code>.</li>
          <li>Giả lập sqlmock:
            <ul>
              <li>Query đếm dòng (Count): Trả về 0.</li>
              <li>Query lấy danh sách (Find): Trả về kết quả rỗng.</li>
            </ul>
          </li>
        </ul>
      </td>
      <td>Khởi tạo request <code>GET /api/v1/profiles/experts?search=Unknown</code>, ghi vào <code>httptest.NewRecorder()</code> và gọi hàm <code>handlers.ListExperts</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = true</code></li>
              <li>Trường <code>data.items</code> là một mảng rỗng <code>[]</code>.</li>
              <li>Thông số phân trang: <code>total_items = 0</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-PROF-EXP-03</b></td>
      <td>Lỗi truy vấn dữ liệu từ database</td>
      <td>
        <ul>
          <li>Gửi request tìm kiếm thông thường.</li>
          <li>Giả lập sqlmock:
            <ul>
              <li>Gặp lỗi kết nối database hoặc lỗi cú pháp truy vấn, trả về lỗi <code>gorm.ErrInvalidDB</code> hoặc lỗi sql bất kỳ.</li>
            </ul>
          </li>
        </ul>
      </td>
      <td>Khởi tạo request <code>GET /api/v1/profiles/experts</code>, ghi vào <code>httptest.NewRecorder()</code> và gọi hàm <code>handlers.ListExperts</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>500 Internal Server Error</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = false</code></li>
              <li>Thông báo lỗi <code>message = "Lỗi truy vấn danh sách chuyên gia"</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-PROF-EXP-04</b></td>
      <td>Xem chi tiết chuyên gia thành công (Happy Case)</td>
      <td>
        <ul>
          <li>Truyền ID của chuyên gia dưới dạng UUID hợp lệ.</li>
          <li>Giả lập sqlmock:
            <ul>
              <li>Query tìm một record (First): Khớp với ID được truyền và trả về 1 dòng dữ liệu đầy đủ (bao gồm preloaded ExpertProfile và Specializations).</li>
            </ul>
          </li>
        </ul>
      </td>
      <td>Thiết lập path parameter <code>id</code> và gửi request <code>GET /api/v1/profiles/experts/:id</code>, gọi hàm <code>handlers.GetExpert</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = true</code></li>
              <li>Trường <code>data</code> chứa đầy đủ chi tiết hồ sơ chuyên gia khớp với UUID đã truyền.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-PROF-EXP-05</b></td>
      <td>Xem chi tiết chuyên gia không tồn tại</td>
      <td>
        <ul>
          <li>Truyền ID của chuyên gia dưới dạng UUID hợp lệ nhưng không tồn tại trong DB.</li>
          <li>Giả lập sqlmock:
            <ul>
              <li>Query tìm một record (First): Trả về lỗi <code>gorm.ErrRecordNotFound</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
      <td>Thiết lập path parameter <code>id</code> và gửi request <code>GET /api/v1/profiles/experts/:id</code>, gọi hàm <code>handlers.GetExpert</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>404 Not Found</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>Trường <code>success = false</code></li>
              <li>Thông báo lỗi <code>message = "Không tìm thấy hồ sơ chuyên gia"</code>.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
  </tbody>
</table>
