-- V1: transactional outbox + consumer inbox for RabbitMQ events (profile.*, expert.*).
-- See app/backend/proto/OUTBOX_MIGRATION_PLAN.md §3.4.
-- profile-service schema is otherwise managed by GORM AutoMigrate (config/database.go). This file
-- holds what AutoMigrate cannot express (partial index, CHECKs). Run it manually before deploying
-- the binary that writes outbox rows.
BEGIN;

CREATE TABLE IF NOT EXISTS outbox_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type VARCHAR(50) NOT NULL,
    aggregate_id VARCHAR(64) NOT NULL,
    aggregate_version BIGINT NOT NULL DEFAULT 0,
    event_type VARCHAR(100) NOT NULL,
    exchange VARCHAR(100) NOT NULL DEFAULT 'profile.exchange',
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
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE outbox_events DROP CONSTRAINT IF EXISTS chk_outbox_events_status;
ALTER TABLE outbox_events ADD CONSTRAINT chk_outbox_events_status
    CHECK (status IN ('PENDING', 'RETRY_WAIT', 'DELIVERED', 'DEAD'));

ALTER TABLE outbox_events DROP CONSTRAINT IF EXISTS chk_outbox_events_attempt_count;
ALTER TABLE outbox_events ADD CONSTRAINT chk_outbox_events_attempt_count
    CHECK (attempt_count >= 0);

ALTER TABLE outbox_events DROP CONSTRAINT IF EXISTS chk_outbox_events_retry_time;
ALTER TABLE outbox_events ADD CONSTRAINT chk_outbox_events_retry_time
    CHECK (status <> 'RETRY_WAIT' OR next_attempt_at IS NOT NULL);

ALTER TABLE outbox_events DROP CONSTRAINT IF EXISTS chk_outbox_events_delivered_time;
ALTER TABLE outbox_events ADD CONSTRAINT chk_outbox_events_delivered_time
    CHECK (status <> 'DELIVERED' OR delivered_at IS NOT NULL);

ALTER TABLE outbox_events DROP CONSTRAINT IF EXISTS chk_outbox_events_content_type;
ALTER TABLE outbox_events ADD CONSTRAINT chk_outbox_events_content_type
    CHECK (content_type IN ('application/x-protobuf', 'application/json'));

-- Relay poll (FOR UPDATE SKIP LOCKED) only scans undelivered rows.
CREATE INDEX IF NOT EXISTS idx_outbox_events_poll
    ON outbox_events (status, next_attempt_at, created_at)
    WHERE status IN ('PENDING', 'RETRY_WAIT');

CREATE INDEX IF NOT EXISTS idx_outbox_events_aggregate
    ON outbox_events (aggregate_type, aggregate_id, created_at);

CREATE TABLE IF NOT EXISTS inbox_events (
    event_id UUID NOT NULL,
    consumer VARCHAR(100) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT pk_inbox_events PRIMARY KEY (event_id, consumer)
);

CREATE INDEX IF NOT EXISTS idx_inbox_events_processed_at
    ON inbox_events (processed_at);

COMMIT;

-- Rollback (manual, only while no relay/consumer is deployed):
-- BEGIN;
-- DROP TABLE IF EXISTS inbox_events;
-- DROP TABLE IF EXISTS outbox_events;
-- COMMIT;
