# app/layers/layer5_filter.py
import re

# Các cụm từ mà một AI hỗ trợ tâm lý tuyệt đối không được dùng
DIAGNOSIS_KEYWORDS = [
    r"\byou have\b.*\b(depression|anxiety|bipolar|schizophrenia|ptsd)\b",
    r"\bi diagnose you\b",
    r"\bprescribe\b",
    r"\bdosage\b",
    r"\btake.*(mg|pill|medication)\b"
]

VIETNAMESE_UNSAFE_PATTERNS = [
    r"(?<!\w)(?:bạn\s+(?:đang\s+)?(?:bị|mắc)|tôi\s+chẩn\s+đoán\s+bạn\s+(?:bị|mắc)?)\s+(?:bệnh\s+)?(?:trầm cảm|lo âu|rối loạn lo âu|rối loạn lưỡng cực|lưỡng cực|tâm thần phân liệt|rối loạn stress sau sang chấn|ptsd)(?!\w)",
    r"(?<!\w)(?:tôi\s+(?:sẽ\s+)?kê\s+đơn(?:\s+cho\s+bạn)?|bạn\s+(?:hãy|nên|cần|phải)\s+(?:uống|dùng|sử dụng)\s+thuốc|(?:hãy|vui lòng)\s+(?:uống|dùng|sử dụng)\s+thuốc)(?!\s+theo\s+đúng\s+chỉ\s+định\s+của\s+bác\s+sĩ\b)(?!\w)",
    r"(?<!\w)(?:uống|dùng|sử dụng)\s+\d+(?:[.,]\d+)?\s*(?:mg|milligram|viên)(?!\w)",
    r"(?<!\w)(?:uống|dùng|sử dụng)\s+\d+(?:[.,]\d+)?\s*ml\s+(?:thuốc|siro|dung\s+dịch\s+thuốc)(?!\w)",
    r"(?<!\w)(?:hãy|vui lòng|bạn\s+(?:nên|hãy|cần|phải))\s+(?:ngừng|dừng|tăng|giảm|đổi|thay\s+đổi)\s+(?:(?:việc|liều)\s+)?(?:(?:uống|dùng|sử dụng)\s+)?(?:thuốc|liều(?:\s+dùng)?)(?!\w)",
]

# Contextual English phrases are used in Vietnamese mode to catch mixed output
# without flagging safe explanations such as "I cannot prescribe medication."
MIXED_LANGUAGE_UNSAFE_PATTERNS = [
    *DIAGNOSIS_KEYWORDS[:2],
    r"\bi\s+(?:will\s+|am going to\s+)?prescribe\b",
    r"\b(?:you should|you must|take)\s+(?:\w+\s+){0,5}(?:\d+(?:[.,]\d+)?\s*mg|(?:a|one|two)\s+(?:pill|tablet)s?)\b",
    r"\b(?:you should|you must)\s+(?:take|use)\s+(?:your\s+|the\s+)?(?:medication|medicine|pills?)\b",
    r"\b(?:you should|you must)\s+(?:stop|increase|decrease|change)\s+(?:taking\s+)?(?:your\s+)?(?:medication|medicine|prescription)\b",
]

ENGLISH_DISCLAIMER = (
    "I'm here to support you, but please note that I cannot provide medical diagnoses or prescribe treatments. "
    "I highly recommend consulting with a licensed healthcare professional for a proper evaluation."
)
VIETNAMESE_DISCLAIMER = (
    "Tôi có thể hỗ trợ bạn, nhưng không thể chẩn đoán bệnh hoặc kê đơn điều trị. "
    "Vui lòng trao đổi với bác sĩ hoặc chuyên gia y tế có chuyên môn để được tư vấn phù hợp."
)


def filter_output(ai_response: str, lang: str = "en") -> str:
    """
    Filter unsafe medical advice. Regex filtering is a limited safeguard and
    cannot detect every unsafe recommendation.
    """
    if not isinstance(lang, str) or lang.lower() not in {"en", "vi"}:
        raise ValueError("Unsupported output language; expected 'en' or 'vi'.")

    language = lang.lower()
    response_lower = ai_response.lower()

    if language == "en":
        patterns = DIAGNOSIS_KEYWORDS
        disclaimer = ENGLISH_DISCLAIMER
    else:
        patterns = VIETNAMESE_UNSAFE_PATTERNS + MIXED_LANGUAGE_UNSAFE_PATTERNS
        disclaimer = VIETNAMESE_DISCLAIMER

    for pattern in patterns:
        if re.search(pattern, response_lower):
            print("[Layer 5 - Filter] ⚠️ Phát hiện câu trả lời vi phạm an toàn y khoa! Đang thay thế...")
            return disclaimer

    return ai_response

# ==========================================
# TEST NHANH
# ==========================================
if __name__ == "__main__":
    test_ai_msg = "Based on what you said, I think you have clinical depression. You should take 20mg of Fluoxetine."
    filtered_msg = filter_output(test_ai_msg)
    print(f"🤖 Output cuối cùng:\n{filtered_msg}")