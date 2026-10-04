# Plan: Outbox / Inbox migration files

Status: **plan only, no migration files created yet**
Implements phase 1 of [OUTBOX_PATTERN_PLAN.md](./OUTBOX_PATTERN_PLAN.md) §4 and §11.
Follows `.claude/skills/sql-migration-generator/SKILL.md` and CODING-CONVENTION §5.

## 1. Goal

Give every service, in its **own database** and with its **own migration mechanism**:

- an **outbox table**: written in the same transaction as the business change, read by the relay
- an **inbox table**: dedupes consumed events by `(event_id, consumer)`

The schema is the same everywhere. Only the naming and time type follow each service.

## 2. Canonical columns

| Column | Type | Null | Notes |
|---|---|---|---|
| `id` | `UUID` PK `DEFAULT gen_random_uuid()` | no | = `EventEnvelope.event_id` = AMQP `message_id` |
| `aggregate_type` | `VARCHAR(50)` | no | `appointment`, `payment_order`, `post`… |
| `aggregate_id` | `VARCHAR(64)` | no | string because forum ids are `BIGINT` (payment keeps its existing `UUID`) |
| `aggregate_version` | `BIGINT` | no, default 0 | consumer ignores versions ≤ last applied |
| `event_type` | `VARCHAR(100)` | no | routing key, e.g. `appointment.cancelled` |
| `exchange` | `VARCHAR(100)` | no | e.g. `booking.exchange` |
| `payload` | `BYTEA` | no | serialized `mindcare.common.v1.EventEnvelope` |
| `payload_type` | `VARCHAR(150)` | no | e.g. `mindcare.booking.v1.AppointmentCancelled` |
| `content_type` | `VARCHAR(50)` | no, default `application/x-protobuf` | `application/json` allowed during transition |
| `correlation_id` | `VARCHAR(64)` | yes | from gRPC metadata / HTTP header |
| `status` | `VARCHAR(20)` + CHECK | no, default `PENDING` | `PENDING`, `RETRY_WAIT`, `DELIVERED`, `DEAD` |
| `attempt_count` | `INTEGER` + CHECK ≥ 0 | no, default 0 | |
| `next_attempt_at` | *time* | yes | NULL = due now |
| `last_attempt_at` | *time* | yes | |
| `delivered_at` | *time* | yes | CHECK: required when DELIVERED |
| `last_error` | `VARCHAR(500)` | yes | |
| `created_at` | *time* | no | |

**Inbox**

| Column | Type | Notes |
|---|---|---|
| `event_id` | `UUID` | PK part 1 |
| `consumer` | `VARCHAR(100)` | PK part 2, e.g. `booking.on_payment_succeeded` |
| `event_type` | `VARCHAR(100)` | |
| `processed_at` | *time* | |

**Indexes and constraints**

- `chk_<t>_status`, `chk_<t>_attempt_count`, `chk_<t>_delivered_time`, `chk_<t>_retry_time` (same rules as payment `migrate_v2.sql`)
- outbox poll: `(status, next_attempt_at, created_at) WHERE status IN ('PENDING','RETRY_WAIT')`
- outbox lookup: `(aggregate_type, aggregate_id, created_at)`
- inbox cleanup: `(processed_at)`

*time* means `BIGINT` Unix ms for payment, booking and chatroom, and `TIMESTAMPTZ` for every
other service. This matches each service's existing columns.

## 3. Files per service

Ordered by rollout phase. Each row is one service PR.

### 3.1 payment-service (phase 4): **extend** the existing outbox

`payment_outbox_events` already exists (`migrate_v2.sql`, `migrate_v3.sql`) and holds live rows.
**File:** `payment-service/migrate_v4.sql` (raw SQL, `BEGIN/COMMIT`, idempotent).

Steps:
1. Add the new columns:
   - `exchange VARCHAR(100)`
   - `payload_type VARCHAR(150)`
   - `content_type VARCHAR(50)`
   - `correlation_id VARCHAR(64)`
   - `aggregate_version BIGINT NOT NULL DEFAULT 0`
   - `payload_bin BYTEA`

   `exchange`, `payload_type` and `content_type` are added **nullable**. Existing `payload TEXT` (JSON) is kept: rows created before the switch have no protobuf bytes.
2. Backfill existing rows:
   - `content_type = 'application/json'`
   - `exchange = 'payment.exchange'`
   - `payload_type` mapped from `event_type`:
     - `booking.appointment.confirm` → `mindcare.payment.v1.PaymentSucceeded`
     - `booking.appointment.fail` → `mindcare.payment.v1.PaymentFailed`
     - `wallet.withdrawal.*` → `mindcare.payment.v1.WithdrawalStatusChanged`
