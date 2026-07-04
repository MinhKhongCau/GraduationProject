"""create assessment schema

Revision ID: 0001_create_assessment_schema
Revises:
Create Date: 2026-07-04

"""
import sqlalchemy as sa
from alembic import op
from sqlalchemy.dialects.postgresql import JSONB, UUID
from sqlalchemy.sql import func

# revision identifiers, used by Alembic.
revision = "0001_create_assessment_schema"
down_revision = None
branch_labels = None
depends_on = None


def upgrade() -> None:
    op.create_table(
        "assess_templates",
        sa.Column(
            "template_id", UUID(as_uuid=True), primary_key=True
        ),
        sa.Column("code", sa.String, unique=True, index=True),
        sa.Column("title", sa.String),
        sa.Column("description", sa.Text),
        sa.Column(
            "is_active", sa.Boolean, server_default=sa.text("true")
        ),
        sa.Column(
            "created_at",
            sa.DateTime(timezone=True),
            server_default=func.now(),
        ),
    )

    op.create_table(
        "assess_option_groups",
        sa.Column("group_id", UUID(as_uuid=True), primary_key=True),
        sa.Column("group_code", sa.String, unique=True),
        sa.Column("group_name", sa.String),
        sa.Column("description", sa.Text),
    )

    op.create_table(
        "assess_options",
        sa.Column("option_id", UUID(as_uuid=True), primary_key=True),
        sa.Column(
            "group_id",
            UUID(as_uuid=True),
            sa.ForeignKey("assess_option_groups.group_id"),
        ),
        sa.Column("label", sa.String),
        sa.Column("value", sa.String),
        sa.Column("score_value", sa.Integer),
        sa.Column("order_index", sa.Integer),
    )

    op.create_table(
        "assess_questions",
        sa.Column("question_id", UUID(as_uuid=True), primary_key=True),
        sa.Column(
            "template_id",
            UUID(as_uuid=True),
            sa.ForeignKey("assess_templates.template_id"),
        ),
        sa.Column(
            "group_id",
            UUID(as_uuid=True),
            sa.ForeignKey("assess_option_groups.group_id"),
        ),
        sa.Column("content", sa.Text),
        sa.Column("dimension", sa.String),
        sa.Column("question_order", sa.Integer),
        sa.Column(
            "is_required", sa.Boolean, server_default=sa.text("true")
        ),
    )

    op.create_table(
        "assess_score_rules",
        sa.Column("rule_id", UUID(as_uuid=True), primary_key=True),
        sa.Column(
            "template_id",
            UUID(as_uuid=True),
            sa.ForeignKey("assess_templates.template_id"),
        ),
        sa.Column("dimension", sa.String, nullable=True),
        sa.Column("min_score", sa.Integer),
        sa.Column("max_score", sa.Integer),
        sa.Column("severity_level", sa.String),
        sa.Column("short_advice", sa.Text),
    )

    op.create_table(
        "assess_results",
        sa.Column("result_id", UUID(as_uuid=True), primary_key=True),
        sa.Column(
            "template_id",
            UUID(as_uuid=True),
            sa.ForeignKey("assess_templates.template_id"),
        ),
        sa.Column("user_id", UUID(as_uuid=True)),
        sa.Column("total_score", sa.Integer),
        sa.Column("dimension_scores", JSONB),
        sa.Column("ai_evaluation", sa.Text),
        sa.Column(
            "created_at",
            sa.DateTime(timezone=True),
            server_default=func.now(),
        ),
    )

    op.create_table(
        "assess_user_answers",
        sa.Column("answer_id", UUID(as_uuid=True), primary_key=True),
        sa.Column(
            "result_id",
            UUID(as_uuid=True),
            sa.ForeignKey("assess_results.result_id"),
        ),
        sa.Column(
            "question_id",
            UUID(as_uuid=True),
            sa.ForeignKey("assess_questions.question_id"),
        ),
        sa.Column(
            "option_id",
            UUID(as_uuid=True),
            sa.ForeignKey("assess_options.option_id"),
        ),
    )


def downgrade() -> None:
    op.drop_table("assess_user_answers")
    op.drop_table("assess_results")
    op.drop_table("assess_score_rules")
    op.drop_table("assess_questions")
    op.drop_table("assess_options")
    op.drop_table("assess_option_groups")
    op.drop_table("assess_templates")
