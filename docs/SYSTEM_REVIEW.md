# MindCare: System Review

- **Date:** 2026-10-29
- **Branch:** `features/PSY-50-be-update-webrtc-with-handle-error`
- **Scope:** all backend services, the gateway, the frontend, infra, and CI/CD
- **Method:** static code review, read-only. No services were run. File references are relative to `app/backend/` unless they start with `app/frontend/` or a repo-root file.
- **Criteria:** [REVIEW_CRITERIA.md](REVIEW_CRITERIA.md)

---

## 1. Executive Summary

MindCare is a feature-rich microservices system. Most of the features in the README are present, and the core booking → payment → meeting flow is well designed: a transactional outbox, escrow-then-settle, idempotent IPN handling, `FOR UPDATE` row locks, and a meeting service that is properly authorised.

The system is **demo-ready but not production-ready**. The main reasons:

1. **Several critical security holes.** Anyone can register as ADMIN. Anyone can mint wallet money. The gateway can be bypassed through the "docs" routes. The socket call registry trusts the client. Assessment accepts `Bearer admin` as an admin token.
2. **The real-time calling path is fragile.** There is no TURN server, no call state machine, and call state is kept in each process.
3. **DevOps is incomplete.** The production deploy workflow uses the dev compose file, there are no backups, and only `:latest` image tags exist.
4. **Testing is uneven.** Go unit tests are strong. Chatroom and chatbot are not in CI. There are no integration, concurrency or gateway tests.

### Scorecard

Scores are 1–5 (see the criteria document).

| Area | Score | One-line verdict |
|---|---|---|
| Functional requirements | **3.5** | Broad coverage. Refunds, reconciliation, payouts, severity scoring and email are stubbed. |
| Security & privacy | **1.5** | Several critical issues that can be exploited remotely. |
| Real-time & WebRTC | **2.5** | Meeting rooms are good. 1-on-1 calls are insecure and fragile. No TURN server. |
| Performance & scalability | **3** | Fine at demo scale. Some pieces only work on a single instance. A few non-sargable queries. |
| Code quality & architecture | **3** | Good layering intent. Inconsistent execution, dead code, duplication. |
| Testing | **2.5** | About 450 unit tests, concentrated in Go. No integration or E2E tests. |
| DevOps & operations | **2** | CI exists. CD is broken for prod and staging. No backups or metrics. |

---

## 2. How the System Meets the Requirements

### 2.1 Architecture

```
Next.js 16 (web + Capacitor) ──HTTPS/WSS──> Kong Gateway (Go plugin: RS256 JWT → X-User-Id/Role)
                                               │
   ┌───────────┬───────────┬──────────┬───────┴──┬──────────┬────────────┬─────────────┐
 auth(Java)  profile(Go) booking(Go) payment(Go) forum(Go) assessment(Py) chatroom(Node) chatbot(Py)
   │            │  gRPC ◄──────┤◄── gRPC ───┤                    │ Gemini      │ Socket.IO     │ Ollama/Whisper
   └────────── RabbitMQ (topic exchanges) ── Redis ── Postgres 16 + pgvector (one DB per service) ──┘
```

- **Gateway:** a single entry point, with RS256 JWT checked in a custom Kong Go plugin.
- **Service-to-service calls:** internal JWTs issued by auth-service (`/internal/auth/token`).
- **Async events:** RabbitMQ topic exchanges.
- **Payment → booking reliability:** a transactional outbox with exponential backoff and a DEAD state (`payment-service/.../messaging/outbox_publisher.go`).

### 2.2 Requirement coverage

