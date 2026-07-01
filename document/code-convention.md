# Quy tắc Viết Code (Code Convention)

Tài liệu này định nghĩa các tiêu chuẩn viết code được áp dụng thống nhất trong toàn bộ dự án **MindCare**. Mọi thành viên cần tuân thủ nghiêm ngặt các quy tắc dưới đây.

---

## 1. Quy tắc Đặt tên File & Thư mục (Naming Convention)

* **Tên Thư mục & File chính (Backend/General)**: Sử dụng kiểu **kebab-case** (chữ thường, ngăn cách bằng dấu gạch ngang).
  * *Ví dụ*: `assessment-calendar`, `booking-service`, `auth-service`.
* **Tên File JavaScript / TypeScript (Frontend)**: Sử dụng kiểu **camelCase** hoặc chữ thường tùy thuộc vào vai trò.
  * *Ví dụ*: `calendar.js`, `authController.ts`.
* **Quy tắc chuyển đổi mẫu**:
  * Thư mục/Tính năng: `assessment-calendar`
  * Interface tương ứng: `Calendar`
  * File JS/TS xử lý: `calendar.js`

---

## 2. Quy tắc Đặt tên Biến (Variables)

* Sử dụng kiểu **camelCase** (chữ cái đầu viết thường, viết hoa chữ cái đầu của các từ tiếp theo).
* Không sử dụng dấu gạch dưới (`_`) làm phân tách trong tên biến.
* *Ví dụ*:
  * Đúng: `appointmentDate`, `userId`, `expertProfile`
  * Sai: `appointment_date`, `user_id`, `ExpertProfile`

---

## 3. Quy tắc Định nghĩa Interface

### 3.1. Định dạng Tên Interface
* Tên Interface phải viết hoa chữ cái đầu của mỗi từ (**PascalCase**).
* *Ví dụ*: `Calendar`, `ExpertTimeOff`, `AccountRepository`.

### 3.2. Interface cho Yêu cầu (Request) và Phản hồi (Response)
* **Bắt buộc** định nghĩa các Interface cụ thể cho dữ liệu đầu vào (Request payload) và dữ liệu đầu ra (Response payload) để đảm bảo tính an toàn kiểu (type safety) và kiểm chuẩn dữ liệu (validation).
* Tên của các interface này phải kết thúc bằng hậu tố `Request` hoặc `Response`.
* *Ví dụ*:
  ```typescript
  interface LoginRequest {
    email: string;
    password: string;
  }

  interface LoginResponse {
    token: string;
    refreshToken: string;
    expiresIn: number;
  }
  ```

---

## 4. Các Quy tắc Bổ sung

* **Hàm (Functions / Methods)**: Sử dụng **camelCase** (ví dụ: `getExpertById`, `calculateTotalAmount`).
* **Hằng số (Constants)**: Sử dụng **UPPER_SNAKE_CASE** (ví dụ: `MAX_RETRY_COUNT`, `DEFAULT_PAGE_SIZE`).
* **Lớp (Classes)**: Sử dụng **PascalCase** (ví dụ: `AuthService`, `PaymentController`).
