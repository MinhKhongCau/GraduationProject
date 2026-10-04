-- V4: extend payment_outbox_events for RabbitMQ relay (protobuf EventEnvelope) and add the inbox.
-- See app/backend/proto/OUTBOX_MIGRATION_PLAN.md §3.1.
--
-- Run BEFORE deploying the binary that writes protobuf events. The current binary keeps
-- working after this file: every new column is either nullable or has a default that matches
-- what it writes today (JSON payload in `payload`, exchange payment.exchange).
-- Withdrawal events marked DEAD by the current relay are replayed by migrate_v4_replay_withdrawals.sql,
-- which runs AFTER the new relay is deployed.
BEGIN;

-- 1. New columns (nullable first, CODING-CONVENTION §5.1).
ALTER TABLE payment_outbox_events
    ADD COLUMN IF NOT EXISTS aggregate_version BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS exchange VARCHAR(100),
    ADD COLUMN IF NOT EXISTS payload_bin BYTEA,
    ADD COLUMN IF NOT EXISTS payload_type VARCHAR(150),
    ADD COLUMN IF NOT EXISTS content_type VARCHAR(50),
    ADD COLUMN IF NOT EXISTS correlation_id VARCHAR(64);

-- 2. Backfill. Every existing row carries a JSON payload in `payload`.
UPDATE payment_outbox_events
SET exchange = 'payment.exchange'
WHERE exchange IS NULL;

UPDATE payment_outbox_events
SET content_type = 'application/json'
WHERE content_type IS NULL;

UPDATE payment_outbox_events
SET payload_type = CASE
    WHEN event_type = 'booking.appointment.confirm' THEN 'mindcare.payment.v1.PaymentSucceeded'
    WHEN event_type = 'booking.appointment.fail' THEN 'mindcare.payment.v1.PaymentFailed'
    WHEN event_type LIKE 'wallet.withdrawal.%' THEN 'mindcare.payment.v1.WithdrawalStatusChanged'
END
WHERE payload_type IS NULL;

-- 3. Required from now on. content_type defaults to JSON until every writer emits protobuf;
--    a later migration flips the default to application/x-protobuf.
ALTER TABLE payment_outbox_events
    ALTER COLUMN exchange SET DEFAULT 'payment.exchange',
    ALTER COLUMN exchange SET NOT NULL,
    ALTER COLUMN content_type SET DEFAULT 'application/json',
    ALTER COLUMN content_type SET NOT NULL;

-- 4. A protobuf row must carry its bytes and its message type. JSON rows keep using `payload`,
--    which becomes optional for protobuf rows (drop the `not null` tag on entity.OutboxEvent.Payload
--    in the same PR that starts writing protobuf).
ALTER TABLE payment_outbox_events
    ALTER COLUMN payload DROP NOT NULL;

ALTER TABLE payment_outbox_events
    DROP CONSTRAINT IF EXISTS chk_payment_outbox_events_json_payload;
ALTER TABLE payment_outbox_events
    ADD CONSTRAINT chk_payment_outbox_events_json_payload
    CHECK (content_type <> 'application/json' OR payload IS NOT NULL);

ALTER TABLE payment_outbox_events
    DROP CONSTRAINT IF EXISTS chk_payment_outbox_events_content_type;
ALTER TABLE payment_outbox_events
    ADD CONSTRAINT chk_payment_outbox_events_content_type
    CHECK (content_type IN ('application/json', 'application/x-protobuf'));

ALTER TABLE payment_outbox_events
    DROP CONSTRAINT IF EXISTS chk_payment_outbox_events_protobuf_payload;
ALTER TABLE payment_outbox_events
    ADD CONSTRAINT chk_payment_outbox_events_protobuf_payload
    CHECK (content_type = 'application/json' OR (payload_bin IS NOT NULL AND payload_type IS NOT NULL));

-- 5. Lookup by aggregate (debugging, replay, "what happened to order X").
CREATE INDEX IF NOT EXISTS ix_payment_outbox_events_aggregate
    ON payment_outbox_events (aggregate_type, aggregate_id, created_at);

-- 6. Inbox: dedupe consumed events (booking.appointment.cancelled, consultation.ended, ...).
CREATE TABLE IF NOT EXISTS payment_inbox_events (
    event_id UUID NOT NULL,
    consumer VARCHAR(100) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    processed_at BIGINT NOT NULL,
    CONSTRAINT pk_payment_inbox_events PRIMARY KEY (event_id, consumer)
);

CREATE INDEX IF NOT EXISTS ix_payment_inbox_events_processed_at
    ON payment_inbox_events (processed_at);

COMMIT;

-- Rollback (manual):
-- BEGIN;
-- DROP TABLE IF EXISTS payment_inbox_events;
-- DROP INDEX IF EXISTS ix_payment_outbox_events_aggregate;
-- ALTER TABLE payment_outbox_events
--     DROP CONSTRAINT IF EXISTS chk_payment_outbox_events_protobuf_payload,
--     DROP CONSTRAINT IF EXISTS chk_payment_outbox_events_content_type,
--     DROP CONSTRAINT IF EXISTS chk_payment_outbox_events_json_payload,
--     ALTER COLUMN payload SET NOT NULL,
--     DROP COLUMN IF EXISTS correlation_id,
--     DROP COLUMN IF EXISTS content_type,
--     DROP COLUMN IF EXISTS payload_type,
--     DROP COLUMN IF EXISTS payload_bin,
--     DROP COLUMN IF EXISTS exchange,
--     DROP COLUMN IF EXISTS aggregate_version;
-- COMMIT;
-- Only safe while no row has content_type = 'application/x-protobuf' (those rows have no JSON payload).
