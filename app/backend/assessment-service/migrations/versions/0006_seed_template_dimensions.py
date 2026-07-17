"""seed dimensions for the full clinical test catalog

A dimension belongs to many questions (one-to-many via
assess_questions.dimension_id); it has no direct relation to a template.
Which dimensions a given test "covers" is derived by looking at the
dimensions of its questions, not stored as a separate template<->dimension
link — so this migration only seeds the dimension catalog, it does not
create or link any templates.

Revision ID: 0006_seed_template_dimensions
Revises: 0005_add_dimensions
Create Date: 2026-07-12

"""
import hashlib
import uuid

import sqlalchemy as sa
from alembic import op

# revision identifiers, used by Alembic.
revision = "0006_seed_template_dimensions"
down_revision = "0005_add_dimensions"
branch_labels = None
depends_on = None


def slugify(text: str) -> str:
    if not text:
        text = ""
    return hashlib.sha256(text.encode()).hexdigest()[:10]


# Every dimension code used across the full test catalog, deduplicated —
# dimensions are global/shared, several templates reuse the same code
# (e.g. "avoidance", "dependence", "fear", "suicidal_ideation").
_ALL_DIMENSION_CODES = [
    "intrusion", "avoidance", "negative_cognition_mood", "arousal_reactivity",
    "anhedonia", "depressed_mood", "sleep", "fatigue", "appetite", "self_worth",
    "concentration", "psychomotor", "suicidal_ideation",
    "nervousness", "uncontrollable_worry", "excessive_worry", "relaxation",
    "restlessness", "irritability", "fear",
    "depression", "anxiety", "stress",
    "inattention", "hyperactivity", "impulsivity",
    "dieting", "bulimia_food_preoccupation", "oral_control",
    "body_image", "weight_loss", "loss_of_control_eating",
    "self_induced_vomiting", "food_preoccupation",
    "alcohol_consumption", "dependence", "alcohol_related_harm",
    "drug_use", "drug_related_problems",
    "sleep_onset", "sleep_maintenance", "early_awakening", "sleep_satisfaction",
    "daytime_impairment", "quality_of_life", "sleep_related_distress",
    "perceived_helplessness", "perceived_self_efficacy",
    "positive_mood", "vitality", "general_interest",
    "washing", "checking", "ordering", "obsessing", "hoarding", "neutralizing",
    "obsessions", "compulsions",
    "panic_attack", "anticipatory_anxiety", "functional_impairment",
    "physiological_discomfort",
    "mania_hypomania", "suicidal_behavior",
    "abuse", "neglect", "household_dysfunction",
    "re_experiencing", "hyperarousal",
]


def upgrade() -> None:
    connection = op.get_bind()

    # Ensure every dimension used by the catalog exists.
    for dim_code in _ALL_DIMENSION_CODES:
        exists = connection.execute(
            sa.text("SELECT 1 FROM assess_dimensions WHERE code = :code"),
            {"code": dim_code},
        ).first()
        if exists:
            continue
        connection.execute(
            sa.text(
                "INSERT INTO assess_dimensions "
                "(dimension_id, code, name, slug, is_active, created_at) "
                "VALUES (:dimension_id, :code, :name, :slug, true, now())"
            ),
            {
                "dimension_id": uuid.uuid4(),
                "code": dim_code,
                "name": dim_code.replace("_", " ").title(),
                "slug": slugify(dim_code),
            },
        )


def downgrade() -> None:
    connection = op.get_bind()

    # Only drop dimensions this migration created and that no question ended
    # up referencing in the meantime.
    dim_placeholders = "'" + "','".join(_ALL_DIMENSION_CODES) + "'"
    connection.execute(
        sa.text(
            f"DELETE FROM assess_dimensions "
            f"WHERE code IN ({dim_placeholders}) "
            f"AND dimension_id NOT IN ("
            f"  SELECT dimension_id FROM assess_questions WHERE dimension_id IS NOT NULL"
            f")"
        )
    )