| Requirement | Status | Evidence / gap |
|---|---|---|
| Register / login / RBAC | ✅ with a critical bug | Refresh-token rotation, bcrypt, RS256. But the `role` field from the client is accepted at registration (`auth-service/.../RegisterUseCase.java:30`). Email is sent to the console only, and verification is forced to true (`:31`). The 2FA fields are unused. |
| Profiles & expert search | ✅ / ⚠️ | CRUD, S3 avatar (5 MB cap), gRPC. Guests cannot browse experts because Kong requires auth on all of `/api/v1/profiles` (`gateway/kong.yml:110-118`). |
| Availability → slots → 15-min lock → appointment | ✅ | `create_appointment.go:77-110` uses `SELECT … FOR UPDATE`. The docs say 5 minutes but the code uses 15 (`domain/slot/slot_lock.go:3`). |
| Wallet + VNPay + IPN | ✅ | IPN is idempotent and writes the order, the escrow credit and the outbox row in one transaction (`process_ipn.go:68-212`). |
| Refunds / compensation | ❌ stub | `ResolveCompensationCase` only changes the status (`manage_orders.go:131-147`). No money is returned. |
| Withdrawals | ⚠️ mock | The payout is mocked (`withdrawal/usecase.go:419-450`). Withdrawal outbox events always end up DEAD. |
| VNPay reconciliation | ❌ stub | TODO at `wallet/reconciliation_worker.go:83`. |
| Psychological tests (PHQ-9/ASRS/MDQ) | ⚠️ partial | Only raw sums are computed. `severity_level` is never written. Answers are not validated against the template, so scores can be forged. |
| AI advice (Gemini) | ✅ | Falls back on errors. No timeout. No consent or anonymisation notice. |
| Real-time DM, voice messages, presence | ✅ | JWT at the handshake. Messages are written to Postgres first and Redis as a cache. Base64 audio up to 8 MB with no validation. |
| 1-on-1 WebRTC calls | ⚠️ | Works on a single instance only. No auth on the call registry. STUN only. |
| Appointment meeting rooms (PSY-41/50) | ✅ best part | Booking-backed authorization, Redis Lua presence, typed ack errors, tests. |
| AI chatbot (RAG + emotion) | ⚠️ | A 6-layer pipeline. The LLM goes through a hardcoded **ngrok** URL (`chatbot-service/app/layers/layer4_llm.py:11`). Crisis detection is English-only. History is in memory. |
| Forum | ✅ | ltree comment trees, likes, bookmarks, owner-or-admin checks. |
| Mobile app | ✅ | Capacitor android/ios with a hardened config. |

---

## 3. Pros and Cons

### Pros
- **Clear domain decomposition.** Each service owns its own database, and the Go services follow a domain/application/infrastructure layout.
- **The payment design is strong:**
  - integer money;
  - state machines for orders, captures, fulfilment and compensation;
  - partial unique indexes;
  - escrow-then-settle;
  - duplicate-capture detection;
  - a transactional outbox.
- **Sound security basics:**
  - RS256 with the algorithm pinned at the gateway and at the socket handshake;
  - bcrypt;
  - refresh tokens stored as hashes and rotated, with replay detection;
  - parameterised queries everywhere (no SQL injection found).
- **The meeting rooms (latest work) are well built.** The frontend client queues ICE candidates, falls back to audio-only, cleans up on unmount and rejoins after reconnect.
- **Good frontend foundation.** TanStack Query with sensible defaults, de-duplicated token refresh, `strict` TypeScript, almost no console noise, decent ARIA use.
- **A sizeable unit-test suite.** About 270 tests in booking and payment, 46 in auth, 43 in profile, 41 in chatroom.
- **Developer experience.** Swagger for every service, a per-service CI matrix, multi-stage non-root Docker images.

### Cons
- **Critical security flaws** (see §4.1).
- **Docs and code disagree** (lock time, create-appointment request body, who holds the money, `.env.example` that does not exist).
- **Several features look done but are stubs** (refunds, reconciliation, payout, email, severity scoring).
- **Real-time calls are not production-grade** (§5).
- **Broken production and staging deploys, no backups, no metrics.**
- **Inconsistent code style** across services, dead code, about 700 duplicated lines in the frontend.

---

## 4. Technical Debt

### 4.1 Critical: fix before any public demo

