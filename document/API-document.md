# API Contract: Psychological Counseling System
***General Information***

- Base URL: http://localhost:8080/api/v1
- Date: Wednesday, March 4, 2026
- Authentication: JWT token required for protected endpoints, passed in the Authorization header as Bearer <token>.
- Roles: CLIENT, EXPERT, ADMIN
- Content-Type: application/json
- Version: v1.0
```
Enums

export enum UserRole {
  CLIENT = 'CLIENT',
  EXPERT = 'EXPERT',
  ADMIN = 'ADMIN'
}

export enum AppointmentStatus {
  PENDING = 'PENDING',
  LOCKED = 'LOCKED', // Temporary lock for 15 mins during payment
  CONFIRMED = 'CONFIRMED',
  CANCELED = 'CANCELED',
  COMPLETED = 'COMPLETED'
}

export enum PaymentStatus {
  PENDING = 'PENDING',
  SUCCESS = 'SUCCESS',
  FAILED = 'FAILED',
  REFUNDED = 'REFUNDED'
}

export enum AssessmentType {
  DEPRESSION = 'DEPRESSION',
  ANXIETY = 'ANXIETY',
  STRESS = 'STRESS'
}
```
## 1. Authentication Service
### 1.1 Register Account

- Endpoint: POST /auth/register
- Description: Create a new Client or Expert account.
- Authentication: None
```
Request Body:
{
    "email": "user@example.com",
    "password": "SecurePassword123!",
    "fullName": "Nguyen Van A",
    "role": "CLIENT",
    "phone": "0901234567"
}

Response:

    201 Created: { "message": "User registered successfully", "userId": 101 }
```
### 1.2 Login

Endpoint: POST /auth/login
Request Body: { "email": "user@example.com", "password": "password" }
Response:
    200 OK: { "token": "eyJhbGci...", "role": "CLIENT", "userId": 101 }

## 2. Booking & Expert Service
### 2.1 Search Experts

- Endpoint: GET /experts
- Description: List experts with filters.
- Query Params: specialization, rating, availabilityDate

Response: 200 OK:
```    
{
    "data": [
        {
            "expertId": 501,
            "fullName": "Dr. Smith",
            "specialization": "Clinical Psychology",
            "hourlyRate": 500000,
            "rating": 4.9
        }
    ]
}
```
### 2.2 Seek Expert Schedule

- Endpoint: GET /experts/:id/slots
- Description: Get available time slots for a specific expert.
Response: 200 OK:
```
{
    "expertId": 501,
    "availableSlots": [
    { "slotId": 1001, "startTime": "2026-03-05T09:00:00Z", "status": "AVAILABLE" }
    ]
}
```
### 2.3 Book Appointment (Locking Slot)

- Endpoint: POST /bookings/lock
- Description: Temporarily lock a slot for 15 minutes to proceed with payment.
- Permissions: CLIENT
```
Request Body: { "slotId": 1001, "expertId": 501 }
Response:
201 Created: { "bookingId": 999, "lockExpiresAt": "2026-03-04T19:15:00Z" }
```
## 3. Assessment & AI Service
### 3.1 Submit Test Results

- Endpoint: POST /assessments/submit
- Description: Client submits answers; Assessment Service calls AI Service for analysis.
- Permissions: CLIENT
```
Request Body:
{
  "testType": "DEPRESSION",
  "answers": [ { "questionId": 1, "score": 3 }, { "questionId": 2, "score": 4 } ]
}
Response:

    200 OK:
    JSON

        {
          "score": 15,
          "severity": "MODERATE",
          "ai_recommendation": "It is recommended to speak with an expert specializing in CBT.",
          "suggestedExperts": [501, 505]
        }
```

## 4. Payment Service
### 4.1 Create Payment Intent

- Endpoint: POST /payments/create
- Description: Integrate with VNPay/MoMo to pay for a locked booking.
- Request Body: { "bookingId": 999, "paymentGateway": "VNPAY" }
- Response:
- 200 OK: 
```
{ "paymentUrl": "https://vnpay.vn/pay?..." }
```
## 5. Chat & Notification Service
### 5.1 Send Message (REST Fallback)

- Endpoint: POST /chat/messages
- Description: Send a message in a 1-1 session (Primary communication is via Socket.io).
```
Request Body: 

{ "receiverId": 501, "content": "Hello Doctor" }
Response:
    200 OK:
    JSON
    { "message": "Message sent successfully" }
```
### 5.2 Get Notifications

- Endpoint: GET /notifications
- Description: Fetch unread notifications for the user.
- Response: 200 OK

## 📋 API Summary

| Service     | Endpoint              | Method | Permission | Description              |
|------------|----------------------|--------|------------|--------------------------|
| Auth        | `/auth/login`         | POST   | Public     | Authenticate user        |
| Booking     | `/experts`            | GET    | Public     | List counselors          |
| Booking     | `/bookings/lock`      | POST   | CLIENT     | Reserve slot (15m)       |
| Assessment  | `/assessments/submit` | POST   | CLIENT     | Submit test to AI        |
| Payment     | `/payments/create`    | POST   | CLIENT     | Pay for booking          |
| Community   | `/forum/posts`        | GET    | Public     | View blog/discussions    |

---

## ⚙️ Implementation Notes

- **Slot Locking:** Slot is reserved for 15 minutes. If payment is not completed successfully, the slot is released automatically.
- **AI Processing:** Assessment data is sent to AI Service (Python) via REST API and results are stored in the database.
- **Payment Flow:** Payment Service integrates with external gateways (VNPay, MoMo) and updates booking status after confirmation.
- **Security:** All protected endpoints require JWT validation.
- Slot Locking Logic: The Booking Service must implement a TTL (Time-To-Live) mechanism using Redis or a scheduled task to unlock slots if payment is not confirmed within 15 minutes.

- AI Integration: The AI Service (Python) should be called asynchronously via a Feign Client or WebClient from the Assessment Service.

- Real-time: Ensure the Notification Service pushes a BOOKING_CONFIRMED event via WebSocket as soon as the Payment Service receives a success callback from the gateway.
