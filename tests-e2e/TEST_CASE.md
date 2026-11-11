# Kịch bản Kiểm thử Đơn vị (Unit Test) - Hệ thống MindCare

Tài liệu này tổng hợp chi tiết từng kịch bản kiểm thử (Unit Test) theo từng Use Case từ Chương V của báo cáo Đồ án tốt nghiệp[cite: 1].

---

## 1. UC-01: Login & Logout

### 1.1. Chức năng Login (Bảng 5.1)[cite: 1]

| ID | Tên Test Case | Điều kiện (Setup Data) | Kết quả mong đợi |
| :--- | :--- | :--- | :--- |
| `TC-AUTH-LGN-CON-01` | Đăng nhập thành công (Happy Case) | Gửi yêu cầu đăng nhập với tài khoản và mật khẩu hợp lệ. | Đăng nhập thành công, hiển thị thông tin xác thực chính xác.[cite: 1] |
| `TC-AUTH-LGN-CON-02` | Lỗi validate: Trống email | Gửi yêu cầu đăng nhập nhưng để trống trường email. | Báo lỗi dữ liệu không hợp lệ do thiếu email.[cite: 1] |
| `TC-AUTH-LGN-CON-03` | Lỗi validate: Email sai định dạng | Gửi yêu cầu đăng nhập với email sai định dạng. | Báo lỗi dữ liệu không hợp lệ do định dạng email không đúng.[cite: 1] |
| `TC-AUTH-LGN-SVC-01` | Xác thực thành công (Service) | Thông tin đăng nhập hợp lệ, tài khoản tồn tại, mật khẩu chính xác và hệ thống tạo, lưu phiên đăng nhập thành công. | Tạo thông tin xác thực thành công và lưu phiên làm việc vào hệ thống.[cite: 1] |
| `TC-AUTH-LGN-SVC-02` | Lỗi tài khoản không tồn tại | Đăng nhập với tài khoản chưa được đăng ký trên hệ thống. | Báo lỗi không tìm thấy tài khoản.[cite: 1] |
| `TC-AUTH-LGN-SVC-03` | Lỗi sai mật khẩu | Đăng nhập với tài khoản có tồn tại nhưng nhập sai mật khẩu. | Báo lỗi mật khẩu không chính xác.[cite: 1] |

### 1.2. Chức năng Logout (Bảng 5.2)[cite: 1]
*(Lưu ý: Trong báo cáo gốc, nội dung Bảng 5.2 đang bị trùng lặp với Bảng 5.1)*[cite: 1]

| ID | Tên Test Case | Điều kiện (Setup Data) | Kết quả mong đợi |
| :--- | :--- | :--- | :--- |
| `TC-AUTH-LGN-CON-01` | Đăng nhập thành công (Happy Case) | Gửi yêu cầu đăng nhập với tài khoản và mật khẩu hợp lệ. | Đăng nhập thành công, hiển thị thông tin xác thực chính xác.[cite: 1] |
| `TC-AUTH-LGN-CON-02` | Lỗi validate: Trống email | Gửi yêu cầu đăng nhập nhưng để trống trường email. | Báo lỗi dữ liệu không hợp lệ do thiếu email.[cite: 1] |
| `TC-AUTH-LGN-CON-03` | Lỗi validate: Email sai định dạng | Gửi yêu cầu đăng nhập với email sai định dạng. | Báo lỗi dữ liệu không hợp lệ do định dạng email không đúng.[cite: 1] |
| `TC-AUTH-LGN-SVC-01` | Xác thực thành công (Service) | Thông tin đăng nhập hợp lệ, tài khoản tồn tại, mật khẩu chính xác và hệ thống tạo, lưu phiên đăng nhập thành công. | Tạo thông tin xác thực thành công và lưu phiên làm việc vào hệ thống.[cite: 1] |
| `TC-AUTH-LGN-SVC-02` | Lỗi tài khoản không tồn tại | Đăng nhập với tài khoản chưa được đăng ký trên hệ thống. | Báo lỗi không tìm thấy tài khoản.[cite: 1] |
| `TC-AUTH-LGN-SVC-03` | Lỗi sai mật khẩu | Đăng nhập với tài khoản có tồn tại nhưng nhập sai mật khẩu. | Báo lỗi mật khẩu không chính xác.[cite: 1] |

