# API Contract: Psychological Counseling System

***General Information***

* **Base URL:** `http://localhost:8080/api/v1`

* **Date:** Wednesday, March 4, 2026

* **Version:** `v1.4`

* **Authentication:** JWT token is required for protected endpoints.
  The token must be passed in the request header:

  ```
  Authorization: Bearer <token>
  ```

* **Roles:** `CLIENT`, `EXPERT`, `ADMIN`

* **Content-Type:** `application/json`

---

# 0. Enums & Shared Schemas

```ts
export enum UserRole {
  CLIENT = 'CLIENT',
  EXPERT = 'EXPERT',
  ADMIN = 'ADMIN'
}

export enum AppointmentStatus {
  PENDING = 'PENDING',
  LOCKED = 'LOCKED',     // Temporarily locked for 15 minutes during payment
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

export enum LeaveRequestStatus {
  PENDING = 'PENDING',
  APPROVED = 'APPROVED',
  REJECTED = 'REJECTED'
}
```

---

# 1. Authentication Service

## 1.1 Register & Login

### Register

```
POST /auth/register
```

Create a new account (Client or Expert).

---

### Login

```
POST /auth/login
```

Authenticate a user and return a **JWT token**.

---

### Google OAuth Login

```
POST /auth/google
```

Authenticate using **Google OAuth2**.

---

## 1.2 Password Management

### Forgot Password

```
POST /auth/forgot-password
```

Send a password reset link to the user's email.

---

### Reset Password

```
POST /auth/reset-password
```

Update the password using a valid reset token.

---

# 2. Expert & Booking Service

## 2.1 Search Experts

```
GET /experts
```

Retrieve a list of experts with optional filters:

* `specialization`
* `rating`
* `availabilityDate`

---

# 2.2 Expert Profile Management

### Get Current Expert Profile

```
GET /experts/me/profile
```

Retrieve the profile information of the currently authenticated expert.

---

### Create / Update Expert Profile

```
PUT /experts/me/profile
```

Create or update the expert’s professional profile.

**Permissions:** `EXPERT`

**Request Body Example**

```json
{
  "bio": "Psychological counselor with 10 years of experience in behavioral therapy.",
  "specializations": ["Clinical Psychology", "CBT", "Family Therapy"],
  "experienceYears": 10,
  "hourlyRate": 550000,
  "education": "Master of Psychology - University of Social Sciences and Humanities",
  "languages": ["Vietnamese", "English"]
}
```

---

# 2.3 Schedule Management

```
POST /experts/me/schedule
```

Create or update the expert’s **weekly working schedule**.

**Permissions:** `EXPERT`

**Request Body Example**

```json
{
  "weeklySlots": [
    { "dayOfWeek": "MONDAY", "startTime": "08:00", "endTime": "12:00" },
    { "dayOfWeek": "WEDNESDAY", "startTime": "14:00", "endTime": "18:00" }
  ]
}
```

---

# 2.4 Sudden Leave Request

```
POST /experts/leave-requests
```

Submit a sudden leave request for emergency situations.

**Permissions:** `EXPERT`

**Request Body Example**

```json
{
  "startDate": "2026-03-10T00:00:00Z",
  "endDate": "2026-03-11T23:59:59Z",
  "reason": "Family emergency",
  "cancelAffectedBookings": true
}
```

**Response**

```json
{
  "requestId": 45,
  "status": "PENDING",
  "affectedAppointments": 3
}
```

---

# 2.5 Booking Actions

### Lock Appointment Slot

```
POST /bookings/lock
```

Reserve a time slot for **15 minutes** while the client completes the payment process.

**Permissions:** `CLIENT`

---

### Booking History

```
GET /bookings/history
```

Retrieve appointment history.

**Permissions**

* `CLIENT`
* `EXPERT`

---

# 3. Clinical Records & Assessment

## 3.1 Assessment Templates & Questions

### Get Assessment Templates

```
GET /assessments/templates
```

Retrieve all active psychological assessment templates.

**Permissions:** `Public`

---

### Create Assessment Template

```
POST /assessments/templates
```

Create a new assessment template.

**Permissions:** `Public` / `ADMIN`

**Request Body Example**

```json
{
  "code": "DASS_21",
  "title": "DASS-21 Assessment",
  "description": "Depression, Anxiety and Stress Scale - 21 Items"
}
```

---

### Create Option Group

```
POST /assessments/option-groups
```

Create a new option group (response options with scores) for questions.

**Permissions:** `Public` / `ADMIN`

**Request Body Example**

```json
{
  "group_code": "OSS_2",
  "group_name": "OSS_2",
  "description": "Yes/No Option Group",
  "options": [
    {
      "label": "1",
      "value": "Co",
      "score_value": "10",
      "order_index": 1
    },
    {
      "label": "2",
      "value": "Khong",
      "score_value": "0",
      "order_index": 2
    }
  ]
}
```

