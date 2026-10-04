# MindCare Frontend — Design Document

MindCare is the Next.js frontend for a psychological-counseling microservices platform (`app/backend/{auth,booking,profile,payment,assessment}-service`). This document explains the architecture decisions behind the codebase — what to read before making structural changes. For the feature/page inventory and getting-started instructions, see [README.md](./README.md).

The functional spec for the patient/expert/admin flows was migrated from a legacy PHP reference app at `../edoc-doctor-appointment-system` (kept in the repo for context only — it is not run or deployed).

## Tech stack

Next.js 16 (App Router) · React 19 · TypeScript · Tailwind CSS v4 · TanStack React Query v5 · axios · react-hook-form + zod · Radix UI primitives (Dialog, Dropdown Menu) · Vitest + Testing Library · Capacitor (config only, see below) · `@react-oauth/google`.

**Next.js 16 note**: this version renamed the `middleware.ts` file convention to `proxy.ts` (same purpose — the root `proxy.ts` in this repo, not `middleware.ts`). If you're used to older Next.js docs, check `node_modules/next/dist/docs/` before assuming a convention still applies — see `AGENTS.md`.

## Ground truth about the backend (why some things are mocked)

Roles are `PATIENT` / `EXPERT` / `ADMIN` — this is `auth-service`'s actual `Account.role` enum. Note that `document/API-document.md` and the root Vietnamese `README.md` say `CLIENT` instead of `PATIENT`; the running code is the source of truth here, not those docs.

All services sit behind a Kong API gateway at a single base URL (`NEXT_PUBLIC_API_URL`, for example `https://api.qmcloud.io.vn/api/v1`), not on individual ports — `api/http/instances.ts` points every client at that one `baseURL`. Not every documented endpoint is implemented backend-side yet:

| Area | Status |
|---|---|
| auth-service: register/login/refresh/logout/me/profile/change-password | **Implemented** |
| auth-service: Google OAuth (`POST /auth/google`), forgot/reset password | Documented only — not implemented. The frontend calls them anyway (behind `NEXT_PUBLIC_ENABLE_GOOGLE_AUTH` for Google) and handles the failure gracefully. |
| profile-service: patient/expert profiles, medical histories, specializations | **Implemented** |
| payment-service: wallet init/get/top-up/pay/withdraw/process-withdrawal | **Implemented**. No VNPay/MoMo gateway integration exists — it's an internal ledger only. No transaction-history listing endpoint. |
| booking-service: available-dates/times, slot lock, appointment create/list/cancel (patient + expert) | **Implemented**. Patient's own bookings and an expert's own appointments each have a real endpoint; there's no admin-wide "all appointments" listing. No queue-number concept — that was a mock-only idea, dropped once the real API was wired in. `Appointment` has no `expertName`/`topic` (no join to profile-service), so pages resolve expert names client-side via `expertApi.getExpertProfile`. |
| booking-service: weekly-schedule POST, leave-requests | Documented only for that exact shape — the real service models this differently, as shift `templates` + per-expert `availabilities` + `time-off` (see its swagger). `expert/schedule` still targets the old shape and stays mocked. |
| clinical records (any service) | Not implemented at all. |
| chat / messaging (`chatroom-service`, Node/Socket.IO + Redis) | **Implemented**: 1:1 patient↔expert DM (text + recorded voice messages), typing indicator, emoji reactions, poll-based online presence. No live WebRTC calls and no public multi-user rooms (out of scope — the reference project this was ported from has both, but the product here is 1:1 messaging, not Discord-style chat). The Socket.IO handshake verifies the real JWT (RS256, same keys `auth-service` signs with) and addresses DMs by `accountId`, not a client-supplied name. Its transport isn't proxied through Kong (only `/api/v1/chatroom` REST is gateway-wired) — the frontend connects directly via `NEXT_PUBLIC_CHATROOM_WS_URL`. GAP: chat contacts are derived client-side from booking history (no dedicated "my experts"/"my patients" endpoint); an expert's patient contacts show a placeholder name (`Patient #xxxxxxxx`) because `profile-service`'s patient lookup is ADMIN-only. |
| forum (`forum-service`, Go/Gin) | **Implemented** (browse/filter posts, post detail + nested comments, create post, comment/reply, like, bookmark, my-bookmarks). No admin category-management UI and no post edit/archive/delete controls yet. Its ids are numeric (Postgres `int64`), unlike every other service's UUID strings — except `authorId`/`userId`, which are the account's UUID (this required a same-session backend fix: `X-User-Id` is the JWT's UUID `accountId`, but forum-service's auth middleware originally only accepted a parseable int64, so every authenticated write 401'd for a real account until its `author_id`/`user_id` columns were migrated from `BIGINT` to `UUID`). Response envelope is `{success, message, data, error}` (`forumClient`, `transformCase: false` — its DTOs are already camelCase). GAP: patient-authored posts/comments show "Community member" instead of a real name (same profile-service lookup gap as chat); like/bookmark buttons can't reflect prior-session state since no GET endpoint exposes whether the current user already liked/bookmarked a given post. |

Wherever a real endpoint doesn't exist, the frontend falls back to an in-memory mock in `/data` (see below) rather than leaving the page broken. This is a deliberate, temporary bridge — as each backend endpoint ships, the corresponding `api/*.ts` function should be pointed at it and its `/data` fallback deleted.

## Folder conventions

```
app/            Next.js routes only — thin pages that compose page-local components
constants/      ROUTES, API endpoint paths, nav items, static copy, validation limits
types/          Request/Response interfaces (PascalCase, Request/Response suffix — see document/code-convention.md)
api/            axios call functions, one file per backend resource group + the http/ client layer
hooks/          Custom hooks (useApiQuery/useApiMutation, useAuth, useBreakpoint, ...)
context/        React context providers (Auth, Error/toast, Locale, React Query)
data/           In-memory mock data for endpoints not implemented backend-side yet
router/         Route→role config + the client-side ProtectedRoute guard
locales/        en/vi translation JSON, one file per feature domain
components/     Shared/cross-page UI — layout shells (client-shell, admin-shell, landing) and ui/ primitives
test/           Vitest unit tests
```

Every barrel is a plain `index.ts` re-export (`export * from './x'`) — `.tsx` is reserved for files that actually author JSX. Each page's own extracted UI pieces live in a sibling `component/` folder (singular), e.g. `app/patient/wallet/component/BalanceCard.tsx`; the top-level `components/` (plural) holds things shared *across* pages, like `ClientShell` or `Button`.

Route segments are named after the real role (`app/patient/`, `app/expert/`, `app/admin/`), not route groups — this is what lets `proxy.ts` and `router/routes.config.ts` do simple prefix matching (`/patient/*` requires `PATIENT`), and avoids a `/dashboard` collision between the patient and expert portals sharing one nav shell. `app/(landing)/` and `app/auth/` are the two segments outside that role-prefix scheme (public).

## API client layer (`api/http/`)

**Five axios instances**, one per backend service (`api/http/instances.ts`), not one shared instance with per-call `baseURL` overrides — each service has an independently movable base URL and, critically, a different JSON casing convention:

- `authClient` (Spring Boot / Jackson) returns **camelCase** JSON.
- `profileClient`, `paymentClient`, `bookingClient` (Go/Gin) and `assessmentClient` (FastAPI/Pydantic) all return **snake_case** JSON.

`api/http/client.ts`'s `createHttpClient({ baseURL, transformCase })` factory attaches interceptors:

- **Request**: attaches `Authorization: Bearer <token>` from `api/http/session.ts`; if `transformCase`, deep-converts outgoing `camelCase → snake_case`.
- **Response**: if `transformCase`, deep-converts incoming `snake_case → camelCase` — so every `types/*.ts` interface and every `api/*.ts` call site can stay camelCase (per `document/code-convention.md`) with zero manual per-DTO mapping. The converter lives in `api/http/caseTransform.ts` and is unit-tested in `test/caseTransform.test.ts`.
- **401 handling**: a de-duplicated refresh-and-retry against `auth-service`'s `/auth/refresh`; on failure, clears the session and redirects to login.
- **Error normalization**: `api/http/errorNormalizer.ts` always produces `{ message, statusCode, details }`, because backend error shapes are inconsistent — `auth-service`'s register failure is a raw string body, its other failures are `{ message }`, and the Go services use `{ error }`. This normalized shape is the only one `hooks/useApiQuery`/`useApiMutation` and `ErrorContext` ever see.
- Go services (`profile-service`, at least) wrap list/detail responses as `{ message, data }` — see `ServiceEnvelope<T>` in `types/common.ts`. This is assumed for `payment-service` too since it's the same team/stack, but hasn't been independently confirmed against its source — check if a payment call ever returns unexpectedly-shaped data.

## React Query + automatic error handling

The brief was: page code should only ever handle the success path; errors are handled once, centrally. `hooks/useApiQuery.ts` and `hooks/useApiMutation.ts` wrap `useQuery`/`useMutation` and route every error through `normalizeError()` into `ErrorContext.showError()` (a toast), automatically. Callers only ever pass `onSuccess`.

Two exceptions use plain `useQuery` directly instead: `WalletBadge` (header) and the wallet/medical-history pages' initial profile fetch. A 404 there means "this user hasn't set up a wallet/profile yet" — an expected state, not an error worth toasting on every page load. Real user-initiated actions (top-up, withdraw, profile save) still go through `useApiMutation` and do toast on failure.

React Query v5 note: it removed `onSuccess`/`onError` from `useQuery` — `useApiQuery` replicates `onSuccess` via an internal effect. `useMutation` kept both, but `onError`'s signature grew a 4th parameter (`onMutateResult`) that older tutorials won't mention.

## Auth/session storage — a named tradeoff

There is no BFF or API gateway issuing first-party cookies (`auth-service` is a separate origin), so `accessToken`/`refreshToken`/role are stored in `localStorage` (`api/http/session.ts`, driven by `context/AuthContext.tsx`). A small **non-httpOnly** cookie (`mc_session`, `mc_role` — presence and role only, never the token) is mirrored on login/logout purely so `proxy.ts`, which can only read cookies, can do a coarse, flicker-free redirect before the page renders.

**This cookie provides no XSS protection beyond localStorage** — it's a routing convenience, not a security boundary. The real authorization boundary is each backend service validating the JWT on every request. `router/ProtectedRoute.tsx` is the authoritative client-side guard (exact role check, once `AuthContext` has rehydrated from localStorage); `proxy.ts` + `router/routes.config.ts` is only ever a fast, coarse first line. Production hardening would mean a real BFF (Next.js Route Handlers proxying auth calls) issuing true httpOnly cookies so the JWT never touches client JS — noted here as a deliberate MVP scope cut, not an oversight.

## Responsive "client shell" (`components/layout/client-shell/`)

One shared shell renders three nav variants and lets Tailwind's own breakpoints decide which is visible — no `useMediaQuery`-gated conditional rendering, which would cause a hydration flash (server doesn't know the client's viewport):

- **Desktop** (`lg:` — `DesktopHeaderNav.tsx`): sticky top header with inline horizontal nav links.
- **Tablet** (`sm:`–`lg:` — `TabletDrawerNav.tsx`): slim top bar with a hamburger that opens a left Radix Dialog sheet.
- **Mobile** (`<sm:` — `MobileBottomNav.tsx`): fixed bottom tab bar with the 4–5 highest-value nav items (`PATIENT_BOTTOM_NAV_ITEMS`/`EXPERT_BOTTOM_NAV_ITEMS` in `constants/nav.ts`).

**Core nav vs. dashboard features**: every nav surface (desktop header, tablet drawer, mobile bottom bar, admin sidebar) shows only a short core list per role (`*_HEADER_NAV_ITEMS` in `constants/nav.ts`). Every other page is reached from that role's dashboard, which renders the grouped `*_DASHBOARD_FEATURES` lists through `components/dashboard/FeatureGrid.tsx`. To add a page, add it to the role's `*_NAV_ITEMS` with a `description`, then put its id in either the header list or a dashboard group. `*_NAV_ITEMS` itself stays the full list, used for lookups such as the admin topbar title.

`ClientShell` is used by both `app/patient/layout.tsx` and `app/expert/layout.tsx`, each passing its own `navItems`/`settingsItem` — this is the "client" layout the requirements referred to (a shared shell for both patient and expert "clients" of the platform), as distinct from `landing` (public marketing) and `admin` (a separate, desktop-first sidebar layout in `components/layout/admin-shell/`, since admin panels are conventionally desktop tools and weren't worth building a second three-way responsive system for in this pass).

`hooks/useBreakpoint.ts`/`useMediaQuery.ts` still exist (built on `useSyncExternalStore`, SSR-safe) for genuine JS-only behavior that CSS can't express — they are not used to decide what to render in the shell.

## Design tokens (`app/index.css` + `tailwind.config.js`)

Tailwind v4 is CSS-first: `app/index.css`'s `@theme inline` block is the **single real source of truth** for color/radius/shadow tokens (`--color-primary`, `--color-danger`, etc.), migrated from the ad hoc hex values scattered across the original prototype pages. `tailwind.config.js` exists because it was explicitly requested, and is kept intentionally thin — `theme.extend` mirrors the same CSS variables for tooling that expects a JS config, plus a `safelist` for dynamically-built class names Tailwind's static analyzer can't see (e.g. status-pill colors chosen by a JS switch). **Edit `app/index.css` to change a color, not `tailwind.config.js`.**

UI conventions (aligned with the `ui-ux-pro-max` skill's accessibility/touch rules):

- **Radius scale** — controls (buttons, inputs, selects, nav links) `rounded-lg`; cards, panels, list rows `rounded-xl`; modals and hero blocks `rounded-2xl`; `rounded-full` only for avatars, status pills and dots, never for text buttons.
- **Contrast** — solid `primary`/`danger`/`success`/`warning` fills carry white text, so their token values are chosen to reach 4.5:1 against white.
- **Primitives** in `components/ui/` — `Button` (+ `buttonClasses()` for a `Link` styled as a button; never nest `<Button>` in `<Link>`), `Input`/`Select`/`Textarea`/`Label`/`FieldError` (`fieldClasses()` for one-offs), `Card`/`CardHeader`, `PageHeader` (title + description + actions at the top of every portal page), `Badge` (status pills), `BrandLogo`/`BrandMark`.
- Focus rings, pointer cursor on buttons, and `prefers-reduced-motion` are handled globally in `app/index.css`'s `@layer base`.

## i18n (`context/LocaleContext.tsx` + `hooks/useTranslation.ts` + `locales/`)

No URL locale prefixing (`/en/...`, `/vi/...`) — that would double every route under the already-established `/patient`, `/expert`, `/admin` prefix scheme. Instead, `t(key, defaultText, vars?)` takes a **required** English default so every page reads correctly even before a translation exists for a given key — this is why the signature isn't the more familiar `t(key, vars?)`.

The locale is read server-side from the `mc_locale` cookie in `app/layout.tsx` (via `next/headers`'s `cookies()`) and passed into `AppProviders`/`LocaleProvider` as `initialLocale`, so SSR output already matches a returning visitor's saved language — the client only re-detects (via `localStorage`/`navigator.language`) for the one case the server can't see: a first-ever visit with no cookie set yet. Getting this wrong (defaulting to a hardcoded locale in `useState` and only correcting after a `useEffect`) causes a visible flash of the wrong language on every load; this was caught and fixed during development, not by accident.

Dictionaries are split by domain (`locales/{en,vi}/{common,auth,patient,expert,admin,landing}.json`) and statically merged in `locales/index.ts`. Translation coverage is not exhaustive: navigation labels, the booking-topic grid, and the landing page are fully wired through `t()`; most auth/patient/expert page body copy (form labels, headings) is still hardcoded English, tracked as a follow-up in `README.md`. No ICU plural support — only `{{var}}` string interpolation.

## `/data` mock-fallback convention

Each file under `/data` backs exactly the endpoints the "ground truth" table above marks as not implemented — `clinical-records.ts`, `schedule.ts` (expert weekly availability), `transactions.ts` (wallet history), `notifications.ts` (unused this pass, kept for the future `/notifications` page). `experts.ts` is the one exception — it seeds the public landing page's "Featured Experts" section with curated sample data by design, not because the real endpoint is missing (`GET /profiles/experts/` is real and is what `find-experts` actually calls). `booking-history.ts` used to back appointment create/cancel/history (with a legacy-PHP-inspired mock queue number) — deleted once `api/booking.ts` was pointed at the real booking-service endpoints; there is no queue-number equivalent server-side. `messages.ts` (chat UI mock) was deleted the same way once `app/patient/messages` and `app/expert/messages` were pointed at the real `chatroom-service` Socket.IO backend.

Mutations against the remaining mocks (save a weekly schedule) only persist for the current browser session/module lifetime — a full page reload resets them. This is intentional and documented at each call site; don't mistake it for a bug.

## Known bugs deliberately not carried over from the legacy PHP app

The reference app (`../edoc-doctor-appointment-system`) had: SQL injection via string interpolation, plaintext passwords, auth checks that redirect but don't `exit`/halt (so protected logic still ran), no ownership check before cancelling a booking (any patient could cancel any booking by ID), no capacity check against a session's `nop` limit before booking, and orphaned bookings left behind when a session was deleted. None of these patterns exist in this codebase — every protected route is guarded server-side (whatever backend implements it) and client-side (`ProtectedRoute`), and mock mutations don't skip ownership logic even though nothing currently enforces it across users in the mock layer.

## Capacitor

`@capacitor/core` + `@capacitor/cli` only — `npx cap add ios`/`android` was deliberately not run this pass (see `README.md`). `capacitor.config.ts` points `server.url` at the deployed web app (`NEXT_PUBLIC_APP_URL`) rather than a static export, so it stays compatible with `next.config.ts`'s `output: "standalone"` Docker deploy.