---

## 2. UC-02: Đăng ký (Bảng 5.3)[cite: 1]

| ID | Tên Test Case | Điều kiện (Setup Data) | Kết quả mong đợi (Expected Output) |
| :--- | :--- | :--- | :--- |
| `TC-AUTH-REG-CON-01` | Đăng ký thành công (Controller) | Gửi yêu cầu đăng ký với thông tin hợp lệ. | Đăng ký tài khoản mới thành công.[cite: 1] |
| `TC-AUTH-REG-CON-02` | Lỗi validate: Trống họ tên | Gửi yêu cầu đăng ký nhưng để trống trường họ tên. | Báo lỗi dữ liệu không hợp lệ do thiếu họ tên.[cite: 1] |
| `TC-AUTH-REG-CON-03` | Lỗi validate: Trống email | Gửi yêu cầu đăng ký nhưng để trống trường email. | Báo lỗi dữ liệu không hợp lệ do thiếu email.[cite: 1] |
| `TC-AUTH-REG-SVC-01` | Đăng ký thành công với vai trò chỉ định | Thông tin đăng ký hợp lệ, email chưa tồn tại, mật khẩu được mã hóa và tài khoản mới được lưu thành công với vai trò cụ thể. | Lưu tài khoản mới thành công với đầy đủ thông tin và vai trò được chỉ định.[cite: 1] |
| `TC-AUTH-REG-SVC-02` | Đăng ký thành công dùng vai trò mặc định | Yêu cầu đăng ký hợp lệ không chỉ định vai trò rõ ràng. | Lưu tài khoản mới thành công với vai trò mặc định của hệ thống.[cite: 1] |

---

## 3. UC-03: Tìm kiếm chuyên gia (Bảng 5.4)[cite: 1]

| ID | Tên Test Case | Điều kiện (Setup Data) | Kết quả mong đợi (Expected Output) |
| :--- | :--- | :--- | :--- |
| `TC-PROF-EXP-01` | Tìm kiếm thành công có kết quả (Happy Case) | Gửi yêu cầu tìm kiếm chuyên gia kèm theo từ khóa hợp lệ và thông tin phân trang. | Trả về danh sách hồ sơ chuyên gia phù hợp với từ khóa tìm kiếm.[cite: 1] |
| `TC-PROF-EXP-02` | Tìm kiếm chuyên gia không ra kết quả | Gửi yêu cầu tìm kiếm chuyên gia với từ khóa không tồn tại trên hệ thống. | Trả về danh sách kết quả trống.[cite: 1] |
| `TC-PROF-EXP-03` | Lỗi truy vấn dữ liệu từ database | Gửi yêu cầu tìm kiếm chuyên gia nhưng hệ thống gặp sự cố kết nối hoặc lỗi truy vấn cơ sở dữ liệu. | Báo lỗi truy vấn danh sách chuyên gia.[cite: 1] |

---

## 4. UC-04: Quản lý hồ sơ (Bảng 5.5)[cite: 1]

| ID | Tên Test Case | Điều kiện (Setup Data) | Kết quả mong đợi (Expected Output) |
| :--- | :--- | :--- | :--- |
| `TC-PROF-MNG-01` | Tạo profile nội bộ thành công (Happy Case) | Gửi yêu cầu tạo hồ sơ mới với đầy đủ thông tin hợp lệ, thông tin tài khoản chưa từng tồn tại trên hệ thống. | Tạo và lưu thông tin hồ sơ mới thành công.[cite: 1] |
| `TC-PROF-MNG-02` | Tạo profile thất bại do đã có profile (Conflict) | Gửi yêu cầu tạo hồ sơ với mã định danh tài khoản đã tồn tại trong hệ thống. | Báo lỗi hồ sơ cho tài khoản này đã tồn tại.[cite: 1] |
| `TC-PROF-MNG-03` | Xem profile cá nhân thành công (Happy Case) | Hệ thống xác thực người dùng hợp lệ, sau đó tìm kiếm và truy xuất thông tin hồ sơ cá nhân. | Hiển thị thông tin chi tiết hồ sơ cá nhân của tài khoản đang đăng nhập.[cite: 1] |

