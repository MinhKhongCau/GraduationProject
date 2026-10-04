-- V4 (post-deploy): replay withdrawal events the old REST relay marked DEAD as unsupported.
-- See app/backend/proto/OUTBOX_MIGRATION_PLAN.md §3.1 step 5.
--
-- Run ONLY AFTER the RabbitMQ relay that publishes wallet.withdrawal.* is deployed.
-- Running it earlier is harmless but useless: the old relay marks the rows DEAD again.
-- Guarded on event_type so DEAD booking deliveries (handled by compensation cases) are never replayed.
BEGIN;

UPDATE payment_outbox_events
SET status = 'PENDING',
    attempt_count = 0,
    next_attempt_at = NULL,
    last_error = NULL,
    terminal_reason_code = NULL
WHERE status = 'DEAD'
  AND terminal_reason_code = 'unsupported_event_type'
  AND event_type LIKE 'wallet.withdrawal.%';

COMMIT;

-- No rollback: once replayed, the events have been published and consumers dedupe by event id.
