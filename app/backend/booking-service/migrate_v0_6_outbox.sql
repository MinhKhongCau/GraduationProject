-- V0.6-OUTBOX: transactional outbox + consumer inbox for RabbitMQ events.
-- See app/backend/proto/OUTBOX_MIGRATION_PLAN.md §3.2.
-- New tables only, no existing data touched. Times are Unix ms (BIGINT) like the other Booking_* tables.
-- Run before deploying the binary that registers the outbox entities in AutoMigrate:
-- AutoMigrate cannot create the partial poll index or the CHECK constraints.
BEGIN;

CREATE TABLE IF NOT EXISTS "Booking_Outbox_Events" (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type VARCHAR(50) NOT NULL,
    aggregate_id VARCHAR(64) NOT NULL,
    aggregate_version BIGINT NOT NULL DEFAULT 0,
    event_type VARCHAR(100) NOT NULL,
    exchange VARCHAR(100) NOT NULL DEFAULT 'booking.exchange',
    payload BYTEA NOT NULL,
    payload_type VARCHAR(150) NOT NULL,
    content_type VARCHAR(50) NOT NULL DEFAULT 'application/x-protobuf',
    correlation_id VARCHAR(64),
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    attempt_count INTEGER NOT NULL DEFAULT 0,
    next_attempt_at BIGINT,
    last_attempt_at BIGINT,
    delivered_at BIGINT,
    last_error VARCHAR(500),
    created_at BIGINT NOT NULL
);

ALTER TABLE "Booking_Outbox_Events"
    DROP CONSTRAINT IF EXISTS chk_booking_outbox_events_status;
ALTER TABLE "Booking_Outbox_Events"
    ADD CONSTRAINT chk_booking_outbox_events_status
    CHECK (status IN ('PENDING', 'RETRY_WAIT', 'DELIVERED', 'DEAD'));

ALTER TABLE "Booking_Outbox_Events"
    DROP CONSTRAINT IF EXISTS chk_booking_outbox_events_attempt_count;
ALTER TABLE "Booking_Outbox_Events"
    ADD CONSTRAINT chk_booking_outbox_events_attempt_count
    CHECK (attempt_count >= 0);

ALTER TABLE "Booking_Outbox_Events"
    DROP CONSTRAINT IF EXISTS chk_booking_outbox_events_retry_time;
ALTER TABLE "Booking_Outbox_Events"
    ADD CONSTRAINT chk_booking_outbox_events_retry_time
    CHECK (status <> 'RETRY_WAIT' OR next_attempt_at IS NOT NULL);

ALTER TABLE "Booking_Outbox_Events"
    DROP CONSTRAINT IF EXISTS chk_booking_outbox_events_delivered_time;
ALTER TABLE "Booking_Outbox_Events"
    ADD CONSTRAINT chk_booking_outbox_events_delivered_time
    CHECK (status <> 'DELIVERED' OR delivered_at IS NOT NULL);

ALTER TABLE "Booking_Outbox_Events"
    DROP CONSTRAINT IF EXISTS chk_booking_outbox_events_content_type;
ALTER TABLE "Booking_Outbox_Events"
    ADD CONSTRAINT chk_booking_outbox_events_content_type
    CHECK (content_type IN ('application/x-protobuf', 'application/json'));

-- Relay poll: ... WHERE status IN ('PENDING','RETRY_WAIT') AND (next_attempt_at IS NULL OR next_attempt_at <= $now)
--             ORDER BY created_at LIMIT n FOR UPDATE SKIP LOCKED
CREATE INDEX IF NOT EXISTS idx_booking_outbox_events_poll
    ON "Booking_Outbox_Events" (status, next_attempt_at, created_at)
    WHERE status IN ('PENDING', 'RETRY_WAIT');

CREATE INDEX IF NOT EXISTS idx_booking_outbox_events_aggregate
    ON "Booking_Outbox_Events" (aggregate_type, aggregate_id, created_at);

CREATE TABLE IF NOT EXISTS "Booking_Inbox_Events" (
    event_id UUID NOT NULL,
    consumer VARCHAR(100) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    processed_at BIGINT NOT NULL,
    CONSTRAINT pk_booking_inbox_events PRIMARY KEY (event_id, consumer)
);

CREATE INDEX IF NOT EXISTS idx_booking_inbox_events_processed_at
    ON "Booking_Inbox_Events" (processed_at);

COMMIT;

-- Rollback (manual, only while no relay/consumer is deployed):
-- BEGIN;
-- DROP TABLE IF EXISTS "Booking_Inbox_Events";
-- DROP TABLE IF EXISTS "Booking_Outbox_Events";
-- COMMIT;