---

## 5. UC-05: Đặt lịch (Bảng 5.6)[cite: 1]

| ID | Tên Test Case | Điều kiện (Setup Data) | Kết quả mong đợi (Expected Output) |
| :--- | :--- | :--- | :--- |
| `TC-BOOK-LCK-01` | Giữ chỗ tạm thời thành công (Happy Case - Lock Slot) | Người dùng với vai trò bệnh nhân gửi yêu cầu khóa một lịch hẹn hợp lệ. | Giữ chỗ lịch hẹn thành công trong thời gian quy định.[cite: 1] |
| `TC-BOOK-LCK-02` | Giữ chỗ thất bại do slot đã bị người khác khóa (Conflict) | Bệnh nhân gửi yêu cầu khóa lịch hẹn nhưng lịch hẹn này trước đó đã bị người khác khóa hoặc đã được đặt chỗ. | Báo lỗi lịch hẹn đã có người khác khóa.[cite: 1] |
| `TC-BOOK-APT-01` | Tạo cuộc hẹn thành công (Happy Case - Create Appointment) | Bệnh nhân gửi yêu cầu đặt lịch hẹn hợp lệ với chuyên gia. | Tạo lịch hẹn thành công ở trạng thái chờ thanh toán.[cite: 1] |
| `TC-BOOK-APT-02` | Tạo cuộc hẹn thất bại do chưa giữ chỗ slot | Bệnh nhân gửi yêu cầu đặt lịch hẹn với mã lịch hẹn chưa được giữ chỗ hoặc đang ở trạng thái trống. | Báo lỗi chưa thực hiện giữ chỗ lịch hẹn.[cite: 1] |
| `TC-BOOK-APT-03` | Tạo cuộc hẹn thất bại do phiên giữ chỗ hết hạn | Bệnh nhân gửi yêu cầu đặt lịch hẹn nhưng phiên giữ chỗ của lịch hẹn đã hết hạn. | Báo lỗi phiên giữ chỗ đã hết hạn.[cite: 1] |

---

## 6. UC-06: Quản lý lịch hẹn (Bảng 5.7)[cite: 1]

| ID | Tên Test Case | Điều kiện (Setup Data) | Kết quả mong đợi (Expected Output) |
| :--- | :--- | :--- | :--- |
| `TC-BOOK-MNG-01` | Bệnh nhân xem danh sách lịch hẹn thành công | Bệnh nhân gửi yêu cầu truy vấn danh sách cuộc hẹn của mình kèm theo thông tin phân trang hợp lệ. | Hiển thị danh sách cuộc hẹn của bệnh nhân thành công.[cite: 1] |
| `TC-BOOK-MNG-02` | Chuyên gia xem danh sách lịch hẹn thành công | Chuyên gia gửi yêu cầu truy vấn danh sách cuộc hẹn của mình kèm theo thông tin phân trang hợp lệ. | Hiển thị danh sách cuộc hẹn tương ứng của chuyên gia thành công.[cite: 1] |
| `TC-BOOK-MNG-03` | Xem chi tiết lịch hẹn thành công | Bệnh nhân gửi yêu cầu xem thông tin chi tiết của một cuộc hẹn bằng mã định danh hợp lệ. | Hiển thị thông tin chi tiết của cuộc hẹn thành công.[cite: 1] |

---

## 7. UC-07: Thanh toán (Bảng 5.8)[cite: 1]

