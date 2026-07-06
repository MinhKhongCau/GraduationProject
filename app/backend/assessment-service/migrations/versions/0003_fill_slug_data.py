"""fill slug data to assessment

Revision ID: 0003_fill_slug_data
Revises: 0002_seed_assessment_data
Create Date: 2026-07-06

"""
import hashlib
import sqlalchemy as sa
from alembic import op

# revision identifiers, used by Alembic.
revision = "0003_fill_slug_data"
down_revision = "0002_seed_assessment_data"
branch_labels = None
depends_on = None

def slugify(text: str) -> str:
    if not text:
        text = ""
    return hashlib.sha256(text.encode()).hexdigest()[:10]

def upgrade() -> None:
    connection = op.get_bind()

    # 1. Fill templates
    templates = connection.execute(sa.text("SELECT template_id, code FROM assess_templates")).fetchall()
    for row in templates:
        template_id = row[0]
        code = row[1]
        slug = slugify(code)
        connection.execute(
            sa.text("UPDATE assess_templates SET slug = :slug WHERE template_id = :template_id"),
            {"slug": slug, "template_id": template_id}
        )

    # 2. Fill option groups
    groups = connection.execute(sa.text("SELECT group_id, group_code FROM assess_option_groups")).fetchall()
    for row in groups:
        group_id = row[0]
        group_code = row[1]
        slug = slugify(group_code)
        connection.execute(
            sa.text("UPDATE assess_option_groups SET slug = :slug WHERE group_id = :group_id"),
            {"slug": slug, "group_id": group_id}
        )

    # 3. Fill options (unique constraint: group slug + option label slug)
    options = connection.execute(
        sa.text(
            "SELECT o.option_id, g.group_code, o.label "
            "FROM assess_options o "
            "JOIN assess_option_groups g ON o.group_id = g.group_id"
        )
    ).fetchall()
    for row in options:
        option_id = row[0]
        group_code = row[1]
        label = row[2]
        slug = slugify(f"{group_code}-{label}")
        connection.execute(
            sa.text("UPDATE assess_options SET slug = :slug WHERE option_id = :option_id"),
            {"slug": slug, "option_id": option_id}
        )

    # 4. Fill questions (unique constraint: template code + question order)
    questions = connection.execute(
        sa.text(
            "SELECT q.question_id, t.code, q.question_order "
            "FROM assess_questions q "
            "JOIN assess_templates t ON q.template_id = t.template_id"
        )
    ).fetchall()
    for row in questions:
        question_id = row[0]
        code = row[1]
        order = row[2]
        slug = slugify(f"{code}-q{order}")
        connection.execute(
            sa.text("UPDATE assess_questions SET slug = :slug WHERE question_id = :question_id"),
            {"slug": slug, "question_id": question_id}
        )


def downgrade() -> None:
    connection = op.get_bind()
    connection.execute(sa.text("UPDATE assess_templates SET slug = NULL"))
    connection.execute(sa.text("UPDATE assess_option_groups SET slug = NULL"))
    connection.execute(sa.text("UPDATE assess_options SET slug = NULL"))
    connection.execute(sa.text("UPDATE assess_questions SET slug = NULL"))
