-- V2: durable REST delivery state for payment outbox events.
BEGIN;

ALTER TABLE payment_outbox_events
    ADD COLUMN IF NOT EXISTS status VARCHAR(20),
    ADD COLUMN IF NOT EXISTS attempt_count INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS next_attempt_at BIGINT,
    ADD COLUMN IF NOT EXISTS last_attempt_at BIGINT,
    ADD COLUMN IF NOT EXISTS delivered_at BIGINT,
    ADD COLUMN IF NOT EXISTS last_error VARCHAR(500);

-- Legacy published=true is the only persisted evidence that the old worker
-- completed an event. Preserve that history conservatively as DELIVERED.
UPDATE payment_outbox_events
SET status = 'DELIVERED',
    delivered_at = COALESCE(delivered_at, created_at)
WHERE published IS TRUE
  AND (status IS NULL OR status IN ('PENDING', 'DELIVERED'));

-- Every legacy undelivered row remains eligible after deployment.
UPDATE payment_outbox_events
SET status = 'PENDING'
WHERE status IS NULL
  AND published IS NOT TRUE;

UPDATE payment_outbox_events
SET attempt_count = 0
WHERE attempt_count IS NULL;

ALTER TABLE payment_outbox_events
    ALTER COLUMN status SET DEFAULT 'PENDING',
    ALTER COLUMN status SET NOT NULL,
    ALTER COLUMN attempt_count SET DEFAULT 0,
    ALTER COLUMN attempt_count SET NOT NULL;

ALTER TABLE payment_outbox_events
    DROP CONSTRAINT IF EXISTS chk_payment_outbox_events_status;

ALTER TABLE payment_outbox_events
    ADD CONSTRAINT chk_payment_outbox_events_status
    CHECK (status IN ('PENDING', 'RETRY_WAIT', 'DELIVERED', 'DEAD'));

ALTER TABLE payment_outbox_events
    DROP CONSTRAINT IF EXISTS chk_payment_outbox_events_attempt_count;

ALTER TABLE payment_outbox_events
    ADD CONSTRAINT chk_payment_outbox_events_attempt_count
    CHECK (attempt_count >= 0);

ALTER TABLE payment_outbox_events
    DROP CONSTRAINT IF EXISTS chk_payment_outbox_events_retry_time;

ALTER TABLE payment_outbox_events
    ADD CONSTRAINT chk_payment_outbox_events_retry_time
    CHECK (status <> 'RETRY_WAIT' OR next_attempt_at IS NOT NULL);

ALTER TABLE payment_outbox_events
    DROP CONSTRAINT IF EXISTS chk_payment_outbox_events_delivered_time;

ALTER TABLE payment_outbox_events
    ADD CONSTRAINT chk_payment_outbox_events_delivered_time
    CHECK (status <> 'DELIVERED' OR delivered_at IS NOT NULL);

CREATE INDEX IF NOT EXISTS ix_payment_outbox_events_delivery_poll
    ON payment_outbox_events (status, next_attempt_at, created_at)
    WHERE status IN ('PENDING', 'RETRY_WAIT');

COMMIT;
