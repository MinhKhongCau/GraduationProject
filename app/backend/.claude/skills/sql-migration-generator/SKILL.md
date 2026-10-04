---
name: sql-migration-generator
description: 'Write PostgreSQL schema/data migrations for MindCare backend services in
  the format each service actually uses (raw SQL, golang-migrate, Alembic, TypeORM,
  GORM AutoMigrate, JPA). Use when adding/changing tables, columns, indexes or
  constraints, backfilling data, or creating outbox/inbox tables. Triggers: "migration",
  "migrate", "alter table", "add column", "sql", "schema change", "outbox table".'
allowed-tools: Read, Write, Edit, Grep, Glob
version: 2.0.0
tags:
- backend
- database
- postgresql
- migrations
---
# MindCare SQL migrations

Every service has its own PostgreSQL database on the shared `postgres-db` container (pgvector
pg16): `auth_db`, `profile_db`, `payment_db`, `booking_db`, `assessment_db`, `forum_db`,
`chatbot_db` and `chatroom_db`, created by `init-db.sh`. **Never** touch another service's
database or add cross-database foreign keys.

## 1. Find the service's mechanism first

| Service | Mechanism | Where / naming | Runs |
|---|---|---|---|
| payment (Go) | GORM `AutoMigrate` + hand-written SQL | `payment-service/migrate_vN.sql` (next: `migrate_v5.sql`) | AutoMigrate on boot; SQL run manually |
| booking (Go) | GORM `AutoMigrate` + pre-migration + hand-written SQL | `booking-service/migrate_v<major>_<minor>_<tag>.sql` (latest `migrate_v0_6_outbox.sql`) | AutoMigrate on boot; SQL run manually |
| profile (Go) | GORM `AutoMigrate` only | `profile-service/config/database.go`. Plus `profile-service/migrations/NNN_<name>.sql` (next `002`) for anything AutoMigrate can't do (partial indexes, CHECKs) | boot |
| forum (Go) | **golang-migrate** embedded | `forum-service/internal/repository/db/migrations/NNNNNN_<name>.up.sql` + `.down.sql` (6 digits, next `000012`) | `RunMigrations` on boot |
| assessment (Python) | **Alembic** | `assessment-service/migrations/versions/NNNN_<name>.py`, `revision = "NNNN_<name>"`, `down_revision` = previous file's revision (next `0009`) | `alembic upgrade head` |
| chatroom (Node) | **TypeORM** migrations, `synchronize: false` | `chatroom-service/migrations/<epochMs>-<PascalName>.js`. ESM class with `up`/`down` using `queryRunner.query`, and **register it in `config/db.js` `migrations: [...]`** | `dataSource.runMigrations()` on boot |
| auth (Java) | JPA `ddl-auto: update` + **Flyway** (baseline 0) | `auth-service/src/main/resources/db/migration/V<N>__<name>.sql` (next `V2`). Unquoted `Identity_*` names (stored lower-case, as Hibernate expects) | Flyway on boot, before Hibernate |
| chatbot (Python) | **Alembic** (PGVector tables unmanaged) | `chatbot-service/migrations/versions/NNNN_<name>.py` (next `0002`) | `alembic upgrade head` in Dockerfile CMD |

If the table is also declared as a GORM/JPA/SQLAlchemy/TypeORM model, keep the model and the
migration in sync: same column names, types and nullability.

## 2. SQL style (match existing files)

- First line is a comment: `-- V<N>: <one-line purpose>`, plus why for non-obvious steps.
- Wrap raw SQL files in `BEGIN; ... COMMIT;`. Alembic, TypeORM and Flyway already run in a tx.
  forum's golang-migrate uses `x-multi-statement` (statements run one by one, **no** tx, naive `;`
  split): keep every statement idempotent and never put `;` inside comments or `$$` bodies.
- **Idempotent**: `CREATE TABLE IF NOT EXISTS`, `ADD COLUMN IF NOT EXISTS`,
  `CREATE INDEX IF NOT EXISTS`, and `DROP CONSTRAINT IF EXISTS` before `ADD CONSTRAINT`.
- Names:
  - indexes `ix_<table>_<purpose>` (payment) or `idx_<table>_<cols>` (booking/forum); follow the file you are next to
  - constraints `chk_<table>_<rule>`, `uq_<table>_<cols>`
- Table names follow the service:
  - payment uses `payment_*`
  - booking uses quoted PascalCase `"Booking_*"`
  - forum, assessment and chatroom use plain snake_case
- Keep each service's time convention:
  - payment, booking and chatroom: `BIGINT` Unix **milliseconds**
  - forum, assessment, profile and chatbot: `TIMESTAMPTZ`
- UUID PK: `UUID PRIMARY KEY DEFAULT gen_random_uuid()`.
- Enum-like columns: `VARCHAR(n)` + `CHECK (col IN (...))`. No PostgreSQL `ENUM` types.

## 3. Safety rules

- **Required column on a table with data** (CODING-CONVENTION §5.1), in three steps:
  1. add it nullable
  2. backfill
  3. `SET NOT NULL`

  Separate migrations for golang-migrate/Alembic/TypeORM; separate statements in one tx for raw SQL files.
- Indexes on big hot tables: `CREATE INDEX CONCURRENTLY` in its **own** migration, outside a tx.
  For golang-migrate this means a file with no other statements.
- Never `DROP COLUMN` or `DROP TABLE` in the same release that stops using it. Stop writing first, drop one release later.
- Never rename a column in place. Instead:
  1. add the new column
  2. dual-write
  3. backfill
  4. switch reads
  5. drop the old one
- Every golang-migrate, Alembic and TypeORM migration must have a working `down`.
  For raw SQL files, add a commented rollback block at the end.
- Seed data goes in a separate migration from DDL.
- No secrets in migrations.

## 4. Outbox / inbox tables (see `proto/OUTBOX_PATTERN_PLAN.md`)

Use the plan in `proto/OUTBOX_MIGRATION_PLAN.md`. Each service gets:

- `<prefix>outbox_events`, with columns:
  - identity and routing: `id`, `aggregate_type`, `aggregate_id`, `aggregate_version`, `event_type`, `exchange`
  - payload: `payload BYTEA` (protobuf `EventEnvelope`), `payload_type`, `content_type`
  - tracing: `correlation_id`
  - delivery: `status`, `attempt_count`, `next_attempt_at`, `last_attempt_at`, `delivered_at`, `last_error`
  - `created_at`
  - plus a CHECK on status and a partial poll index `WHERE status IN ('PENDING','RETRY_WAIT')`
- `<prefix>inbox_events`: `(event_id, consumer)` primary key, `event_type`, `processed_at`.

payment already has `payment_outbox_events`: **extend it, don't recreate it**.

## 5. Checklist

- [ ] Used the service's own mechanism, path and next sequence number
- [ ] Idempotent DDL, transaction wrapper / down migration present
- [ ] Model (GORM/JPA/SQLAlchemy/TypeORM entity) updated to match
- [ ] Time and naming conventions of that service respected
- [ ] NOT NULL on existing data done in add → backfill → set-not-null order
- [ ] TypeORM migration registered in `config/db.js`; Alembic `down_revision` chain correct
- [ ] Rollback steps documented
