-- Transactional outbox for forum events (post.published, post.archived, comment.created, post.liked).
-- Written in the same transaction as the business change and published to forum.exchange by the relay.
-- See app/backend/proto/OUTBOX_MIGRATION_PLAN.md section 3.5. aggregate_id is VARCHAR because post and
-- comment ids are BIGINT here while other services use UUID.
-- Every statement is idempotent: x-multi-statement runs them one by one, outside a transaction.
CREATE TABLE IF NOT EXISTS outbox_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type VARCHAR(50) NOT NULL,
    aggregate_id VARCHAR(64) NOT NULL,
    aggregate_version BIGINT NOT NULL DEFAULT 0,
    event_type VARCHAR(100) NOT NULL,
    exchange VARCHAR(100) NOT NULL DEFAULT 'forum.exchange',
    payload BYTEA NOT NULL,
    payload_type VARCHAR(150) NOT NULL,
    content_type VARCHAR(50) NOT NULL DEFAULT 'application/x-protobuf',
    correlation_id VARCHAR(64),
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    attempt_count INTEGER NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMP WITH TIME ZONE,
    last_attempt_at TIMESTAMP WITH TIME ZONE,
    delivered_at TIMESTAMP WITH TIME ZONE,
    last_error VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_outbox_events_status CHECK (status IN ('PENDING', 'RETRY_WAIT', 'DELIVERED', 'DEAD')),
    CONSTRAINT chk_outbox_events_attempt_count CHECK (attempt_count >= 0),
    CONSTRAINT chk_outbox_events_retry_time CHECK (status <> 'RETRY_WAIT' OR next_attempt_at IS NOT NULL),
    CONSTRAINT chk_outbox_events_delivered_time CHECK (status <> 'DELIVERED' OR delivered_at IS NOT NULL),
    CONSTRAINT chk_outbox_events_content_type CHECK (content_type IN ('application/x-protobuf', 'application/json'))
);

-- Relay poll (FOR UPDATE SKIP LOCKED) only scans undelivered rows.
CREATE INDEX IF NOT EXISTS idx_outbox_events_poll
    ON outbox_events (status, next_attempt_at, created_at)
    WHERE status IN ('PENDING', 'RETRY_WAIT');

CREATE INDEX IF NOT EXISTS idx_outbox_events_aggregate
    ON outbox_events (aggregate_type, aggregate_id, created_at);