| ID | Tên Test Case | Điều kiện (Setup Data) | Kết quả mong đợi (Expected Output) |
| :--- | :--- | :--- | :--- |
| `TC-PAY-ORD-01` | Tạo đơn thanh toán VNPay thành công (Happy Case) | Người dùng gửi yêu cầu thanh toán cho một cuộc hẹn hợp lệ. | Tạo đơn hàng thanh toán thành công và cung cấp liên kết thanh toán.[cite: 1] |
| `TC-PAY-ORD-02` | Tạo đơn hàng thất bại do không tìm thấy cuộc hẹn | Người dùng gửi yêu cầu thanh toán nhưng mã cuộc hẹn không tồn tại trên hệ thống. | Báo lỗi không tìm thấy cuộc hẹn.[cite: 1] |
| `TC-PAY-ORD-03` | Tạo đơn thất bại do cuộc hẹn không thuộc người thanh toán | Người dùng gửi yêu cầu thanh toán cho cuộc hẹn nhưng không có quyền thao tác trên cuộc hẹn đó. | Báo lỗi không có quyền thao tác trên cuộc hẹn.[cite: 1] |
| `TC-PAY-WLT-01` | Nạp tiền vào ví điện tử thành công (Happy Case) | Tài khoản bệnh nhân thực hiện yêu cầu nạp tiền vào ví. | Nạp tiền thành công và hiển thị thông báo ví đã được nạp tiền.[cite: 1] |
| `TC-PAY-WLT-02` | Ghi nhận thù lao vào ví chuyên gia thành công | Giao dịch thanh toán hoàn tất, hệ thống tính toán chiết khấu sàn và thù lao cho chuyên gia. | Cập nhật chính xác số dư thù lao vào ví chuyên gia và ghi nhận loại giao dịch phù hợp.[cite: 1] |
| `TC-PAY-WLT-03` | Xem thông tin và số dư ví điện tử thành công | Chuyên gia truy cập xem thông tin ví cá nhân. | Hiển thị chính xác số dư hiện tại của ví điện tử.[cite: 1] |

---

## 8. UC-09: Nhận tư vấn (Bảng 5.10)[cite: 1]

| ID | Tên Test Case | Điều kiện (Setup Data) | Kết quả mong đợi (Expected Output) |
| :--- | :--- | :--- | :--- |
| `TC-CHAT-AUTH-01` | Xác thực JWT Token hợp lệ khi kết nối Socket (Happy Case) | Cấu hình xác thực hợp lệ, kết nối kèm theo thông tin định danh người dùng hợp lệ. | Cho phép kết nối thành công và lưu trữ thông tin người dùng vào phiên kết nối.[cite: 1] |
| `TC-CHAT-AUTH-02` | Từ chối kết nối Socket do thiếu JWT Token | Kết nối được gửi đi nhưng không kèm theo thông tin xác thực định danh. | Từ chối kết nối và thông báo thiếu thông tin xác thực.[cite: 1] |
| `TC-CHAT-AUTH-03` | Từ chối kết nối Socket do JWT Token hết hạn | Kết nối được gửi đi với thông tin xác thực đã hết hạn. | Từ chối kết nối và thông báo phiên xác thực đã hết hạn.[cite: 1] |
| `TC-CHAT-SND-01` | Gửi tin nhắn tư vấn văn bản thành công (Happy Case) | Chuyên gia gửi tin nhắn `"Chào bạn, tôi có thể giúp gì cho bạn?"` tới bệnh nhân. | Tin nhắn được lưu lại và gửi thành công tới cả chuyên gia lẫn bệnh nhân.[cite: 1] |
| `TC-CHAT-SND-02` | Gửi thất bại do văn bản rỗng hoặc toàn khoảng trắng | Chuyên gia gửi tin nhắn chỉ chứa khoảng trắng (`" "`). | Tin nhắn không được lưu và không gửi tới người nhận.[cite: 1] |
| `TC-CHAT-SND-03` | Gửi thất bại do thiếu định danh người gửi/nhận | Thiếu thông tin người gửi hoặc người nhận trong yêu cầu. | Hệ thống hủy thực thi ngay, không lưu hay gửi tin nhắn.[cite: 1] |

---

## 9. UC-11: Nhận thông báo hệ thống (Bảng 5.12)[cite: 1]

