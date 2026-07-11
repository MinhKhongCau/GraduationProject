"""add dimensions table and link questions to it

Revision ID: 0005_add_dimensions
Revises: 0004_add_inst_cert
Create Date: 2026-07-11

"""
import hashlib
import uuid

import sqlalchemy as sa
from alembic import op
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.sql import func

# revision identifiers, used by Alembic.
revision = "0005_add_dimensions"
down_revision = "0004_add_inst_cert"
branch_labels = None
depends_on = None


def slugify(text: str) -> str:
    if not text:
        text = ""
    return hashlib.sha256(text.encode()).hexdigest()[:10]


def upgrade() -> None:
    op.create_table(
        "assess_dimensions",
        sa.Column("dimension_id", UUID(as_uuid=True), primary_key=True),
        sa.Column("code", sa.String, unique=True, index=True),
        sa.Column("name", sa.String),
        sa.Column("description", sa.Text, nullable=True),
        sa.Column(
            "is_active", sa.Boolean, server_default=sa.text("true")
        ),
        sa.Column("slug", sa.String, unique=True, index=True, nullable=True),
        sa.Column(
            "created_at",
            sa.DateTime(timezone=True),
            server_default=func.now(),
        ),
    )

    op.add_column(
        "assess_questions",
        sa.Column(
            "dimension_id",
            UUID(as_uuid=True),
            sa.ForeignKey("assess_dimensions.dimension_id"),
            nullable=True,
        ),
    )

    connection = op.get_bind()

    # Backfill: turn every distinct free-text `dimension` value already used
    # by a question into a real AssessDimension row, then point the question
    # at it.
    distinct_codes = connection.execute(
        sa.text(
            "SELECT DISTINCT dimension FROM assess_questions "
            "WHERE dimension IS NOT NULL AND dimension != ''"
        )
    ).fetchall()

    for row in distinct_codes:
        code = row[0]
        dimension_id = uuid.uuid4()
        slug = slugify(code)
        connection.execute(
            sa.text(
                "INSERT INTO assess_dimensions "
                "(dimension_id, code, name, slug, is_active) "
                "VALUES (:dimension_id, :code, :name, :slug, true)"
            ),
            {
                "dimension_id": dimension_id,
                "code": code,
                "name": code.replace("_", " ").title(),
                "slug": slug,
            },
        )
        connection.execute(
            sa.text(
                "UPDATE assess_questions SET dimension_id = :dimension_id "
                "WHERE dimension = :code"
            ),
            {"dimension_id": dimension_id, "code": code},
        )

    op.drop_column("assess_questions", "dimension")


def downgrade() -> None:
    op.add_column(
        "assess_questions", sa.Column("dimension", sa.String, nullable=True)
    )

    connection = op.get_bind()
    connection.execute(
        sa.text(
            "UPDATE assess_questions q SET dimension = d.code "
            "FROM assess_dimensions d "
            "WHERE q.dimension_id = d.dimension_id"
        )
    )

    op.drop_column("assess_questions", "dimension_id")
    op.drop_table("assess_dimensions")
