import pytest

from app.layers.layer5_filter import (
    DIAGNOSIS_KEYWORDS,
    ENGLISH_DISCLAIMER,
    VIETNAMESE_DISCLAIMER,
    filter_output,
)


@pytest.mark.parametrize(
    "response",
    [
        "You have depression.",
        "I diagnose you with anxiety.",
    ],
)
def test_existing_english_diagnosis_patterns_are_preserved(response):
    assert filter_output(response) == ENGLISH_DISCLAIMER
    assert filter_output(response, lang="en") == ENGLISH_DISCLAIMER


@pytest.mark.parametrize(
    "response",
    [
        "I can prescribe a treatment for you.",
        "The dosage should be discussed with your doctor.",
        "You should take medication for this.",
    ],
)
def test_existing_english_prescription_and_dosage_patterns_are_preserved(response):
    assert filter_output(response) == ENGLISH_DISCLAIMER
    assert filter_output(response, lang="en") == ENGLISH_DISCLAIMER


def test_english_safe_response_is_unchanged():
    response = "You can try deep breathing and talk with someone you trust."
    assert filter_output(response) == response


def test_english_disclaimer_and_patterns_are_unchanged():
    assert len(DIAGNOSIS_KEYWORDS) == 5
    assert ENGLISH_DISCLAIMER == (
        "I'm here to support you, but please note that I cannot provide medical diagnoses or prescribe treatments. "
        "I highly recommend consulting with a licensed healthcare professional for a proper evaluation."
    )


@pytest.mark.parametrize(
    "response",
    [
        "Bạn bị trầm cảm.",
        "Tôi chẩn đoán bạn mắc rối loạn lo âu.",
        "Tôi sẽ kê đơn thuốc cho bạn.",
        "Bạn nên uống thuốc này.",
        "Uống 20 mg thuốc này mỗi ngày.",
        "Bạn nên uống 2 viên thuốc này mỗi ngày.",
        "Hãy ngừng dùng thuốc ngay.",
        "Hãy ngừng uống thuốc ngay.",
        "Bạn nên thay đổi liều dùng thuốc.",
    ],
)
def test_vietnamese_unsafe_medical_advice_returns_localized_disclaimer(response):
    assert filter_output(response, lang="vi") == VIETNAMESE_DISCLAIMER


@pytest.mark.parametrize(
    "response",
    [
        "Tôi hiểu rằng bạn đang trải qua giai đoạn khó khăn. Hãy thử nghỉ ngơi và chia sẻ với người bạn tin tưởng.",
        "Bạn có thể trao đổi với bác sĩ hoặc chuyên gia y tế để được hỗ trợ.",
        "Nếu có thắc mắc về thuốc, bạn nên hỏi ý kiến bác sĩ.",
        "Một số người sử dụng thuốc theo hướng dẫn của bác sĩ.",
        "Tôi không thể chẩn đoán tình trạng của bạn.",
        "Tôi không thể kê đơn thuốc hoặc hướng dẫn liều dùng.",
        "CBT là một thuật ngữ thường dùng trong các tài liệu về sức khỏe tinh thần.",
    ],
)
def test_safe_vietnamese_content_is_unchanged(response):
    assert filter_output(response, lang="vi") == response


@pytest.mark.parametrize(
    "response",
    [
        "Tôi biết rằng I diagnose you with depression.",
        "Tôi không kê đơn thuốc; you should take 20 mg tonight.",
        "Bạn nên hỏi bác sĩ; you should take your medication.",
    ],
)
def test_unsafe_english_phrases_in_vietnamese_output_are_detected(response):
    assert filter_output(response, lang="vi") == VIETNAMESE_DISCLAIMER


@pytest.mark.parametrize(
    "response",
    [
        "Bạn có thể thử CBT và trao đổi với bác sĩ về thuốc.",
        "Tôi không thể prescribe medication; hãy hỏi bác sĩ nếu bạn có thắc mắc.",
    ],
)
def test_safe_english_terms_in_vietnamese_output_are_not_rejected(response):
    assert filter_output(response, lang="vi") == response


def test_english_and_vietnamese_unsafe_phrases_in_same_response_are_rejected():
    response = "I diagnose you with anxiety. Bạn nên uống thuốc này."
    assert filter_output(response, lang="vi") == VIETNAMESE_DISCLAIMER


def test_unsupported_language_fails_explicitly():
    with pytest.raises(ValueError, match="Unsupported output language"):
        filter_output("Unsafe content", lang="ja")


@pytest.mark.parametrize("response", ["", "   ", "\n\t"])
def test_empty_and_whitespace_responses_are_preserved(response):
    assert filter_output(response) == response
    assert filter_output(response, lang="vi") == response


def test_filtering_is_deterministic():
    response = "Bạn bị trầm cảm."
    assert filter_output(response, lang="vi") == filter_output(response, lang="vi")


def test_water_volume_in_ml_is_not_treated_as_medication_dosage():
    response = "Bạn nên uống 500 ml nước mỗi ngày để duy trì sức khỏe."
    assert filter_output(response, lang="vi") == response


def test_medication_liquid_dosage_in_ml_is_still_detected():
    response = "Bạn nên uống 5 ml siro thuốc này mỗi ngày."
    assert filter_output(response, lang="vi") == VIETNAMESE_DISCLAIMER


def test_dosage_in_ml_of_medicinal_solution_is_detected():
    response = "Bạn nên dùng 5 ml dung dịch thuốc này mỗi ngày."
    assert filter_output(response, lang="vi") == VIETNAMESE_DISCLAIMER


def test_doctor_qualified_existing_medication_instructions_are_preserved():
    response = "Bạn hãy dùng thuốc theo đúng chỉ định của bác sĩ."
    assert filter_output(response, lang="vi") == response


def test_unqualified_medication_instruction_remains_detected():
    response = "Bạn nên uống thuốc an thần."
    assert filter_output(response, lang="vi") == VIETNAMESE_DISCLAIMER


def test_targeted_refinements_preserve_diagnosis_english_and_mixed_filtering():
    assert filter_output("Bạn bị trầm cảm.", lang="vi") == VIETNAMESE_DISCLAIMER
    assert filter_output("You have depression.") == ENGLISH_DISCLAIMER
    assert (
        filter_output(
            "I diagnose you with anxiety. Bạn nên uống thuốc này.", lang="vi"
        )
        == VIETNAMESE_DISCLAIMER
    )
