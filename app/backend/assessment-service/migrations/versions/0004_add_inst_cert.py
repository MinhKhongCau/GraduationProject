"""add instruction and certification fields

Revision ID: 0004_add_inst_cert
Revises: 0003_fill_slug_data
Create Date: 2026-07-06

"""
import sqlalchemy as sa
from alembic import op

# revision identifiers, used by Alembic.
revision = "0004_add_inst_cert"
down_revision = "0003_fill_slug_data"
branch_labels = None
depends_on = None

def upgrade() -> None:
    op.add_column("assess_templates", sa.Column("instruction", sa.Text(), nullable=True))
    op.add_column("assess_templates", sa.Column("certification", sa.Text(), nullable=True))

def downgrade() -> None:
    op.drop_column("assess_templates", "certification")
    op.drop_column("assess_templates", "instruction")