| # | Issue | Location |
|---|---|---|
| C1 | **Anyone can register as ADMIN.** The client-supplied `role` is accepted. | `auth-service/.../RegisterUseCase.java:30` |
| C2 | **Money-minting endpoint.** Any user can top up their own wallet, then withdraw; amounts up to 5,000,000 VND are approved automatically. | `payment-service/.../routes/routes.go:25`, `withdrawal/usecase.go:91-97` |
| C3 | **The gateway "docs" routes bypass auth.** `/auth`, `/profile`, `/booking`, `/payment`, `/forum` and `/assessment` use `strip_path: true` with no auth plugin. Services trust spoofed `X-User-Role` headers, and `/profile/internal/...` slips past the `/internal` block. | `gateway/kong.yml:98-101,119-122,142-145,162-165,178-181,224-227` |
| C4 | **Assessment admin bypass.** `Authorization: Bearer admin` grants admin. | `assessment-service/app/infrastructure/http/dependencies.py:24` |
| C5 | **The socket call registry trusts the client.** `user:online {name}` lets a user take over another user's identity for calls. `room:join` lets a user take over another user's DMs. | `chatroom-service/.../call.socket.controller.js:9-15`, `room.socket.controller.js:29-33` |
| C6 | **Chatbot privacy.** `session_id` comes from the client, so any user can read another user's history. Path traversal in the voice upload. Mental-health data goes through a public ngrok tunnel. | `chatbot-service/app/main.py:43,70,123`, `layer4_llm.py:11` |
| C7 | **VNPay secret falls back to `"MOCK_SECRET"`**, so IPNs can be forged if the env var is missing. The comparison is not constant-time. | `payment-service/.../vnpay/vnpay_client.go:26,142` |
| C8 | **M2M secrets.** Plaintext client secrets are in `application.yaml` and logged at INFO. Profile's internal HTTP and gRPC endpoints need no token. | `auth-service/src/main/resources/application.yaml:75-86`, `InternalTokenUseCase.java:56,61` |
| C9 | **Production deploy uses `.deploy/dev/docker.yml`** and has no check that CI succeeded. | `.github/workflows/deploy-production.yml:72,77` |

### 4.2 High

- **Tokens:**
  - Access-token TTL is **50 days** (`application.yaml:59`, marked "change to 900000 at go-live").
  - Refresh tokens are accepted as access tokens (no `typ` claim).
  - The gateway caches the public key forever.
- **Withdrawals:**
  - Creating a withdrawal is not atomic: funds can stay locked with no request.
  - Withdrawal outbox events always end up DEAD.
- **Redis wallet lock:**
  - no owner token;
  - treated as "acquired" when Redis is down;
  - every IPN contends on the single system-wallet key.
- **Double-booking protection was removed from the DB.** The unique constraint on `appointments.slot_id` is dropped on every boot and nothing replaces it (`booking-service/config/database.go:181`).
- **SQL migrations with the critical indexes are never applied automatically.** Services run GORM AutoMigrate plus ad-hoc `ALTER`s.
- **XSS:** `dangerouslySetInnerHTML` on unescaped AI and markdown output (`app/frontend/app/patient/assessment/component/MarkdownRenderer.tsx:24-33`). Combined with tokens in localStorage and no CSP, this could let an attacker steal tokens.
- **Mental-health safety:**
  - crisis regexes are English-only;
  - the output filter runs after streaming, so users already saw the text;
  - the PHQ-9 suicidal-ideation item is never flagged.
- **`torch.load` is globally patched to `weights_only=False`**, which turns off PyTorch's protection against pickle code execution (`chatbot-service/app/layers/layer1_nlu.py:14-22`).
- **Logging and exposure:**
  - DB passwords are logged at startup (booking and payment `config/database.go`);
  - Dozzle (with `docker.sock`) is public with no auth.

### 4.3 Medium / maintenance

- **Two architectural styles in profile-service.** Use cases vs. handlers that use the global `config.DB` (~30 call sites).
- **Inverted dependencies in forum.** The application layer imports infrastructure.
- **Leaks in auth's hexagonal design.**
  - Use cases import `JwtUtils`.
  - Every `RuntimeException` becomes a 400.
- **Dead code:**
  - forum `sysadmin/`;
  - auth `*Request` DTOs and demo endpoints;
  - chatbot `routers.py`/`config.py`;
  - chatroom `webrtc.socket.controller.js` (a second relay, duplicating events) and `typing.socket.controller.js`.
