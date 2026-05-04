# 🌿 Psychological Counseling System - Microservices Architecture

Hệ thống cung cấp nền tảng tư vấn tâm lý trực tuyến toàn diện, giúp kết nối **Client** với các **Chuyên gia (Expert)** thông qua quy trình đặt lịch, thanh toán và tư vấn thời gian thực. Hệ thống được thiết kế theo kiến trúc Microservices để đảm bảo tính sẵn sàng cao và khả năng mở rộng linh hoạt.

---

## 1. Tổng quan hệ thống (System Overview)
Mục đích chính của hệ thống là cung cấp một hệ sinh thái chăm sóc sức khỏe tinh thần khép kín:
* **Đặt lịch & Tư vấn:** Quy trình đặt lịch tự động, thực hiện tư vấn trực tuyến.
* **Tương tác Real-time:** Hệ thống Chat 1-1 và thông báo tức thời.
* **Khoa học & AI:** Làm bài trắc nghiệm tâm lý và nhận đánh giá từ AI Server.
* **Cộng đồng:** Diễn đàn chia sẻ kiến thức, thảo luận giữa người dùng và chuyên gia.

**Đối tượng sử dụng:** Client (Khách hàng), Expert (Chuyên gia), Admin (Quản trị viên).

---

## 2. Mục tiêu & Giá trị cốt lõi (System Objectives)
* **Tự động hóa:** Quy trình giữ chỗ (Slot locking) và thanh toán minh bạch.
* **Bảo mật:** Xác thực JWT, phân quyền chi tiết (RBAC) và bảo mật dữ liệu cá nhân.
* **Hiệu năng:** Xử lý hàng nghìn kết nối đồng thời nhờ cơ chế Microservices.
* **Khả năng mở rộng:** Dễ dàng thêm mới service hoặc scale độc lập từng thành phần.

---

## 3. Kiến trúc hệ thống (System Architecture)
Hệ thống sử dụng mô hình **Client-Server** dựa trên các Microservices độc lập:
* **Communication:** Giao tiếp qua REST API (đồng bộ) và WebSockets (bất đồng bộ/thời gian thực).
* **Decoupling:** Các dịch vụ không phụ thuộc trực tiếp vào nhau, đảm bảo tính ổn định của toàn hệ thống.

---

## 4. Công nghệ sử dụng (Technology Stack)

### **Frontend**
* **React:** Xây dựng giao diện người dùng Single Page Application (SPA).
* **WebSocket Client:** Duy trì kết nối real-time cho Chat và Notification.

### **Backend - Microservices**
| Service | Công nghệ | Trách nhiệm chính |
| :--- | :--- | :--- |
| **Auth Service** | Spring Boot | Quản lý User, Expert, JWT và phân quyền. |
| **Booking Service** | Go | Quản lý lịch trình, khóa Slot (15p), trạng thái cuộc hẹn. |
| **Community Service** | Spring Boot | Quản lý Forum, Blog và hệ thống Comment Tree. |
| **Assessment Service** | Python (FastAPI) | Quản lý bộ câu hỏi trắc nghiệm và lưu trữ kết quả. |
| **Chat Service** | Node.js / Socket.io | Xử lý tin nhắn Real-time 1-1 và Group. |
| **Notification Service** | WebSocket | Đẩy thông báo tức thời (Push notifications). |
| **Payment Service** | Go | Xử lý nạp/rút tiền, tích hợp cổng VNPay/MoMo. |
| **AI Service** | Python (FastAPI) | Phân tích kết quả trắc nghiệm bằng mô hình AI. |

---

## 5. Quản lý dữ liệu (Database & Storage)
* **PostgreSQL:** Cơ sở dữ liệu quan hệ chính đảm bảo tính nhất quán (ACID) cho giao dịch và booking.
* **Data Integrity:** Thiết kế chuẩn hóa dữ liệu, hỗ trợ mở rộng sang Redis (Cache) để tăng tốc độ truy vấn.

---

## 6. Xử lý thời gian thực & Sự kiện (Real-time & Events)
* **WebSockets:** Sử dụng cho Chat và Notification để đảm bảo độ trễ thấp nhất.
* **Event-driven:** Các sự kiện quan trọng như *Thanh toán thành công* sẽ kích hoạt các logic liên quan tại Booking Service và Notification Service.

---

## 7. Bảo mật (Security)
* **JWT (JSON Web Token):** Xác thực người dùng trên mọi yêu cầu API.
* **RBAC (Role-based Access Control):** * `ADMIN`: Kiểm duyệt nội dung, quản lý tài chính hệ thống.
    * `EXPERT`: Cấu hình ca làm việc, xem thu nhập và tư vấn.
    * `CLIENT`: Tìm kiếm chuyên gia, đặt lịch, làm bài test và thảo luận.

---

## 8. Triển khai & Mở rộng (Deployment & Scalability)
* **Containerization:** Toàn bộ service được đóng gói bằng **Docker**.
* **Independent Scaling:** Có thể tăng số lượng instance của Chat Service hoặc Payment Service vào giờ cao điểm mà không tốn tài nguyên cho các service khác.
* **CI/CD:** Quy trình triển khai tự động giúp hệ thống luôn được cập nhật liên tục.

---

## 9. Định hướng phát triển tương lai (Future Improvements)
* Tích hợp **API Gateway** (Spring Cloud Gateway) để quản lý luồng tập trung.
* Sử dụng **Message Queue** (Kafka/RabbitMQ) để xử lý dữ liệu lớn.
* Phát triển ứng dụng di động (**Mobile App**) đa nền tảng.
* Nâng cấp AI để cá nhân hóa lộ trình điều trị cho từng Client.

---
© 2026 Psychological Counseling System - Technical Documentation.
