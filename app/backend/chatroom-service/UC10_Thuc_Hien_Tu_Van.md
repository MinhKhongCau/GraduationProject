# Tài Liệu Bàn Giao Thiết Kế Unit Test - UC-10: Thực Hiện Tư Vấn

Tài liệu này chứa Ma trận dò vết và Bảng kế hoạch Unit Test chi tiết bằng NodeJS cho chức năng Thực hiện tư vấn (Gửi tin nhắn văn bản, tin nhắn thoại và thả cảm xúc) thuộc dịch vụ `chatroom-service`.

---

## 1. MA TRẬN DÒ VẾT (TRACEABILITY MATRIX) - UC-10

| Mã Yêu Cầu (Req ID) | Nhóm Chức Năng | Use Case (UC) | Use Case Name | Test Case Liên Kết (Test Case ID) | Trạng thái |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **REQ-CHAT-03** | Gửi tin nhắn tư vấn | UC-10 | Thực hiện tư vấn | TC-CHAT-SND-01<br>TC-CHAT-SND-02<br>TC-CHAT-SND-03<br>TC-CHAT-SND-04<br>TC-CHAT-SND-05 | **READY TO RUN** |

---

## 2. BẢNG KẾ HOẠCH UNIT TEST CHI TIẾT - UC-10

Dưới đây là bảng kế hoạch chi tiết kiểm thử bộ điều khiển Socket tư vấn tin nhắn (`dmSocketController` & `chatSocketController`):

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
      <td><b>TC-CHAT-SND-01</b></td>
      <td>Chuyên gia gửi tin nhắn tư vấn dạng văn bản thành công (Happy Case - Send DM)</td>
      <td>
        <ul>
          <li>Chuyên gia có `socket.data.user.id = "expert-uuid-1"`.</li>
          <li>Payload phát: <code>toUser = "patient-uuid-1"</code>, <code>text = "Chào bạn, tôi có thể giúp gì cho bạn?"</code>.</li>
          <li>Mock `pushDM` lưu dữ liệu vào Redis và mock `getUserSocket` trả về Socket ID của bệnh nhân.</li>
        </ul>
      </td>
      <td>Phát sự kiện Socket <code>dm:send</code> từ client của Chuyên gia.</td>
      <td>
        <ul>
          <li>Hàm `pushDM` được gọi để lưu tin nhắn vào Redis.</li>
          <li>Phát sự kiện `dm:message` cho cả Chuyên gia (sender) và Bệnh nhân (receiver).</li>
          <li>Nội dung tin nhắn chứa `type: "chat"`, `fromId`, `toId` và chuỗi text được làm sạch (`trim`).</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-CHAT-SND-02</b></td>
      <td>Gửi tin nhắn thất bại do nội dung văn bản bị rỗng hoặc chỉ toàn khoảng trắng</td>
      <td>
        <ul>
          <li>`socket.data.user.id = "expert-uuid-1"`.</li>
          <li>Payload phát: <code>toUser = "patient-uuid-1"</code>, <code>text = "   "</code>.</li>
        </ul>
      </td>
      <td>Phát sự kiện Socket <code>dm:send</code> với chuỗi khoảng trắng.</td>
      <td>
        <ul>
          <li>Hàm `pushDM` KHÔNG được gọi.</li>
          <li>Không phát bất kỳ sự kiện `dm:message` nào tới người nhận.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-CHAT-SND-03</b></td>
      <td>Gửi tin nhắn thất bại do thiếu thông tin định danh người gửi hoặc người nhận</td>
      <td>
        <ul>
          <li>`socket.data.user` bị thiếu (undefined) hoặc payload <code>toUser</code> không được truyền.</li>
        </ul>
      </td>
      <td>Phát sự kiện Socket <code>dm:send</code> không có `toUser`.</td>
      <td>
        <ul>
          <li>Controller dừng thực thi ngay lập tức.</li>
          <li>Không lưu dữ liệu hay gửi sự kiện tới bất kỳ ai.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-CHAT-SND-04</b></td>
      <td>Chuyên gia gửi tin nhắn thoại tư vấn thành công (Send Voice DM)</td>
      <td>
        <ul>
          <li>`socket.data.user.id = "expert-uuid-1"`.</li>
          <li>Payload chứa: <code>toUser = "patient-uuid-1"</code>, <code>audio = "data:audio/webm;base64,..."</code>, <code>duration = 12</code>, <code>mimeType = "audio/webm"</code>.</li>
        </ul>
      </td>
      <td>Phát sự kiện Socket <code>dm:send:voice</code> với dữ liệu âm thanh dạng base64.</td>
      <td>
        <ul>
          <li>Hàm `pushDM` lưu đối tượng tin nhắn thoại có `type: "voice"`.</li>
          <li>Phát sự kiện `dm:message` chứa đầy đủ trường `audio`, `duration`, `mimeType`.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-CHAT-SND-05</b></td>
      <td>Thả biểu cảm (Reaction) trên tin nhắn tư vấn thành công (React DM)</td>
      <td>
        <ul>
          <li>`socket.data.user.id = "expert-uuid-1"`.</li>
          <li>Payload chứa: <code>toUser = "patient-uuid-1"</code>, <code>messageId = "msg-123"</code>, <code>emoji = "❤️"</code>.</li>
          <li>Mock `toggleReaction` xử lý lưu cảm xúc vào Redis.</li>
        </ul>
      </td>
      <td>Phát sự kiện Socket <code>dm:react</code>.</td>
      <td>
        <ul>
          <li>Hàm `toggleReaction` được gọi với đúng `dmId`, `messageId`, `emoji`, `fromId`.</li>
          <li>Phát sự kiện `dm:reaction` cập nhật biểu cảm cho cả 2 phía.</li>
        </ul>
      </td>
    </tr>
  </tbody>
</table>
