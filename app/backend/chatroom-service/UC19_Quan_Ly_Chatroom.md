# Tài Liệu Bàn Giao Thiết Kế Unit Test - UC-19: Quản Lý Chatroom & DM Store

Tài liệu này chứa Ma trận dò vết và Bảng kế hoạch Unit Test chi tiết bằng NodeJS cho logic Quản lý phòng chat và Lưu trữ lịch sử tin nhắn trong Redis Store (`makeDmId`, `pushDM`, `getDMHistory`) thuộc dịch vụ `chatroom-service`.

---

## 1. MA TRẬN DÒ VẾT (TRACEABILITY MATRIX) - UC-19

| Mã Yêu Cầu (Req ID) | Nhóm Chức Năng | Use Case (UC) | Use Case Name | Test Case Liên Kết (Test Case ID) | Trạng thái |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **REQ-CHAT-04** | Lịch sử hội thoại & Đảo ngược thời gian | UC-19 | Quản lý Chatroom | TC-CHAT-MNG-01<br>TC-CHAT-MNG-02<br>TC-CHAT-MNG-03 | **READY TO RUN** |
| **REQ-CHAT-05** | Mã định danh DM hai chiều | UC-19 | Quản lý Chatroom | TC-CHAT-MNG-04<br>TC-CHAT-MNG-05 | **READY TO RUN** |

---

## 2. BẢNG KẾ HOẠCH UNIT TEST CHI TIẾT - UC-19

Dưới đây là bảng kế hoạch chi tiết kiểm thử tầng DM Store (`stores/dm.store.js`) tương tác với Redis Cache:

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
      <td><b>TC-CHAT-MNG-01</b></td>
      <td>Tạo DM ID đồng bộ hai chiều bất kể thứ tự tham số (Happy Case - makeDmId)</td>
      <td>
        <ul>
          <li>Hai mã tài khoản: <code>userA = "account-111"</code> và <code>userB = "account-222"</code>.</li>
        </ul>
      </td>
      <td>
        <ul>
          <li>Gọi <code>makeDmId("account-111", "account-222")</code>.</li>
          <li>Gọi <code>makeDmId("account-222", "account-111")</code>.</li>
        </ul>
      </td>
      <td>
        <ul>
          <li>Cả 2 lệnh đều trả về cùng một mã chuỗi DM ID kết hợp duy nhất: <code>"account-111:account-222"</code>.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-CHAT-MNG-02</b></td>
      <td>Lấy lịch sử tin nhắn từ Redis và đảo ngược mảng theo thứ tự thời gian tăng dần (getDMHistory)</td>
      <td>
        <ul>
          <li>Redis `lRange` trả về 2 tin nhắn JSON theo cơ chế LIFO (tin mới nhất ở index 0, tin cũ ở index 1).</li>
          <li>Mock `r.lRange("dm:id:messages", 0, 99)` trả về mảng chuỗi JSON.</li>
        </ul>
      </td>
      <td>Gọi hàm <code>getDMHistory(dmId)</code>.</td>
      <td>
        <ul>
          <li>Hàm thực hiện `JSON.parse` trên từng phần tử.</li>
          <li>Mảng kết quả được gọi hàm `.reverse()` để chuyển danh sách tin nhắn về đúng thứ tự thời gian tăng dần (tin cũ xếp trước, tin mới xếp sau).</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-CHAT-MNG-03</b></td>
      <td>Lưu tin nhắn DM mới vào Redis và giới hạn tối đa 100 tin nhắn (pushDM & lTrim)</td>
      <td>
        <ul>
          <li>Mã `dmId = "userA:userB"` và đối tượng tin nhắn hợp lệ `message`.</li>
          <li>Mock Redis client `r.lPush`, `r.lTrim`, `r.expire`.</li>
        </ul>
      </td>
      <td>Gọi hàm <code>pushDM(dmId, message)</code>.</td>
      <td>
        <ul>
          <li>Gọi `r.lPush` với key <code>"dm:userA:userB:messages"</code> và dữ liệu `JSON.stringify(message)`.</li>
          <li>Gọi `r.lTrim` với khoảng `(0, 99)` để duy trì tối đa 100 tin nhắn mới nhất trong danh sách.</li>
          <li>Gọi `r.expire` thiết lập TTL 7 ngày (604,800 giây).</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-CHAT-MNG-04</b></td>
      <td>Tạo DM ID với các chuỗi UUID thực tế của hệ thống</td>
      <td>
        <ul>
          <li>UUID 1: <code>"b1111111-1111-1111-1111-111111111111"</code>.</li>
          <li>UUID 2: <code>"a2222222-2222-2222-2222-222222222222"</code>.</li>
        </ul>
      </td>
      <td>Gọi <code>makeDmId(UUID1, UUID2)</code>.</td>
      <td>
        <ul>
          <li>Đoạn mã tự động sắp xếp theo thứ tự bảng chữ cái và trả về <code>"a2222222-2222-2222-2222-222222222222:b1111111-1111-1111-1111-111111111111"</code>.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><b>TC-CHAT-MNG-05</b></td>
      <td>Xử lý lấy lịch sử DM khi Redis trả về danh sách rỗng</td>
      <td>
        <ul>
          <li>Mock `r.lRange` trả về mảng rỗng `[]`.</li>
        </ul>
      </td>
      <td>Gọi <code>getDMHistory("new-dm-id")</code>.</td>
      <td>
        <ul>
          <li>Trả về mảng rỗng `[]` mà không gây ra ngoại lệ hay lỗi runtime.</li>
        </ul>
      </td>
    </tr>
  </tbody>
</table>
