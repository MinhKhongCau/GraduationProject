# app/tests/test_translation_unit.py
"""Unit tests for M2M100 translation service using mocks.

Covers:
- Vietnamese -> English translation
- English -> Vietnamese translation
- Japanese, Korean, and Chinese language code validation
- Same-language passthrough (no model load)
- Empty and whitespace input handling
- Unsupported language validation (raises TranslationError)
- Lazy loading (model loaded on demand)
- Single model instance (loaded only once)
- Missing model validation
- Model initialization failure handling
- Inference failure handling
- Original input preservation
- No raw content in logs

Designed to run cleanly with both:
    python3 -m unittest app/tests/test_translation_unit.py
    pytest app/tests/test_translation_unit.py
"""

from __future__ import annotations

import logging
import unittest
from unittest.mock import MagicMock, patch

import app.translation.translation_service as svc
from app.translation.translation_service import TranslationError, translate


def create_mock_tokenizer_and_model():
    mock_tok = MagicMock()
    mock_tok.src_lang = None
    mock_tok.get_lang_id = MagicMock(return_value=128001)

    def fake_call(text, return_tensors="pt", padding=True, **kwargs):
        return {"input_ids": MagicMock(), "attention_mask": MagicMock()}

    mock_tok.side_effect = fake_call
    mock_tok.encode = MagicMock(return_value=[1, 2, 3, 4, 5])
    mock_tok.batch_decode = MagicMock(return_value=["Mocked translation"])

    mock_model = MagicMock()
    mock_model.config.max_position_embeddings = 1024
    mock_model.eval = MagicMock(return_value=mock_model)
    mock_model.generate = MagicMock(return_value=MagicMock())

    return mock_tok, mock_model


