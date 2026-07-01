# Quy tắc Sử dụng Git (Git Convention)

Tài liệu này quy định luồng làm việc với Git, cách đặt tên nhánh (branch) và cấu trúc thông điệp commit (commit message) áp dụng cho dự án **MindCare**.

---

## 1. Quy tắc Đặt tên Nhánh (Branch Naming)

Mọi nhánh phát triển tính năng mới hoặc chỉnh sửa code phải tuân thủ định dạng sau:

- **Tính năng mới (Features)**:
  - Định dạng: `features/PSY-<task_id>-<description>`
  - _Ví dụ_: `features/PSY-102-integrate-payment-gateway`
- **Cấu trúc lại code (Refactoring)**:
  - Định dạng: `refactor/PSY-<task_id>-<description>`
  - _Ví dụ_: `refactor/PSY-105-optimize-database-queries`

_Lưu ý_: `<description>` viết thường không dấu, các từ cách nhau bằng dấu gạch ngang `-`.

---

## 2. Hệ thống Nhánh Chính (Core Branches)

Dự án sử dụng 3 nhánh chính ứng với các môi trường hoạt động cụ thể:

| Tên Nhánh | Mục đích sử dụng                                                                     | Môi trường triển khai             |
| :-------- | :----------------------------------------------------------------------------------- | :-------------------------------- |
| `main`    | Nhánh tích hợp chính để chạy quy trình kiểm thử tự động (**Build & Test**).          | Môi trường CI/CD (GitHub Actions) |
| `staging` | Nhánh chứa mã nguồn ổn định dùng thử nghiệm trước khi bàn giao (**Sandbox Server**). | Sandbox / Staging Server          |
| `product` | Nhánh chứa mã nguồn chính thức đang chạy trên môi trường thực tế (**Production**).   | Production Server                 |

---

## 3. Quy tắc Viết Commit Message

Mỗi commit đẩy lên kho lưu trữ phải tuân theo cấu trúc mẫu tiêu chuẩn dưới đây:

```text
Task Id: PSY-<task id>
Type: <Type>
Description:
- <description 1>
- <description 2>
```

### Các trường thông tin:

- **Task Id**: Định danh của thẻ công việc trên Jira/Trello (ví dụ: `PSY-102`).
- **Type**: Loại thay đổi (ví dụ: `Development`, `Refactor`, `Bugfix`, `Documentation`, `CI/CD`).
- **Description**: Danh sách các gạch đầu dòng mô tả chi tiết, ngắn gọn những thay đổi được thực hiện trong commit.

### Ví dụ mẫu commit hoàn chỉnh:

```text
Task Id: PSY-102
Type: Development
Description:
- Thêm API thanh toán Momo cho payment-service
- Định nghĩa interface PayRequest và PayResponse nhận dữ liệu
- Bổ sung cấu hình kết nối Redis
```
