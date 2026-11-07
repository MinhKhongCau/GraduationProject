# app/tests/test_language_detector.py
"""
Unit and integration tests for LanguageDetector module.

Designed to be executed cleanly with Python's built-in unittest:
    python3 -m unittest app/tests/test_language_detector.py
"""

from __future__ import annotations

import os
import unittest
from pathlib import Path
from unittest.mock import MagicMock, patch

from app.language.language_detector import (
    DEFAULT_MIN_CONFIDENCE,
    STATUS_DETECTED,
    STATUS_UNCERTAIN,
    STATUS_UNKNOWN,
    LanguageAlternative,
    LanguageDetectionResult,
    LanguageDetector,
    ModelNotFoundError,
    check_vietnamese_diacritics,
    get_language_detector,
    resolve_model_path,
)


class MockFastTextModel:
    """Mock FastText model mimicking lid.176.ftz predictions for deterministic unit testing."""

    def predict(self, text: str, k: int = 3):
        # Deterministic responses for testing
        mapping = {
            "Tôi đang cảm thấy rất buồn.": (
                ["__label__vi", "__label__en", "__label__th"],
                [0.9850, 0.0100, 0.0050],
            ),
            "I feel really sad today.": (
                ["__label__en", "__label__fr", "__label__de"],
                [0.9910, 0.0040, 0.0020],
            ),
            "Toi dang cam thay rat buon.": (
                ["__label__vi", "__label__en", "__label__id"],
                [0.8800, 0.0800, 0.0200],
            ),
            "今日はとても悲しいです。": (
                ["__label__ja", "__label__zh", "__label__ko"],
                [0.9950, 0.0030, 0.0010],
            ),
            "오늘 너무 슬퍼요.": (
                ["__label__ko", "__label__ja", "__label__en"],
                [0.9980, 0.0010, 0.0005],
            ),
            "我今天感到很难过。": (
                ["__label__zh", "__label__ja", "__label__en"],
                [0.9920, 0.0050, 0.0010],
            ),
            "Buồn": (
                ["__label__vi", "__label__en"],
                [0.5500, 0.3200],
            ),
            "OK": (
                ["__label__en", "__label__de"],
                [0.4200, 0.2800],
            ),
            "Tôi cảm thấy rất sad today.": (
                ["__label__vi", "__label__en"],
                [0.7200, 0.2500],
            ),
        }

        if text in mapping:
            labels, probs = mapping[text]
            return tuple(labels[:k]), tuple(probs[:k])

        # Default fallback prediction for other mocked inputs
        return ("__label__en", "__label__vi")[:k], (0.5000, 0.3000)[:k]