- **Repo hygiene:**
  - a 38 MB Go binary is committed (`app/backend/profile-service/main`);
  - `docker-compose.dev.huy.yml` is tracked even though it is git-ignored;
  - no root `.env.example`;
  - `init-db.sh` differs between dev and prod.
- **Mojibake** (garbled text from double-encoded UTF-8) in 11 Go files, including user-facing errors and a value stored in the DB.
- **Convention breaches:**
  - 47 `any` in the frontend;
  - 38 `console.*` in chatroom;
  - `print` in the Python services;
  - commits don't follow the documented format.
- **Frontend:**
  - patient and expert `messages/` and `forum/` pages are duplicated;
  - `ChatContext.tsx` has grown to 549 lines;
  - no `error.tsx` boundaries;
  - i18n is about 15% done;
  - the landing page still uses mock data.
- **Infra:**
  - no resource limits or log rotation;
  - `:latest`-only image tags;
  - `ollama:latest`;
  - `go mod tidy` runs inside the Docker build.

---

## 5. Performance & Real-time

### 5.1 Real-time (chatroom-service + frontend)

| Aspect | Current state | Risk |
|---|---|---|
| Socket auth | RS256 checked at the handshake ✅. No refresh after expiry, and the client does not reconnect after a middleware rejection. | Chat stays dead until the page is reloaded. |
| Identity binding | Meeting rooms use the JWT ✅. Calls and rooms use names sent by the client ❌. | Calls and DMs can be hijacked. |
| NAT traversal | Google STUN only, **no TURN** (`app/frontend/hooks/useMeetRoom.ts:78-81`, `app/frontend/context/ChatContext.tsx:123-126`). | Calls fail behind symmetric NAT or carrier-grade NAT, which is common on Vietnamese 4G. |
| Call state machine | The server keeps no call state: no busy, no ring timeout, no glare handling. When one side disconnects, the peer is not notified. | Calls hang, and calls get stuck in "calling". |
| Client recovery | Meeting rooms handle reconnect ✅. A socket blip tears down working media, and there is no ICE restart. The 1-on-1 call path has no ICE queue (a race condition). | Calls drop during short network glitches. |
| Scaling | Redis adapter wired ✅. But `onlineUsers` is a per-process `Map`, the meet same-room check uses the local adapter, and presence is one socket per user. | Calls break with more than one instance. Second tabs miss DMs. |
| Delivery | Message ids are `Date.now()-random`. No acks or dedupe. History cursor is by `created_at`. | Messages sent in the same millisecond can be skipped. No offline delivery. |
| Payload | 8 MB `maxHttpBufferSize`. Base64 audio is stored in Postgres and in the Redis list (up to 100 per DM). The chatbot resends the whole text on every chunk (O(n²)). | Memory and bandwidth blow-up. |
| Flood control | Kong rate-limits only the HTTP upgrade. Socket events are unlimited, and each chatbot message triggers LLM inference. | Easy to abuse for denial of service. |
| Error safety | Async socket handlers have no try/catch. | An unhandled rejection can kill the Node process. |

### 5.2 Backend performance

- **Booking:**
  - `TO_CHAR(date_slot, …) = ?` predicates can't use indexes (`slot/repository.go:43,99,116`);
  - an in-process `sync.Mutex` is used for slot overlap checks;
  - slot generation runs sequentially for each expert.
- **Payment:** the reconciliation worker is labelled "daily" but runs **every 45 s**, with N+1 `SUM`s over all wallets. The outbox has no `SKIP LOCKED` and no jitter, so running two replicas would deliver events twice.
- **Forum:**
  - no upper bound on `pageSize`;
  - the whole comment tree is loaded at once;
  - a synchronous `UPDATE view_count` on every post view;
  - every list request makes a synchronous RPC to profile over RabbitMQ (3 s timeout) on a fixed reply queue, which breaks with more than one replica;
  - like/unlike is not transactional.