---

### Bulk Create Questions

```
POST /assessments/questions/bulk
```

Bulk import questions associated with a specific template and option group.

**Permissions:** `Public` / `ADMIN`

**Request Body Example**

```json
{
  "template_id": "99cd6a3e-1e1e-4ec0-9d2f-b78d3a65e7c6",
  "group_id": "2badb8be-9368-4297-976b-b6c554b0d148",
  "questions": [
    {
      "content": "Have you found it difficult to wind down?",
      "dimension": "STRESS",
      "question_order": 0
    }
  ]
}
```

---

### Get Questions by Template

```
GET /assessments/templates/:templateId/questions
```

Retrieve all questions and their response options for a specific template.

**Permissions:** `Public`

---

## 3.2 Submit Psychological Test

```
POST /assessments/submit
```

Submit psychological assessment answers for **AI analysis and recommendations**.

**Permissions:** `CLIENT`

---

# 3.2 Clinical Records

## 3.2.1 Create Clinical Record

```
POST /clinical-records
```

Create clinical notes after a counseling session.

**Permissions:** `EXPERT`

**Request Body**

```json
{
  "appointmentId": 998,
  "clientId": 101,
  "diagnosis": "Generalized Anxiety Disorder (GAD)",
  "notes": "The patient shows persistent anxiety symptoms and requires further monitoring.",
  "recommendations": "Practice daily meditation and schedule a follow-up session in two weeks.",
  "attachments": [
    "https://storage.link/record_75.pdf"
  ]
}
```

---

## 3.2.2 Get Client Medical History (Expert)

```
GET /clinical-records/client/:clientId
```

Allow an expert to view the full medical history of a specific client.

**Permissions:** `EXPERT`

---

## 3.2.3 Get My Medical History (Client)

```
GET /clinical-records/my-history
```

Allow a client to view their own medical history and recommendations.

**Permissions:** `CLIENT`

---

# 4. Payment Service

### Create Payment Request

```
POST /payments/create
```

Create a payment request for a **locked appointment**.

Supported payment gateways:

* **VNPay**
* **MoMo**

**Permissions:** `CLIENT`

---

# API Summary Table

| Service  | Endpoint                       | Method    | Permission     | Description                   |
| -------- | ------------------------------ | --------- | -------------- | ----------------------------- |
| Auth     | `/auth/login`                  | POST      | Public         | Authenticate user             |
| Expert   | `/experts`                     | GET       | Public         | Search expert list            |
| Expert   | `/experts/me/profile`          | GET / PUT | EXPERT         | View / update profile         |
| Expert   | `/experts/me/schedule`         | POST      | EXPERT         | Configure weekly schedule     |
| Expert   | `/experts/leave-requests`      | POST      | EXPERT         | Submit sudden leave request   |
| Booking  | `/bookings/lock`               | POST      | CLIENT         | Lock slot for 15 minutes      |
| Booking  | `/bookings/history`            | GET       | CLIENT, EXPERT | View booking history          |
| Clinical | `/assessments/templates`       | GET       | Public         | Get all assessment templates  |
| Clinical | `/assessments/templates`       | POST      | ADMIN          | Create assessment template    |
| Clinical | `/assessments/option-groups`   | POST      | ADMIN          | Create option group           |
| Clinical | `/assessments/questions/bulk`  | POST      | ADMIN          | Bulk create questions         |
| Clinical | `/assessments/templates/:id/questions` | GET | Public       | Get questions by template     |
| Clinical | `/assessments/submit`          | POST      | CLIENT         | Submit psychological test     |
| Clinical | `/clinical-records`            | POST      | EXPERT         | Create clinical record        |
| Clinical | `/clinical-records/my-history` | GET       | CLIENT         | View personal medical history |
| Clinical | `/clinical-records/client/:id` | GET       | EXPERT         | View client medical history   |
| Payment  | `/payments/create`             | POST      | CLIENT         | Create payment request        |
| System   | `/notifications`               | GET       | ALL            | Retrieve notifications        |

---

# Business Logic: Sudden Leave Handling

When an expert submits a sudden leave request, the system performs the following steps:

### 1. Scan Appointments

The system scans all **CONFIRMED appointments** within the leave period.

---

### 2. Cancel & Refund

If

```
cancelAffectedBookings = true
```

The system will:

* Change appointment status to `CANCELED`
* Trigger a **refund request** to the Payment Service
* Process automatic refunds

---

### 3. Notify Clients

The system sends notifications via:

* **Push Notification**
* **Email**

to all affected clients.

---

### 4. Update Slots

All available slots during the leave period will be changed from

```
AVAILABLE
```

to

```
UNAVAILABLE
```
