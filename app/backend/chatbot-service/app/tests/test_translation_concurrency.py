# app/tests/test_translation_concurrency.py
"""Concurrency tests for translation service.

Ensures:
- Thread safety of translation calls under high concurrent load.
- _inference_lock prevents cross-request tokenizer language contamination.
- Multiple threads requesting different source languages (vi, ja, zh, ko)
  never interleave or corrupt tokenizer.src_lang.
- Lazy model loading occurs exactly once even when multiple threads hit it simultaneously.

Runs cleanly with:
    python3 -m unittest app/tests/test_translation_concurrency.py
    pytest app/tests/test_translation_concurrency.py
"""

from __future__ import annotations

import threading
import time
import unittest
from unittest.mock import MagicMock, patch

import app.translation.translation_service as svc
from app.translation.translation_service import translate


class TestTranslationConcurrency(unittest.TestCase):

    def setUp(self):
        svc._model = None
        svc._tokenizer = None

    def tearDown(self):
        svc._model = None
        svc._tokenizer = None

    def test_concurrent_different_source_languages_no_contamination(self):
        """Concurrent requests with different source languages must not contaminate tokenizer.src_lang."""
        mock_tok = MagicMock()
        mock_tok.src_lang = None
        mock_tok.get_lang_id = MagicMock(return_value=128001)

        def fake_call(text, return_tensors="pt", padding=True, **kwargs):
            return {"input_ids": MagicMock()}

        mock_tok.side_effect = fake_call
        mock_tok.encode = MagicMock(return_value=[1, 2, 3])

        def fake_decode(tokens, skip_special_tokens=True):
            # Decode returns text reflecting what src_lang was active
            current_lang = mock_tok.src_lang
            return [f"translated_from_{current_lang}"]

        mock_tok.batch_decode = MagicMock(side_effect=fake_decode)

        mock_model = MagicMock()
        mock_model.config.max_position_embeddings = 1024
        mock_model.eval = MagicMock(return_value=mock_model)

        def fake_generate(**kwargs):
            time.sleep(0.01)
            return MagicMock()

        mock_model.generate = MagicMock(side_effect=fake_generate)

        mock_tok_cls = MagicMock()
        mock_tok_cls.from_pretrained.return_value = mock_tok
        mock_model_cls = MagicMock()
        mock_model_cls.from_pretrained.return_value = mock_model

        requests = [
            ("Xin chào bạn", "vi", "en"),
            ("こんにちは", "ja", "en"),
            ("你好吗", "zh", "en"),
            ("안녕하세요", "ko", "en"),
            ("Good morning", "en", "vi"),
        ]

        results = {}
        errors = []

        def worker(idx, text, s_lang, t_lang):
            try:
                res = translate(text, s_lang, t_lang)
                results[idx] = (s_lang, res)
            except Exception as e:
                errors.append(e)

        threads = []
        with patch.object(svc, "_model_is_ready", return_value=True), \
             patch.object(svc, "M2M100Tokenizer", mock_tok_cls), \
             patch.object(svc, "M2M100ForConditionalGeneration", mock_model_cls):

            for i, (text, s_lang, t_lang) in enumerate(requests):
                t = threading.Thread(target=worker, args=(i, text, s_lang, t_lang))
                threads.append(t)

            for t in threads:
                t.start()
            for t in threads:
                t.join()

        self.assertFalse(errors, f"Errors occurred during concurrent execution: {errors}")
        self.assertEqual(len(results), len(requests))

        for idx, (expected_lang, output_text) in results.items():
            self.assertEqual(
                output_text,
                f"translated_from_{expected_lang}",
                f"Contamination detected! Expected language '{expected_lang}', got '{output_text}'",
            )

    def test_concurrent_initialization_loads_only_once(self):
        """Simultaneous calls on uninitialized service should only initialize model once."""
        mock_tok = MagicMock()
        mock_tok.src_lang = None
        mock_tok.get_lang_id = MagicMock(return_value=128001)
        mock_tok.side_effect = lambda text, **kw: {"input_ids": MagicMock()}
        mock_tok.encode = MagicMock(return_value=[1, 2])
        mock_tok.batch_decode = MagicMock(return_value=["output"])

        mock_model = MagicMock()
        mock_model.config.max_position_embeddings = 1024
        mock_model.eval = MagicMock(return_value=mock_model)
        mock_model.generate = MagicMock(return_value=MagicMock())

        load_count = {"tok": 0, "model": 0}

        def counting_load_tok(*args, **kwargs):
            time.sleep(0.02)
            load_count["tok"] += 1
            return mock_tok

        def counting_load_mod(*args, **kwargs):
            time.sleep(0.02)
            load_count["model"] += 1
            return mock_model

        mock_tok_cls = MagicMock()
        mock_tok_cls.from_pretrained.side_effect = counting_load_tok
        mock_model_cls = MagicMock()
        mock_model_cls.from_pretrained.side_effect = counting_load_mod

        threads = []
        errors = []

        def worker():
            try:
                translate("Test parallel init", "vi", "en")
            except Exception as e:
                errors.append(e)

        with patch.object(svc, "_model_is_ready", return_value=True), \
             patch.object(svc, "M2M100Tokenizer", mock_tok_cls), \
             patch.object(svc, "M2M100ForConditionalGeneration", mock_model_cls):

            for _ in range(8):
                t = threading.Thread(target=worker)
                threads.append(t)

            for t in threads:
                t.start()
            for t in threads:
                t.join()

        self.assertFalse(errors)
        self.assertEqual(load_count["tok"], 1, f"Tokenizer loaded {load_count['tok']} times instead of 1")
        self.assertEqual(load_count["model"], 1, f"Model loaded {load_count['model']} times instead of 1")


if __name__ == "__main__":
    unittest.main()
