## MindCare - E2E Blackbox Testing (Selenium & Python)

Tài liệu tổng hợp toàn bộ Use Case và Test Case phục vụ kiểm thử hộp đen (Blackbox Testing) trên giao diện web cho hệ thống tư vấn tâm lý trực tuyến **MindCare** (Link demo: `https://sandbox.qmcloud.io.vn`).

## 1. Tổng hợp Danh sách Test Case theo Use Case

| UC ID | Tên Use Case | Test Case IDs | Điều kiện / Kịch bản chính | Kết quả mong đợi | Trạng thái (Theo báo cáo) |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **UC-01** | Login / Logout | `TC-AUTH-LGN-CON-01` đến `06`<br>`TC-AUTH-LGN-SVC-01` đến `03` | - Đăng nhập hợp lệ (Happy Case)<br>- Trống email / Email sai định dạng<br>- Tài khoản không tồn tại / Sai mật khẩu | - Đăng nhập thành công, chuyển vào Dashboard.<br>- Báo lỗi validate hoặc sai thông tin tương ứng (khóa 15p nếu sai 5 lần). | PASS |
| **UC-02** | Register (Đăng ký) | `TC-AUTH-REG-CON-01` đến `07`<br>`TC-AUTH-REG-SVC-01` đến `04` | Đăng ký tài khoản mới cho Client / Expert với dữ liệu hợp lệ và không hợp lệ. | Tạo tài khoản thành công và chuyển sang bước xác thực/đăng nhập. | PASS |
| **UC-03** | Tìm kiếm chuyên gia | `TC-PROF-EXP-01` đến `05` | - Tìm kiếm theo từ khóa/chuyên khoa hợp lệ<br>- Tìm kiếm từ khóa không tồn tại | - Trả về danh sách chuyên gia phù hợp.<br>- Trả về danh sách rỗng nếu không có kết quả. | PASS |
| **UC-04** | Quản lý hồ sơ | `TC-PROF-MNG-01` đến `05` | - Tạo profile mới / Conflict đã có profile<br>- Xem & cập nhật hồ sơ cá nhân | - Lưu và hiển thị thông tin chính xác.<br>- Báo lỗi nếu thiếu trường bắt buộc hoặc sai định dạng SĐT. | PASS |
| **UC-05** | Đặt lịch (Booking) | `TC-BOOK-LCK-01` đến `04` | - Chọn khung giờ trống (Available) để giữ chỗ tạm thời (15 phút)<br>- Xung đột giữ chỗ từ Client khác | - Khóa slot thành công.<br>- Báo lỗi nếu khung giờ đã bị người khác chọn. | PASS |
| **UC-06** | Quản lý lịch hẹn | `TC-BOOK-MNG-01` đến `04` | - Lọc lịch hẹn theo Ngày/Trạng thái<br>- Expert truy cập lịch hẹn của Expert khác qua URL | - Hiển thị đúng danh sách lọc.<br>- Chặn quyền truy cập (403) nếu xem lịch người khác. | PASS |
| **UC-07** | Thanh toán | `TC-PAY-ORD-01` đến `06` | - Tạo đơn thanh toán VNPay cho cuộc hẹn hợp lệ / không tồn tại / sai quyền<br>- Nạp tiền vào ví | - Chuyển hướng sang VNPay hoặc nạp ví thành công.<br>- Báo lỗi số dư không đủ / không có quyền. | PASS |
| **UC-08** | Ghi nhận thù lao | `TC-PAY-WLT-01` đến `05` | - Thanh toán hoàn tất -> tính chiết khấu (15% sàn, 85% chuyên gia)<br>- Xem số dư ví | - Cộng đúng số dư thù lao vào ví chuyên gia và ghi lịch sử thu nhập. | PASS |
| **UC-09** | Nhận tư vấn (Chat/Video) | `TC-CHAT-AUTH-01` đến `04` | - Kết nối Socket với JWT hợp lệ / thiếu token / token hết hạn<br>- Gửi tin nhắn văn bản | - Cho phép kết nối & lưu tin nhắn nếu hợp lệ.<br>- Từ chối kết nối nếu thiếu/hết hạn JWT. | PASS |
| **UC-10** | Thực hiện tư vấn | `TC-CHAT-SND-01` đến `05` | - Gửi tin nhắn tư vấn hợp lệ / rỗng / thiếu định danh<br>- Xem thông tin Client & ghi chú kết quả | - Tin nhắn gửi thành công tới 2 bên; lưu được ghi chú tư vấn. | PASS |
| **UC-11** | Nhận thông báo | `TC-NOTIF-01` đến `03` | - Sự kiện đặt lịch, thanh toán thành công<br>- Đánh dấu thông báo đã đọc | - Gửi thông báo JSON real-time đúng người dùng; cập nhật trạng thái "Đã đọc". | **FAIL** *(Cần kiểm tra lại)* |
| **UC-12** | Trắc nghiệm tâm lý | `TC-ASM-SUB-01` đến `05` | - Làm bài test, bỏ trống câu hỏi bắt buộc<br>- Nộp bài & nhận phân tích AI | - Cuộn đến câu hỏi thiếu và báo lỗi.<br>- Chấm điểm và hiển thị nhận xét trong < 2s. | PASS |
| **UC-13** | Hỏi đáp cộng đồng | `TC-FORUM-QNA-01` đến `04` | - Đăng câu hỏi / bình luận chứa từ khóa vi phạm<br>- Đăng ẩn danh | - Chặn từ ngữ vi phạm.<br>- Hiển thị câu hỏi ẩn danh thành công. | PASS |
| **UC-14** | Trả/Giải ngân thù lao | `TC-PAY-WTD-01` đến `07` | - Liên kết ngân hàng đúng / sai tên chủ TK<br>- Tạo phiếu rút tiền & Admin duyệt chi trả | - Tạm khóa tiền khi rút; Admin duyệt chuyển trạng thái "Đã chi trả" (disable nút tránh trùng lặp). | PASS |
| **UC-15** | Quản lý dịch vụ | `TC-SVC-01` đến `03` | - Admin tạo dịch vụ mới (hợp lệ / giá âm / trống tên)<br>- Cập nhật giá dịch vụ | - Tạo/cập nhật thành công (201/200) hoặc báo lỗi validate (400). | PASS |
| **UC-16** | Quản lý người dùng | `TC-ADM-01` đến `05` | - Admin khóa / mở khóa tài khoản vi phạm<br>- Thao tác với tài khoản không tồn tại | - Khóa thành công và hủy phiên đăng nhập; báo 404 nếu không tồn tại. | PASS |
| **UC-17** | Quản lý bài viết | `TC-FORUM-MNG-01` đến `04` | - Tác giả/Admin sửa, ẩn, xóa bài viết<br>- Sửa bài viết của người khác | - Cập nhật/ẩn thành công; báo 403 Forbidden nếu sửa bài người khác. | PASS |
| **UC-18** | Quản lý bài trắc nghiệm | `TC-ASM-TPL-01` đến `05` | - Tạo bài test mới / trùng slug<br>- Công khai bài test (đủ câu hỏi & thang điểm) | - Tạo thành công (200) hoặc báo lỗi trùng mã (400). | PASS |
| **UC-19** | Quản lý chatroom | *(Theo dõi UC-19)* | Admin xem danh sách, khóa phòng hoặc xử lý báo cáo vi phạm | Cập nhật trạng thái chatroom và thông báo kết quả. | - |

