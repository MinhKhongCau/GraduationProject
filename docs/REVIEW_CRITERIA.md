# MindCare: Review Criteria

These criteria are used to review the MindCare system, an online psychological counseling platform built as microservices. Each criterion is scored from 1 to 5:

- **1**: missing or broken
- **2**: present but unsafe or unreliable
- **3**: works for a demo
- **4**: solid, with minor gaps
- **5**: production-grade

The findings and scores from the current review are in [SYSTEM_REVIEW.md](SYSTEM_REVIEW.md).

---

## 1. Functional Requirements

| # | Criterion | What to check |
|---|---|---|
| F1 | Account & RBAC | Register, login, refresh, logout, password reset, and email verification work. The ADMIN, EXPERT and PATIENT roles are enforced on the server. A user cannot choose their own role. |
| F2 | Profiles | Patient and expert profiles, specializations, medical history, avatar upload, and public expert search for guests. |
| F3 | Booking | Expert availability leads to generated slots. Slots are locked with an expiry. Appointments go through a state machine. Time-off is respected. Double booking is impossible. |
| F4 | Payment | Wallet, VNPay order, and IPN reconciliation. Money is held in escrow and then settled. Refunds and compensation are carried out, not just recorded. Withdrawals work. |
| F5 | Assessment | PHQ-9, ASRS and MDQ scoring follows the official scales, including severity bands. Questions and options are validated against the template. AI feedback is added. |
| F6 | Real-time chat | DMs, history, voice messages, presence, and typing indicators. |
| F7 | Voice/Video calls | 1-on-1 calls and appointment meeting rooms. Ringing, accept, reject, busy, timeout, and hang-up are handled. |
| F8 | AI chatbot | Context-aware replies, RAG, emotion detection, and crisis detection in **Vietnamese and English**. |
| F9 | Forum | Posts, comment tree, likes, bookmarks, and moderation. |
| F10 | Docs and code agree | The README and flow guides match the code: endpoints, lock times, request bodies, and the money model. |

## 2. Security & Privacy

| # | Criterion | What to check |
|---|---|---|
| S1 | Authentication | JWT algorithm is pinned. Access tokens are short-lived (≤15 min). The token type (access or refresh) is separated. Keys can be rotated. |
| S2 | Authorization | Every route goes through the gateway auth. Services never trust client-supplied `X-User-*` headers. `/internal/*` cannot be reached by any path prefix. |
| S3 | Service-to-service (M2M) | Internal HTTP and gRPC calls require the internal token. No plaintext secrets appear in the repo or in logs. |
| S4 | Secrets management | Nothing is hard-coded and there are no unsafe fallbacks (for example `MOCK_SECRET`, `secretpassword`). A `.env.example` exists. Passwords are never logged. |
| S5 | Input validation | Request and socket payloads are checked against a schema. Sizes are limited (audio, page size, arrays). File paths are sanitised. |
| S6 | Payment integrity | Signatures are compared in constant time. Amounts and status are verified. No endpoint can create money. The database enforces idempotency. |
| S7 | Frontend security | No unescaped `dangerouslySetInnerHTML`. CSP and other security headers are set. Token storage is chosen with the XSS risk in mind. |
| S8 | Health-data privacy | Data sent to third-party AI is minimised and the user consents. No user can read another user's chat or assessment history. No traffic goes through public tunnels. |
| S9 | Abuse protection | Rate limits on login and on each socket event. Account lockout. No user enumeration. |

## 3. Real-time & WebRTC

| # | Criterion | What to check |
|---|---|---|
| R1 | Socket authentication | JWT is checked at the handshake. Expired tokens cause a refresh and reconnect. |
| R2 | Identity binding | Every event uses `socket.data.user` from the JWT, never a name or id sent by the client. |
| R3 | Room/call authorization | A user can only join their own DM, call, or meeting. Signals are relayed only to peers in the same room. |
| R4 | NAT traversal | A TURN server is set up (coturn) next to STUN, with temporary credentials. |
| R5 | Call state machine | Busy, ring timeout, glare (both sides calling at once), and notifying the peer when someone disconnects mid-call. ICE restart. |
| R6 | Client robustness | ICE candidates are queued. Media errors (permission denied, no device, device busy) are told apart. Resources are cleaned up on unmount and logout. Acks have timeouts. |
| R7 | Delivery guarantees | Message ids are monotonic. Acks and dedupe. Offline delivery or push. |
| R8 | Horizontal scaling | The Redis adapter is used. Nothing important lives in per-process state (call maps, chatbot history). Presence handles several tabs. |
| R9 | Latency & payload | Media and large blobs go to object storage, not inline Base64. Chatbot streaming sends only new text (deltas). |

