# Hướng Dẫn Sử Dụng API & Đăng Nhập (Cho Team Phát Triển)

Tài liệu này hướng dẫn các thành viên trong team cách truy cập tài liệu API (Swagger UI) và sử dụng các tài khoản test mặc định của hệ thống.

## 1. Truy cập Swagger UI

Hệ thống cung cấp Swagger UI để tra cứu và test API trực tiếp trên trình duyệt.

- **Khi chạy qua API Gateway (Khuyên dùng)**: 
  👉 [http://localhost:8000/auth/swagger-ui/index.html](http://localhost:8000/auth/swagger-ui/index.html)
  *(Đảm bảo cả API Gateway và Auth Service đều đang chạy)*

- **Khi chạy Local (Chạy trực tiếp Auth Service)**:
  👉 [http://localhost:8080/swagger-ui/index.html](http://localhost:8080/swagger-ui/index.html)



## 2. Tài khoản Đăng nhập Test (Mặc định)

Khi ứng dụng khởi động lần đầu, Database sẽ tự động sinh ra 3 tài khoản test đại diện cho 3 quyền (Role) khác nhau. 

**Quy tắc chung: MẬT KHẨU GIỐNG HỆT TÊN ĐĂNG NHẬP (Tài khoản = Mật khẩu)**

Dưới đây là danh sách tài khoản:

| Vai trò (Role) | Tài khoản đăng nhập (Email) | Mật khẩu (Password) |
| :--- | :--- | :--- |
| **Quản trị viên** (ADMIN) | `admin@mindcare.com` | `admin@mindcare.com` |
| **Chuyên gia/Bác sĩ** (EXPERT) | `expert@mindcare.com` | `expert@mindcare.com` |
| **Bệnh nhân/Khách hàng** (PATIENT) | `patient@mindcare.com` | `patient@mindcare.com` |

---

## 3. Cách Test API cần Xác thực (Authorization) trên Swagger

1. Gửi request POST tới endpoint `/api/v1/auth/login` với email và password ở trên.
2. Trong body response trả về, copy chuỗi token (thường nằm trong trường `accessToken`).
3. Kéo lên đầu trang Swagger, bấm vào nút **Authorize** (ổ khóa màu xanh lá).
4. Dán chuỗi token vừa copy vào ô **Value** và bấm **Authorize**.
5. Bây giờ bạn có thể test các API yêu cầu quyền đăng nhập tương ứng.