## 2. Hướng dẫn chạy Test

### Cài đặt môi trường
```bash
cd tests-e2e
python -m venv venv
source venv/bin/activate  # Windows: venv\Scripts\activate
pip install -r requirements.txt
```

### Chạy kiểm thử tự động E2E với pytest

```bash
# Chạy toàn bộ các test suites (UC-01 đến UC-08, UC-09, UC-11 đến UC-14, UC-16 đến UC-18)
pytest tests/ -v

# Chạy test theo từng Use Case cụ thể:
pytest tests/test_uc01_auth.py -v                  # UC-01: Authentication (Login/Logout)
pytest tests/test_uc02_register.py -v              # UC-02: Register
pytest tests/test_uc03_expert_search.py -v         # UC-03: Expert Search
pytest tests/test_uc04_profile.py -v               # UC-04: Profile Management
pytest tests/test_uc05_booking.py -v               # UC-05: Booking & Slot Lock
pytest tests/test_uc06_appointment.py -v           # UC-06: Appointment Management
pytest tests/test_uc07_payment.py -v               # UC-07: Payment Orders
pytest tests/test_uc08_wallet.py -v                # UC-08: Wallet & Earnings
pytest tests/test_uc09_chat.py -v                  # UC-09: Consultation Chat & Video Meet
pytest tests/test_uc11_notification.py -v          # UC-11: System Notifications (Fails if not implemented)
pytest tests/test_uc12_assessment_submission.py -v # UC-12: Psychological Assessment Submission
pytest tests/test_uc13_community_forum.py -v       # UC-13: Community Q&A & Forum Posting
pytest tests/test_uc14_payout_withdraw.py -v       # UC-14: Payout & Wallet Withdrawals
pytest tests/test_uc16_admin_users.py -v           # UC-16: Admin User Management
pytest tests/test_uc17_forum.py -v                 # UC-17: Forum Management
pytest tests/test_uc18_assessment.py -v            # UC-18: Assessment Management

# Xuất báo cáo HTML
pytest tests/ -v --html=report.html --self-contained-html
```