| ID | Tên Test Case | Điều kiện (Setup Data) | Kết quả mong đợi (Expected Output) |
| :--- | :--- | :--- | :--- |
| `TC-NOTIF-01` | Gửi thông báo sự kiện đặt lịch hẹn thành công (Happy Case) | Hệ thống phát sinh sự kiện đặt lịch hẹn mới (Mã lịch: 100, Bệnh nhân: `patient-uuid-1`, Thời gian: 25/07/2026 09:00). | Gửi thông báo JSON thành công đến đúng bệnh nhân, không phát sinh lỗi.[cite: 1] |
| `TC-NOTIF-02` | Gửi thông báo sự kiện thanh toán thành công (Happy Case) | Hệ thống phát sinh sự kiện thanh toán thành công (Đơn hàng: `ORD-888`, Số tiền: 500.000 VNĐ, Bệnh nhân: `patient-uuid-1`). | Gửi thông báo ghi nhận chính xác mã đơn hàng, số tiền thanh toán và không có lỗi.[cite: 1] |
| `TC-NOTIF-03` | Đánh dấu thông báo đã đọc thành công | Trong hệ thống đang có 1 thông báo (ID = 5) ở trạng thái "Chưa đọc" của bệnh nhân. | Thông báo chuyển sang trạng thái "Đã đọc" và cập nhật thời gian đọc.[cite: 1] |

---

## 10. UC-14: Trả/Giải ngân thù lao (Bảng 5.15)[cite: 1]

| ID | Tên Test Case | Điều kiện (Setup Data) | Kết quả mong đợi (Expected Output) |
| :--- | :--- | :--- | :--- |
| `TC-PAY-WTD-01` | Liên kết ngân hàng thành công (Happy Case) | Chuyên gia nhập đúng thông tin tài khoản ngân hàng MB Bank (STK: `999999999`, Tên: `NGUYEN VAN A`). | Liên kết thành công.[cite: 1] |
| `TC-PAY-WTD-02` | Liên kết thất bại do sai tên chủ tài khoản | Chuyên gia nhập sai tên chủ tài khoản so với dữ liệu thực tế tại ngân hàng. | Trả về mã HTTP `400 Bad Request` kèm thông báo lỗi không khớp tên chủ tài khoản từ ngân hàng.[cite: 1] |
| `TC-PAY-WTD-03` | Tạo phiếu yêu cầu rút tiền thành công | Chuyên gia có đủ số dư và gửi yêu cầu rút 2.000.000 VNĐ về tài khoản đã liên kết. | Trả về mã HTTP `200 OK`, số tiền 2.000.000 VNĐ được tạm khóa và tạo phiếu rút tiền thành công.[cite: 1] |

---

## 11. UC-15: Quản lý dịch vụ tư vấn (Bảng 5.16)[cite: 1]

| ID | Tên Test Case | Điều kiện (Setup Data) | Kết quả mong đợi (Expected Output) |
| :--- | :--- | :--- | :--- |
| `TC-SVC-01` | Tạo dịch vụ tư vấn mới thành công (Happy Case) | Admin nhập đầy đủ thông tin dịch vụ mới (Tên: `"Tư vấn tâm lý trực tuyến 1:1"`, Giá: 300.000 VNĐ, Thời lượng: 60 phút). | Tạo dịch vụ thành công, trả về mã HTTP `201 Created` kèm thông tin chi tiết dịch vụ vừa tạo.[cite: 1] |
| `TC-SVC-02` | Tạo thất bại do trống tên hoặc giá tiền âm | Admin để trống tên dịch vụ hoặc nhập giá tiền âm (ví dụ: -50.000 VNĐ). | Trả về mã HTTP `400 Bad Request` kèm thông báo lỗi dữ liệu không hợp lệ (tên không được trống, giá phải lớn hơn 0).[cite: 1] |
| `TC-SVC-03` | Cập nhật giá và mô tả dịch vụ thành công | Admin thay đổi giá dịch vụ ID = 1 từ 300.000 VNĐ lên 350.000 VNĐ và cập nhật mô tả mới. | Trả về mã HTTP `200 OK` kèm thông tin dịch vụ đã được cập nhật giá mới thành 350.000 VNĐ.[cite: 1] |

