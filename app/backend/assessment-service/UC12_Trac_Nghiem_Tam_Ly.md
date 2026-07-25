# Tài Liệu Bàn Giao Thiết Kế Unit Test - UC-12: Làm Trắc Nghiệm Tâm Lý

Tài liệu này chứa Ma trận dò vết và Bảng kế hoạch Unit Test chi tiết cho chức năng Làm trắc nghiệm tâm lý, nộp bài làm, tính điểm phân loại và xem lại kết quả (`Assessments Submit & Self Results`) thuộc dịch vụ `assessment-service`.

---

## 1. MA TRẬN DÒ VẾT (TRACEABILITY MATRIX) - UC-12

| Mã Yêu Cầu (Req ID) | Nhóm Chức Năng | Use Case (UC) | Use Case Name | Test Case Liên Kết (Test Case ID) | Trạng thái |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **REQ-ASM-01** | Nộp bài & Chấm điểm | UC-12 | Làm trắc nghiệm | TC-ASM-SUB-01<br>TC-ASM-SUB-02<br>TC-ASM-SUB-03<br>TC-ASM-SUB-04<br>TC-ASM-SUB-05 | **READY TO RUN** |
| **REQ-ASM-02** | Xem kết quả cá nhân | UC-12 | Làm trắc nghiệm | TC-ASM-HIS-01<br>TC-ASM-HIS-02 | **READY TO RUN** |

---

## 2. BẢNG KẾ HOẠCH UNIT TEST CHI TIẾT - UC-12

Dưới đây là bảng kế hoạch chi tiết kiểm thử các API Endpoint nộp bài và truy vấn kết quả đánh giá tâm lý cá nhân (`POST /api/v1/assessments/submit` & `GET /api/v1/assessments/self`):

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
      <td><b>TC-ASM-SUB-01</b></td>
      <td>Nộp bài trắc nghiệm và chấm điểm thành công (Happy Case - Submit Assessment)</td>
      <td>
        <ul>
          <li>Bài test DASS-21 có `slug = "dass-21"` tồn tại trong CSDL.</li>
          <li>Câu hỏi `q1` thuộc thang đo `Depression`, phương án trả lựa chọn có `score_value = 3`.</li>
          <li>Header `X-User-Id = "11111111-1111-1111-1111-111111111111"`.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/assessments/submit</code> với danh sách câu trả lời hợp lệ.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code></li>
          <li>Response JSON chứa:
            <ul>
              <li>`message`: <code>"Nộp bài và chấm điểm thành công!"</code></li>
              <li>`total_score = 3`</li>
              <li>`dimension_scores`: <code>{"Depression": 3}</code></li>
              <li>`result_id` được khởi tạo và lưu vào CSDL.</li>
            </ul>
          </li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-ASM-SUB-02</b></td>
      <td>Nộp bài thất bại do mã bài trắc nghiệm không tồn tại (Template Not Found)</td>
      <td>
        <ul>
          <li>Header `X-User-Id = "11111111-1111-1111-1111-111111111111"`.</li>
          <li>Payload chứa `template_id = "non-existent-slug"`.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/assessments/submit</code> với mã bài test không hợp lệ.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>404 Not Found</code></li>
          <li>Response JSON chứa `detail`: <code>"Không tìm thấy bài test tương ứng!"</code>.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-ASM-SUB-03</b></td>
      <td>Nộp bài có chứa câu hỏi hoặc phương án trả lời bị ẩn/xóa (Invalid Answer Ignored)</td>
      <td>
        <ul>
          <li>Một câu hỏi có `is_active = False` hoặc `option_id` không tồn tại.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/assessments/submit</code> kèm đáp án không hợp lệ.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code></li>
          <li>Hệ thống bỏ qua câu hỏi không hợp lệ và chỉ tính điểm trên các đáp án hợp lệ.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-ASM-SUB-04</b></td>
      <td>Nộp bài thất bại do thiếu header định danh người dùng (Unauthorized)</td>
      <td>
        <ul>
          <li>Thiếu header `X-User-Id` hoặc truyền chuỗi rỗng.</li>
        </ul>
      </td>
      <td>Gửi request <code>POST /api/v1/assessments/submit</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>401 Unauthorized</code> hoặc <code>400 Bad Request</code>.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-ASM-SUB-05</b></td>
      <td>Tích hợp gọi dịch vụ AI đánh giá nhận xét tâm lý tự động khi nộp bài</td>
      <td>
        <ul>
          <li>Mock hàm `generate_psychological_advice` trả về nhận xét tâm lý mẫu.</li>
        </ul>
      </td>
      <td>Nộp bài trắc nghiệm thành công.</td>
      <td>
        <ul>
          <li>Trường `ai_evaluation` trong kết quả lưu đúng nội dung văn bản nhận xét từ AI.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-ASM-HIS-01</b></td>
      <td>Xem danh sách kết quả trắc nghiệm cá nhân thành công (Get My Assessments)</td>
      <td>
        <ul>
          <li>Người dùng đã hoàn thành 2 bài test trước đó và lưu trong CSDL.</li>
          <li>Header `X-User-Id = "11111111-1111-1111-1111-111111111111"`.</li>
        </ul>
      </td>
      <td>Gửi request <code>GET /api/v1/assessments/self</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>200 OK</code></li>
          <li>Trả về danh sách các kết quả làm bài xếp theo thứ tự thời gian giảm dần (`created_at desc`).</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-ASM-HIS-02</b></td>
      <td>Xem danh sách kết quả thất bại do User ID không đúng định dạng UUID</td>
      <td>
        <ul>
          <li>Header `X-User-Id = "invalid-uuid-format"`.</li>
        </ul>
      </td>
      <td>Gửi request <code>GET /api/v1/assessments/self</code>.</td>
      <td>
        <ul>
          <li>HTTP Status Code: <code>400 Bad Request</code></li>
          <li>Response JSON chứa `detail`: <code>"User ID extract from token is not a valid UUID!"</code>.</li>
        </ul>
      </td>
    </tr>
  </tbody>
</table>
