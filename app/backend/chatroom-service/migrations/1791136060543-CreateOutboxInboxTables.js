// Transactional outbox (consultation.started / consultation.ended -> chatroom.exchange) and consumer
// inbox (appointment.confirmed, appointment.cancelled, profile.updated), deduped by (event_id, consumer).
// Times are Unix ms (BIGINT) like messages.created_at. See app/backend/proto/OUTBOX_MIGRATION_PLAN.md §3.7.
export class CreateOutboxInboxTables1791136060543 {
  name = "CreateOutboxInboxTables1791136060543";

  async up(queryRunner) {
    // 1. Create outbox_events table
    await queryRunner.query(`
      CREATE TABLE IF NOT EXISTS outbox_events (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        aggregate_type VARCHAR(50) NOT NULL,
        aggregate_id VARCHAR(64) NOT NULL,
        aggregate_version BIGINT NOT NULL DEFAULT 0,
        event_type VARCHAR(100) NOT NULL,
        exchange VARCHAR(100) NOT NULL DEFAULT 'chatroom.exchange',
        payload BYTEA NOT NULL,
        payload_type VARCHAR(150) NOT NULL,
        content_type VARCHAR(50) NOT NULL DEFAULT 'application/x-protobuf',
        correlation_id VARCHAR(64),
        status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
        attempt_count INT NOT NULL DEFAULT 0,
        next_attempt_at BIGINT,
        last_attempt_at BIGINT,
        delivered_at BIGINT,
        last_error VARCHAR(500),
        created_at BIGINT NOT NULL,
        CONSTRAINT chk_outbox_events_status CHECK (status IN ('PENDING', 'RETRY_WAIT', 'DELIVERED', 'DEAD')),
        CONSTRAINT chk_outbox_events_attempt_count CHECK (attempt_count >= 0),
        CONSTRAINT chk_outbox_events_retry_time CHECK (status <> 'RETRY_WAIT' OR next_attempt_at IS NOT NULL),
        CONSTRAINT chk_outbox_events_delivered_time CHECK (status <> 'DELIVERED' OR delivered_at IS NOT NULL),
        CONSTRAINT chk_outbox_events_content_type CHECK (content_type IN ('application/x-protobuf', 'application/json'))
      )
    `);

    // 2. Relay poll (FOR UPDATE SKIP LOCKED) only scans undelivered rows
    await queryRunner.query(`
      CREATE INDEX IF NOT EXISTS idx_outbox_events_poll
        ON outbox_events (status, next_attempt_at, created_at)
        WHERE status IN ('PENDING', 'RETRY_WAIT')
    `);
    await queryRunner.query(`CREATE INDEX IF NOT EXISTS idx_outbox_events_aggregate ON outbox_events (aggregate_type, aggregate_id, created_at)`);

    // 3. Create inbox_events table
    await queryRunner.query(`
      CREATE TABLE IF NOT EXISTS inbox_events (
        event_id UUID NOT NULL,
        consumer VARCHAR(100) NOT NULL,
        event_type VARCHAR(100) NOT NULL,
        processed_at BIGINT NOT NULL,
        CONSTRAINT pk_inbox_events PRIMARY KEY (event_id, consumer)
      )
    `);
    await queryRunner.query(`CREATE INDEX IF NOT EXISTS idx_inbox_events_processed_at ON inbox_events (processed_at)`);
  }

  async down(queryRunner) {
    await queryRunner.query(`DROP INDEX IF EXISTS idx_inbox_events_processed_at`);
    await queryRunner.query(`DROP TABLE IF EXISTS inbox_events`);
    await queryRunner.query(`DROP INDEX IF EXISTS idx_outbox_events_aggregate`);
    await queryRunner.query(`DROP INDEX IF EXISTS idx_outbox_events_poll`);
    await queryRunner.query(`DROP TABLE IF EXISTS outbox_events`);
  }
}
