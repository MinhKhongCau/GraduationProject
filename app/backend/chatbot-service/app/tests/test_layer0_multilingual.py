"""
test_layer0_multilingual.py

Focused unit tests for the multilingual extension of Layer 0 input safety.

Design:
  - unittest.TestCase only (no pytest dependency) → runs on host venv.
  - No ML model, no I/O, no network.
  - Covers: English regression, all 4 new languages, negation handling,
    localized response text, unknown-lang fallback, backward-compat call sig,
    Traditional Chinese, Vietnamese case/Unicode, Japanese complex negation,
    Korean negation scope, mixed positive+negative expressions.

Known limitations (documented, not bugs):
  - Regex-only detection produces false positives for educational or quoted
    references (e.g. "this article discusses suicide prevention").
  - Complex sentential negation in Japanese (e.g. 死にたいとは思わない) is not
    reliably excluded because the pattern matches the substring 死にたい.
    Layer 0 intentionally errs on the side of safety (high recall).
  - Supported-language patterns do not provide complete semantic coverage.
  - Unsupported languages receive English-only pattern matching.
  - Native patterns are inactive in production until orchestrator passes lang.
"""

import os
import re
import subprocess
import sys
import unicodedata
import unittest
from pathlib import Path


# ---------------------------------------------------------------------------
# Derive project root portably (works on macOS and Docker /workspace mounts)
# ---------------------------------------------------------------------------
# This file: .../chatbot-service/app/tests/test_layer0_multilingual.py
# Project root: .../chatbot-service/
_THIS_FILE = Path(__file__).resolve()
_CHATBOT_SERVICE_ROOT = _THIS_FILE.parent.parent.parent  # app/tests → app → chatbot-service


# ---------------------------------------------------------------------------
# 0. Lightweight import guard: Layer 0 must NOT import any ML framework.
#    Uses a fresh subprocess so other test-suite imports (torch, transformers)
#    already in sys.modules cannot pollute this check.
# ---------------------------------------------------------------------------
class TestNoMLImport(unittest.TestCase):
    """
    Layer 0 must remain lightweight — no torch / transformers / fasttext.

    Uses a subprocess so the global sys.modules from other test suites that
    load torch/transformers cannot produce a false failure here.
    """

    def test_no_heavy_ml_imports(self):
        code = (
            "import sys; "
            "import app.layers.layer0_safety; "
            "heavy = {'torch', 'transformers', 'fasttext', 'sentence_transformers'}; "
            "found = heavy & set(sys.modules.keys()); "
            "print('FOUND:' + ','.join(sorted(found)) if found else 'CLEAN')"
        )
        env = {**os.environ, "PYTHONPATH": "."}
        result = subprocess.run(
            [sys.executable, "-c", code],
            capture_output=True,
            text=True,
            cwd=str(_CHATBOT_SERVICE_ROOT),
            env=env,
        )
        output = result.stdout.strip()
        self.assertEqual(
            output,
            "CLEAN",
            f"Layer 0 imported a heavy ML module in isolation: {output}\n"
            f"stderr: {result.stderr[:500]}",
        )


# ---------------------------------------------------------------------------
# Helper import
# ---------------------------------------------------------------------------
from app.layers.layer0_safety import (  # noqa: E402
    CRISIS_KEYWORDS_BY_LANG,
    EMERGENCY_RESPONSES,
    check_input_risk,
)