## 3. Cấu trúc Page Object Model (POM)

- `pages/base_page.py`: BasePage cung cấp các hàm tiện ích (`find_element`, `click_element`, `type_text`, `wait_until_visible`, `wait_until_clickable`, `get_toast_message`). Hỗ trợ tương thích chuẩn cho React Hook Form & Controlled Inputs.
- `pages/login_page.py`: Page object cho luồng Đăng nhập / Đăng xuất (`/auth/login`), bao gồm helper methods đăng nhập tự động theo vai trò (`login_as_patient`, `login_as_expert`, `login_as_admin`).
- `pages/register_page.py`: Page object cho luồng Đăng ký (`/auth/register`), hỗ trợ chọn vai trò (Client / Expert) và validate form.
- `pages/expert_search_page.py`: Page object cho Tìm kiếm & Lọc chuyên gia theo từ khóa và chuyên khoa (`/experts`).
- `pages/profile_page.py`: Page object cho Quản lý hồ sơ cá nhân (`/client/profile` & `/expert/profile`).
- `pages/booking_page.py`: Page object cho Đặt lịch hẹn và khóa slot thời gian (`/booking`).
- `pages/appointment_page.py`: Page object cho Quản lý lịch hẹn của Bệnh nhân & Chuyên gia (`/client/appointments`, `/expert/appointments`).
- `pages/payment_page.py`: Page object cho kết quả thanh toán VNPay và đơn hàng (`/payment/result`).
- `pages/wallet_page.py`: Page object cho Quản lý ví điện tử, nạp tiền, rút tiền và thu nhập chuyên gia (`/patient/wallet`, `/expert/wallet`, `/admin/withdrawals`).
- `pages/chat_page.py`: Page object cho Tin nhắn tư vấn trực tuyến và phòng họp video (`/patient/messages`, `/expert/messages`, `/meet/:id`).
- `pages/notification_page.py`: Page object cho Trung tâm thông báo và chuông thông báo hệ thống (`/notifications`).
- `pages/admin_users_page.py`: Page object cho Admin Quản lý & Xác thực người dùng (`/admin/users`).
- `pages/forum_page.py`: Page object cho Diễn đàn cộng đồng, đăng bài & Quản lý bài viết (`/patient/forum`, `/community`).
- `pages/assessment_page.py`: Page object cho Quản lý bài trắc nghiệm và quy trình làm bài đánh giá tâm lý (`/patient/assessment`, `/admin/templates`).
 