- **Search:** `LIKE '%x%'` with no trigram index.
- **Pools and caches:** connection pools are not tuned and nothing is cached.
- **AI:**
  - Whisper transcription runs synchronously inside `async def`, so it blocks the event loop;
  - a new `PGVector` engine is created for every request;
  - Gemini is called synchronously before the DB commit, with no timeout;
  - assessment submit makes 2 queries per answer (N+1);
  - ingest re-runs on every container start.
- **Frontend:** 157 `"use client"` files, so the app barely uses server components. Heavy editor libraries should be loaded on demand. Node versions differ: 22 in Docker, 24 in CI.

---

## 6. Code Quality

| Area | Assessment |
|---|---|
| **booking / payment (Go)** | Best in the repo: pure domain planners, state machines, about 270 unit tests. Leaks: the application layer imports `gorm` and the global config, and ports are discovered with runtime type assertions. Pre-migrations ignore errors. |
| **auth (Java)** | Clean use-case split and 46 tests. JPA entities sit in the domain layer, `System.out.println` is used, unused DTOs remain, and the exception mapping is too coarse. |
| **profile / forum (Go)** | Mixed styles and global DB handles. `err.Error()` text is returned to clients. Forum has only 6 tests, and 3 of them test dead code. |
| **chatroom (Node)** | Layered layout. The meet module is high quality. The legacy call and room code is unsafe and duplicated. 38 `console.*` calls while `utils/logger.js` sits unused. No payload validation. |
| **chatbot (Py)** | A neatly layered pipeline, but dead modules, hardcoded URLs, `print` logging, and one test fails against the current code (`test_health_endpoint`). |
| **assessment (Py)** | Small and readable. Weak validation, no scoring rules, the admin bypass, and CI runs only `compileall`. |
| **frontend (TS)** | Strict mode and a consistent data layer. 47 `any`, duplicated role pages, no error boundaries, `alert()`/`confirm()` in the call flows, and only 3 test files. |
| **Tests overall** | Unit tests use fakes or `sqlmock` only. **No** real-DB integration, concurrency, gateway, contract or E2E tests. Chatroom and chatbot are not in CI. |

---

## 7. Recommended Roadmap

**Phase 1 (security hot-fixes, about 1–2 days)**
1. Ignore `role` at registration. Remove or env-gate `/wallets/top-up`. Remove the `Bearer admin` check.
2. Put the auth plugin on the docs routes or remove them in prod. Strip incoming `X-User-*` headers at Kong.
3. Bind call and room identity to `socket.data.user`. Delete `webrtc.socket.controller.js`.
4. Derive the chatbot `session_id` from the JWT. Sanitise upload paths. Move the LLM URL to env.
5. Fail fast when `VNP_HASH_SECRET` is missing, and use `hmac.Equal`. Remove plaintext secrets from the repo and logs. Set access TTL to 15 min and add a `typ` claim.
6. Fix `deploy-production.yml` (use the product compose file and require CI success).

**Phase 2 (correctness, about 1 week)**
- Add a partial unique index on `appointments(slot_id) WHERE status IN (active)`. Run migrations automatically (golang-migrate / Flyway).
- Make withdrawals atomic, route withdrawal events, and drop the Redis wallet lock in favour of `FOR UPDATE`.
- Implement refund execution and VNPay QueryDR reconciliation.
- Add assessment validation and severity bands, and a suicidal-ideation flag that alerts an expert.
- Add Vietnamese crisis keywords and filter the chatbot output **before** streaming.

**Phase 3 (real-time hardening)**
- Deploy coturn with temporary credentials. Add a server-side call state machine (busy, timeout, disconnect → `call:ended`). Add ICE restart.
- Move per-process state to Redis. Support multiple sockets per user for presence. Rate-limit and validate socket events.
- Store voice messages in object storage. Use ULIDs for message ids. Stream only new chatbot text.

**Phase 4 (quality & ops)**
- Add chatroom, chatbot and lint/tsc to CI. Add testcontainers integration tests and a book-pay-meet E2E test.
- Tag images by commit SHA. Add healthchecks, resource limits and log rotation. Add Prometheus/OTel. Put Dozzle behind auth. Back up Postgres nightly.
- Remove dead code and the committed binary. Fix the mojibake. Extract shared frontend chat and forum components. Finish i18n.
