# Plan: Transactional Outbox with gRPC + RabbitMQ

Status: **proposal** · Scope: all 8 backend services · Contracts: `proto/**/events.proto`

## 1. Why

Current state (from the code, October 2026):

| Service | How it notifies others today | Problem |
|---|---|---|
| auth | `user.created` published to RabbitMQ **after** `accountPort.save()`, error only logged | Account saved but event lost → user has no profile |
| profile | raw-JSON `profile.sync_seed_authors`, RPC replies over `forum.events` | No envelope, no DB write around it, RPC over AMQP |
| forum | `forum.post.created / comment.created / post.liked` fire-and-forget after commit, non-persistent | Lost events, broker down = "[RABBITMQ MOCK]" |
| payment | **Real outbox** (`payment_outbox_events`) but delivered as REST to booking's webhook; `wallet.withdrawal.*` rows go DEAD (`unsupported_event_type`) | Point-to-point, only booking can react, withdrawal events never leave |
| booking | Nothing published; cancel of a paid appointment never reaches payment | No refund, no escrow trigger |
| assessment, chatbot, chatroom | Nothing | — |

Goal: **every state change that another service cares about is written to an outbox
table in the same DB transaction as the change, then relayed to RabbitMQ at-least-once.
Consumers are idempotent. gRPC is used only for synchronous reads/commands.**

## 2. Rules: gRPC vs RabbitMQ

| Use **gRPC** when… | Use **outbox → RabbitMQ event** when… |
|---|---|
| The caller needs the answer to continue (eligibility, access check, enrichment) | Something already happened and others may react (confirmed, cancelled, paid) |
| Read-only or idempotent command | Multiple consumers, or consumer may be down |
| Streaming (chatbot answers) | Cross-service workflow step (saga) |

Hard rules:

1. **Never call gRPC inside a DB transaction** (holds locks during network I/O).
   Pattern: gRPC reads *before* the transaction → transaction (business rows + outbox row) → commit.
2. **Never publish to RabbitMQ from business code.** Only the relay publishes.
3. Events are **facts** (`appointment.cancelled`), not commands (`refund.payment`).
4. Payload = proto message from `events.proto`, wrapped in `common.v1.EventEnvelope`,
   never contains secrets or chat/clinical text.

## 3. Architecture

```text
            ┌─────────────── service A (producer) ────────────────┐
 gRPC/REST  │  usecase                                            │
 ─────────▶ │   BEGIN                                             │
            │     UPDATE appointments ...                         │
            │     INSERT INTO outbox_events (envelope bytes)      │
            │   COMMIT                                            │
            │                                                     │
            │  outbox relay (goroutine / @Scheduled / task)       │
            │   SELECT ... FOR UPDATE SKIP LOCKED LIMIT 100       │
            │   basic.publish(exchange, routing_key, bytes)       │
            │   wait publisher-confirm → UPDATE status=DELIVERED  │
            └───────────────────────────┬─────────────────────────┘
                                        ▼
                         <domain>.exchange (topic, durable)
                                        │ routing key = event_type
                    ┌───────────────────┴────────────────────┐
                    ▼                                        ▼
         <consumer>.queue (quorum, DLX)          <consumer2>.queue
                    │  manual ack
            ┌───────┴──────── service B (consumer) ────────────────┐
            │   BEGIN                                              │
            │     INSERT INTO inbox_events(event_id) ON CONFLICT → │ duplicate → ack & skip
            │     apply business change                            │
            │     (optional) INSERT INTO outbox_events (next event)│
            │   COMMIT → basic.ack                                 │
            │   error → nack → <consumer>.retry (TTL) → back       │
            │   after N attempts → <consumer>.dlq                  │
            └──────────────────────────────────────────────────────┘
```

## 4. Database schema (PostgreSQL, one copy per service DB)

Reference shape below. Per-service table names, time types (`BIGINT` ms vs `TIMESTAMPTZ`)
and migration files are in [OUTBOX_MIGRATION_PLAN.md](./OUTBOX_MIGRATION_PLAN.md).

