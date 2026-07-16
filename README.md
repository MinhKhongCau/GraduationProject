# 🌿 Psychological Counseling System - Microservices Architecture

Hệ thống cung cấp nền tảng tư vấn tâm lý trực tuyến, giúp kết nối **Client** với các **Chuyên gia (Expert)** thông qua quy trình đặt lịch, thanh toán ví nội bộ và làm bài trắc nghiệm tâm lý có đánh giá AI hỗ trợ. Hệ thống được thiết kế theo kiến trúc Microservices để đảm bảo tính sẵn sàng cao và khả năng mở rộng linh hoạt.

---

## 🔗 Liên kết dịch vụ & Tài liệu API (Service Links & API Docs)

- **Frontend Sandbox**: [https://sandbox.qmcloud.io.vn](https://sandbox.qmcloud.io.vn)
- **Auth Service API Docs**: [https://api.qmcloud.io.vn/auth/swagger-ui/index.html#/](https://api.qmcloud.io.vn/auth/swagger-ui/index.html#/)
- **Booking Service API Docs**: [https://api.qmcloud.io.vn/booking/swagger-ui/index.html#/](https://api.qmcloud.io.vn/booking/swagger-ui/index.html#/)
- **Payment Service API Docs**: [https://api.qmcloud.io.vn/payment/swagger-ui/index.html#/](https://api.qmcloud.io.vn/payment/swagger-ui/index.html#/)
- **Assessment Service API Docs**: [https://api.qmcloud.io.vn/assessment/swagger-ui](https://api.qmcloud.io.vn/assessment/swagger-ui)
- **Profile Service API Docs**: [https://api.qmcloud.io.vn/profile/swagger-ui/index.html#/](https://api.qmcloud.io.vn/profile/swagger-ui/index.html#/)
- **Forum Service API Docs**: [https://api.qmcloud.io.vn/forum/swagger-ui/index.html#/](https://api.qmcloud.io.vn/forum/swagger-ui/index.html#/)

Tất cả traffic đi qua **API Gateway (Kong)**, gateway route request tới 6 service backend phía trên qua các prefix `/api/v1/*`.

---

## 1. Tổng quan hệ thống (System Overview)

Mục đích chính của hệ thống là cung cấp một hệ sinh thái chăm sóc sức khỏe tinh thần:

- **Đặt lịch & Tư vấn:** Quy trình đặt lịch, khóa slot, quản lý ca làm việc của Expert.
- **Ví & Thanh toán nội bộ:** Nạp/rút tiền, thanh toán buổi tư vấn qua ví (wallet) nội bộ.
- **Khoa học & AI:** Làm bài trắc nghiệm tâm lý (PHQ, ASRS, MDQ...) và nhận đánh giá gợi ý từ AI (Google Gemini) tích hợp trong Assessment Service.
- **Hồ sơ người dùng:** Quản lý hồ sơ Patient/Expert/Admin, lịch sử bệnh án, chuyên môn.

> ⚠️ **Chưa triển khai:** Chat 1-1/real-time, Diễn đàn (Forum/Community) và Thông báo (Notification) hiện **chưa có backend**, mới chỉ tồn tại dưới dạng UI/mock data ở frontend. Xem mục [9. Định hướng phát triển](#9-định-hướng-phát-triển-tương-lai-future-improvements).

**Đối tượng sử dụng:** Client (Khách hàng), Expert (Chuyên gia), Admin (Quản trị viên).

---

## 2. Mục tiêu & Giá trị cốt lõi (System Objectives)

- **Tự động hóa:** Quy trình giữ chỗ (Slot locking) và ví thanh toán nội bộ minh bạch.
- **Bảo mật:** Xác thực JWT qua Gateway (Kong custom plugin), phân quyền chi tiết (RBAC) và bảo mật dữ liệu cá nhân.
- **Hiệu năng:** Xử lý nhiều kết nối đồng thời nhờ cơ chế Microservices, mỗi service có database riêng.
- **Khả năng mở rộng:** Dễ dàng thêm mới service hoặc scale độc lập từng thành phần.

---

## 3. Kiến trúc hệ thống (System Architecture)

Hệ thống sử dụng mô hình **Client-Server** dựa trên các Microservices độc lập, đứng sau một **API Gateway**:

- **API Gateway:** [Kong](https://konghq.com/) (declarative config) tích hợp plugin JWT xác thực tự viết bằng Go (`mindcare-auth`), route request tới từng service theo path prefix, xử lý CORS/rate-limiting.
- **Communication:** Giao tiếp qua REST API (đồng bộ). Chưa có event bus / message broker nào được các service thực sự sử dụng (xem mục 6).
- **Decoupling:** Mỗi service sở hữu database riêng (theo schema `*_DB_NAME` trên cùng một Postgres instance), không truy cập trực tiếp DB của service khác.

---

## 4. Công nghệ sử dụng (Technology Stack)

### **Frontend**

- **Next.js 16 (App Router) + React 19 + TypeScript:** Xây dựng giao diện người dùng.
- **TanStack React Query:** Quản lý server state / data fetching.
- **React Hook Form + Zod:** Quản lý và validate form.
- **Radix UI + Tailwind CSS v4:** UI primitives và styling.
- **Axios**, **@react-oauth/google:** Gọi API và đăng nhập Google OAuth.
- **Capacitor:** Cấu hình sẵn cho mobile wrapper (chưa triển khai app di động thật).

### **API Gateway**

| Thành phần  | Công nghệ                                 | Vai trò                                                                    |
| :---------- | :---------------------------------------- | :------------------------------------------------------------------------- |
| **Gateway** | Kong + custom Go plugin (`mindcare-auth`) | Định tuyến request, xác thực JWT, CORS, rate-limiting cho toàn bộ backend. |

### **Backend - Microservices (đã triển khai)**

| Service                | Công nghệ                     | Trách nhiệm chính                                                                       |
| :--------------------- | :---------------------------- | :-------------------------------------------------------------------------------------- |
| **Auth Service**       | Java 21, Spring Boot          | Quản lý tài khoản người dùng, đăng ký/đăng nhập, phát hành JWT.                         |
| **Profile Service**    | Go (Gin + GORM)               | Quản lý hồ sơ Patient/Expert/Admin, chuyên môn, lịch sử bệnh án.                        |
| **Booking Service**    | Go (Gin + GORM)               | Quản lý lịch trình, khóa slot, ca làm việc/nghỉ phép, trạng thái cuộc hẹn.              |
| **Payment Service**    | Go (Gin + GORM)               | Quản lý ví nội bộ (wallet): nạp/rút, giao dịch, xử lý webhook thanh toán buổi tư vấn.   |
| **Assessment Service** | Python (FastAPI + SQLAlchemy) | Quản lý bộ câu hỏi trắc nghiệm, chấm điểm, sinh gợi ý đánh giá bằng AI (Google Gemini). |

### **Dịch vụ đã lên kế hoạch nhưng chưa triển khai (Planned, not implemented)**

| Service                     | Công nghệ dự kiến   | Trách nhiệm dự kiến                                                  |
| :-------------------------- | :------------------ | :------------------------------------------------------------------- |
| **Chat Service**            | Node.js / Socket.io | Nhắn tin real-time 1-1 giữa Client và Expert.                        |
| **Community/Forum Service** | Spring Boot         | Diễn đàn, blog, hệ thống comment tree giữa người dùng và chuyên gia. |
| **Notification Service**    | WebSocket           | Đẩy thông báo tức thời (push notification).                          |

Hiện tại các tính năng trên chỉ tồn tại dưới dạng UI với mock data ở frontend (`app/frontend/data/messages.ts`, `notifications.ts`, `app/patient/messages/`), chưa có endpoint hoặc service backend tương ứng.

---

## 5. Quản lý dữ liệu (Database & Storage)

- **PostgreSQL 16:** Cơ sở dữ liệu quan hệ chính, mỗi service (`auth`, `profile`, `booking`, `payment`, `assessment`) có database logic riêng trên cùng một Postgres instance, đảm bảo tính nhất quán (ACID).
- **Redis & RabbitMQ:** Đã được khai báo sẵn trong docker-compose (hạ tầng), nhưng **chưa có service nào thực sự kết nối/sử dụng** — dự kiến dùng cho cache và message queue trong tương lai.

---

## 6. Xử lý thời gian thực & Sự kiện (Real-time & Events)

Hiện hệ thống **chưa có** cơ chế real-time (WebSocket) hay event-driven thực sự nào được triển khai — toàn bộ giao tiếp giữa các service là REST API đồng bộ. Redis và RabbitMQ đã được provision sẵn trong hạ tầng để phục vụ cho Chat/Notification Service và xử lý sự kiện (ví dụ _thanh toán thành công_ kích hoạt cập nhật Booking) khi các service này được xây dựng.

---

## 7. Bảo mật (Security)

- **JWT (JSON Web Token):** Xác thực người dùng, được xác minh tại API Gateway (Kong) thông qua plugin Go `mindcare-auth` trước khi request tới các service downstream.
- **RBAC (Role-based Access Control):**
  - `ADMIN`: Kiểm duyệt nội dung, quản lý tài chính hệ thống.
  - `EXPERT`: Cấu hình ca làm việc, xem thu nhập và tư vấn.
  - `CLIENT`: Tìm kiếm chuyên gia, đặt lịch, làm bài test.

---

## 8. Triển khai & Mở rộng (Deployment & Scalability)

- **Containerization:** Toàn bộ service (frontend, gateway, auth, profile, booking, payment, assessment) được đóng gói bằng **Docker**, orchestrate qua `docker-compose` (`.deploy/dev`, `.deploy/product`).
- **Independent Scaling:** Có thể tăng số lượng instance của từng service độc lập mà không tốn tài nguyên cho các service khác.
- **CI/CD:** GitHub Actions build & test riêng cho từng service (Go matrix cho profile/payment/booking, Maven cho auth, pip cho assessment, npm cho frontend, build plugin cho gateway) trước khi deploy dev/staging/production.

---

## 9. Định hướng phát triển tương lai (Future Improvements)

- Xây dựng **Chat Service** (Node.js/Socket.io) cho nhắn tin real-time 1-1 giữa Client và Expert.
- Xây dựng **Community/Forum Service** (Spring Boot) cho diễn đàn, blog và comment tree.
- Xây dựng **Notification Service** (WebSocket) để đẩy thông báo tức thời, liên kết với các sự kiện booking/payment.
- Đưa **Redis** vào sử dụng thực tế cho caching, và **RabbitMQ** cho xử lý sự kiện bất đồng bộ giữa các service.
- Tích hợp cổng thanh toán bên ngoài (VNPay/MoMo) cho Payment Service (hiện chỉ có ví nội bộ).
- Phát triển ứng dụng di động (**Mobile App**) đa nền tảng dựa trên Capacitor đã cấu hình sẵn ở frontend.
- Nâng cấp AI (hiện dùng Google Gemini trong Assessment Service) để cá nhân hóa lộ trình điều trị cho từng Client.

---

© 2026 Psychological Counseling System - Technical Documentation.