3. `SET NOT NULL` on `exchange`, `payload_type` and `content_type`, and set the defaults
   (`'payment.exchange'`, `'application/x-protobuf'`).
4. Add a CHECK: `payload_bin IS NOT NULL OR content_type = 'application/json'`.
5. Resurrect the withdrawal rows wrongly marked `DEAD` with `terminal_reason_code` = unsupported event type:
   - `status = 'PENDING'`, `attempt_count = 0`, `next_attempt_at = NULL`, `last_error = NULL`
   - **Guarded by a WHERE on `event_type LIKE 'wallet.withdrawal.%'`**, so DEAD booking deliveries are not replayed.
6. Add the index `ix_payment_outbox_events_aggregate (aggregate_type, aggregate_id, created_at)`.
7. Create `payment_inbox_events` (BIGINT ms `processed_at`) + `ix_payment_inbox_events_processed_at`.
8. Commented rollback block: drop the inbox table and the new columns/constraints, but **not** the step 5 data fix.

Code to update in the same PR:
- `internal/domain/entity/outbox.go`: new fields
- a new `entity/inbox.go` added to `pkg/database/postgres.go` AutoMigrate list

Run order: `migrate_v4.sql` **before** deploying the new binary, so AutoMigrate only sees existing columns.

Deliberately not done: renaming `payload` → `payload_json` or dropping `published`. That waits for a later release, after all rows are protobuf.

### 3.2 booking-service (phase 4–5): new tables

**File:** `booking-service/migrate_v0_6_outbox.sql` (raw SQL, `BEGIN/COMMIT`).

- `"Booking_Outbox_Events"` with all canonical columns, BIGINT ms times
- `"Booking_Inbox_Events"`
- indexes `idx_booking_outbox_events_poll`, `idx_booking_outbox_events_aggregate`, `idx_booking_inbox_events_processed_at`

Code: add `domain/outbox_event.go` and `domain/inbox_event.go` with `TableName()` returning the
quoted names, and register them in `pkg/database/postgres.go` AutoMigrate.
AutoMigrate can't create the partial index and CHECKs, so the SQL file is the source of truth.
Run it before deploy.

### 3.3 auth-service (phase 2): `user.created` must stop being lost

JPA `ddl-auto: update` can create the tables but not the partial index or CHECKs.

- **Option A (recommended): introduce Flyway.**
  - Add the `flyway-core` and `flyway-database-postgresql` dependencies.
  - Set `spring.flyway.baseline-on-migrate: true`, `baseline-version: 0`.
  - Add `src/main/resources/db/migration/V1__create_outbox_inbox.sql`.
  - Keep `ddl-auto: update` for now and switch to `validate` in a later PR.
  - **Needs team approval** (new dependency, changes how auth schema is managed).
- **Option B:** JPA entities `OutboxEvent` / `InboxEvent` only, plus a manual SQL file
  `auth-service/sql/outbox_indexes.sql` for the partial index and CHECKs.

Tables:
- `Identity_Outbox_Events` and `Identity_Inbox_Events`, following the existing `Identity_*` prefix (`@Table(name = "Identity_Outbox_Events")`)
- `TIMESTAMPTZ` times (auth uses `Long` epoch for its own entities; the outbox is new, so `Instant` + `TIMESTAMPTZ` is fine. Choose one and keep it.)

Code: the `@Transactional` register use case writes the account plus the outbox row.

### 3.4 profile-service (phase 3)

GORM AutoMigrate only, no migrations folder yet.

**File:** `profile-service/migrations/001_create_outbox_inbox.sql`. This creates the folder and is run manually like payment/booking. Contents:
- `outbox_events`, `inbox_events` (TIMESTAMPTZ)
- the partial poll index and CHECKs

Code: `internal/models/outbox.go`, `inbox.go`, added to `config/database.go` AutoMigrate.

### 3.5 forum-service (phase 7)

golang-migrate, embedded, runs on boot.

**Files:**
- `internal/repository/db/migrations/000010_create_outbox_events_table.up.sql` / `.down.sql`
- `internal/repository/db/migrations/000011_create_inbox_events_table.up.sql` / `.down.sql`

Tables: `outbox_events`, `inbox_events`, TIMESTAMPTZ, `aggregate_id VARCHAR(64)` (post ids are BIGINT).
`down` = `DROP TABLE IF EXISTS ...`.