class TestLanguageDetectorMocked(unittest.TestCase):
    """Unit tests using a mocked FastText engine (100% offline, zero network/disk dependencies)."""

    def setUp(self):
        self.mock_model = MockFastTextModel()
        self.detector = LanguageDetector(model=self.mock_model, min_confidence=0.70, top_k=3)

    # -------------------------------------------------------------
    # Normal detection tests
    # -------------------------------------------------------------
    def test_normal_detection_vietnamese_with_diacritics(self):
        input_text = "Tôi đang cảm thấy rất buồn."
        res = self.detector.detect(input_text)

        self.assertEqual(res.status, STATUS_DETECTED)
        self.assertEqual(res.language, "vi")
        self.assertEqual(res.candidate_language, "vi")
        self.assertGreaterEqual(res.confidence, 0.70)
        self.assertTrue(res.is_reliable)
        self.assertTrue(res.has_diacritics)
        self.assertEqual(res.raw_text_length, len(input_text))
        self.assertEqual(res.alternatives[0].language, "vi")

    def test_normal_detection_english(self):
        input_text = "I feel really sad today."
        res = self.detector.detect(input_text)

        self.assertEqual(res.status, STATUS_DETECTED)
        self.assertEqual(res.language, "en")
        self.assertGreaterEqual(res.confidence, 0.70)
        self.assertTrue(res.is_reliable)
        self.assertFalse(res.has_diacritics)

    def test_normal_detection_vietnamese_without_diacritics(self):
        input_text = "Toi dang cam thay rat buon."
        res = self.detector.detect(input_text)

        self.assertEqual(res.status, STATUS_DETECTED)
        self.assertEqual(res.language, "vi")
        self.assertFalse(res.has_diacritics)
        self.assertGreaterEqual(res.confidence, 0.70)

    def test_normal_detection_japanese(self):
        input_text = "今日はとても悲しいです。"
        res = self.detector.detect(input_text)

        self.assertEqual(res.status, STATUS_DETECTED)
        self.assertEqual(res.language, "ja")
        self.assertGreaterEqual(res.confidence, 0.70)

    def test_normal_detection_korean(self):
        input_text = "오늘 너무 슬퍼요."
        res = self.detector.detect(input_text)

        self.assertEqual(res.status, STATUS_DETECTED)
        self.assertEqual(res.language, "ko")
        self.assertGreaterEqual(res.confidence, 0.70)

    def test_normal_detection_chinese(self):
        input_text = "我今天感到很难过。"
        res = self.detector.detect(input_text)

        self.assertEqual(res.status, STATUS_DETECTED)
        self.assertEqual(res.language, "zh")
        self.assertGreaterEqual(res.confidence, 0.70)

    # -------------------------------------------------------------
    # Short and invalid input tests
    # -------------------------------------------------------------
    def test_short_and_invalid_inputs_empty_string(self):
        res = self.detector.detect("")
        self.assertEqual(res.status, STATUS_UNKNOWN)
        self.assertIsNone(res.language)
        self.assertEqual(res.confidence, 0.0)
        self.assertEqual(len(res.alternatives), 0)
        self.assertFalse(res.is_reliable)

    def test_short_and_invalid_inputs_whitespace_only(self):
        res = self.detector.detect("   \n\t  ")
        self.assertEqual(res.status, STATUS_UNKNOWN)
        self.assertIsNone(res.language)
        self.assertEqual(res.confidence, 0.0)
        self.assertEqual(len(res.alternatives), 0)

    def test_short_and_invalid_inputs_emojis_only(self):
        res = self.detector.detect("😊😊😊 🔥❤️👍")
        self.assertEqual(res.status, STATUS_UNKNOWN)
        self.assertIsNone(res.language)
        self.assertEqual(res.confidence, 0.0)
        self.assertEqual(len(res.alternatives), 0)

    def test_short_and_invalid_inputs_numbers_only(self):
        res = self.detector.detect("12345 67890 2026/11/04")
        self.assertEqual(res.status, STATUS_UNKNOWN)
        self.assertIsNone(res.language)
        self.assertEqual(res.confidence, 0.0)

    def test_short_and_invalid_inputs_punctuation_and_symbols(self):
        res = self.detector.detect("??? !!! ... --- @#$%^&*()")
        self.assertEqual(res.status, STATUS_UNKNOWN)
        self.assertIsNone(res.language)
        self.assertEqual(res.confidence, 0.0)

    def test_none_input_safely_handled(self):
        # pyrefly: ignore [arg-type]
        res = self.detector.detect(None)
        self.assertEqual(res.status, STATUS_UNKNOWN)
        self.assertIsNone(res.language)
        self.assertEqual(res.confidence, 0.0)

    # -------------------------------------------------------------
    # Confidence and uncertain prediction handling
    # -------------------------------------------------------------
    def test_confidence_below_threshold_is_marked_uncertain(self):
        # "Buồn" returns confidence 0.55 < threshold 0.70
        res = self.detector.detect("Buồn")

        self.assertEqual(res.status, STATUS_UNCERTAIN)
        self.assertIsNone(res.language)  # By default, unreliable predictions are not returned as .language
        self.assertEqual(res.candidate_language, "vi")
        self.assertEqual(res.confidence, 0.55)
        self.assertFalse(res.is_reliable)
        self.assertTrue(res.has_diacritics)  # Diacritics detected, but does NOT override confidence!

    def test_confidence_below_threshold_with_allow_uncertain_language(self):
        # When caller explicitly requests candidate language on uncertain status
        res = self.detector.detect("Buồn", allow_uncertain_language=True)

        self.assertEqual(res.status, STATUS_UNCERTAIN)
        self.assertEqual(res.language, "vi")  # Populated because allow_uncertain_language=True
        self.assertEqual(res.candidate_language, "vi")
        self.assertEqual(res.confidence, 0.55)
        self.assertFalse(res.is_reliable)

    def test_short_ambiguous_ok_is_uncertain_not_forced_english_or_vietnamese(self):
        res = self.detector.detect("OK")

        self.assertEqual(res.status, STATUS_UNCERTAIN)
        self.assertIsNone(res.language)  # Not forced to English or Vietnamese
        self.assertEqual(res.confidence, 0.42)
        self.assertFalse(res.is_reliable)

    def test_configurable_confidence_threshold_runtime_override(self):
        # With standard threshold 0.70, confidence 0.55 is uncertain
        res_default = self.detector.detect("Buồn")
        self.assertEqual(res_default.status, STATUS_UNCERTAIN)

        # With lowered threshold 0.50, confidence 0.55 becomes detected
        res_lowered = self.detector.detect("Buồn", min_confidence=0.50)
        self.assertEqual(res_lowered.status, STATUS_DETECTED)
        self.assertEqual(res_lowered.language, "vi")

    def test_multiple_candidates_in_alternatives(self):
        res = self.detector.detect("Tôi đang cảm thấy rất buồn.", top_k=3)

        self.assertEqual(len(res.alternatives), 3)
        self.assertEqual(res.alternatives[0].language, "vi")
        self.assertEqual(res.alternatives[1].language, "en")
        self.assertEqual(res.alternatives[2].language, "th")

    def test_to_dict_serialization(self):
        res = self.detector.detect("I feel really sad today.")
        d = res.to_dict()

        self.assertIsInstance(d, dict)
        self.assertEqual(d["language"], "en")
        self.assertEqual(d["status"], STATUS_DETECTED)
        self.assertIn("alternatives", d)
        self.assertIn("confidence", d)
        self.assertIn("has_diacritics", d)

    # -------------------------------------------------------------
    # Input integrity & edge cases
    # -------------------------------------------------------------
    def test_input_text_is_never_modified(self):
        original = "  Tôi đang cảm thấy rất buồn.\n\n"
        copied = str(original)

        res = self.detector.detect(original)

        self.assertEqual(original, copied)  # Strict immutability check
        self.assertEqual(res.language, "vi")

    def test_unicode_nfc_handling(self):
        # Combining character sequence for 'ế'
        text_decomposed = "Tôi đang ca\u0309m tha\u0301y ra\u0302\u0301t buo\u0300n."
        # Should not raise exception and should normalize safely
        res = self.detector.detect(text_decomposed)
        self.assertIsNotNone(res)

    def test_mixed_language_input(self):
        res = self.detector.detect("Tôi cảm thấy rất sad today.")
        self.assertEqual(res.status, STATUS_DETECTED)
        self.assertEqual(res.language, "vi")
        self.assertEqual(res.alternatives[1].language, "en")

    def test_vietnamese_diacritics_never_forces_confidence_to_one(self):
        # Even with diacritics, confidence comes strictly from FastText (0.55 here), never hardcoded 1.0!
        res = self.detector.detect("Buồn")
        self.assertEqual(res.confidence, 0.55)
        self.assertNotEqual(res.confidence, 1.0)