# ---------------------------------------------------------------------------
# 1. English regression — existing behavior must be fully preserved
# ---------------------------------------------------------------------------
class TestEnglishRegression(unittest.TestCase):
    """Existing English patterns must continue to trigger exactly as before."""

    EN_TRIGGERS = [
        "I want to die",
        "I want to kill myself",
        "I am thinking about suicide",
        "I want to end my life",
        "I keep wanting to hurt myself",
        "I have no reason to live",
        "I am so tired of everything, I just want to die.",
    ]

    EN_SAFE = [
        "I am feeling fine today.",
        "What is the meaning of life?",
        # NOTE: "This article discusses suicide prevention strategies." triggers
        # r"\bsuicide\b" and IS flagged as a false positive by this regex-only layer.
        # This is a documented limitation; Layer 0 errs on the side of safety.
        "She said she had no reason to stay angry.",
    ]

    def test_english_triggers(self):
        for msg in self.EN_TRIGGERS:
            with self.subTest(msg=msg):
                is_safe, resp = check_input_risk(msg)
                self.assertFalse(is_safe, f"Expected crisis detection for: {msg!r}")
                self.assertNotEqual(resp, "")

    def test_english_safe_messages(self):
        for msg in self.EN_SAFE:
            with self.subTest(msg=msg):
                is_safe, resp = check_input_risk(msg)
                self.assertTrue(is_safe, f"Expected safe for: {msg!r}")
                self.assertEqual(resp, "")

    def test_english_emergency_contains_hotline_text(self):
        """Existing test_layers.py checks 'psychiatric hospital hotline' — must remain."""
        _, resp = check_input_risk("I want to die")
        self.assertIn("psychiatric hospital hotline", resp.lower())

    def test_explicit_en_lang_param(self):
        """Explicitly passing lang='en' must behave identically to no lang."""
        is_safe1, resp1 = check_input_risk("I want to die")
        is_safe2, resp2 = check_input_risk("I want to die", lang="en")
        self.assertEqual(is_safe1, is_safe2)
        self.assertEqual(resp1, resp2)


# ---------------------------------------------------------------------------
# 2. Vietnamese triggers
# ---------------------------------------------------------------------------
class TestVietnameseTriggers(unittest.TestCase):
    VI_TRIGGERS = [
        "Tôi muốn chết",
        "Tôi muốn tự tử",
        "Tôi không muốn sống nữa",
        "Tôi muốn kết thúc cuộc đời",
        "Tôi muốn tự làm hại bản thân",
        "Con muốn tự tử thật sự",
        "Mình muốn chết đi cho xong",
    ]

    def test_vi_triggers_detected(self):
        for msg in self.VI_TRIGGERS:
            with self.subTest(msg=msg):
                is_safe, resp = check_input_risk(msg, lang="vi")
                self.assertFalse(is_safe, f"Expected crisis for vi: {msg!r}")
                self.assertNotEqual(resp, "")

    def test_vi_response_is_vietnamese(self):
        _, resp = check_input_risk("Tôi muốn tự tử", lang="vi")
        self.assertIn("bạn", resp)

    def test_vi_response_no_user_visible_disclaimer(self):
        """Developer review note must NOT appear in the user-visible response."""
        _, resp = check_input_risk("Tôi muốn chết", lang="vi")
        self.assertNotIn("LƯU Ý", resp,
                         "User-visible Vietnamese response must not contain developer disclaimer.")

    def test_vi_uppercase_trigger(self):
        """Uppercase Vietnamese input must still be detected (IGNORECASE matching)."""
        # "MUỐN TỰ TỬ" — Vietnamese uppercase; re.IGNORECASE handles ASCII portions;
        # the accented chars remain the same code points in upper/lower for these words.
        is_safe, _ = check_input_risk("TÔI MUỐN TỰ TỬ", lang="vi")
        self.assertFalse(is_safe, "Uppercase Vietnamese crisis phrase must be detected.")

    def test_vi_mixed_case_trigger(self):
        """Mixed-case Vietnamese input must still be detected."""
        is_safe, _ = check_input_risk("Tôi Muốn Tự Tử", lang="vi")
        self.assertFalse(is_safe, "Mixed-case Vietnamese crisis phrase must be detected.")

    def test_vi_nfd_normalisation(self):
        """NFD-decomposed Unicode must match after NFC normalisation inside check_input_risk."""
        msg_nfd = unicodedata.normalize("NFD", "Tôi muốn tự tử")
        is_safe, _ = check_input_risk(msg_nfd, lang="vi")
        self.assertFalse(is_safe, "NFD-decomposed Vietnamese crisis phrase must be detected.")

    def test_vi_nfc_normalisation(self):
        """NFC-composed Unicode must match."""
        msg_nfc = unicodedata.normalize("NFC", "Tôi muốn tự tử")
        is_safe, _ = check_input_risk(msg_nfc, lang="vi")
        self.assertFalse(is_safe, "NFC Vietnamese crisis phrase must be detected.")


