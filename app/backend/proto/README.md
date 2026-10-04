# MindCare internal contracts (gRPC + events)

Single source of truth for **service-to-service** communication. Browser/mobile traffic
keeps using REST through Kong (`:8000`) and Socket.IO; everything east-west moves to:

| Need | Transport | Defined in |
|---|---|---|
| Synchronous query / command that needs an answer now | **gRPC** (unary or streaming) | `<service>/v1/<service>.proto` |
| A business fact other services react to | **RabbitMQ event**, published through the **transactional outbox** | `<service>/v1/events.proto` + `common/v1/event.proto` |

See [OUTBOX_PATTERN_PLAN.md](./OUTBOX_PATTERN_PLAN.md) for the outbox and RabbitMQ design.

## Layout

```text
proto/
├── buf.yaml / buf.gen.yaml / Makefile
├── common/v1/      common.proto (Role, Money, paging, UserSummary), event.proto (EventEnvelope)
├── auth/v1/        AuthService            + user.*            events
├── profile/v1/     ProfileService         + profile.*, expert.* events
├── assessment/v1/  AssessmentService      + assessment.*      events
├── booking/v1/     BookingService         + appointment.*, medical_record.*, review.* events
├── payment/v1/     PaymentService         + payment.*, refund.*, wallet.*, withdrawal.* events
├── chat-room/v1/   ChatRoomService        + consultation.*    events  (package mindcare.chatroom.v1)
├── chatbot/v1/     ChatbotService         + crisis.*          events
└── forum/v1/       ForumService           + post.*, comment.* events
```

Package naming: `mindcare.<service>.v1`. A breaking change means a new `v2` folder, never an
edit of `v1` (`make breaking` enforces this against `main`).

## What each RPC replaces

| Today | gRPC |
|---|---|
| `POST auth:/internal/auth/token` | `AuthService.IssueServiceToken` |
| `GET auth:/api/v1/auth/public-key` (internal callers) | `AuthService.GetPublicKey` |
| `POST profile:/internal/api/v1/profiles/create` | `ProfileService.CreateProfile` (fallback; main path is `user.created`) |
| RabbitMQ RPC `profile.get_batch.request/response` (forum) | `ProfileService.BatchGetUserSummaries` |
| `POST profile:/internal/api/v1/profiles/sync-seed-authors` | `ProfileService.GetSeedExpert` + `profile.seed_authors_synced` event |
| `POST booking:/internal/appointments/:id/payment-eligibility` | `BookingService.GetPaymentEligibility` |
| `POST booking:/internal/appointments/:id/webhook` (outbox → REST) | `payment.succeeded` / `payment.failed` events; `BookingService.ApplyPaymentResult` kept for reconciliation |
| `GET booking:/internal/appointments/:id` | `BookingService.GetAppointment` |
| `POST chatbot:/api/v1/chat/stream` (from chatroom) | `ChatbotService.Chat` (server streaming) |
| `POST chatbot:/api/v1/chat/voice` (from chatroom) | `ChatbotService.VoiceChat` (bidi streaming) |

New RPCs with no REST predecessor: `BookingService.CheckConsultationAccess`,
`AssessmentService.GetLatestResults`, `ChatRoomService.GetPresence`, `PaymentService.RequestRefund`…

## gRPC conventions

- **Ports** — every service keeps its HTTP port and adds a gRPC port = HTTP port + 1000
  (auth `9080`, profile `9081`, payment `9082`, booking `9083`, forum `9084`, chatroom `9085`,
  chatbot `9086`, assessment `6000`). gRPC ports are only published on the Docker
  `app-network`, never through Kong.
- **Auth** — the M2M JWT from `AuthService.IssueServiceToken` is sent in metadata
  `authorization: Bearer <token>`. A server interceptor verifies it with the RSA public key
  (same logic as `pkg/internal_auth/middleware.go`) and exposes the caller `client_id`.
  End-user identity, when needed, is passed in metadata `x-user-id` / `x-user-role`.
- **Tracing** — metadata `x-correlation-id` is propagated on every hop and copied into
  `EventEnvelope.correlation_id` when an event is written to the outbox.
- **Deadlines** — every client call sets one (default 3 s, `Chat`/`VoiceChat` 60 s).
- **Retries** — client-side retry only for idempotent RPCs (`Get*`, `List*`, `Check*`,
  and RPCs carrying an `idempotency_key`) on `UNAVAILABLE` / `DEADLINE_EXCEEDED`.
- **Errors** — use gRPC status codes: `NOT_FOUND`, `INVALID_ARGUMENT`, `PERMISSION_DENIED`,
  `FAILED_PRECONDITION` (business rule), `ABORTED` (state conflict), `UNAVAILABLE` (retryable).
- **Health** — each server registers `grpc.health.v1.Health` for Docker healthchecks.
- **Times / money** — `google.protobuf.Timestamp` on the wire (services keep their internal
  Unix-ms columns); money is `common.v1.Money` in VND (no minor unit).
- **Enums** — zero value is always `*_UNSPECIFIED`. Booking/payment smallint statuses are
  shifted by one on the wire (e.g. DB `0 PENDING_PAYMENT` ⇄ proto `1`), map them in the adapter.

## Generate stubs

Stubs are **not committed** (`gen/` is git-ignored) and generation is not wired into any
build yet. When a service starts implementing gRPC:

```bash
cd app/backend/proto
make lint        # buf lint
make breaking    # buf breaking vs main
make generate    # buf generate -> gen/{go,java,python,ts}
```

| Service | Language | Recommended way |
|---|---|---|
| booking, payment, profile, forum | Go | `gen/go` + `replace github.com/MinhKhongCau/GraduationProject/app/backend/proto/gen/go => ../proto/gen/go` in `go.mod` |
| auth | Java 21 | `protobuf-maven-plugin` pointing `protoSourceRoot` at `../proto` + `grpc-spring-boot-starter` |
| assessment, chatbot | Python | `grpcio-tools` (`python -m grpc_tools.protoc -I ../proto ...`) at Docker build time |
| chatroom | Node.js | `@grpc/grpc-js` + `@grpc/proto-loader` (loads `.proto` at runtime, no codegen) |

Because the Dockerfiles build each service from its own folder, the build context must be
widened to `app/backend` (or `proto/` copied in) once a service consumes these contracts.