```sql
-- Producer side
CREATE TABLE outbox_events (
  id                UUID PRIMARY KEY,               -- = EventEnvelope.event_id
  aggregate_type    VARCHAR(50)  NOT NULL,          -- appointment, payment_order, post...
  aggregate_id      VARCHAR(64)  NOT NULL,
  aggregate_version BIGINT       NOT NULL DEFAULT 0,
  event_type        VARCHAR(100) NOT NULL,          -- routing key, e.g. appointment.confirmed
  exchange          VARCHAR(100) NOT NULL,          -- e.g. booking.exchange
  payload           BYTEA        NOT NULL,          -- serialized common.v1.EventEnvelope
  payload_type      VARCHAR(150) NOT NULL,          -- mindcare.booking.v1.AppointmentConfirmed
  correlation_id    VARCHAR(64),
  status            VARCHAR(20)  NOT NULL DEFAULT 'PENDING', -- PENDING|RETRY_WAIT|DELIVERED|DEAD
  attempt_count     INT          NOT NULL DEFAULT 0,
  next_attempt_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
  last_error        VARCHAR(500),
  created_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
  delivered_at      TIMESTAMPTZ
);
CREATE INDEX idx_outbox_poll ON outbox_events (next_attempt_at, created_at)
  WHERE status IN ('PENDING', 'RETRY_WAIT');

-- Consumer side (idempotency)
CREATE TABLE inbox_events (
  event_id     UUID         NOT NULL,
  consumer     VARCHAR(100) NOT NULL,             -- handler name, e.g. booking.on_payment_succeeded
  event_type   VARCHAR(100) NOT NULL,
  processed_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
  PRIMARY KEY (event_id, consumer)
);
```

payment-service already has `payment_outbox_events` with the same delivery columns
(`status`, `attempt_count`, `next_attempt_at`, `last_error`, `terminal_reason_code`).
Plan: add `exchange`, `payload_type`, `correlation_id`, change `payload` to `BYTEA`
(or keep JSON via `protojson` during the transition) instead of creating a new table.

Housekeeping: a daily job deletes `DELIVERED` rows older than 7 days and `inbox_events` older
than 30 days (longer than the max redelivery window).

## 5. Outbox relay