# ---------------------------------------------------------------------------
# 2b. Vietnamese negation scope — per-pattern, not global
# ---------------------------------------------------------------------------
class TestVietnameseNegationScope(unittest.TestCase):
    """
    Critical: negation exclusion must be scoped PER PATTERN.
    A negated form elsewhere must NOT cancel an independent crisis expression.
    """

    def test_negation_does_not_suppress_independent_crisis(self):
        """
        'Tôi không muốn chết, nhưng tôi muốn tự tử.'
        'không muốn chết' applies only to the 'muốn chết' pattern.
        'muốn tự tử' is a separate pattern with no negation exclusion → still a crisis.
        """
        msg = "Tôi không muốn chết, nhưng tôi muốn tự tử."
        is_safe, _ = check_input_risk(msg, lang="vi")
        self.assertFalse(
            is_safe,
            "Independent crisis expression 'muốn tự tử' must still be detected "
            "even when 'không muốn chết' appears elsewhere in the message.",
        )

    def test_negation_suppresses_only_its_own_pattern(self):
        """
        'Tôi không muốn chết' alone → safe (negation suppresses 'muốn chết').
        """
        is_safe, _ = check_input_risk("Tôi không muốn chết", lang="vi")
        self.assertTrue(is_safe, "Tôi không muốn chết should be safe (negation).")

    def test_same_pattern_negated_then_positive_is_crisis(self):
        """A later 'muốn chết' must not be hidden by the earlier negated one."""
        msg = "Tôi không muốn chết, nhưng sau đó tôi lại muốn chết."
        is_safe, _ = check_input_risk(msg, lang="vi")
        self.assertFalse(is_safe)

    def test_negated_death_then_suicide_is_crisis(self):
        msg = "Tôi không muốn chết, nhưng tôi vẫn muốn tự tử."
        is_safe, _ = check_input_risk(msg, lang="vi")
        self.assertFalse(is_safe)

    def test_mixed_negation_and_crisis_kết_thúc(self):
        """
        'Không muốn chết nhưng muốn kết thúc cuộc đời' → still a crisis.
        'kết thúc cuộc đời' has no negation exclusion.
        """
        msg = "Không muốn chết nhưng muốn kết thúc cuộc đời."
        is_safe, _ = check_input_risk(msg, lang="vi")
        self.assertFalse(is_safe,
                         "'muốn kết thúc cuộc đời' must trigger even if 'không muốn chết' present.")

    def test_negation_does_not_suppress_khong_muon_song_nua(self):
        """
        'không muốn chết' must NOT suppress 'không muốn sống nữa'
        (the latter has no negation exclusion).
        """
        msg = "Tôi không muốn chết nhưng tôi không muốn sống nữa."
        is_safe, _ = check_input_risk(msg, lang="vi")
        self.assertFalse(is_safe,
                         "'không muốn sống nữa' must trigger despite 'không muốn chết' also present.")


# ---------------------------------------------------------------------------
# 3. Vietnamese negation — these must NOT trigger crisis detection
# ---------------------------------------------------------------------------
class TestVietnameseNegation(unittest.TestCase):
    def test_vi_negation_khong_muon_chet(self):
        """Tôi không muốn chết (don't want to die) → safe."""
        is_safe, _ = check_input_risk("Tôi không muốn chết", lang="vi")
        self.assertTrue(is_safe, "Tôi không muốn chết should be safe (negation)")

    def test_vi_khong_muon_song_nua_is_crisis(self):
        """'không muốn sống nữa' means 'done with living' → IS a crisis signal."""
        is_safe, resp = check_input_risk("Tôi không muốn sống nữa", lang="vi")
        self.assertFalse(is_safe, "Tôi không muốn sống nữa must be detected as crisis")
        self.assertNotEqual(resp, "")

    def test_vi_variation_con_khong_muon_song_nua(self):
        is_safe, _ = check_input_risk("Con không muốn sống nữa", lang="vi")
        self.assertFalse(is_safe)


