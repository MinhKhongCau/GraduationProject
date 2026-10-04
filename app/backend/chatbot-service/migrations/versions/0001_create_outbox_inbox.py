"""create outbox_events and inbox_events for RabbitMQ events

First Alembic revision of chatbot-service. Only creates the new tables:
the langchain-postgres vector tables are owned by PGVector and untouched.

Transactional outbox for crisis.detected (published to chatbot.exchange by
the relay) and a consumer inbox that dedupes redelivered messages
(assessment.completed) by (event_id, consumer).

See app/backend/proto/OUTBOX_MIGRATION_PLAN.md section 3.8.

Revision ID: 0001_create_outbox_inbox
Revises:
Create Date: 2026-10-05

"""
import sqlalchemy as sa
from alembic import op
from sqlalchemy.dialects import postgresql

# revision identifiers, used by Alembic.
revision = "0001_create_outbox_inbox"
down_revision = None
branch_labels = None
depends_on = None


def upgrade() -> None:
    op.create_table(
        "outbox_events",
        sa.Column(
            "id",
            postgresql.UUID(as_uuid=True),
            primary_key=True,
            server_default=sa.text("gen_random_uuid()"),
        ),
        sa.Column("aggregate_type", sa.String(50), nullable=False),
        sa.Column("aggregate_id", sa.String(64), nullable=False),
        sa.Column("aggregate_version", sa.BigInteger(), nullable=False, server_default="0"),
        sa.Column("event_type", sa.String(100), nullable=False),
        sa.Column(
            "exchange", sa.String(100), nullable=False, server_default="chatbot.exchange"
        ),
        sa.Column("payload", sa.LargeBinary(), nullable=False),
        sa.Column("payload_type", sa.String(150), nullable=False),
        sa.Column(
            "content_type",
            sa.String(50),
            nullable=False,
            server_default="application/x-protobuf",
        ),
        sa.Column("correlation_id", sa.String(64), nullable=True),
        sa.Column("status", sa.String(20), nullable=False, server_default="PENDING"),
        sa.Column("attempt_count", sa.Integer(), nullable=False, server_default="0"),
        sa.Column("next_attempt_at", sa.DateTime(timezone=True), nullable=True),
        sa.Column("last_attempt_at", sa.DateTime(timezone=True), nullable=True),
        sa.Column("delivered_at", sa.DateTime(timezone=True), nullable=True),
        sa.Column("last_error", sa.String(500), nullable=True),
        sa.Column(
            "created_at",
            sa.DateTime(timezone=True),
            nullable=False,
            server_default=sa.text("now()"),
        ),
        sa.CheckConstraint(
            "status IN ('PENDING', 'RETRY_WAIT', 'DELIVERED', 'DEAD')",
            name="chk_outbox_events_status",
        ),
        sa.CheckConstraint("attempt_count >= 0", name="chk_outbox_events_attempt_count"),
        sa.CheckConstraint(
            "status <> 'RETRY_WAIT' OR next_attempt_at IS NOT NULL",
            name="chk_outbox_events_retry_time",
        ),
        sa.CheckConstraint(
            "status <> 'DELIVERED' OR delivered_at IS NOT NULL",
            name="chk_outbox_events_delivered_time",
        ),
        sa.CheckConstraint(
            "content_type IN ('application/x-protobuf', 'application/json')",
            name="chk_outbox_events_content_type",
        ),
    )
    # Relay poll (FOR UPDATE SKIP LOCKED) only scans undelivered rows.
    op.create_index(
        "idx_outbox_events_poll",
        "outbox_events",
        ["status", "next_attempt_at", "created_at"],
        postgresql_where=sa.text("status IN ('PENDING', 'RETRY_WAIT')"),
    )
    op.create_index(
        "idx_outbox_events_aggregate",
        "outbox_events",
        ["aggregate_type", "aggregate_id", "created_at"],
    )

    op.create_table(
        "inbox_events",
        sa.Column("event_id", postgresql.UUID(as_uuid=True), nullable=False),
        sa.Column("consumer", sa.String(100), nullable=False),
        sa.Column("event_type", sa.String(100), nullable=False),
        sa.Column(
            "processed_at",
            sa.DateTime(timezone=True),
            nullable=False,
            server_default=sa.text("now()"),
        ),
        sa.PrimaryKeyConstraint("event_id", "consumer", name="pk_inbox_events"),
    )
    op.create_index("idx_inbox_events_processed_at", "inbox_events", ["processed_at"])


def downgrade() -> None:
    op.drop_index("idx_inbox_events_processed_at", table_name="inbox_events")
    op.drop_table("inbox_events")
    op.drop_index("idx_outbox_events_aggregate", table_name="outbox_events")
    op.drop_index("idx_outbox_events_poll", table_name="outbox_events")
    op.drop_table("outbox_events")
