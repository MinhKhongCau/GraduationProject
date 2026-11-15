# app/layers/layer0_safety.py
import re
import unicodedata

# ---------------------------------------------------------------------------
# English crisis patterns (original – preserved for backward compatibility)
# Word-boundary \b works correctly for ASCII/Latin scripts.
# ---------------------------------------------------------------------------
CRISIS_KEYWORDS = [
    r"\bsuicide\b", r"\bkill myself\b", r"\bwant to die\b",
    r"\bend my life\b", r"\bhurt myself\b", r"\bno reason to live\b"
]

# ---------------------------------------------------------------------------
# Multilingual crisis patterns
#
# Design notes:
#   - CJK (ja, ko, zh): match against NFC-normalised ORIGINAL text (no .lower()).
#     \b word-boundary is meaningless for CJK; patterns rely on character
#     sequences and targeted negation handling where applicable.
#   - Vietnamese (vi): NFC-normalised text; re.IGNORECASE to handle
#     uppercase/mixed-case input (Vietnamese accented chars are case-folded by
#     Python's re engine when IGNORECASE is set for Basic Latin – for the small
#     number of ASCII letters in these patterns that is sufficient).
#     "không muốn sống nữa" (not wanting to live) IS a crisis signal.
#     "không muốn chết"     (not wanting to die)   is NOT.
#   - Per-occurrence negation: Vietnamese and Korean negation exclusions are
#     stored alongside each crisis pattern.  A matching exclusion suppresses
#     only a crisis match contained inside that exclusion's span; it cannot
#     suppress a later, independent occurrence of the same pattern.
#
# INTERNAL NOTE (not shown to users):
#   These patterns require qualified clinical and linguistic review before
#   deployment in a production mental-health context. They provide high
#   recall but NOT semantic understanding. False positives are expected for
#   educational or quoted references. Supported-language regexes do not
#   provide complete coverage. Unsupported languages receive English-only
#   pattern matching. Native multilingual patterns are not active in the
#   production chatbot flow until the orchestrator passes the detected
#   language code to check_input_risk().
# ---------------------------------------------------------------------------

# Each entry is a tuple: (crisis_pattern, negation_exclusion_pattern | None).
# If negation_exclusion_pattern is not None and matches the same span/vicinity,
# that specific crisis match is voided — but other patterns still run.
#
# Some CJK negations are encoded directly in a crisis pattern via negative
# lookahead; Korean suicide intent uses an occurrence-scoped exclusion.
_CRISIS_PATTERNS_BY_LANG: dict[str, list[tuple[str, str | None]]] = {
    "vi": [
        # (crisis_pattern, negation_exclusion_pattern | None)
        # A negation exclusion is applied only to the occurrence it contains.
        (r"muốn\s+tự\s+tử",             None),                      # muốn tự tử
        (r"muốn\s+chết",                 r"không\s+muốn\s+chết"),   # muốn chết — excluded if "không muốn chết" present
        (r"không\s+muốn\s+sống\s+nữa",  None),                      # không muốn sống nữa → IS crisis
        (r"kết\s+thúc\s+cuộc\s+đời",     None),                      # kết thúc cuộc đời
        (r"tự\s+làm\s+hại\s+bản\s+thân", None),                      # tự làm hại bản thân
        (r"tự\s+hủy\s+hoại\s+bản\s+thân", None),                     # tự hủy hoại bản thân
        (r"muốn\s+kết\s+thúc\s+tất\s+cả", None),                     # muốn kết thúc tất cả
    ],
    "ja": [
        # Negation handled inline via negative lookahead; negation_exclusion is None.
        # Note: the lookahead r"(?!くない|く無い)" only blocks immediately
        # following chars; the construction 「死にたいとは思わない」 still matches
        # 「死にたい」 as a substring.
        # This is a documented limitation: complex sentential negation in Japanese
        # is not reliably detected by simple regex. Erring on the side of safety.
        (r"死にた(?!くない|く無い)い",   None),   # 死にたい but NOT 死にたくない / 死にたく無い
        (r"自殺したい",                   None),   # 自殺したい
        (r"生きていたくない",              None),   # 生きていたくない
        (r"自分を傷つけたい",              None),   # 自分を傷つけたい
        (r"消えてしまいたい",              None),   # 消えてしまいたい
        (r"死を選びたい",                 None),   # 死を選びたい
    ],
    "ko": [
        # Negation handled inline via negative lookahead for 죽고싶지않아.
        # For 자살하고싶지않아, a separate negation_exclusion pattern is used so that
        # an independent crisis elsewhere in the same message is still caught.
        (r"죽고\s*싶(?!지\s*않)",         None),                        # 죽고 싶어 but NOT 죽고 싶지 않아
        (r"자살하고\s*싶",                r"자살하고\s*싶지\s*않"),       # 자살하고 싶어 — excluded if negated form present
        (r"살기\s*싫",                    None),                        # 살기 싫어
        (r"스스로를\s*해치고\s*싶",        None),                        # 스스로를 해치고 싶어
        (r"삶을\s*끝내고\s*싶",           None),                        # 삶을 끝내고 싶어
        (r"더\s*이상\s*살고\s*싶지\s*않", None),                        # 더 이상 살고 싶지 않아 (explicitly done)
    ],
    "zh": [
        # Covers both Simplified and Traditional Chinese where variants differ.
        (r"我想死(?!亡后)",   None),   # 我想死 (guard against 我想死亡后…)
        (r"我想自[杀殺]",     None),   # 我想自杀 / 我想自殺 (Simplified + Traditional)
        (r"我不想活了",       None),   # 我不想活了 → IS crisis ("done with living")
        (r"我想[伤傷]害自己", None),   # 我想伤害自己 / 我想傷害自己
        (r"不想活下去",       None),   # 不想活下去
        (r"[结結]束生命",     None),   # 结束生命 / 結束生命
        (r"[轻輕]生",         None),   # 轻生 / 輕生 (euphemism for suicide)
    ],
}