# ---------------------------------------------------------------------------
# 4. Japanese triggers
# ---------------------------------------------------------------------------
class TestJapaneseTriggers(unittest.TestCase):
    JA_TRIGGERS = [
        "死にたい",
        "自殺したい",
        "生きていたくない",
        "自分を傷つけたい",
        "消えてしまいたい",
        "もう死を選びたい",
        "私は死にたいと思っている",
    ]

    def test_ja_triggers_detected(self):
        for msg in self.JA_TRIGGERS:
            with self.subTest(msg=msg):
                is_safe, resp = check_input_risk(msg, lang="ja")
                self.assertFalse(is_safe, f"Expected crisis for ja: {msg!r}")
                self.assertNotEqual(resp, "")

    def test_ja_response_is_japanese(self):
        _, resp = check_input_risk("死にたい", lang="ja")
        self.assertIn("あなた", resp)

    def test_ja_response_no_user_visible_disclaimer(self):
        """Developer review note must NOT appear in the user-visible response."""
        _, resp = check_input_risk("死にたい", lang="ja")
        self.assertNotIn("注意", resp,
                         "User-visible Japanese response must not contain developer disclaimer.")

    def test_ja_response_no_disclaimer_bracket(self):
        """The 【注意:…】 bracket block must not appear in user-visible response."""
        _, resp = check_input_risk("死にたい", lang="ja")
        self.assertNotIn("【", resp)


# ---------------------------------------------------------------------------
# 4b. Japanese negation handling
# ---------------------------------------------------------------------------
class TestJapaneseNegation(unittest.TestCase):
    def test_runtime_japanese_pattern_matches_positive_and_excludes_negative_forms(self):
        """Exercise the actual runtime regex rather than only its source literal."""
        death_pattern = CRISIS_KEYWORDS_BY_LANG["ja"][0]
        self.assertEqual(death_pattern, r"死にた(?!くない|く無い)い")
        self.assertIsNotNone(re.search(death_pattern, "死にたい"))
        self.assertIsNone(re.search(death_pattern, "死にたくない"))
        self.assertIsNone(re.search(death_pattern, "死にたく無い"))

    def test_ja_negation_shinitagunai(self):
        """死にたくない (don't want to die) → safe."""
        is_safe, _ = check_input_risk("死にたくない", lang="ja")
        self.assertTrue(is_safe, "死にたくない should be safe (negation)")

    def test_ja_negation_alternate_form(self):
        """死にたく無い (alternate form) → safe."""
        is_safe, _ = check_input_risk("死にたく無い", lang="ja")
        self.assertTrue(is_safe, "死にたく無い should be safe (negation)")

    def test_ja_complex_negation_is_flagged_as_known_limitation(self):
        """
        死にたいとは思わない (I don't think I want to die) contains the substring
        「死にたい」which IS matched by the regex. Layer 0 errs on the side of
        safety for complex sentential negation. This is a documented limitation.
        This test documents the ACTUAL behaviour (crisis detected), not the
        ideal semantic behaviour.
        """
        is_safe, _ = check_input_risk("死にたいとは思わない", lang="ja")
        # Document: false-positive due to substring match — intentional high-recall design.
        # If this changes in a future version, update this test accordingly.
        self.assertFalse(
            is_safe,
            "死にたいとは思わない currently matches 死にたい substring. "
            "This is a documented false-positive / high-recall design decision.",
        )

    def test_ja_shinukoto_wo_kangaeteiru_is_known_coverage_gap(self):
        """死ぬことを考えている is currently not one of the regex triggers."""
        is_safe, response = check_input_risk("死ぬことを考えている", lang="ja")
        self.assertTrue(is_safe)
        self.assertEqual(response, "")