One relay per service instance, polling every 1 s (payment's current 5 s is fine for it):

1. `BEGIN; SELECT * FROM outbox_events WHERE status IN ('PENDING','RETRY_WAIT') AND next_attempt_at <= now() ORDER BY created_at LIMIT 100 FOR UPDATE SKIP LOCKED;`
2. For each row: `basic.publish` with
   - `exchange = row.exchange`, `routing_key = row.event_type`, `mandatory = true`
   - properties: `message_id = id`, `type = payload_type`, `content_type = application/x-protobuf`,
     `delivery_mode = 2`, `correlation_id`, header `x-schema-version`
3. Wait for **publisher confirms** (channel in confirm mode). Ack → `DELIVERED`; nack/return/timeout →
   `RETRY_WAIT`, `attempt_count++`, `next_attempt_at = now() + min(5s·2^n, 5min)`;
   after 10 attempts → `DEAD` + alert.
4. `COMMIT`.

Ordering: rows are read in `created_at` order and `SKIP LOCKED` lets several instances work in
parallel, so strict global order is **not** guaranteed. Consumers must therefore compare
`aggregate_version` (ignore older-or-equal versions) or only rely on state-transition guards
(booking already does: replaying a terminal state is a no-op).

Duplicates: a crash between publish-confirm and `UPDATE` republishes the row → consumers dedupe
with `inbox_events` on `message_id`.

Later option: replace polling with CDC (Debezium on the outbox table → RabbitMQ). Contracts don't
change.

## 6. RabbitMQ topology

Follows `RABBITMQ_CONVENTION.md`. Exchanges are topic + durable; queues are **quorum** queues
with a retry queue and a DLQ each.

| Exchange | Producer | Routing keys (proto message) |
|---|---|---|
| `user.exchange` | auth | `user.created` (UserCreated), `user.updated`, `user.email_verified`, `user.deactivated` |
| `profile.exchange` | profile | `profile.created`, `profile.updated`, `expert.verification_changed`, `profile.seed_authors_synced` |
| `booking.exchange` | booking | `appointment.created/confirmed/cancelled/expired/completed`, `medical_record.created`, `review.created` |
| `payment.exchange` | payment | `payment.succeeded/failed/compensation_required`, `refund.completed`, `wallet.funds_released`, `withdrawal.*` |
| `assessment.exchange` | assessment | `assessment.completed`, `assessment.high_risk_detected` |
| `chatroom.exchange` | chatroom | `consultation.started`, `consultation.ended` |
| `chatbot.exchange` | chatbot | `crisis.detected` |
| `forum.exchange` | forum | `post.published`, `post.archived`, `comment.created`, `post.liked` |
| `mindcare.dlx` | (broker) | dead-letters of every queue |

| Queue | Bindings |
|---|---|
| `profile.queue` | `user.exchange: user.created, user.updated, user.deactivated`; `booking.exchange: review.created` |
| `booking.queue` | `payment.exchange: payment.succeeded, payment.failed`; `chatroom.exchange: consultation.ended`; `profile.exchange: expert.verification_changed` |
| `payment.queue` | `booking.exchange: appointment.cancelled, appointment.expired, appointment.completed` |
| `chatroom.queue` | `booking.exchange: appointment.confirmed, appointment.cancelled`; `profile.exchange: profile.updated` |
| `forum.queue` | `profile.exchange: profile.updated, profile.seed_authors_synced` |
| `chatbot.queue` | `assessment.exchange: assessment.completed` |
| `notification.queue` (future) | `#` on booking, payment, forum, assessment, chatbot exchanges |

Each `<svc>.queue` has `x-dead-letter-exchange = mindcare.dlx`, plus:

- `<svc>.retry` — `x-message-ttl = 10000`, `x-dead-letter-exchange = ""` and
  `x-dead-letter-routing-key = <svc>.queue` (delayed retry back to the main queue)
- `<svc>.dlq` — bound to `mindcare.dlx`

The consumer reads the `x-death` count; when it is ≥ 3 it publishes to `<svc>.dlq`, otherwise to
`<svc>.retry`. This replaces profile-service's current `sleep 2s + republish` loop.
The topology lives in `rabbitmq/definitions.json` (source of truth), and services only *declare
passively* at boot.

## 7. Consumer contract

```text
on message m:
  env = EventEnvelope.decode(m.body)            -- malformed → DLQ immediately (no retry)
  BEGIN
    INSERT INTO inbox_events(event_id, consumer, event_type) VALUES (...) ON CONFLICT DO NOTHING
    if 0 rows inserted: COMMIT; ack; return     -- duplicate
    handle(env.data.unpack())                   -- business change, may INSERT outbox rows
  COMMIT
  ack
on transient error (DB down, gRPC UNAVAILABLE): ROLLBACK; publish to <svc>.retry; ack
on permanent error (validation, FAILED_PRECONDITION): ROLLBACK; publish to <svc>.dlq; ack
```

`prefetch = 10`, manual ack, one channel per consumer goroutine/worker.

## 8. Flows

### 8.1 Register → profile (fixes event loss in auth)

```text
auth  : BEGIN; INSERT account; INSERT outbox(user.created); COMMIT          (@Transactional)
relay : user.exchange / user.created
profile: inbox dedupe → CreateProfileCore (already idempotent) → outbox(profile.created)
```

### 8.2 Booking + payment saga (choreography)

```text
1. patient  → booking  REST  lock slot, create appointment (PENDING_PAYMENT)
                       outbox: appointment.created
2. patient  → payment  REST  POST /orders
   payment  → booking  gRPC  GetPaymentEligibility      (sync, before the tx)
   payment             tx    INSERT payment_order PENDING
3. VNPay    → payment  IPN   tx: order SUCCESS, wallet pending +net, outbox: payment.succeeded
                             (today: outbox type booking.appointment.confirm → REST webhook)
4. booking  ← payment.succeeded
            tx: inbox; appointment CONFIRMED, slot OCCUPIED; outbox: appointment.confirmed
            if conflict (slot gone / appointment cancelled):
               outbox: appointment.cancelled{was_paid=true, refund_percent=100, cancelled_by=SYSTEM}
5. chatroom ← appointment.confirmed → create consultation room (idempotent on appointment_id)
   payment  ← appointment.confirmed (optional) → fulfillment_status = BOOKING_CONFIRMED
```

Failure branch: IPN failed → `payment.failed` → booking cancels and frees the slot.
Timeout branch: booking's expired-lock worker → `appointment.expired` → payment expires the
PENDING order (no more race between order TTL and slot lock).

### 8.3 Cancellation + refund (new)

```text
patient/expert → booking REST cancel
booking  tx: appointment CANCELLED; outbox appointment.cancelled{was_paid, refund_percent}
payment  ← appointment.cancelled (was_paid) → tx: refund record, wallet pending -= net;
           outbox refund.completed (or payment.compensation_required → MANUAL_REVIEW)
chatroom ← appointment.cancelled → close room
```

### 8.4 Consultation end → escrow release

```text
chatroom tx: room ENDED; outbox consultation.ended
booking  ← consultation.ended → appointment COMPLETED; outbox appointment.completed
payment  ← appointment.completed → start hold period; wallet worker releases
           pending → available; outbox wallet.funds_released
```

This replaces "release after `paid_at + HOLD_PERIOD`" with "release after the session really
happened + hold period".

### 8.5 Forum enrichment (removes RPC over RabbitMQ)

`forum → profile gRPC BatchGetUserSummaries` replaces `profile.get_batch.request/response`
(the shared reply queue breaks with >1 forum instance). Forum may cache summaries and refresh
them on `profile.updated`.

## 9. Per-language implementation notes

| Service | Transaction + outbox insert | Relay | Consumer lib |
|---|---|---|---|
| booking / payment / profile / forum (Go) | `db.Transaction(func(tx *gorm.DB) error {...})`, `proto.Marshal(envelope)` | goroutine + `time.Ticker`, `amqp091-go` `Confirm(false)` + `PublishWithDeferredConfirmWithContext` | `amqp091-go`, manual ack |
| auth (Java/Spring) | `@Transactional` on use case, `OutboxRepository.save()` (fixes "no @Transactional anywhere") | `@Scheduled(fixedDelay=1000)` + `RabbitTemplate` with `publisher-confirm-type: correlated` | `@RabbitListener(ackMode=MANUAL)` |
| assessment / chatbot (Python) | same SQLAlchemy `Session` / `with session.begin()` | background asyncio task or separate worker process, `aio-pika` with `publisher_confirms=True` | `aio-pika` |
| chatroom (Node) | TypeORM `dataSource.transaction(async (em) => ...)` | `setInterval` worker, `amqplib` `createConfirmChannel()` | `amqplib` |

Shared code to write once per language (`pkg/outbox`): `Add(tx, aggregate, event proto.Message)`,
`Relay.Run(ctx)`, `Inbox.Handle(tx, envelope, fn)`.

## 10. Observability

- Metrics: `outbox_pending_total`, `outbox_oldest_pending_seconds` (lag), `outbox_dead_total`,
  `consumer_processed_total{result}`, `dlq_messages`.
- Alerts: lag > 60 s, any DEAD row, DLQ depth > 0.
- Logs carry `event_id`, `event_type`, `correlation_id`. The correlation id flows
  HTTP header → gRPC metadata `x-correlation-id` → envelope → AMQP `correlation_id`.
- Admin: payment's compensation-case screens stay; add a "replay DEAD outbox row" admin RPC later.

## 11. Rollout

| Phase | Work | Done when |
|---|---|---|
| 0 | Merge `proto/` contracts, `make lint` in CI (no codegen) | buf lint green |
| 1 | `outbox_events` + `inbox_events` migrations in every service; shared `pkg/outbox` per language; RabbitMQ topology in `definitions.json` | tables exist, relay runs with no rows |
| 2 | **auth**: `@Transactional` register + outbox `user.created` (JSON envelope kept for profile until phase 3) | user.created never lost when broker down |
| 3 | **profile** consumer uses inbox + retry/DLQ queues; switch envelope to protobuf for user.* | old `sleep+republish` loop removed |
| 4 | **payment**: relay publishes `payment.succeeded/failed` and `withdrawal.*` to `payment.exchange` (fix DEAD withdrawal rows); **booking** consumes them via inbox. Keep REST/gRPC `ApplyPaymentResult` as fallback behind a flag for one release | booking confirmed through RabbitMQ in staging, REST delivery disabled |
| 5 | **booking** outbox: appointment.* events; payment consumes cancelled/expired/completed (refund + escrow) | refund flow test passes |
| 6 | gRPC servers: booking (`GetPaymentEligibility`), profile (`BatchGetUserSummaries`), chatbot (`Chat`, `VoiceChat`); clients in payment, forum, chatroom; remove `/internal/*` REST and AMQP-RPC | no `/internal/*` routes left |
| 7 | forum, assessment, chatroom, chatbot outboxes; notification-service consumes `#` | all events in catalog published |

Each phase ships independently; producer and consumer are switched one event type at a time.

## 12. Testing

- Unit: use case writes business row + outbox row in the same tx (rollback → no outbox row).
- Integration (testcontainers Postgres + RabbitMQ): kill broker → events stay PENDING → restart →
  delivered exactly once to the handler (inbox dedupe), delivered ≥ 1 on the wire.
- Chaos: publish the same envelope twice / out of order → state unchanged.
- Contract: `buf breaking` on every PR touching `proto/`.