# Public read-only alias kept for backward compatibility with tests that import it.
CRISIS_KEYWORDS_BY_LANG: dict[str, list[str]] = {
    lang: [pattern for pattern, _ in entries]
    for lang, entries in _CRISIS_PATTERNS_BY_LANG.items()
}

# ---------------------------------------------------------------------------
# Emergency response messages
#
# IMPORTANT: English message is preserved verbatim (existing test checks for
# "psychiatric hospital hotline" substring). Do NOT alter it.
# NOTE: The English template contains a Vietnam-specific hotline (1900 1267)
# inherited from the original implementation. This is NOT an endorsement of
# that number's clinical suitability and does NOT imply language = country.
#
# Localized messages (vi/ja/ko/zh):
#   - Do NOT contain invented or dynamically translated hotline numbers.
#   - Do NOT imply that language identifies the user's country.
#   - Have NOT been clinically validated.
#   - Require qualified clinical and linguistic review before production use.
#   - Developer-facing disclaimers are kept in this source file (above) rather
#     than in the user-visible text strings below.
# ---------------------------------------------------------------------------
_EN_EMERGENCY = (
    "I am so sorry you are feeling this way, but please know you are not alone. "
    "If you are in immediate danger or experiencing a crisis, please reach out for help immediately. "
    "In Vietnam, you can contact the Psychiatric Hospital Hotline at 1900 1267, "
    "or go to the nearest emergency room. Your life is extremely valuable."
)

EMERGENCY_RESPONSES: dict[str, str] = {
    "en": _EN_EMERGENCY,
    "vi": (
        "Tôi rất lo lắng cho bạn và muốn bạn biết rằng bạn không đơn độc. "
        "Nếu bạn đang trong nguy hiểm ngay lập tức, hãy liên hệ với người thân tin cậy "
        "hoặc đến phòng cấp cứu gần nhất. Cuộc sống của bạn rất có giá trị."
    ),
    "ja": (
        "あなたのことがとても心配です。あなたは一人ではありません。"
        "今すぐ危険な状況にある場合は、信頼できる人に連絡するか、最寄りの救急室を受診してください。"
        "あなたの命はとても大切です。"
    ),
    "ko": (
        "당신이 매우 걱정됩니다. 당신은 혼자가 아닙니다. "
        "지금 즉각적인 위험에 처해 있다면, 신뢰할 수 있는 사람에게 연락하거나 "
        "가까운 응급실을 방문하세요. 당신의 삶은 소중합니다."
    ),
    "zh": (
        "我非常担心你，请知道你并不孤单。"
        "如果你现在处于紧急危险中，请联系你信任的人，或前往最近的急诊室。"
        "你的生命非常宝贵。"
    ),
    "zh-Hant": (
        "我非常擔心你，請知道你並不孤單。"
        "如果你現在處於緊急危險中，請聯絡你信任的人，或前往最近的急診室。"
        "你的生命非常寶貴。"
    ),
}


def _normalise(text: str) -> str:
    """Return NFC-normalised text for consistent pattern matching."""
    return unicodedata.normalize("NFC", text)


def _matches_english_crisis(message_lower: str) -> bool:
    """Check English crisis patterns against the lowercased message."""
    for pattern in CRISIS_KEYWORDS:
        if re.search(pattern, message_lower):
            return True
    return False


