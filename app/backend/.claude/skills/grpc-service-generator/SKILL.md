---
name: grpc-service-generator
description: 'Define or change MindCare internal gRPC contracts and RabbitMQ event
  messages in app/backend/proto, and wire a service to them. Use when adding an RPC,
  a service-to-service call, an integration event, or replacing an /internal/* REST
  endpoint or AMQP request/reply with gRPC. Triggers: "grpc", "proto", "protobuf",
  "internal call", "service-to-service", "events.proto", "buf".'
allowed-tools: Read, Write, Edit, Grep, Glob, Bash(buf:*), Bash(make:*), Bash(python -m grpc_tools.protoc:*)
version: 2.0.0
tags:
- backend
- grpc
- protobuf
- microservices
---
# MindCare gRPC contracts

All east-west contracts live in `app/backend/proto/`. Read `proto/README.md` (conventions)
and `proto/OUTBOX_PATTERN_PLAN.md` (events) before changing anything.

## When to use gRPC vs an event

| Situation | Use |
|---|---|
| Caller needs an answer to continue (eligibility, access check, enrichment, streaming) | gRPC RPC in `<svc>/v1/<svc>.proto` |
| A fact happened that other services react to | Event message in `<svc>/v1/events.proto`, published **only** via the outbox |
| Browser/mobile traffic | REST via Kong or Socket.IO — **never** gRPC |

Never call gRPC inside a DB transaction. Never publish to RabbitMQ from business code.

## Layout and naming

```text
proto/<service-folder>/v1/<service>.proto   service + request/response + domain messages
proto/<service-folder>/v1/events.proto      one message per routing key
proto/common/v1/common.proto                Role, Money, PageRequest/PageResponse, UserSummary
proto/common/v1/event.proto                 EventEnvelope (wraps every event)
```

| Item | Rule | Example |
|---|---|---|
| Folder | service name, kebab-case | `chat-room/v1/` |
| Package | `mindcare.<service>.v1`, no `-` | `mindcare.chatroom.v1` |
| File | snake_case | `chat_room.proto` |
| Service | PascalCase + `Service` | `BookingService` |
| RPC | verb first: `Get`, `List`, `BatchGet`, `Create`, `Check`, `Apply`… | `GetPaymentEligibility` |
| Request | `<Rpc>Request` (unique per RPC) | `GetAppointmentRequest` |
| Response | `<Rpc>Response`, or a domain message for Get/Create | `returns (Appointment)` |
| Enum | values prefixed with the enum name; zero value `<ENUM>_UNSPECIFIED` | `APPOINTMENT_STATUS_CONFIRMED` |
| Event message | past tense fact | `AppointmentCancelled` |
| Routing key | `<entity>.<action>` written as a comment above the event | `// routing key: appointment.cancelled` |

Every file needs these options (adapt the folder/package):

```proto
option go_package = "github.com/MinhKhongCau/GraduationProject/app/backend/proto/gen/go/<pkg>/v1;<pkg>v1";
option java_multiple_files = true;
option java_package = "com.mindcare.proto.<pkg>.v1";
```

## Field rules

- IDs are `string` UUIDs (forum post/comment ids are `int64`).
- Times: `google.protobuf.Timestamp` on the wire. Services convert from their own storage
  (booking and payment use Unix ms `BIGINT`).
- Money: `mindcare.common.v1.Money` (VND, no minor unit). Never `double`.
- Booking and payment store statuses as smallint starting at 0. Proto enums start at 1, so map
  them in the adapter and document the mapping in the enum comment.
- Use `oneof` for "id by profile_id OR auth_id" style lookups.
- Commands that may be retried carry `string idempotency_key`.
- Events never contain secrets, JWTs, chat text or clinical notes.
- Comment every service, RPC and non-obvious field, including the REST endpoint it replaces
  and the gRPC error codes it returns.

## Compatibility (v1 is frozen once consumed)

- Allowed in place: add fields, RPCs, enum values or messages.
- Never: change a field's number, type or name; reuse a removed number; rename a package.
- When removing a field, use `reserved <n>; reserved "<name>";`.
- A breaking change goes in a new `v2/` folder, and v1 is kept until all callers migrate.

## Steps

1. Find the real shape of the data first: read the service's entity/model and the handler it
   replaces. Don't invent fields.
2. Edit or add the `.proto` in the right folder and follow the rules above.
3. If it is a new event, also:
   - add the routing key to the exchange/queue tables in `OUTBOX_PATTERN_PLAN.md` §6
   - add the binding to `app/backend/rabbitmq/definitions.json`
4. Update the "What each RPC replaces" table in `proto/README.md` when replacing REST or AMQP RPC.
5. Validate:
   ```bash
   cd app/backend/proto && make lint && make breaking
   # buf not installed? at least check it compiles (output outside the repo):
   python -m grpc_tools.protoc -I . --descriptor_set_out=/tmp/mindcare.pb --include_imports $(find . -name '*.proto')
   ```
6. **Do not commit generated stubs.** `gen/` is git-ignored. Only run `make generate` when the
   user asks to implement a server or client.

## Implementing a server/client (only when asked)

| Service | Language | How |
|---|---|---|
| booking, payment, profile, forum | Go | `buf generate`, `replace .../proto/gen/go => ../proto/gen/go`. Interceptors: M2M JWT verify (reuse `pkg/internal_auth`), `x-correlation-id`, recovery, logging. Register `grpc.health.v1`. |
| auth | Java 21 Spring | `protobuf-maven-plugin` with `protoSourceRoot=../proto`, `grpc-spring-boot-starter` |
| assessment, chatbot | Python FastAPI | `grpcio-tools` at Docker build time, run `grpc.aio` server next to uvicorn |
| chatroom | Node ESM | `@grpc/grpc-js` + `@grpc/proto-loader`, no codegen |

gRPC port = HTTP port + 1000 (assessment 6000). The port is only exposed on `app-network`,
never in `gateway/kong.yml`.

Client rules:
- Always set a deadline: 3 s by default, 60 s for chatbot streams.
- Retry only idempotent RPCs, and only on `UNAVAILABLE` or `DEADLINE_EXCEEDED`.
- Send `authorization: Bearer <M2M token>` from `AuthService.IssueServiceToken`.

Map errors to status codes:

| Situation | Code |
|---|---|
| Missing or bad input | `INVALID_ARGUMENT` |
| Entity doesn't exist | `NOT_FOUND` |
| Caller not allowed | `PERMISSION_DENIED` |
| Business rule fails | `FAILED_PRECONDITION` |
| State conflict | `ABORTED` |
| Transient error | `UNAVAILABLE` |

## Checklist before finishing

- [ ] Package, options and naming follow the table
- [ ] No field number reused or changed; `reserved` used for removals
- [ ] Every RPC and event has a comment (what it replaces, routing key, consumers)
- [ ] Compiles / `buf lint` passes, `buf breaking` passes against main
- [ ] README / OUTBOX plan / `definitions.json` updated if needed
- [ ] No files under `gen/` added to git