class TestTranslationUnit(unittest.TestCase):

    def setUp(self):
        """Reset service singletons before each test."""
        svc._model = None
        svc._tokenizer = None

    def tearDown(self):
        """Reset service singletons after each test."""
        svc._model = None
        svc._tokenizer = None

    def test_same_language_passthrough_en(self):
        """Same source and target returns original text without model loading."""
        text = "Hello, how are you feeling today?"
        result = translate(text, "en", "en")
        self.assertEqual(result, text)
        self.assertIsNone(svc._model)
        self.assertIsNone(svc._tokenizer)

    def test_same_language_passthrough_vi(self):
        """Same source and target (vi -> vi) returns original text unchanged."""
        text = "Tôi cảm thấy hơi lo âu."
        result = translate(text, "vi", "vi")
        self.assertEqual(result, text)
        self.assertIsNone(svc._model)

    def test_empty_and_whitespace_input(self):
        """Empty or whitespace-only inputs are handled safely without loading model."""
        self.assertEqual(translate("", "en", "vi"), "")
        self.assertEqual(translate("   ", "vi", "en"), "   ")
        self.assertIsNone(svc._model)

    def test_unsupported_source_language(self):
        """Unsupported source language raises TranslationError."""
        with self.assertRaises(TranslationError) as ctx:
            translate("Bonjour", "fr", "en")
        self.assertIn("unsupported source language", str(ctx.exception).lower())
        self.assertIsNone(svc._model)

    def test_unsupported_target_language(self):
        """Unsupported target language raises TranslationError."""
        with self.assertRaises(TranslationError) as ctx:
            translate("Hello", "en", "de")
        self.assertIn("unsupported target language", str(ctx.exception).lower())
        self.assertIsNone(svc._model)

    def test_supported_language_codes_accepted(self):
        """Supported languages vi, en, ja, ko, zh are accepted."""
        mock_tok, mock_model = create_mock_tokenizer_and_model()

        mock_tok_cls = MagicMock()
        mock_tok_cls.from_pretrained.return_value = mock_tok
        mock_model_cls = MagicMock()
        mock_model_cls.from_pretrained.return_value = mock_model

        with patch.object(svc, "_model_is_ready", return_value=True), \
             patch.object(svc, "M2M100Tokenizer", mock_tok_cls), \
             patch.object(svc, "M2M100ForConditionalGeneration", mock_model_cls):

            for src in ["vi", "en", "ja", "ko", "zh"]:
                for tgt in ["vi", "en", "ja", "ko", "zh"]:
                    if src == tgt:
                        continue
                    res = translate("Sample text", src, tgt)
                    self.assertEqual(res, "Mocked translation")

    def test_vietnamese_to_english_translation(self):
        """Test vi -> en translation flow."""
        mock_tok, mock_model = create_mock_tokenizer_and_model()
        mock_tok.batch_decode.return_value = ["I feel much better today."]

        mock_tok_cls = MagicMock()
        mock_tok_cls.from_pretrained.return_value = mock_tok
        mock_model_cls = MagicMock()
        mock_model_cls.from_pretrained.return_value = mock_model

        with patch.object(svc, "_model_is_ready", return_value=True), \
             patch.object(svc, "M2M100Tokenizer", mock_tok_cls), \
             patch.object(svc, "M2M100ForConditionalGeneration", mock_model_cls):

            out = translate("Hôm nay tôi thấy ổn hơn nhiều.", "vi", "en")
            self.assertEqual(out, "I feel much better today.")
            self.assertEqual(mock_tok.src_lang, "vi")
            mock_model.generate.assert_called_once()
            self.assertTrue(mock_model.eval.called)

    def test_english_to_vietnamese_translation(self):
        """Test en -> vi translation flow."""
        mock_tok, mock_model = create_mock_tokenizer_and_model()
        mock_tok.batch_decode.return_value = ["Tôi ở đây để lắng nghe bạn."]

        mock_tok_cls = MagicMock()
        mock_tok_cls.from_pretrained.return_value = mock_tok
        mock_model_cls = MagicMock()
        mock_model_cls.from_pretrained.return_value = mock_model

        with patch.object(svc, "_model_is_ready", return_value=True), \
             patch.object(svc, "M2M100Tokenizer", mock_tok_cls), \
             patch.object(svc, "M2M100ForConditionalGeneration", mock_model_cls):

            out = translate("I am here to listen to you.", "en", "vi")
            self.assertEqual(out, "Tôi ở đây để lắng nghe bạn.")
            self.assertEqual(mock_tok.src_lang, "en")

    def test_lazy_loading_and_loaded_only_once(self):
        """Model should not load until needed, and only load once across multiple calls."""
        mock_tok, mock_model = create_mock_tokenizer_and_model()

        mock_tok_cls = MagicMock()
        mock_tok_cls.from_pretrained.return_value = mock_tok
        mock_model_cls = MagicMock()
        mock_model_cls.from_pretrained.return_value = mock_model

        with patch.object(svc, "_model_is_ready", return_value=True), \
             patch.object(svc, "M2M100Tokenizer", mock_tok_cls), \
             patch.object(svc, "M2M100ForConditionalGeneration", mock_model_cls):

            self.assertIsNone(svc._model)
            # First call triggers load
            translate("First call", "vi", "en")
            self.assertEqual(mock_tok_cls.from_pretrained.call_count, 1)
            self.assertEqual(mock_model_cls.from_pretrained.call_count, 1)

            # Second call reuses existing model
            translate("Second call", "en", "vi")
            self.assertEqual(mock_tok_cls.from_pretrained.call_count, 1)
            self.assertEqual(mock_model_cls.from_pretrained.call_count, 1)

    def test_missing_model_raises_translation_error(self):
        """Missing model files trigger a clear TranslationError."""
        with patch.object(svc, "_model_is_ready", return_value=False):
            with self.assertRaises(TranslationError) as ctx:
                translate("Hello", "en", "vi")
            self.assertIn("not found or incomplete", str(ctx.exception).lower())

    def test_model_initialization_failure_raises_translation_error(self):
        """Exception during from_pretrained raises TranslationError."""
        mock_tok_cls = MagicMock()
        mock_tok_cls.from_pretrained.side_effect = RuntimeError("Disk read error")

        with patch.object(svc, "_model_is_ready", return_value=True), \
             patch.object(svc, "M2M100Tokenizer", mock_tok_cls):
            with self.assertRaises(TranslationError) as ctx:
                translate("Hello", "en", "vi")
            self.assertIn("failed to initialise", str(ctx.exception).lower())
            self.assertIsNone(svc._model)
            self.assertIsNone(svc._tokenizer)

    def test_inference_failure_raises_translation_error(self):
        """Inference failure during generate raises TranslationError."""
        mock_tok, mock_model = create_mock_tokenizer_and_model()
        mock_model.generate.side_effect = RuntimeError("CPU execution failure")

        mock_tok_cls = MagicMock()
        mock_tok_cls.from_pretrained.return_value = mock_tok
        mock_model_cls = MagicMock()
        mock_model_cls.from_pretrained.return_value = mock_model

        with patch.object(svc, "_model_is_ready", return_value=True), \
             patch.object(svc, "M2M100Tokenizer", mock_tok_cls), \
             patch.object(svc, "M2M100ForConditionalGeneration", mock_model_cls):

            with self.assertRaises(TranslationError) as ctx:
                translate("Hello", "en", "vi")
            self.assertIn("inference failure", str(ctx.exception).lower())

    def test_no_raw_user_content_in_logs(self):
        """Sensitive mental health input must not appear in service log records."""
        mock_tok, mock_model = create_mock_tokenizer_and_model()
        mock_model.generate.side_effect = RuntimeError("Generation error")

        mock_tok_cls = MagicMock()
        mock_tok_cls.from_pretrained.return_value = mock_tok
        mock_model_cls = MagicMock()
        mock_model_cls.from_pretrained.return_value = mock_model

        secret_text = "SECRET_DEPRESSION_SYMPTOM_DATA_12345"

        with self.assertLogs(level="DEBUG") as cm:
            with patch.object(svc, "_model_is_ready", return_value=True), \
                 patch.object(svc, "M2M100Tokenizer", mock_tok_cls), \
                 patch.object(svc, "M2M100ForConditionalGeneration", mock_model_cls):

                logging.getLogger("app.translation").info("Service initialized without user payload")
                try:
                    translate(secret_text, "en", "vi")
                except TranslationError:
                    pass

        for record in cm.output:
            self.assertNotIn(secret_text, record)


if __name__ == "__main__":
    unittest.main()
