-- V1: transactional outbox + consumer inbox for RabbitMQ events (user.created, user.updated, ...).
-- See app/backend/proto/OUTBOX_MIGRATION_PLAN.md §3.3.
-- Table names follow the Identity_* entities. They are unquoted, so PostgreSQL stores them lower-case,
-- which is what Hibernate's default naming strategy produces for @Table(name = "Identity_Outbox_Events").
-- Flyway runs this file inside a transaction.

CREATE TABLE IF NOT EXISTS Identity_Outbox_Events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type VARCHAR(50) NOT NULL,
    aggregate_id VARCHAR(64) NOT NULL,
    aggregate_version BIGINT NOT NULL DEFAULT 0,
    event_type VARCHAR(100) NOT NULL,
    exchange VARCHAR(100) NOT NULL DEFAULT 'user.exchange',
    payload BYTEA NOT NULL,
    payload_type VARCHAR(150) NOT NULL,
    content_type VARCHAR(50) NOT NULL DEFAULT 'application/x-protobuf',
    correlation_id VARCHAR(64),
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    attempt_count INTEGER NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMPTZ,
    last_attempt_at TIMESTAMPTZ,
    delivered_at TIMESTAMPTZ,
    last_error VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_identity_outbox_events_status CHECK (status IN ('PENDING', 'RETRY_WAIT', 'DELIVERED', 'DEAD')),
    CONSTRAINT chk_identity_outbox_events_attempt_count CHECK (attempt_count >= 0),
    CONSTRAINT chk_identity_outbox_events_retry_time CHECK (status <> 'RETRY_WAIT' OR next_attempt_at IS NOT NULL),
    CONSTRAINT chk_identity_outbox_events_delivered_time CHECK (status <> 'DELIVERED' OR delivered_at IS NOT NULL),
    CONSTRAINT chk_identity_outbox_events_content_type CHECK (content_type IN ('application/x-protobuf', 'application/json'))
);

-- Relay poll (FOR UPDATE SKIP LOCKED) only scans undelivered rows.
CREATE INDEX IF NOT EXISTS idx_identity_outbox_events_poll
    ON Identity_Outbox_Events (status, next_attempt_at, created_at)
    WHERE status IN ('PENDING', 'RETRY_WAIT');

CREATE INDEX IF NOT EXISTS idx_identity_outbox_events_aggregate
    ON Identity_Outbox_Events (aggregate_type, aggregate_id, created_at);

CREATE TABLE IF NOT EXISTS Identity_Inbox_Events (
    event_id UUID NOT NULL,
    consumer VARCHAR(100) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT pk_identity_inbox_events PRIMARY KEY (event_id, consumer)
);

CREATE INDEX IF NOT EXISTS idx_identity_inbox_events_processed_at
    ON Identity_Inbox_Events (processed_at);

-- Rollback (Flyway Community has no undo): run manually, then
-- DELETE FROM flyway_schema_history WHERE version = '1';
--   DROP TABLE IF EXISTS Identity_Inbox_Events;
--   DROP TABLE IF EXISTS Identity_Outbox_Events;