# ---------------------------------------------------------------------------
# 5. Korean triggers
# ---------------------------------------------------------------------------
class TestKoreanTriggers(unittest.TestCase):
    KO_TRIGGERS = [
        "죽고 싶어",
        "자살하고 싶어",
        "살기 싫어",
        "스스로를 해치고 싶어",
        "삶을 끝내고 싶어",
        "더 이상 살고 싶지 않아",
    ]

    def test_ko_triggers_detected(self):
        for msg in self.KO_TRIGGERS:
            with self.subTest(msg=msg):
                is_safe, resp = check_input_risk(msg, lang="ko")
                self.assertFalse(is_safe, f"Expected crisis for ko: {msg!r}")
                self.assertNotEqual(resp, "")

    def test_ko_response_is_korean(self):
        _, resp = check_input_risk("죽고 싶어", lang="ko")
        self.assertIn("당신", resp)

    def test_ko_response_no_user_visible_disclaimer(self):
        """Developer review note must NOT appear in the user-visible response."""
        _, resp = check_input_risk("죽고 싶어", lang="ko")
        self.assertNotIn("참고", resp,
                         "User-visible Korean response must not contain developer disclaimer.")


# ---------------------------------------------------------------------------
# 5b. Korean negation handling — per-pattern scope
# ---------------------------------------------------------------------------
class TestKoreanNegation(unittest.TestCase):
    def test_runtime_korean_patterns_use_optional_whitespace(self):
        """Guard actual regex literals against the erroneous literal ``\\s\\*``."""
        patterns = CRISIS_KEYWORDS_BY_LANG["ko"]
        self.assertIn(r"죽고\s*싶(?!지\s*않)", patterns)
        self.assertIn(r"자살하고\s*싶", patterns)
        self.assertTrue(all(r"\s\*" not in pattern for pattern in patterns))

    def test_ko_jukgo_sipeo_is_crisis(self):
        """죽고 싶어 (want to die) → crisis."""
        is_safe, _ = check_input_risk("죽고 싶어", lang="ko")
        self.assertFalse(is_safe, "죽고 싶어 must be detected as crisis")

    def test_ko_jukgo_sipji_anha_is_safe(self):
        """죽고 싶지 않아 (don't want to die) → safe (inline negative lookahead)."""
        is_safe, _ = check_input_risk("죽고 싶지 않아", lang="ko")
        self.assertTrue(is_safe, "죽고 싶지 않아 should be safe (negation)")

    def test_ko_jasalha_sipeo_is_crisis(self):
        """자살하고 싶어 (want to commit suicide) → crisis."""
        is_safe, _ = check_input_risk("자살하고 싶어", lang="ko")
        self.assertFalse(is_safe, "자살하고 싶어 must be detected as crisis")

    def test_ko_jasalha_sipji_anha_is_safe(self):
        """자살하고 싶지 않아 (don't want to commit suicide) → safe."""
        is_safe, _ = check_input_risk("자살하고 싶지 않아", lang="ko")
        self.assertTrue(is_safe, "자살하고 싶지 않아 should be safe (negation)")

    def test_ko_negation_does_not_suppress_independent_crisis(self):
        """
        '자살하고 싶지 않아, 그냥 죽고 싶어.'
        Negation on 자살하고 must not cancel the independent 죽고 싶어 expression.
        """
        msg = "자살하고 싶지 않아, 그냥 죽고 싶어."
        is_safe, _ = check_input_risk(msg, lang="ko")
        self.assertFalse(
            is_safe,
            "Independent crisis '죽고 싶어' must trigger even when "
            "'자살하고 싶지 않아' appears elsewhere in the same message.",
        )

    def test_same_pattern_negated_then_positive_is_crisis(self):
        """The second 자살하고 싶어 occurrence remains positive and must trigger."""
        msg = "자살하고 싶지 않아. 하지만 자살하고 싶어."
        is_safe, _ = check_input_risk(msg, lang="ko")
        self.assertFalse(is_safe)

    def test_negated_suicide_then_positive_death_is_crisis(self):
        msg = "자살하고 싶지 않아. 그래도 죽고 싶어."
        is_safe, _ = check_input_risk(msg, lang="ko")
        self.assertFalse(is_safe)

    def test_runtime_patterns_match_spaced_positive_and_reject_spaced_negation(self):
        """Exercise the exact runtime literals, not comments describing them."""
        death_pattern = CRISIS_KEYWORDS_BY_LANG["ko"][0]
        suicide_pattern = CRISIS_KEYWORDS_BY_LANG["ko"][1]
        self.assertIsNotNone(re.search(death_pattern, "죽고 싶어"))
        self.assertIsNone(re.search(death_pattern, "죽고 싶지 않아"))
        self.assertIsNotNone(re.search(suicide_pattern, "자살하고 싶어"))
        self.assertIsNotNone(re.search(suicide_pattern, "자살하고 싶지 않아"))