### 3.6 assessment-service (phase 7)

Alembic.

**File:** `migrations/versions/0008_create_outbox_inbox.py`
- `revision = "0008_create_outbox_inbox"`, `down_revision = "0007_link_question_dimensions"`
- `op.create_table(...)` for `outbox_events` and `inbox_events`
- `op.create_index(..., postgresql_where=sa.text("status IN ('PENDING','RETRY_WAIT')"))`
- `op.create_check_constraint(...)`
- `downgrade()` drops both tables

Code: SQLAlchemy models in `app/models.py`.

### 3.7 chatroom-service (phase 7)

TypeORM, `synchronize: false`.

**File:** `migrations/<Date.now()>-CreateOutboxInboxTables.js`
- class `CreateOutboxInboxTables<ts>`
- `up` creates the `outbox_events` and `inbox_events` tables via `queryRunner.query` (BIGINT ms, matching `messages.created_at`)
- `down` drops them

Code:
- `entities/outbox-event.entity.js` and `inbox-event.entity.js`
- **register the migration and entities in `config/db.js`**

### 3.8 chatbot-service (phase 7)

No migration tool today. **Needs team approval:**
- add `alembic` to `requirements.txt` and `chatbot-service/alembic.ini`
- add `migrations/versions/0001_create_outbox_inbox.py` (only the new tables; existing `chat_sessions` / `chat_messages` and langchain tables are not touched)
- run `alembic upgrade head` in the Dockerfile CMD before `ingest_vector_db.py`

Fallback if Alembic is rejected: `chatbot-service/migrations/001_create_outbox_inbox.sql`, run manually.

## 4. Summary

| # | Service | File(s) | Mechanism | Tables | Time type | Needs approval |
|---|---|---|---|---|---|---|
| 1 | payment | `migrate_v4.sql` | raw SQL (manual) | extend `payment_outbox_events`, new `payment_inbox_events` | BIGINT ms | no |
| 2 | booking | `migrate_v0_6_outbox.sql` | raw SQL (manual) + AutoMigrate | `"Booking_Outbox_Events"`, `"Booking_Inbox_Events"` | BIGINT ms | no |
| 3 | auth | `db/migration/V1__create_outbox_inbox.sql` | Flyway (new) | `Identity_Outbox_Events`, `Identity_Inbox_Events` | TIMESTAMPTZ | **yes** (Flyway) |
| 4 | profile | `migrations/001_create_outbox_inbox.sql` | raw SQL (manual) + AutoMigrate | `outbox_events`, `inbox_events` | TIMESTAMPTZ | no |
| 5 | forum | `000010_*`, `000011_*` up/down | golang-migrate | `outbox_events`, `inbox_events` | TIMESTAMPTZ | no |
| 6 | assessment | `0008_create_outbox_inbox.py` | Alembic | `outbox_events`, `inbox_events` | TIMESTAMPTZ | no |
| 7 | chatroom | `<ts>-CreateOutboxInboxTables.js` | TypeORM | `outbox_events`, `inbox_events` | BIGINT ms | no |
| 8 | chatbot | `0001_create_outbox_inbox.py` | Alembic (new) | `outbox_events`, `inbox_events` | TIMESTAMPTZ | **yes** (Alembic) |

## 5. Verification per migration

1. Run against a fresh DB (`docker compose -f docker-compose.dev.yml up postgres-db`) **and**
   against a copy of current data (payment especially: check the backfill counts with
   `SELECT status, content_type, count(*) ... GROUP BY 1,2` before and after).
2. Run it twice: the second run must be a no-op (idempotency).
3. Run `down`/rollback, then `up` again.
4. Boot the service: AutoMigrate/JPA must not try to alter the new tables (no diff).
5. `EXPLAIN` the relay query uses the partial poll index:
   `SELECT ... WHERE status IN ('PENDING','RETRY_WAIT') AND (next_attempt_at IS NULL OR next_attempt_at <= $1) ORDER BY created_at LIMIT 100 FOR UPDATE SKIP LOCKED`.
6. Insert a duplicate `(event_id, consumer)` into the inbox → unique violation (dedupe works).

## 6. Retention (later migration, not in phase 1)

Retention is a scheduled job, not a migration. It deletes:
- `DELIVERED` outbox rows older than 7 days
- inbox rows older than 30 days

If tables grow large, a later migration may switch the outbox to `PARTITION BY RANGE (created_at)`.
