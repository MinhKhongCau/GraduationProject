"""assign dimension_id to each question of the already-seeded tests

Maps every question of PHQ-9, GAD-7, MDQ, OCI-R, ASRS v1.1, EAT-26 and
PCL-5 (seeded by 0002_seed_assessment_data) to a specific
assess_dimensions row, by (template code, question_order). All the
dimension codes referenced here were already seeded by
0006_seed_template_dimensions.

AUDIT, DAST-10, ISI, PSS-10, WHO-5, ACE and PSS-SR are left untouched —
they only exist as template shells with no question rows yet.

MDQ questions 16-17 (family history / prior diagnosis) aren't part of
any of the 15 mapped MDQ items and are intentionally left unmapped.

Revision ID: 0007_link_question_dimensions
Revises: 0006_seed_template_dimensions
Create Date: 2026-07-11

"""
import sqlalchemy as sa
from alembic import op

# revision identifiers, used by Alembic.
revision = "0007_link_question_dimensions"
down_revision = "0006_seed_template_dimensions"
branch_labels = None
depends_on = None


# {template_code: {question_order: dimension_code}}
_QUESTION_DIMENSIONS = {
    "PHQ_9": {
        1: "anhedonia",
        2: "depressed_mood",
        3: "sleep",
        4: "fatigue",
        5: "appetite",
        6: "self_worth",
        7: "concentration",
        8: "psychomotor",
        9: "suicidal_ideation",
    },
    "GAD_7": {
        1: "nervousness",
        2: "uncontrollable_worry",
        3: "excessive_worry",
        4: "relaxation",
        5: "restlessness",
        6: "irritability",
        7: "fear",
    },
    "MDQ": {
        1: "mania_hypomania",
        2: "mania_hypomania",
        3: "mania_hypomania",
        4: "mania_hypomania",
        5: "mania_hypomania",
        6: "mania_hypomania",
        7: "mania_hypomania",
        8: "mania_hypomania",
        9: "mania_hypomania",
        10: "mania_hypomania",
        11: "mania_hypomania",
        12: "mania_hypomania",
        13: "mania_hypomania",
        14: "mania_hypomania",  # "Các hiện tượng xảy ra cùng lúc?"
        15: "functional_impairment",  # "Mức độ ảnh hưởng"
        # 16 (family history) and 17 (prior diagnosis) left unmapped.
    },
    # Official OCI-R item-to-subscale assignment (Foa et al., 2002).
    "OCI_R": {
        1: "hoarding",
        2: "checking",
        3: "ordering",
        4: "neutralizing",
        5: "washing",
        6: "obsessing",
        7: "hoarding",
        8: "checking",
        9: "ordering",
        10: "neutralizing",
        11: "washing",
        12: "obsessing",
        13: "hoarding",
        14: "checking",
        15: "ordering",
        16: "neutralizing",
        17: "washing",
        18: "obsessing",
    },
    "ASRS_V1_1": {
        1: "inattention",
        2: "inattention",
        3: "inattention",
        4: "inattention",
        5: "hyperactivity",
        6: "hyperactivity",
        7: "inattention",
        8: "inattention",
        9: "inattention",
        10: "inattention",
        11: "inattention",
        12: "hyperactivity",
        13: "hyperactivity",
        14: "impulsivity",
        15: "impulsivity",
        16: "impulsivity",
        17: "impulsivity",
        18: "impulsivity",
    },
    # EAT-26 has no single official mapping onto these 8 fine-grained
    # dimensions (they cut across the standard 3-factor structure), so
    # each item below is matched to its dimension by content.
    "EAT_26": {
        1: "dieting",
        2: "oral_control",
        3: "food_preoccupation",
        4: "loss_of_control_eating",
        5: "oral_control",
        6: "dieting",
        7: "dieting",
        8: "oral_control",
        9: "self_induced_vomiting",
        10: "bulimia_food_preoccupation",
        11: "weight_loss",
        12: "weight_loss",
        13: "body_image",
        14: "body_image",
        15: "oral_control",
        16: "dieting",
        17: "dieting",
        18: "food_preoccupation",
        19: "oral_control",
        20: "oral_control",
        21: "food_preoccupation",
        22: "dieting",
        23: "dieting",
        24: "dieting",
        25: "self_induced_vomiting",
        26: "food_preoccupation",
    },
    "PCL_5": {
        1: "intrusion",
        2: "intrusion",
        3: "intrusion",
        4: "intrusion",
        5: "intrusion",
        6: "avoidance",
        7: "avoidance",
        8: "negative_cognition_mood",
        9: "negative_cognition_mood",
        10: "negative_cognition_mood",
        11: "negative_cognition_mood",
        12: "negative_cognition_mood",
        13: "negative_cognition_mood",
        14: "negative_cognition_mood",
        15: "arousal_reactivity",
        16: "arousal_reactivity",
        17: "arousal_reactivity",
        18: "arousal_reactivity",
        19: "arousal_reactivity",
        20: "arousal_reactivity",
    },
}


def _apply(connection, mapping) -> None:
    for template_code, order_to_dimension in mapping.items():
        for order, dimension_code in order_to_dimension.items():
            connection.execute(
                sa.text(
                    "UPDATE assess_questions AS q "
                    "SET dimension_id = d.dimension_id "
                    "FROM assess_templates AS t, assess_dimensions AS d "
                    "WHERE q.template_id = t.template_id "
                    "AND t.code = :template_code "
                    "AND q.question_order = :order "
                    "AND d.code = :dimension_code"
                ),
                {
                    "template_code": template_code,
                    "order": order,
                    "dimension_code": dimension_code,
                },
            )


def upgrade() -> None:
    connection = op.get_bind()
    _apply(connection, _QUESTION_DIMENSIONS)


def downgrade() -> None:
    connection = op.get_bind()
    for template_code, order_to_dimension in _QUESTION_DIMENSIONS.items():
        for order in order_to_dimension:
            connection.execute(
                sa.text(
                    "UPDATE assess_questions AS q "
                    "SET dimension_id = NULL "
                    "FROM assess_templates AS t "
                    "WHERE q.template_id = t.template_id "
                    "AND t.code = :template_code "
                    "AND q.question_order = :order"
                ),
                {"template_code": template_code, "order": order},
            )
