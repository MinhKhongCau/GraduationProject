"""seed dimensions for the full clinical test catalog and new template shells

A dimension belongs to many questions (one-to-many via
assess_questions.dimension_id); it has no direct relation to a template.
Which dimensions a given test "covers" is derived by looking at the
dimensions of its questions, not stored as a separate template<->dimension
link — so this migration only seeds the dimension catalog and the new
template shells, it does not link them together.

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


# Note: PHQ_9, GAD_7, MDQ, OCI_R, ASRS_V1_1, EAT_26 and PCL_5 are already
# seeded by 0002_seed_assessment_data — not touched by this migration.

# Templates this migration introduces. Only the template shell (code/title/
# description) is created — question banks for these are to be authored
# later via the admin Questions UI, since no question text was provided.
_NEW_TEMPLATES = [
    ("DASS_21", "DASS-21", "Thang đo trầm cảm, lo âu và căng thẳng (Depression Anxiety Stress Scales - 21 câu)"),
    ("SCOFF", "SCOFF", "Bảng câu hỏi sàng lọc rối loạn ăn uống (SCOFF)"),
    ("AUDIT", "AUDIT", "Bảng câu hỏi sàng lọc mức độ sử dụng rượu bia (Alcohol Use Disorders Identification Test)"),
    ("DAST_10", "DAST-10", "Bảng câu hỏi sàng lọc mức độ sử dụng chất gây nghiện (Drug Abuse Screening Test - 10 câu)"),
    ("ISI", "ISI (Insomnia Severity Index)", "Thang đo mức độ nghiêm trọng của chứng mất ngủ"),
    ("PSS_10", "PSS-10 (Perceived Stress Scale)", "Thang đo mức độ căng thẳng cảm nhận được"),
    ("WHO_5", "WHO-5", "Chỉ số hạnh phúc của Tổ chức Y tế Thế giới (WHO-5 Well-Being Index)"),
    ("YBOCS_SC", "Y-BOCS Symptom Checklist", "Danh mục triệu chứng ám ảnh cưỡng chế Yale-Brown"),
    ("PDSS", "PDSS (Panic Disorder Severity Scale)", "Thang đo mức độ nghiêm trọng của rối loạn hoảng sợ"),
    ("LSAS", "LSAS (Social Anxiety)", "Thang đo lo âu xã hội Liebowitz"),
    ("SPIN", "SPIN", "Bảng câu hỏi ám sợ xã hội (Social Phobia Inventory)"),
    ("C_SSRS", "C-SSRS", "Thang đánh giá mức độ nghiêm trọng tự sát Columbia"),
    ("ACE", "ACE (Adverse Childhood Experiences)", "Bảng câu hỏi về trải nghiệm bất lợi thời thơ ấu"),
    ("PSS_SR", "PSS-SR / PTSD khác", "Thang đo triệu chứng PTSD tự đánh giá (PTSD Symptom Scale - Self Report)"),
]

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

    # 1. Ensure every new template exists (skip ones already seeded by 0002).
    for code, title, description in _NEW_TEMPLATES:
        exists = connection.execute(
            sa.text("SELECT 1 FROM assess_templates WHERE code = :code"),
            {"code": code},
        ).first()
        if exists:
            continue
        connection.execute(
            sa.text(
                "INSERT INTO assess_templates "
                "(template_id, code, title, description, is_active, slug, created_at) "
                "VALUES (:template_id, :code, :title, :description, true, :slug, now())"
            ),
            {
                "template_id": uuid.uuid4(),
                "code": code,
                "title": title,
                "description": description,
                "slug": slugify(code),
            },
        )

    # 2. Ensure every dimension used by the catalog exists.
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

    new_codes = [code for code, _, _ in _NEW_TEMPLATES]
    template_placeholders = "'" + "','".join(new_codes) + "'"
    connection.execute(
        sa.text(f"DELETE FROM assess_templates WHERE code IN ({template_placeholders})")
    )

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