# ---------------------------------------------------------------------------
# 6. Chinese triggers — Simplified and Traditional
# ---------------------------------------------------------------------------
class TestChineseTriggers(unittest.TestCase):
    # Simplified Chinese
    ZH_SIMPLIFIED_TRIGGERS = [
        "我想死",
        "我想自杀",
        "我不想活了",
        "我想伤害自己",
        "不想活下去",
        "我想结束生命",
        "有轻生的念头",
    ]

    # Traditional Chinese variants
    ZH_TRADITIONAL_TRIGGERS = [
        "我想自殺",         # Traditional: 殺 instead of 杀
        "我想傷害自己",      # Traditional: 傷 instead of 伤
        "我想結束生命",      # Traditional: 結 instead of 结
        "有輕生的念頭",      # Traditional: 輕 instead of 轻, 頭 instead of 头
    ]

    def test_zh_simplified_triggers_detected(self):
        for msg in self.ZH_SIMPLIFIED_TRIGGERS:
            with self.subTest(msg=msg):
                is_safe, resp = check_input_risk(msg, lang="zh")
                self.assertFalse(is_safe, f"Expected crisis for zh (Simplified): {msg!r}")
                self.assertNotEqual(resp, "")

    def test_zh_traditional_triggers_detected(self):
        for msg in self.ZH_TRADITIONAL_TRIGGERS:
            with self.subTest(msg=msg):
                is_safe, resp = check_input_risk(msg, lang="zh")
                self.assertFalse(is_safe, f"Expected crisis for zh (Traditional): {msg!r}")
                self.assertNotEqual(resp, "")

    def test_zh_response_contains_expected_text(self):
        _, resp = check_input_risk("我想死", lang="zh")
        self.assertIn("孤单", resp)

    def test_zh_response_no_user_visible_disclaimer(self):
        """Developer review note must NOT appear in the user-visible response."""
        _, resp = check_input_risk("我想死", lang="zh")
        self.assertNotIn("注意", resp,
                         "User-visible Chinese response must not contain developer disclaimer.")

    def test_zh_negation_wo_bu_xiang_si(self):
        """我不想死 (don't want to die) → safe."""
        is_safe, _ = check_input_risk("我不想死", lang="zh")
        self.assertTrue(is_safe, "我不想死 should be safe (negation)")

    def test_zh_bu_xiang_huo_le_is_crisis(self):
        """我不想活了 (done with living) → IS a crisis signal."""
        is_safe, _ = check_input_risk("我不想活了", lang="zh")
        self.assertFalse(is_safe, "我不想活了 must be detected as crisis")

    def test_zh_traditional_qingsheng_detected(self):
        """輕生 (Traditional: euphemism for suicide) → crisis."""
        is_safe, _ = check_input_risk("有輕生的念頭", lang="zh")
        self.assertFalse(is_safe, "輕生 (Traditional) must be detected as crisis")