def _matches_native_crisis(user_message: str, lang: str) -> bool:
    """
    Check native-language crisis patterns against the NFC-normalised original text.

    Per-occurrence negation:
      - Each pattern entry carries an optional negation_exclusion regex.
      - An exclusion voids only a crisis match whose span it contains. Other
        occurrences of the same pattern, and other patterns, keep being checked.

    Returns True as soon as one un-voided pattern matches.
    """
    entries = _CRISIS_PATTERNS_BY_LANG.get(lang, [])
    if not entries:
        return False

    # Vietnamese uses IGNORECASE for mixed-case/uppercase input.
    # CJK scripts are not case-sensitive; IGNORECASE has no effect but is harmless.
    flags = re.UNICODE | re.IGNORECASE

    for crisis_pattern, negation_exclusion in entries:
        exclusion_spans = []
        if negation_exclusion is not None:
            exclusion_spans = [
                match.span()
                for match in re.finditer(negation_exclusion, user_message, flags)
            ]

        for crisis_match in re.finditer(crisis_pattern, user_message, flags):
            crisis_start, crisis_end = crisis_match.span()
            is_negated = any(
                exclusion_start <= crisis_start and crisis_end <= exclusion_end
                for exclusion_start, exclusion_end in exclusion_spans
            )
            if not is_negated:
                return True

    return False


def _normalise_lang(lang: str) -> tuple[str, str]:
    """Return ``(pattern_lang, response_lang)`` for a supported language tag.

    Short codes ``en``, ``vi``, ``ja``, ``ko``, and ``zh`` remain supported.
    Regional BCP-47 tags are case-insensitive and use their primary subtag for
    matching (for example, ``vi-VN`` uses Vietnamese patterns).  ``zh-TW``,
    ``zh-Hant``, ``zh-HK``, and ``zh-MO`` use Chinese patterns plus the
    Traditional Chinese response. Unsupported tags fall back to English.
    """
    canonical = lang.replace("_", "-").lower()
    primary = canonical.split("-", 1)[0]
    if primary not in {"en", "vi", "ja", "ko", "zh"}:
        return "en", "en"
    if primary == "zh" and any(
        tag in canonical.split("-") for tag in ("hant", "tw", "hk", "mo")
    ):
        return "zh", "zh-Hant"
    return primary, primary


def check_input_risk(user_message: str, lang: str = "en") -> tuple[bool, str]:
    """
    Scan the input message for crisis signals in the specified language.

    Args:
        user_message: The raw user input string.
        lang: A case-insensitive BCP-47 language tag. ``en``, ``vi``, ``ja``,
              ``ko``, and ``zh`` plus their regional variants are supported.
              ``zh-TW``, ``zh-Hant``, ``zh-HK``, and ``zh-MO`` receive a
              Traditional Chinese response. Defaults to ``en``; unsupported
              tags fall back to English-only patterns and response.

    Returns:
        (is_safe, emergency_response)
        - is_safe=True,  emergency_response=""                when no crisis detected.
        - is_safe=False, emergency_response=<localised string> when crisis detected.

    Limitations (must be understood before production use):
        - Existing callers that omit `lang` receive identical behaviour to before.
        - Layer 0 is intentionally lightweight: regex only, no ML model, no I/O.
        - Regex-only detection produces false positives for educational or quoted
          references (e.g. "this article discusses suicide prevention").
        - Supported-language regexes do not provide complete semantic or clinical
          understanding. Complex sentential negation may not be handled.
        - English regexes retain their existing high-recall false-positive
          limitation; English negation is intentionally unchanged in this task.
        - Unsupported languages receive English-only pattern matching only.
        - Native multilingual patterns are not active in the production chatbot
          flow until the orchestrator explicitly passes the detected language code
          to this function.
        - Emergency response templates have NOT been clinically validated.
          They require qualified clinical and linguistic review.
        - Language code does NOT identify the user's country.
    """
    # NFC-normalise once; reuse for all pattern checks.
    normalised = _normalise(user_message)
    message_lower = normalised.lower()
    pattern_lang, response_lang = _normalise_lang(lang)

    crisis_detected = False

    # Always run English patterns (also covers unknown/unsupported lang as fallback).
    if _matches_english_crisis(message_lower):
        crisis_detected = True

    # Run native-language patterns when supported (English is already covered above).
    if not crisis_detected and pattern_lang in _CRISIS_PATTERNS_BY_LANG:
        if _matches_native_crisis(normalised, pattern_lang):
            crisis_detected = True

    if crisis_detected:
        response = EMERGENCY_RESPONSES.get(response_lang, _EN_EMERGENCY)
        return False, response

    return True, ""


# ==========================================
# TEST NHANH
# ==========================================
if __name__ == "__main__":
    test_msg = "I am so tired of everything, I just want to die."
    is_safe, response = check_input_risk(test_msg)
    print(f"Message safe: {is_safe}")
    if not is_safe:
        print(f"🚨 Kích hoạt Crisis Protocol:\n{response}")