class TestLanguageDetectorModelHandling(unittest.TestCase):
    """Tests for model loading lifecycle, error handling, and path resolution."""

    def test_missing_model_file_raises_clear_model_not_found_error(self):
        missing_path = Path("/tmp/non_existent_fasttext_model_12345.ftz")
        detector = LanguageDetector(model_path=missing_path)

        with self.assertRaises(ModelNotFoundError) as ctx:
            detector.get_model()

        err_msg = str(ctx.exception)
        self.assertIn("FastText language identification model not found", err_msg)
        self.assertIn("lid.176.ftz", err_msg)
        self.assertIn("curl", err_msg)

    def test_model_is_loaded_only_once_lazy_singleton(self):
        mock_raw_model = MagicMock()
        detector = LanguageDetector(model_path="/dummy/path")

        # Mock the internal loader
        with patch.object(detector, "_load_model", return_value=mock_raw_model) as mock_loader:
            # First access loads the model
            m1 = detector.get_model()
            # Second access reuses the cached model
            m2 = detector.get_model()

            self.assertIs(m1, mock_raw_model)
            self.assertIs(m2, mock_raw_model)
            mock_loader.assert_called_once()

    def test_model_prediction_exception_handled_safely(self):
        faulty_model = MagicMock()
        faulty_model.predict.side_effect = RuntimeError("FastText internal C++ memory error")

        detector = LanguageDetector(model=faulty_model)
        res = detector.detect("Hello world")

        # Should not raise exception; should degrade gracefully to unknown
        self.assertEqual(res.status, STATUS_UNKNOWN)
        self.assertIsNone(res.language)
        self.assertEqual(res.confidence, 0.0)

    def test_resolve_model_path_default_and_env(self):
        # Default
        default_p = resolve_model_path()
        self.assertTrue(default_p.is_absolute())
        self.assertTrue(str(default_p).endswith("models/language/lid.176.ftz"))

        # Explicit relative
        rel_p = resolve_model_path("custom/dir/lid.ftz")
        self.assertTrue(rel_p.is_absolute())
        self.assertTrue(str(rel_p).endswith("custom/dir/lid.ftz"))

        # Environment variable override
        with patch.dict(os.environ, {"FASTTEXT_LID_MODEL_PATH": "/opt/models/lid.176.ftz"}):
            env_p = resolve_model_path()
            self.assertEqual(env_p, Path("/opt/models/lid.176.ftz"))

    def test_get_language_detector_singleton(self):
        d1 = get_language_detector()
        d2 = get_language_detector()
        self.assertIs(d1, d2)


class TestLanguageDetectorRealModelIntegration(unittest.TestCase):
    """
    Integration tests against the real binary 'lid.176.ftz' if supplied locally.
    Automatically skipped if the model binary is not present on disk.
    """

    def setUp(self):
        self.model_path = resolve_model_path()
        if not self.model_path.is_file():
            self.skipTest(f"Pretrained FastText binary not found at '{self.model_path}'. Skipping real model test.")

        try:
            import fasttext  # noqa: F401
        except ImportError:
            self.skipTest("fasttext module not installed. Skipping real model integration test.")

        self.detector = LanguageDetector(model_path=self.model_path)

    def test_real_model_detection_vietnamese(self):
        res = self.detector.detect("Tôi cảm thấy lo lắng và bất an.")
        self.assertEqual(res.language, "vi")
        self.assertEqual(res.status, STATUS_DETECTED)
        self.assertGreater(res.confidence, 0.80)

    def test_real_model_detection_english(self):
        res = self.detector.detect("I am feeling overwhelmed with stress today.")
        self.assertEqual(res.language, "en")
        self.assertEqual(res.status, STATUS_DETECTED)
        self.assertGreater(res.confidence, 0.80)


if __name__ == "__main__":
    unittest.main()