# ---------------------------------------------------------------------------
# 7. Emergency response content requirements
# ---------------------------------------------------------------------------
class TestEmergencyResponseContent(unittest.TestCase):
    def test_emergency_responses_dict_has_all_languages(self):
        for lang in ("en", "vi", "ja", "ko", "zh"):
            self.assertIn(lang, EMERGENCY_RESPONSES, f"Missing EMERGENCY_RESPONSES['{lang}']")

    def test_en_response_contains_hotline_text(self):
        """English hotline text preserved for backward compat with test_layers.py."""
        self.assertIn("psychiatric hospital hotline", EMERGENCY_RESPONSES["en"].lower())

    def test_en_response_is_english(self):
        _, resp = check_input_risk("I want to die", lang="en")
        self.assertIn("not alone", resp)

    # --- Per-language: no invented hotlines (no 4+ digit sequences) ---
    def _assert_no_phone_number(self, lang: str) -> None:
        import re as _re
        resp = EMERGENCY_RESPONSES[lang]
        match = _re.search(r"\d{4,}", resp)
        self.assertIsNone(
            match,
            f"EMERGENCY_RESPONSES['{lang}'] must not contain invented hotline numbers "
            f"(found: {match.group() if match else ''}).",
        )

    def test_no_invented_hotline_in_vi_response(self):
        self._assert_no_phone_number("vi")

    def test_no_invented_hotline_in_ja_response(self):
        self._assert_no_phone_number("ja")

    def test_no_invented_hotline_in_ko_response(self):
        self._assert_no_phone_number("ko")

    def test_no_invented_hotline_in_zh_response(self):
        self._assert_no_phone_number("zh")

    # --- Localized responses are distinct from English ---
    def test_vi_response_different_from_en(self):
        _, resp_en = check_input_risk("I want to die", lang="en")
        _, resp_vi = check_input_risk("Tôi muốn tự tử", lang="vi")
        self.assertNotEqual(resp_en, resp_vi)

    def test_ja_response_different_from_en(self):
        _, resp_en = check_input_risk("I want to die", lang="en")
        _, resp_ja = check_input_risk("死にたい", lang="ja")
        self.assertNotEqual(resp_en, resp_ja)

    def test_ko_response_different_from_en(self):
        _, resp_en = check_input_risk("I want to die", lang="en")
        _, resp_ko = check_input_risk("죽고 싶어", lang="ko")
        self.assertNotEqual(resp_en, resp_ko)

    def test_zh_response_different_from_en(self):
        _, resp_en = check_input_risk("I want to die", lang="en")
        _, resp_zh = check_input_risk("我想死", lang="zh")
        self.assertNotEqual(resp_en, resp_zh)


# ---------------------------------------------------------------------------
# 8. Unknown / unsupported language falls back to English patterns + English response
# ---------------------------------------------------------------------------
class TestUnknownLanguageFallback(unittest.TestCase):
    def test_unknown_lang_safe_message(self):
        """Non-crisis message with unsupported lang is safe."""
        is_safe, resp = check_input_risk("Bonjour, comment allez-vous?", lang="fr")
        self.assertTrue(is_safe)
        self.assertEqual(resp, "")

    def test_unknown_lang_english_crisis_still_detected(self):
        """English crisis phrase is caught even when lang='fr' (unknown)."""
        is_safe, resp = check_input_risk("I want to die", lang="fr")
        self.assertFalse(is_safe)
        self.assertIn("not alone", resp)

    def test_unknown_lang_response_is_english(self):
        _, resp = check_input_risk("I want to kill myself", lang="de")
        self.assertEqual(resp, EMERGENCY_RESPONSES["en"])

    def test_empty_lang_string_treated_as_unknown(self):
        """Empty lang string → English fallback only (no native patterns)."""
        is_safe, _ = check_input_risk("I am feeling great today.", lang="")
        self.assertTrue(is_safe)


# ---------------------------------------------------------------------------
# 8b. BCP-47 language-tag contract
# ---------------------------------------------------------------------------
class TestLanguageCodeContract(unittest.TestCase):
    def test_regional_tags_use_primary_language_patterns(self):
        for lang, message in (
            ("vi-VN", "Tôi muốn tự tử"),
            ("ja-JP", "死にたい"),
            ("ko-KR", "죽고 싶어"),
            ("zh-CN", "我想死"),
        ):
            with self.subTest(lang=lang):
                is_safe, _ = check_input_risk(message, lang=lang)
                self.assertFalse(is_safe)

    def test_traditional_chinese_tags_receive_traditional_response(self):
        for lang in ("zh-TW", "zh-Hant"):
            with self.subTest(lang=lang):
                is_safe, response = check_input_risk("我想自殺", lang=lang)
                self.assertFalse(is_safe)
                self.assertIn("擔心", response)
                self.assertNotIn("担心", response)


