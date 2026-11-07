# app/tests/test_translation_integration.py
"""Integration tests for Translation API and real-model execution.

Distinguishes between:
1. API tests with mocked translation backend (verifying HTTP endpoints, status codes, schemas).
2. Real-model integration tests (requiring downloaded M2M100 weights; skipped when weights absent).

Runs cleanly with:
    python3 -m unittest app/tests/test_translation_integration.py
    pytest app/tests/test_translation_integration.py
"""

from __future__ import annotations

import unittest
from unittest.mock import MagicMock, patch
from fastapi import FastAPI
from fastapi.testclient import TestClient

from app.routers.translation_router import router
import app.translation.translation_service as svc

# Create isolated test application with translation router attached at canonical prefix
app = FastAPI(title="Translation Test App")
app.include_router(router, prefix="/api/v1/translation")

client = TestClient(app)


class TestTranslationAPIMocked(unittest.TestCase):
    """API endpoint tests using mocked translation model."""

    def setUp(self):
        svc._model = None
        svc._tokenizer = None

    def tearDown(self):
        svc._model = None
        svc._tokenizer = None

    def test_translate_api_success_schema(self):
        """Valid translation request returns 200 and expected schema fields."""
        mock_tok = MagicMock()
        mock_tok.src_lang = None
        mock_tok.get_lang_id = MagicMock(return_value=128001)
        mock_tok.side_effect = lambda text, **kw: {"input_ids": MagicMock()}
        mock_tok.encode = MagicMock(return_value=[1, 2, 3])
        mock_tok.batch_decode = MagicMock(return_value=["Hello"])

        mock_model = MagicMock()
        mock_model.config.max_position_embeddings = 1024
        mock_model.eval = MagicMock(return_value=mock_model)
        mock_model.generate = MagicMock(return_value=MagicMock())

        mock_tok_cls = MagicMock()
        mock_tok_cls.from_pretrained.return_value = mock_tok
        mock_model_cls = MagicMock()
        mock_model_cls.from_pretrained.return_value = mock_model

        payload = {
            "text": "Xin chào",
            "source_lang": "vi",
            "target_lang": "en"
        }

        with patch.object(svc, "_model_is_ready", return_value=True), \
             patch.object(svc, "M2M100Tokenizer", mock_tok_cls), \
             patch.object(svc, "M2M100ForConditionalGeneration", mock_model_cls):

            response = client.post("/api/v1/translation/translate", json=payload)

        self.assertEqual(response.status_code, 200)
        data = response.json()
        self.assertEqual(data["source_text"], "Xin chào")
        self.assertEqual(data["translated_text"], "Hello")
        self.assertEqual(data["source_language"], "vi")
        self.assertEqual(data["target_language"], "en")
        self.assertEqual(data["model_name"], "facebook/m2m100_418M")
        self.assertEqual(data["status"], "success")

    def test_translate_api_unsupported_language_returns_422(self):
        """Unsupported language code triggers Pydantic HTTP 422 error."""
        payload = {
            "text": "Bonjour tout le monde",
            "source_lang": "fr",
            "target_lang": "en"
        }
        response = client.post("/api/v1/translation/translate", json=payload)
        self.assertEqual(response.status_code, 422)

    def test_translate_api_missing_fields_returns_422(self):
        """Missing required JSON fields returns HTTP 422."""
        payload = {
            "text": "Hello"
        }
        response = client.post("/api/v1/translation/translate", json=payload)
        self.assertEqual(response.status_code, 422)

    def test_translate_api_same_language_passthrough_200(self):
        """Same source and target language returns original text directly."""
        payload = {
            "text": "I feel slightly stressed.",
            "source_lang": "en",
            "target_lang": "en"
        }
        response = client.post("/api/v1/translation/translate", json=payload)
        self.assertEqual(response.status_code, 200)
        data = response.json()
        self.assertEqual(data["source_text"], "I feel slightly stressed.")
        self.assertEqual(data["translated_text"], "I feel slightly stressed.")
        self.assertEqual(data["status"], "success")

    def test_translate_api_missing_model_returns_503_without_path_leak(self):
        """Missing model returns HTTP 503 without revealing filesystem paths in response body."""
        payload = {
            "text": "Xin chào",
            "source_lang": "vi",
            "target_lang": "en"
        }

        with patch.object(svc, "_model_is_ready", return_value=False):
            response = client.post("/api/v1/translation/translate", json=payload)

        self.assertEqual(response.status_code, 503)
        data = response.json()
        self.assertIn("detail", data)
        # Ensure internal filesystem paths or stack traces are not exposed
        self.assertNotIn("/models/translation", data["detail"])
        self.assertNotIn("traceback", data["detail"].lower())


class TestTranslationRealModelIntegration(unittest.TestCase):
    """Real-model integration tests.

    These tests require the actual M2M100 418M weights to be present in the
    mounted model directory. They are automatically skipped if weights are absent.
    """

    @unittest.skipIf(
        not svc._model_is_ready(),
        "M2M100 model weights are not present on disk; real model test skipped."
    )
    def test_real_model_inference_vi_to_en(self):
        """Real inference test (only runs when actual model weights are present)."""
        payload = {
            "text": "Xin chào, tôi cần được hỗ trợ tâm lý.",
            "source_lang": "vi",
            "target_lang": "en"
        }
        response = client.post("/api/v1/translation/translate", json=payload)
        self.assertEqual(response.status_code, 200)
        data = response.json()
        self.assertGreater(len(data["translated_text"]), 0)
        self.assertEqual(data["source_language"], "vi")
        self.assertEqual(data["target_language"], "en")


if __name__ == "__main__":
    unittest.main()
