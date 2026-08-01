# MindCare — Frontend

Next.js frontend for **MindCare**, a psychological-counseling platform built on a microservices backend (`app/backend/{auth,booking,profile,payment,assessment}-service`). This app's UI/feature scope was migrated from a legacy PHP reference app, [`../edoc-doctor-appointment-system`](../edoc-doctor-appointment-system) (a generic doctor-appointment booking system), adapted to MindCare's actual roles and domain (psychological counseling rather than general medicine) and to what the real backend currently supports.

For architecture decisions (why things are structured this way, what's mocked and why), see **[DESIGN.md](./DESIGN.md)**.

## Getting started

```bash
npm install
cp .env.local.example .env.local   # fill in service URLs / feature flags
npm run dev
```

Open [http://localhost:3000](http://localhost:3000). By default the app expects the local Kong API gateway at port 8000 (`NEXT_PUBLIC_API_URL`) and the chatroom Socket.IO service at port 8085 (`NEXT_PUBLIC_CHATROOM_WS_URL`).

### Capacitor development

`package.json` is strict JSON, so it cannot contain comments. The Capacitor scripts below are documented here instead:

- `npm run cap:sync:ios` syncs iOS with `http://localhost:3000`, which iOS Simulator resolves to the host machine.
- `npm run cap:sync:android` syncs Android with `http://10.0.2.2:3000`, Android Emulator's host-machine alias.
- `npm run cap:sync` runs both platform-specific syncs.
- `npm run cap:dev` syncs both platforms, then starts Next.js at `0.0.0.0` so emulators and physical devices can reach it.

```bash
npm run lint       # eslint
npm run test:run   # vitest, once
npm run test       # vitest, watch mode
npm run build       # production build (output: "standalone", for Docker)
```

## Roles

The backend's `auth-service` defines three account roles: `PATIENT`, `EXPERT`, `ADMIN`. (Note: some docs elsewhere in this repo, e.g. `document/API-document.md` and the root README, use `CLIENT` instead of `PATIENT` — the running code is the source of truth; see DESIGN.md.)

## Feature / page checklist

Legend: **Full** = real UI wired to a hook/API call (real backend where implemented, an in-memory `/data` mock otherwise — see DESIGN.md). **Stub** = routed placeholder page ("Coming soon"), not built this pass. **edoc source** = the legacy PHP page(s) this was migrated from, where applicable.

### Public

| Page | Route | Status | edoc source |
|---|---|---|---|
| Landing / marketing home | `/` | Full | `index.html` |
| Login | `/auth/login` | Full | `login.php` |
| Register (single-step wizard; edoc split this into 2 pages) | `/auth/register` | Full | `signup.php` + `create-account.php` |
| Forgot password | `/auth/forgot-password` | Full (backend endpoint documented, not implemented yet — see DESIGN.md) | *(not in edoc)* |
| Reset password | `/auth/reset-password` | Full (same caveat) | *(not in edoc)* |
| Google sign-in | *(in login/register)* | Full, feature-flagged off by default (`NEXT_PUBLIC_ENABLE_GOOGLE_AUTH`) — backend has no `/auth/google` handler yet | *(not in edoc)* |
| Logout | *(header/sidebar action)* | Full | `logout.php` |

### Patient portal (`/patient/*`)

| Page | Route | Status | edoc source |
|---|---|---|---|
| Dashboard | `/patient/dashboard` | Full | `patient/index.php` |
| Find experts (browse/search/filter by specialization) | `/patient/find-experts` | Full | `patient/doctors.php` |
| Expert detail | `/patient/experts/[expertId]` | Full | *(doctor "View" popup in `patient/doctors.php`)* |
| Book appointment (topic → expert → slot → review w/ queue # & fee → confirm → success) | `/patient/book-appointment` | Full — confirm step uses the `/data` mock (booking-service has no create/lock endpoint yet) | `patient/schedule.php` → `booking.php` → `booking-complete.php` |
| My Bookings (history + cancel) | `/patient/my-bookings` | Full — backed by the `/data` mock (no booking-history endpoint yet) | `patient/appointment.php` + `patient/delete-appointment.php` |
| Wallet (balance, top-up, withdraw, transactions) | `/patient/wallet` | Full — balance/top-up/withdraw are real payment-service calls; transaction list is a `/data` mock (no history endpoint exists) | *(not in edoc — new for MindCare)* |
| Assessment list | `/patient/assessment` | Full | *(not in edoc — new for MindCare)* |
| Take assessment + result | `/patient/assessment/[templateId]` | Full | *(not in edoc — new for MindCare)* |
| Medical history / health profile | `/patient/medical-history` | Full | *(not in edoc — new for MindCare)* |
| Messages (chat) | `/patient/messages` | Full — real-time 1:1 DM + voice messages via `chatroom-service` (Socket.IO) | *(not in edoc)* |
| Community forum (browse/post/comment/like/bookmark) | `/patient/forum` | Full — real `forum-service` calls; no post edit/delete UI yet | *(not in edoc)* |
| Settings (profile edit, change password, delete account) | `/patient/settings` | Full — delete-account is disabled (no backend endpoint) | `patient/settings.php` + `edit-user.php` |

### Expert portal (`/expert/*`)

| Page | Route | Status | edoc source |
|---|---|---|---|
| Dashboard | `/expert/dashboard` | Full | `doctor/index.php` |
| Weekly schedule / availability | `/expert/schedule` | Full — backed by the `/data` mock (no weekly-schedule endpoint yet) | `admin/schedule.php` (session mgmt was admin-only in edoc; MindCare gives experts self-service availability instead) |
| My appointments | `/expert/appointments` | Full | `doctor/appointment.php` |
| My patients | `/expert/patients` | **Stub** | `doctor/patient.php` |
| Messages (chat) | `/expert/messages` | Full — real-time 1:1 DM + voice messages via `chatroom-service` (Socket.IO) | *(not in edoc)* |
| Community forum (browse/post/comment/like/bookmark) | `/expert/forum` | Full — real `forum-service` calls; no post edit/delete UI yet | *(not in edoc)* |
| Clinical records (create/view per patient) | `/expert/clinical-records` | **Stub** | *(not in edoc — new for MindCare)* |
| Wallet / earnings | `/expert/wallet` | **Stub** | *(not in edoc)* |
| Settings | `/expert/settings` | **Stub** | `doctor/settings.php` |

### Admin portal (`/admin/*`)

| Page | Route | Status | edoc source |
|---|---|---|---|
| Dashboard (stat cards) | `/admin/dashboard` | Full — expert/specialization counts are real, patient count has no backend source and stays a labeled placeholder | `admin/index.php` |
| Experts (CRUD: add/edit/view/remove) | `/admin/experts` | **Stub** | `admin/doctors.php` + `add-new.php` + `edit-doc.php` + `delete-doctor.php` |
| Specializations (lookup CRUD) | `/admin/specializations` | **Stub** | edoc's `specialties` table (no dedicated admin UI in edoc — it was seeded directly) |
| Patients (list/view) | `/admin/patients` | **Stub** | `admin/patient.php` |
| Appointments (global list, cancel) | `/admin/appointments` | **Stub** | `admin/appointment.php` |
| Schedules (global session/slot overview) | `/admin/schedules` | **Stub** | `admin/schedule.php` + `add-session.php` + `delete-session.php` |
| Withdrawal approvals | `/admin/withdrawals` | **Stub** | *(not in edoc — new for MindCare, maps to payment-service's withdrawal-processing endpoint)* |

### Fallback pages

`/forbidden` (wrong role) and the default Next.js `not-found` (404) — both full.

## What's intentionally not migrated from edoc

The legacy app had several bugs that were deliberately **not** ported — see "Known bugs deliberately not carried over" in [DESIGN.md](./DESIGN.md) (SQL injection, plaintext passwords, missing server-side ownership checks, missing session-capacity checks, orphaned bookings on delete, auth checks without an early return).

## Known gaps / next pass

- All **Stub** pages above (expert: patients/clinical-records/wallet/settings; admin: experts/specializations/patients/appointments/schedules/withdrawals).
- i18n coverage: navigation, the booking-topic grid, and the landing page are fully translated (EN/VI); most other page body copy is still English-only text, not yet routed through `t()`.
- No native iOS/Android Capacitor projects yet (`npx cap add ios|android` not run — see DESIGN.md's Capacitor section).
- Google OAuth and forgot/reset-password are wired frontend-side but call backend endpoints that don't exist yet (`auth-service` has no `/auth/google`, `/auth/forgot-password`, or `/auth/reset-password` handlers) — see DESIGN.md.
- Booking confirm/history, expert weekly-schedule, wallet transaction-history, and clinical records all run against `/data` mocks pending the corresponding backend endpoints.