## 4. Performance & Scalability

| # | Criterion | What to check |
|---|---|---|
| P1 | Query efficiency | No N+1 queries. Predicates can use indexes (no `TO_CHAR(col)=?`). Text search uses a trigram or full-text index. |
| P2 | Pagination | Every list has a maximum page size. Large trees are paginated. |
| P3 | Concurrency control | Database constraints back up application checks. Locks belong to their owner. No `sync.Mutex` is relied on across replicas. |
| P4 | Background workers | Intervals are sensible. Workers shut down gracefully. Outbox rows are claimed with `SKIP LOCKED` and retried with jitter. |
| P5 | AI inference | Inference doesn't block the event loop. Models are cached and warmed up. Timeouts and fallbacks are set. Clients are reused. |
| P6 | Caching & pooling | Database connection pools are tuned. Hot reads are cached. |
| P7 | Frontend performance | Server components are used where possible. Heavy libraries are loaded on demand. `next/image` is used with remote patterns. |

## 5. Code Quality & Architecture

| # | Criterion | What to check |
|---|---|---|
| Q1 | Layering | Dependencies point inward: domain ← application ← infrastructure. No global database handle is used inside handlers or use cases. |
| Q2 | Consistency | Each service follows one architectural style. Shared concerns (token manager, response envelope) are not copied around. |
| Q3 | Error handling | Errors are typed and mapped to the correct status codes. No swallowed errors. No internal error text is shown to clients. No unhandled promise rejections. |
| Q4 | Conventions | Follows `CODING-CONVENTION.md`: no `any`, no `console.log`/`print`, commit format. Each language has its own style guide. |
| Q5 | Dead code & hygiene | No dead modules, demo endpoints, committed binaries, or mojibake (garbled text from double-encoded UTF-8). |
| Q6 | Frontend structure | No duplicated pages between roles. Contexts have a single purpose. Error boundaries exist. i18n is complete. |
| Q7 | Typing | TypeScript is strict with no `any`. Python has type hints and Pydantic models. Go uses typed errors. |

## 6. Testing

| # | Criterion | What to check |
|---|---|---|
| T1 | Unit tests | Domain and use-case coverage for every service. |
| T2 | Integration tests | Tests run against a real database (for example with testcontainers) for repositories, locks, and migrations. |
| T3 | Concurrency tests | Slot locking, IPN races, wallet locks. |
| T4 | Contract and gateway tests | Kong routing, auth bypass attempts, inter-service contracts. |
| T5 | Real-time tests | Socket handlers, call and WebRTC flows. |
| T6 | E2E | The critical journey: register → book → pay → meet → settle. |

## 7. DevOps & Operations

| # | Criterion | What to check |
|---|---|---|
| D1 | CI coverage | Every service is built, linted, type-checked and tested. Dependencies and images are scanned. |
| D2 | CD correctness | Each environment uses its own compose file and env. Deploys are gated on CI success. Images are tagged immutably (by commit SHA). Rollback is possible. |
| D3 | Migrations | Versioned and applied automatically (golang-migrate, Flyway, Alembic). No `ddl-auto: update` or AutoMigrate in production. |
| D4 | Runtime config | Healthchecks on every service. Resource limits. Log rotation. Pinned image tags. Minimal exposed ports. |
| D5 | Observability | Structured logs, metrics (Prometheus), tracing (OpenTelemetry), error tracking. Admin tools are behind auth. |
| D6 | Data safety | Automated backups with tested restore. Each service gets its own database role. |
| D7 | Reproducibility | Lock files exist. No `go mod tidy` inside the Docker build. Runtime versions match (Node in CI and in Docker). |