---

## 12. UC-16: Quản lý người dùng (Bảng 5.17)[cite: 1]

| ID | Tên Test Case | Điều kiện (Setup Data) | Kết quả mong đợi (Expected Output) |
| :--- | :--- | :--- | :--- |
| `TC-ADM-01` | Admin thực hiện khóa tài khoản vi phạm thành công | Admin thực hiện khóa một tài khoản đang hoạt động do vi phạm tiêu chuẩn (spam). | Khóa tài khoản thành công, trả về mã HTTP `200 OK` và hủy toàn bộ phiên đăng nhập của tài khoản đó.[cite: 1] |
| `TC-ADM-02` | Admin thực hiện mở khóa tài khoản thành công | Admin thực hiện mở khóa cho một tài khoản đang bị khóa. | Mở khóa thành công, trả về mã HTTP `200 OK` và tài khoản hoạt động trở lại bình thường.[cite: 1] |
| `TC-ADM-03` | Khóa/Mở khóa thất bại do tài khoản không tồn tại | Admin thao tác khóa hoặc mở khóa trên một tài khoản không có trong hệ thống. | Trả về mã HTTP `404 Not Found` kèm thông báo không tìm thấy tài khoản người dùng.[cite: 1] |

---

## 13. UC-17: Quản lý bài viết & bình luận (Bảng 5.18)[cite: 1]

| ID | Tên Test Case | Điều kiện (Setup Data) | Kết quả mong đợi (Expected Output) |
| :--- | :--- | :--- | :--- |
| `TC-FORUM-MNG-01` | Chủ bài viết chỉnh sửa bài viết thành công (Happy Case) | Tác giả thực hiện chỉnh sửa tiêu đề và nội dung bài viết của chính mình. | Trả về mã HTTP `200 OK`, hệ thống cập nhật bài viết thành công và hiển thị tiêu đề, nội dung mới.[cite: 1] |
| `TC-FORUM-MNG-02` | Từ chối chỉnh sửa bài viết của người khác | Người dùng cố tình chỉnh sửa bài viết do người khác đăng. | Trả về mã HTTP `403 Forbidden` kèm thông báo không có quyền chỉnh sửa bài viết này.[cite: 1] |
| `TC-FORUM-MNG-03` | Xóa bài viết cá nhân thành công (Soft Delete) | Tác giả thực hiện xóa bài viết của chính mình. | Trả về mã HTTP `204 No Content`, bài viết được ẩn/đánh dấu xóa khỏi hệ thống.[cite: 1] |

---

## 14. UC-18: Quản lý bài trắc nghiệm (Bảng 5.19)[cite: 1]

| ID | Tên Test Case | Điều kiện (Setup Data) | Kết quả mong đợi (Expected Output) |
| :--- | :--- | :--- | :--- |
| `TC-ASM-TPL-01` | Tạo bài trắc nghiệm mới thành công (Happy Case) | Admin nhập thông tin bài trắc nghiệm mới với mã chưa từng tồn tại trong hệ thống. | Tạo bài trắc nghiệm thành công, trả về mã HTTP `200 OK` kèm thông báo thành công và mã định danh (slug) tự tạo.[cite: 1] |
| `TC-ASM-TPL-02` | Tạo bài trắc nghiệm thất bại do trùng lặp mã bài test | Admin nhập mã bài trắc nghiệm trùng với một bài test đang hoạt động trên hệ thống. | Trả về mã HTTP `400 Bad Request` kèm thông báo lỗi mã bài test đã tồn tại.[cite: 1] |
| `TC-ASM-TPL-03` | Xem thông tin chi tiết bài trắc nghiệm theo slug | Người dùng truy cập xem chi tiết bài trắc nghiệm qua mã định danh (slug) hợp lệ. | Trả về mã HTTP `200 OK` cùng đầy đủ tiêu đề, mô tả và hướng dẫn làm bài.[cite: 1] |