# ---------------------------------------------------------------------------
# 9. Backward compatibility — no lang param call signature
# ---------------------------------------------------------------------------
class TestBackwardCompatibility(unittest.TestCase):
    def test_no_lang_param_safe(self):
        is_safe, resp = check_input_risk("How are you?")
        self.assertTrue(is_safe)
        self.assertEqual(resp, "")

    def test_no_lang_param_crisis(self):
        is_safe, resp = check_input_risk("I want to die")
        self.assertFalse(is_safe)
        self.assertIn("psychiatric hospital hotline", resp.lower())

    def test_signature_returns_tuple(self):
        result = check_input_risk("test message")
        self.assertIsInstance(result, tuple)
        self.assertEqual(len(result), 2)
        self.assertIsInstance(result[0], bool)
        self.assertIsInstance(result[1], str)

    def test_safe_always_returns_empty_string(self):
        is_safe, resp = check_input_risk("The weather is lovely today.")
        self.assertTrue(is_safe)
        self.assertEqual(resp, "")


# ---------------------------------------------------------------------------
# 10. Cross-language: English crisis phrase detected regardless of declared lang
# ---------------------------------------------------------------------------
class TestCrossLanguageEnglishDetection(unittest.TestCase):
    """
    Even when lang='ja'/'ko'/'vi'/'zh', an English crisis phrase in the message
    must still be caught (English patterns always run as a base layer).
    """

    def test_english_phrase_with_ja_lang(self):
        is_safe, _ = check_input_risk("I want to die", lang="ja")
        self.assertFalse(is_safe)

    def test_english_phrase_with_vi_lang(self):
        is_safe, _ = check_input_risk("I want to kill myself", lang="vi")
        self.assertFalse(is_safe)

    def test_english_phrase_with_ko_lang(self):
        is_safe, _ = check_input_risk("I want to end my life", lang="ko")
        self.assertFalse(is_safe)

    def test_english_phrase_with_zh_lang(self):
        is_safe, _ = check_input_risk("no reason to live", lang="zh")
        self.assertFalse(is_safe)


# ---------------------------------------------------------------------------
# 11. Edge cases
# ---------------------------------------------------------------------------
class TestEdgeCases(unittest.TestCase):
    def test_empty_string_is_safe(self):
        is_safe, resp = check_input_risk("")
        self.assertTrue(is_safe)
        self.assertEqual(resp, "")

    def test_whitespace_only_is_safe(self):
        is_safe, resp = check_input_risk("   \n\t  ")
        self.assertTrue(is_safe)

    def test_mixed_language_message_vi_crisis(self):
        """Mixed English/Vietnamese message with crisis in Vietnamese part."""
        is_safe, _ = check_input_risk("Hello everyone, Tôi muốn tự tử", lang="vi")
        self.assertFalse(is_safe)

    def test_mixed_language_message_ja_crisis(self):
        """Mixed English/Japanese message with crisis in Japanese part."""
        is_safe, _ = check_input_risk("I am fine but 死にたい", lang="ja")
        self.assertFalse(is_safe)

    def test_very_long_message_does_not_error(self):
        """Layer 0 must handle long messages without raising exceptions."""
        long_msg = "I am feeling okay. " * 1000
        try:
            is_safe, resp = check_input_risk(long_msg, lang="en")
            self.assertTrue(is_safe)
        except Exception as e:
            self.fail(f"check_input_risk raised {type(e).__name__} on long message: {e}")

    def test_unicode_nfc_vi(self):
        """NFC-composed Vietnamese text must match."""
        msg_nfc = unicodedata.normalize("NFC", "Tôi muốn tự tử")
        is_safe, _ = check_input_risk(msg_nfc, lang="vi")
        self.assertFalse(is_safe)

    def test_unicode_nfd_vi(self):
        """NFD-decomposed Vietnamese text must match after internal NFC normalisation."""
        msg_nfd = unicodedata.normalize("NFD", "Tôi muốn tự tử")
        # Confirm input is genuinely NFD (different from NFC)
        self.assertNotEqual(msg_nfd, unicodedata.normalize("NFC", msg_nfd),
                            "Test setup: NFD form should differ from NFC for accented text.")
        is_safe, _ = check_input_risk(msg_nfd, lang="vi")
        self.assertFalse(is_safe, "NFD-decomposed text must be detected after NFC normalisation.")


if __name__ == "__main__":
    unittest.main(verbosity=2)
