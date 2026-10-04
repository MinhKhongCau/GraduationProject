-- Consumer inbox: one row per (event_id, consumer) so a redelivered RabbitMQ message is applied once
-- (profile.updated, profile.seed_authors_synced). See app/backend/proto/OUTBOX_MIGRATION_PLAN.md section 3.5.
CREATE TABLE IF NOT EXISTS inbox_events (
    event_id UUID NOT NULL,
    consumer VARCHAR(100) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (event_id, consumer)
);

CREATE INDEX IF NOT EXISTS idx_inbox_events_processed_at ON inbox_events (processed_at);
