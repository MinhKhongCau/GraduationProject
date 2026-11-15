"""Unit tests for chat_orchestrator_stream focusing on Layer 5 streaming safety.

Verifies:
1. Normal generation: complete aggregation and filtering before yield.
2. Filtering order: Layer 5 is invoked before any text is emitted to the client.
3. Filter transformation: replacement by Layer 5 is respected.
4. Conversation history: exact filtered response is saved to history.
5. LLM failure: partial chunks are discarded; safe fallback yielded and saved.
6. Layer 5 failure: raw model output is discarded; safe fallback yielded and saved.
7. Layer 0 compatibility: emergency protocol aborts before LLM invocation.
8. Sanitized logging: no sensitive user message or model output logged on error.
9. Empty LLM output: empty response replaced with safe fallback without calling Layer 5.
10. Whitespace-only LLM output: whitespace replaced with safe fallback without calling Layer 5.
11. History save order: history is persisted before the first yield occurs.
12. History storage exception: does not suppress safe filtered output, logs sanitized error.
"""

from __future__ import annotations

import sys
import unittest
from contextlib import ExitStack
from types import SimpleNamespace
from unittest.mock import MagicMock, call, patch

from app.translation.translation_service import TranslationError

# Mock host-venv missing packages to enable running unit tests in isolation
for missing_pkg in [
    "langchain_ollama",
    "langchain_postgres",
    "langchain_postgres.vectorstores",
]:
    if missing_pkg not in sys.modules:
        sys.modules[missing_pkg] = MagicMock()

# Prevent network calls during layer1_nlu import
with patch("transformers.AutoFeatureExtractor.from_pretrained", return_value=MagicMock()), \
     patch("transformers.AutoModelForAudioClassification.from_pretrained", return_value=MagicMock()):
    from app.core.orchestrator import (
        SAFE_FALLBACK_RESPONSE,
        VIETNAMESE_OUTPUT_TRANSLATION_FAILURE_RESPONSE,
        VIETNAMESE_OUTPUT_VALIDATION_FAILURE_RESPONSE,
        VIETNAMESE_SAFE_FALLBACK_RESPONSE,
        _LANGUAGE_CLARIFICATION_RESPONSE,
        _UNSUPPORTED_LANGUAGE_RESPONSE,
        chat_orchestrator_stream,
    )


class TestOrchestratorStreamingSafety(unittest.TestCase):
    """Test suite for orchestrator streaming safety and Layer 5 enforcement."""

    def setUp(self):
        self.session_id = "test-session-456"
        self.user_message = "I have been feeling stressed about exams."
        self.detector = MagicMock()
        self.detector.detect.return_value = SimpleNamespace(
            status="detected", language="en", candidate_language="en"
        )
        self.get_detector = patch(
            "app.core.orchestrator.get_language_detector", return_value=self.detector
        )
        self.get_detector.start()
        self.addCleanup(self.get_detector.stop)

    @patch("app.core.orchestrator.layer2_dialogue.save_chat_turn")
    @patch("app.core.orchestrator.layer5_filter.filter_output")
    @patch("app.core.orchestrator.layer4_llm.generate_response_stream")
    @patch("app.core.orchestrator.layer3_rag.retrieve_interventions")
    @patch("app.core.orchestrator.layer2_dialogue.get_chat_history")
    @patch("app.core.orchestrator.layer1_nlu.analyze_audio_emotion")
    @patch("app.core.orchestrator.layer1_nlu.analyze_text_emotion")
    @patch("app.core.orchestrator.layer0_safety.check_input_risk")
    def test_normal_generation_aggregates_before_yielding(
        self,
        mock_l0,
        mock_l1_text,
        mock_l1_audio,
        mock_l2_get,
        mock_l3,
        mock_l4_stream,
        mock_l5_filter,
        mock_l2_save,
    ):
        """1. Verify multiple chunks are aggregated and Layer 5 receives the complete text."""
        mock_l0.return_value = (True, "")
        mock_l1_text.return_value = "anxiety"
        mock_l2_get.return_value = ""
        mock_l3.return_value = "RAG context on breathing."
        
        # Simulate LLM producing 3 chunks
        mock_l4_stream.return_value = iter(["It is normal ", "to feel exam stress. ", "Try deep breathing."])
        mock_l5_filter.side_effect = lambda text, lang="en": text

        generator = chat_orchestrator_stream(self.session_id, self.user_message)
        yielded_chunks = list(generator)

        # Must yield exactly 1 consolidated block, not 3 individual chunks
        self.assertEqual(len(yielded_chunks), 1)
        expected_full = "It is normal to feel exam stress. Try deep breathing."
        self.assertEqual(yielded_chunks[0], expected_full)

        # Layer 5 must have received the full consolidated text
        mock_l5_filter.assert_called_once_with(expected_full, lang="en")

    @patch("app.core.orchestrator.layer2_dialogue.save_chat_turn")
    @patch("app.core.orchestrator.layer5_filter.filter_output")
    @patch("app.core.orchestrator.layer4_llm.generate_response_stream")
    @patch("app.core.orchestrator.layer3_rag.retrieve_interventions")
    @patch("app.core.orchestrator.layer2_dialogue.get_chat_history")
    @patch("app.core.orchestrator.layer1_nlu.analyze_text_emotion")
    @patch("app.core.orchestrator.layer0_safety.check_input_risk")
    def test_filtering_order_layer5_called_before_first_yield(
        self,
        mock_l0,
        mock_l1_text,
        mock_l2_get,
        mock_l3,
        mock_l4_stream,
        mock_l5_filter,
        mock_l2_save,
    ):
        """2. Verify that Layer 5 is called BEFORE the client observes any output."""
        mock_l0.return_value = (True, "")
        mock_l1_text.return_value = "neutral"
        mock_l2_get.return_value = ""
        mock_l3.return_value = ""
        mock_l4_stream.return_value = iter(["Chunk A ", "Chunk B"])

        call_order = []

        def tracked_filter(text, lang="en"):
            call_order.append("layer5_filtered")
            return text

        mock_l5_filter.side_effect = tracked_filter

        gen = chat_orchestrator_stream(self.session_id, self.user_message)
        
        # Generator created, but not advanced yet
        self.assertEqual(call_order, [])

        # Advance generator to first yield
        first_item = next(gen)
        call_order.append("client_received")

        # Layer 5 must have executed before the client received the item
        self.assertEqual(call_order, ["layer5_filtered", "client_received"])
        self.assertEqual(first_item, "Chunk A Chunk B")

    @patch("app.core.orchestrator.layer2_dialogue.save_chat_turn")
    @patch("app.core.orchestrator.layer5_filter.filter_output")
    @patch("app.core.orchestrator.layer4_llm.generate_response_stream")
    @patch("app.core.orchestrator.layer3_rag.retrieve_interventions")
    @patch("app.core.orchestrator.layer2_dialogue.get_chat_history")
    @patch("app.core.orchestrator.layer1_nlu.analyze_text_emotion")
    @patch("app.core.orchestrator.layer0_safety.check_input_risk")
    def test_filter_transformation_emits_only_replacement(
        self,
        mock_l0,
        mock_l1_text,
        mock_l2_get,
        mock_l3,
        mock_l4_stream,
        mock_l5_filter,
        mock_l2_save,
    ):
        """3. When Layer 5 replaces an unsafe model response, only the replacement reaches the client."""
        mock_l0.return_value = (True, "")
        mock_l1_text.return_value = "neutral"
        mock_l2_get.return_value = ""
        mock_l3.return_value = ""
        
        # Unsafe model output containing medical diagnosis
        mock_l4_stream.return_value = iter(["I diagnose you with major depression. ", "Take 20mg Prozac daily."])
        safe_disclaimer = (
            "I'm here to support you, but please note that I cannot provide medical diagnoses or prescribe treatments."
        )
        mock_l5_filter.return_value = safe_disclaimer

        yielded = list(chat_orchestrator_stream(self.session_id, self.user_message))

        self.assertEqual(len(yielded), 1)
        self.assertEqual(yielded[0], safe_disclaimer)
        self.assertNotIn("major depression", yielded[0])
        self.assertNotIn("Prozac", yielded[0])

    @patch("app.core.orchestrator.layer2_dialogue.save_chat_turn")
    @patch("app.core.orchestrator.layer5_filter.filter_output")
    @patch("app.core.orchestrator.layer4_llm.generate_response_stream")
    @patch("app.core.orchestrator.layer3_rag.retrieve_interventions")
    @patch("app.core.orchestrator.layer2_dialogue.get_chat_history")
    @patch("app.core.orchestrator.layer1_nlu.analyze_text_emotion")
    @patch("app.core.orchestrator.layer0_safety.check_input_risk")
    def test_conversation_history_stores_exact_filtered_response(
        self,
        mock_l0,
        mock_l1_text,
        mock_l2_get,
        mock_l3,
        mock_l4_stream,
        mock_l5_filter,
        mock_l2_save,
    ):
        """4. Verify save_chat_turn() stores the exact filtered response that was yielded."""
        mock_l0.return_value = (True, "")
        mock_l1_text.return_value = "neutral"
        mock_l2_get.return_value = ""
        mock_l3.return_value = ""
        mock_l4_stream.return_value = iter(["Unsafe diagnosis text."])
        
        safe_replacement = "Safe medical disclaimer."
        mock_l5_filter.return_value = safe_replacement

        yielded = list(chat_orchestrator_stream(self.session_id, self.user_message))

        mock_l2_save.assert_called_once_with(self.session_id, self.user_message, safe_replacement)
        self.assertEqual(mock_l2_save.call_args[0][2], yielded[0])

    @patch("app.core.orchestrator.layer2_dialogue.save_chat_turn")
    @patch("app.core.orchestrator.layer5_filter.filter_output")
    @patch("app.core.orchestrator.layer4_llm.generate_response_stream")
    @patch("app.core.orchestrator.layer3_rag.retrieve_interventions")
    @patch("app.core.orchestrator.layer2_dialogue.get_chat_history")
    @patch("app.core.orchestrator.layer1_nlu.analyze_text_emotion")
    @patch("app.core.orchestrator.layer0_safety.check_input_risk")
    def test_llm_generation_failure_discards_partial_chunks(
        self,
        mock_l0,
        mock_l1_text,
        mock_l2_get,
        mock_l3,
        mock_l4_stream,
        mock_l5_filter,
        mock_l2_save,
    ):
        """5. Simulate LLM failure mid-stream; verify partial chunks are never yielded or saved."""
        mock_l0.return_value = (True, "")
        mock_l1_text.return_value = "neutral"
        mock_l2_get.return_value = ""
        mock_l3.return_value = ""

        # Generator yields partial chunk then raises error
        def failing_stream():
            yield "Partial half-sentence before crash... "
            raise ConnectionError("Ollama connection severed")

        mock_l4_stream.return_value = failing_stream()

        yielded = list(chat_orchestrator_stream(self.session_id, self.user_message))

        self.assertEqual(len(yielded), 1)
        self.assertEqual(yielded[0], SAFE_FALLBACK_RESPONSE)
        self.assertNotIn("Partial", yielded[0])
        mock_l5_filter.assert_not_called()

        # Verify history saved the fallback, NOT the partial chunk
        mock_l2_save.assert_called_once_with(self.session_id, self.user_message, SAFE_FALLBACK_RESPONSE)

    @patch("app.core.orchestrator.layer2_dialogue.save_chat_turn")
    @patch("app.core.orchestrator.layer5_filter.filter_output")
    @patch("app.core.orchestrator.layer4_llm.generate_response_stream")
    @patch("app.core.orchestrator.layer3_rag.retrieve_interventions")
    @patch("app.core.orchestrator.layer2_dialogue.get_chat_history")
    @patch("app.core.orchestrator.layer1_nlu.analyze_text_emotion")
    @patch("app.core.orchestrator.layer0_safety.check_input_risk")
    def test_llm_generation_failure_before_first_chunk_yields_only_fallback(
        self,
        mock_l0,
        mock_l1_text,
        mock_l2_get,
        mock_l3,
        mock_l4_stream,
        mock_l5_filter,
        mock_l2_save,
    ):
        mock_l0.return_value = (True, "")
        mock_l1_text.return_value = "neutral"
        mock_l2_get.return_value = ""
        mock_l3.return_value = ""

        def failing_stream():
            raise ConnectionError("Generation failed before first chunk")
            yield "unreachable"

        mock_l4_stream.return_value = failing_stream()

        yielded = list(chat_orchestrator_stream(self.session_id, self.user_message))

        self.assertEqual(yielded, [SAFE_FALLBACK_RESPONSE])
        mock_l5_filter.assert_not_called()
        mock_l2_save.assert_called_once_with(
            self.session_id, self.user_message, SAFE_FALLBACK_RESPONSE
        )

    @patch("app.core.orchestrator.layer2_dialogue.save_chat_turn")
    @patch("app.core.orchestrator.layer5_filter.filter_output")
    @patch("app.core.orchestrator.layer4_llm.generate_response_stream")
    @patch("app.core.orchestrator.layer3_rag.retrieve_interventions")
    @patch("app.core.orchestrator.layer2_dialogue.get_chat_history")
    @patch("app.core.orchestrator.layer1_nlu.analyze_text_emotion")
    @patch("app.core.orchestrator.layer0_safety.check_input_risk")
    def test_layer5_failure_discards_raw_model_output(
        self,
        mock_l0,
        mock_l1_text,
        mock_l2_get,
        mock_l3,
        mock_l4_stream,
        mock_l5_filter,
        mock_l2_save,
    ):
        """6. When Layer 5 raises an exception, raw model output is never yielded or saved."""
        mock_l0.return_value = (True, "")
        mock_l1_text.return_value = "neutral"
        mock_l2_get.return_value = ""
        mock_l3.return_value = ""
        mock_l4_stream.return_value = iter(["Raw sensitive ", "AI generation."])
        
        # Layer 5 crashes
        mock_l5_filter.side_effect = RuntimeError("Regex engine out of memory")

        yielded = list(chat_orchestrator_stream(self.session_id, self.user_message))

        self.assertEqual(len(yielded), 1)
        self.assertEqual(yielded[0], SAFE_FALLBACK_RESPONSE)
        self.assertNotIn("Raw sensitive", yielded[0])

        # A filter failure must not persist a normal conversation turn.
        mock_l2_save.assert_not_called()

    @patch("app.core.orchestrator.layer2_dialogue.save_chat_turn")
    @patch("app.core.orchestrator.layer5_filter.filter_output")
    @patch("app.core.orchestrator.layer4_llm.generate_response_stream")
    @patch("app.core.orchestrator.layer3_rag.retrieve_interventions")
    @patch("app.core.orchestrator.layer2_dialogue.get_chat_history")
    @patch("app.core.orchestrator.layer1_nlu.analyze_text_emotion")
    @patch("app.core.orchestrator.layer0_safety.check_input_risk")
    def test_layer0_compatibility_unsafe_input_triggers_immediate_abort(
        self,
        mock_l0,
        mock_l1_text,
        mock_l2_get,
        mock_l3,
        mock_l4_stream,
        mock_l5_filter,
        mock_l2_save,
    ):
        """7. Unsafe crisis input returns Layer 0 emergency response without calling Qwen or Layer 5."""
        crisis_message = "I just want to end my life."
        crisis_response = "I am so sorry you are feeling this way... Hotline: 1900 1267"
        mock_l0.return_value = (False, crisis_response)

        yielded = list(chat_orchestrator_stream(self.session_id, crisis_message))

        self.assertEqual(yielded, [crisis_response])
        mock_l4_stream.assert_not_called()
        mock_l5_filter.assert_not_called()
        mock_l2_save.assert_not_called()

    @patch("app.core.orchestrator.layer2_dialogue.save_chat_turn")
    @patch("app.core.orchestrator.layer5_filter.filter_output")
    @patch("app.core.orchestrator.layer4_llm.generate_response_stream")
    @patch("app.core.orchestrator.layer3_rag.retrieve_interventions")
    @patch("app.core.orchestrator.layer2_dialogue.get_chat_history")
    @patch("app.core.orchestrator.layer1_nlu.analyze_text_emotion")
    @patch("app.core.orchestrator.layer0_safety.check_input_risk")
    def test_sanitized_logging_does_not_leak_raw_content(
        self,
        mock_l0,
        mock_l1_text,
        mock_l2_get,
        mock_l3,
        mock_l4_stream,
        mock_l5_filter,
        mock_l2_save,
    ):
        """8. Verify error logs do not contain raw user messages or raw LLM output."""
        mock_l0.return_value = (True, "")
        mock_l1_text.return_value = "neutral"
        mock_l2_get.return_value = ""
        mock_l3.return_value = ""
        sensitive_llm_text = "HighlySensitiveDiagnosisDrugName123"
        mock_l4_stream.return_value = iter([sensitive_llm_text])
        mock_l5_filter.side_effect = ValueError("Filter processing failed")

        with self.assertLogs("app.core.orchestrator", level="ERROR") as log_capture:
            list(chat_orchestrator_stream(self.session_id, "SensitiveUserMessageXYZ"))

        all_log_text = " ".join(log_capture.output)
        self.assertNotIn(sensitive_llm_text, all_log_text)
        self.assertNotIn("SensitiveUserMessageXYZ", all_log_text)
        self.assertIn("ValueError", all_log_text)

    @patch("app.core.orchestrator.layer2_dialogue.save_chat_turn")
    @patch("app.core.orchestrator.layer5_filter.filter_output")
    @patch("app.core.orchestrator.layer4_llm.generate_response_stream")
    @patch("app.core.orchestrator.layer3_rag.retrieve_interventions")
    @patch("app.core.orchestrator.layer2_dialogue.get_chat_history")
    @patch("app.core.orchestrator.layer1_nlu.analyze_text_emotion")
    @patch("app.core.orchestrator.layer0_safety.check_input_risk")
    def test_empty_llm_generator_yields_and_saves_safe_fallback(
        self,
        mock_l0,
        mock_l1_text,
        mock_l2_get,
        mock_l3,
        mock_l4_stream,
        mock_l5_filter,
        mock_l2_save,
    ):
        """9. When LLM produces an empty generator, yield and save SAFE_FALLBACK_RESPONSE without calling Layer 5."""
        mock_l0.return_value = (True, "")
        mock_l1_text.return_value = "neutral"
        mock_l2_get.return_value = ""
        mock_l3.return_value = ""
        mock_l4_stream.return_value = iter([])  # Empty stream

        yielded = list(chat_orchestrator_stream(self.session_id, self.user_message))

        self.assertEqual(len(yielded), 1)
        self.assertEqual(yielded[0], SAFE_FALLBACK_RESPONSE)
        mock_l5_filter.assert_not_called()
        mock_l2_save.assert_called_once_with(self.session_id, self.user_message, SAFE_FALLBACK_RESPONSE)

    @patch("app.core.orchestrator.layer2_dialogue.save_chat_turn")
    @patch("app.core.orchestrator.layer5_filter.filter_output")
    @patch("app.core.orchestrator.layer4_llm.generate_response_stream")
    @patch("app.core.orchestrator.layer3_rag.retrieve_interventions")
    @patch("app.core.orchestrator.layer2_dialogue.get_chat_history")
    @patch("app.core.orchestrator.layer1_nlu.analyze_text_emotion")
    @patch("app.core.orchestrator.layer0_safety.check_input_risk")
    def test_whitespace_only_llm_output_yields_and_saves_safe_fallback(
        self,
        mock_l0,
        mock_l1_text,
        mock_l2_get,
        mock_l3,
        mock_l4_stream,
        mock_l5_filter,
        mock_l2_save,
    ):
        """10. When LLM produces whitespace-only chunks, yield and save SAFE_FALLBACK_RESPONSE without calling Layer 5."""
        mock_l0.return_value = (True, "")
        mock_l1_text.return_value = "neutral"
        mock_l2_get.return_value = ""
        mock_l3.return_value = ""
        mock_l4_stream.return_value = iter(["   \n", "\t  ", " \n\n "])

        yielded = list(chat_orchestrator_stream(self.session_id, self.user_message))

        self.assertEqual(len(yielded), 1)
        self.assertEqual(yielded[0], SAFE_FALLBACK_RESPONSE)
        mock_l5_filter.assert_not_called()
        mock_l2_save.assert_called_once_with(self.session_id, self.user_message, SAFE_FALLBACK_RESPONSE)

    @patch("app.core.orchestrator.layer2_dialogue.save_chat_turn")
    @patch("app.core.orchestrator.layer5_filter.filter_output")
    @patch("app.core.orchestrator.layer4_llm.generate_response_stream")
    @patch("app.core.orchestrator.layer3_rag.retrieve_interventions")
    @patch("app.core.orchestrator.layer2_dialogue.get_chat_history")
    @patch("app.core.orchestrator.layer1_nlu.analyze_text_emotion")
    @patch("app.core.orchestrator.layer0_safety.check_input_risk")
    def test_history_save_occurs_before_first_yield(
        self,
        mock_l0,
        mock_l1_text,
        mock_l2_get,
        mock_l3,
        mock_l4_stream,
        mock_l5_filter,
        mock_l2_save,
    ):
        """11. Verify save_chat_turn executes strictly before the first yielded response is emitted to the client."""
        mock_l0.return_value = (True, "")
        mock_l1_text.return_value = "neutral"
        mock_l2_get.return_value = ""
        mock_l3.return_value = ""
        mock_l4_stream.return_value = iter(["Valid response."])
        mock_l5_filter.return_value = "Safe filtered response."

        call_sequence = []

        def tracked_save(sid, msg, resp):
            call_sequence.append("history_saved")

        mock_l2_save.side_effect = tracked_save

        gen = chat_orchestrator_stream(self.session_id, self.user_message)
        self.assertEqual(call_sequence, [])

        item = next(gen)
        call_sequence.append("client_received_yield")

        # History save must occur BEFORE the client receives the yielded text
        self.assertEqual(call_sequence, ["history_saved", "client_received_yield"])
        self.assertEqual(item, "Safe filtered response.")

    @patch("app.core.orchestrator.layer2_dialogue.save_chat_turn")
    @patch("app.core.orchestrator.layer5_filter.filter_output")
    @patch("app.core.orchestrator.layer4_llm.generate_response_stream")
    @patch("app.core.orchestrator.layer3_rag.retrieve_interventions")
    @patch("app.core.orchestrator.layer2_dialogue.get_chat_history")
    @patch("app.core.orchestrator.layer1_nlu.analyze_text_emotion")
    @patch("app.core.orchestrator.layer0_safety.check_input_risk")
    def test_history_storage_exception_does_not_suppress_filtered_response(
        self,
        mock_l0,
        mock_l1_text,
        mock_l2_get,
        mock_l3,
        mock_l4_stream,
        mock_l5_filter,
        mock_l2_save,
    ):
        """12. If dialogue history storage fails, do not expose raw text, do not crash, and do not suppress safe output."""
        mock_l0.return_value = (True, "")
        mock_l1_text.return_value = "neutral"
        mock_l2_get.return_value = ""
        mock_l3.return_value = ""
        mock_l4_stream.return_value = iter(["Valid text."])
        mock_l5_filter.return_value = "Safe filtered text."
        
        # History save raises an error
        mock_l2_save.side_effect = RuntimeError("Database disk full")

        with self.assertLogs("app.core.orchestrator", level="ERROR") as log_capture:
            yielded = list(chat_orchestrator_stream(self.session_id, self.user_message))

        # Client still receives the safety-filtered response
        self.assertEqual(yielded, ["Safe filtered text."])

        # Sanitized log recorded without leaking message content
        all_logs = " ".join(log_capture.output)
        self.assertIn("RuntimeError", all_logs)
        self.assertNotIn("Safe filtered text.", all_logs)
        self.assertNotIn(self.user_message, all_logs)


class TestOrchestratorLanguageDetection(unittest.TestCase):
    """Language-resolution and Layer 0 ordering tests without real FastText."""

    def setUp(self):
        self.session_id = "language-test-session"
        self.original_message = "Original user text must remain unchanged."
        self.detector = MagicMock()
        self.get_detector = patch(
            "app.core.orchestrator.get_language_detector", return_value=self.detector
        )
        self.get_detector.start()
        self.addCleanup(self.get_detector.stop)

    @staticmethod
    def _result(status, language=None, candidate_language=None):
        return SimpleNamespace(
            status=status,
            language=language,
            candidate_language=candidate_language,
        )

    def _run_with_pipeline_mocks(self, layer0_side_effect):
        with ExitStack() as stack:
            mocks = {
                "layer0": stack.enter_context(
                    patch("app.core.orchestrator.layer0_safety.check_input_risk")
                ),
                "translate": stack.enter_context(patch("app.core.orchestrator.translate")),
                "layer1_text": stack.enter_context(
                    patch("app.core.orchestrator.layer1_nlu.analyze_text_emotion")
                ),
                "layer1_audio": stack.enter_context(
                    patch("app.core.orchestrator.layer1_nlu.analyze_audio_emotion")
                ),
                "layer2_get": stack.enter_context(
                    patch("app.core.orchestrator.layer2_dialogue.get_chat_history")
                ),
                "layer2_save": stack.enter_context(
                    patch("app.core.orchestrator.layer2_dialogue.save_chat_turn")
                ),
                "layer3": stack.enter_context(
                    patch("app.core.orchestrator.layer3_rag.retrieve_interventions")
                ),
                "layer4": stack.enter_context(
                    patch("app.core.orchestrator.layer4_llm.generate_response_stream")
                ),
                "layer5": stack.enter_context(
                    patch("app.core.orchestrator.layer5_filter.filter_output")
                ),
            }
            mocks["layer0"].side_effect = layer0_side_effect
            mocks["translate"].return_value = "translated"
            mocks["layer1_text"].return_value = "neutral"
            mocks["layer1_audio"].return_value = "anxiety"
            mocks["layer2_get"].return_value = ""
            mocks["layer3"].return_value = "context"
            mocks["layer4"].return_value = iter(["complete response"])
            mocks["layer5"].side_effect = lambda text, lang="en": text
            result = list(chat_orchestrator_stream(self.session_id, self.original_message))
        return result, mocks

    @patch("app.core.orchestrator.layer0_safety.check_input_risk")
    def test_confident_supported_languages_are_passed_to_layer0(self, mock_l0):
        """FastText's supported confident short codes activate native Layer 0 patterns."""
        for language in ("en", "vi"):
            with self.subTest(language=language):
                self.detector.detect.return_value = self._result(
                    "detected", language=language, candidate_language=language
                )
                mock_l0.return_value = (False, f"{language} emergency")

                yielded = list(chat_orchestrator_stream(self.session_id, self.original_message))

                self.assertEqual(yielded, [f"{language} emergency"])
                mock_l0.assert_called_once_with(self.original_message, lang=language)
                self.detector.detect.assert_called_once_with(self.original_message)
                mock_l0.reset_mock()
                self.detector.detect.reset_mock()

    def test_uncertain_candidates_do_not_override_full_supported_scan(self):
        for candidate in ("vi", "ja"):
            with self.subTest(candidate=candidate):
                self.detector.detect.return_value = self._result(
                    "uncertain", candidate_language=candidate
                )
                yielded, mocks = self._run_with_pipeline_mocks(
                    lambda message, lang: (True, "")
                )

                self.assertEqual(yielded, [_LANGUAGE_CLARIFICATION_RESPONSE])
                self.assertEqual(
                    mocks["layer0"].call_args_list,
                    [
                        call(self.original_message, lang="en"),
                        call(self.original_message, lang="vi"),
                    ],
                )
                mocks["translate"].assert_not_called()
                mocks["layer1_text"].assert_not_called()
                mocks["layer2_get"].assert_not_called()
                mocks["layer2_save"].assert_not_called()
                self.detector.detect.assert_called_once_with(self.original_message)
                self.detector.detect.reset_mock()

    def test_unresolved_detection_states_run_full_scan_until_crisis(self):
        class MissingLanguage:
            status = "detected"

        class RaisingLanguage:
            status = "uncertain"
            candidate_language = "vi"

            @property
            def language(self):
                raise RuntimeError("private detector metadata")

        cases = (
            ("none", None, _LANGUAGE_CLARIFICATION_RESPONSE),
            ("unknown", self._result("unknown"), _LANGUAGE_CLARIFICATION_RESPONSE),
            ("uncertain-without-candidate", self._result("uncertain"), _LANGUAGE_CLARIFICATION_RESPONSE),
            ("missing-field", MissingLanguage(), _LANGUAGE_CLARIFICATION_RESPONSE),
            ("raising-property", RaisingLanguage(), _LANGUAGE_CLARIFICATION_RESPONSE),
            ("unsupported-fr", self._result("detected", "fr", "fr"), _UNSUPPORTED_LANGUAGE_RESPONSE),
            ("unsupported-ja", self._result("detected", "ja", "ja"), _UNSUPPORTED_LANGUAGE_RESPONSE),
            ("unsupported-ko", self._result("detected", "ko", "ko"), _UNSUPPORTED_LANGUAGE_RESPONSE),
            ("unsupported-zh", self._result("detected", "zh", "zh"), _UNSUPPORTED_LANGUAGE_RESPONSE),
            ("unsupported-zh-hant", self._result("detected", "zh-Hant"), _UNSUPPORTED_LANGUAGE_RESPONSE),
        )

        for name, result, expected_response in cases:
            with self.subTest(case=name):
                self.detector.detect.return_value = result

                def safe_scan(message, lang):
                    self.assertEqual(message, self.original_message)
                    return True, ""

                if name == "raising-property":
                    with self.assertLogs("app.core.orchestrator", level="WARNING") as logs:
                        yielded, mocks = self._run_with_pipeline_mocks(safe_scan)
                    self.assertIn("RuntimeError", " ".join(logs.output))
                    self.assertNotIn("private detector metadata", " ".join(logs.output))
                else:
                    yielded, mocks = self._run_with_pipeline_mocks(safe_scan)

                self.assertEqual(yielded, [expected_response])
                self.assertEqual(
                    mocks["layer0"].call_args_list,
                    [
                        call(self.original_message, lang="en"),
                        call(self.original_message, lang="vi"),
                    ],
                )
                self.detector.detect.assert_called_once_with(self.original_message)
                mocks["translate"].assert_not_called()
                mocks["layer1_text"].assert_not_called()
                mocks["layer2_get"].assert_not_called()
                mocks["layer2_save"].assert_not_called()
                self.detector.detect.reset_mock()
                self.detector.detect.side_effect = None

    def test_detector_execution_failure_runs_full_scan(self):
        self.detector.detect.side_effect = RuntimeError("private detector failure")

        def detect_at_vi(message, lang):
            self.assertEqual(message, self.original_message)
            return (False, "Vietnamese emergency") if lang == "vi" else (True, "")

        with self.assertLogs("app.core.orchestrator", level="WARNING") as logs:
            yielded, mocks = self._run_with_pipeline_mocks(detect_at_vi)

        self.assertEqual(yielded, ["Vietnamese emergency"])
        self.assertEqual(
            mocks["layer0"].call_args_list,
            [
                call(self.original_message, lang="en"),
                call(self.original_message, lang="vi"),
            ],
        )
        self.detector.detect.assert_called_once_with(self.original_message)
        self.assertIn("RuntimeError", " ".join(logs.output))
        self.assertNotIn("private detector failure", " ".join(logs.output))
        self.assertNotIn(self.original_message, " ".join(logs.output))

    def test_detector_failures_complete_safe_scan_and_skip_translation(self):
        expected_calls = [
            call(self.original_message, lang=language)
            for language in ("en", "vi")
        ]

        self.detector.detect.side_effect = RuntimeError("private detector failure")
        yielded, mocks = self._run_with_pipeline_mocks(lambda message, lang: (True, ""))
        self.assertEqual(yielded, [_LANGUAGE_CLARIFICATION_RESPONSE])
        self.assertEqual(mocks["layer0"].call_args_list, expected_calls)
        mocks["translate"].assert_not_called()
        mocks["layer1_text"].assert_not_called()
        mocks["layer2_save"].assert_not_called()
        self.detector.detect.assert_called_once_with(self.original_message)

        self.detector.detect.reset_mock()
        self.detector.detect.side_effect = None
        self.get_detector.stop()
        with patch(
            "app.core.orchestrator.get_language_detector",
            side_effect=RuntimeError("private initialization failure"),
        ) as get_detector:
            yielded, mocks = self._run_with_pipeline_mocks(
                lambda message, lang: (True, "")
            )

        self.assertEqual(yielded, [_LANGUAGE_CLARIFICATION_RESPONSE])
        self.assertEqual(mocks["layer0"].call_args_list, expected_calls)
        mocks["translate"].assert_not_called()
        mocks["layer1_text"].assert_not_called()
        mocks["layer2_save"].assert_not_called()
        get_detector.assert_called_once_with()

    def test_detector_initialization_failure_runs_full_scan(self):
        self.get_detector.stop()

        def detect_at_vi(message, lang):
            self.assertEqual(message, self.original_message)
            return (False, "Vietnamese emergency") if lang == "vi" else (True, "")

        with patch(
            "app.core.orchestrator.get_language_detector",
            side_effect=RuntimeError("private initialization failure"),
        ), self.assertLogs("app.core.orchestrator", level="WARNING") as logs:
            yielded, mocks = self._run_with_pipeline_mocks(detect_at_vi)

        self.assertEqual(yielded, ["Vietnamese emergency"])
        self.assertEqual(
            mocks["layer0"].call_args_list,
            [call(self.original_message, lang=lang) for lang in ("en", "vi")],
        )
        self.assertIn("RuntimeError", " ".join(logs.output))
        self.assertNotIn("private initialization failure", " ".join(logs.output))
        self.assertNotIn(self.original_message, " ".join(logs.output))

    def test_incorrect_candidate_does_not_hide_vietnamese_crisis(self):
        self.detector.detect.return_value = self._result(
            "uncertain", candidate_language="ko"
        )

        def vietnamese_crisis(message, lang):
            self.assertEqual(message, self.original_message)
            return (False, "Vietnamese emergency response") if lang == "vi" else (True, "")

        yielded, mocks = self._run_with_pipeline_mocks(vietnamese_crisis)

        self.assertEqual(yielded, ["Vietnamese emergency response"])
        self.assertEqual(
            mocks["layer0"].call_args_list,
            [
                call(self.original_message, lang="en"),
                call(self.original_message, lang="vi"),
            ],
        )
        for name in (
            "translate", "layer1_text", "layer1_audio", "layer2_get", "layer3",
            "layer4", "layer5", "layer2_save",
        ):
            mocks[name].assert_not_called()

    def test_crisis_at_each_full_scan_position_stops_with_detecting_response(self):
        languages = ("en", "vi")
        self.detector.detect.return_value = self._result("unknown")

        for crisis_index, crisis_language in enumerate(languages):
            with self.subTest(language=crisis_language):
                seen = []

                def detect_at_position(message, lang):
                    self.assertEqual(message, self.original_message)
                    seen.append(lang)
                    if lang == crisis_language:
                        return False, f"{crisis_language} localized emergency"
                    return True, ""

                yielded, mocks = self._run_with_pipeline_mocks(detect_at_position)

                self.assertEqual(yielded, [f"{crisis_language} localized emergency"])
                self.assertEqual(seen, list(languages[:crisis_index + 1]))
                self.assertEqual(
                    mocks["layer0"].call_args_list,
                    [call(self.original_message, lang=lang) for lang in seen],
                )
                self.detector.detect.assert_called_once_with(self.original_message)
                self.detector.detect.reset_mock()
                for name in (
                    "translate", "layer1_text", "layer1_audio", "layer2_get", "layer3",
                    "layer4", "layer5", "layer2_save",
                ):
                    mocks[name].assert_not_called()

    def test_layer0_exceptions_fail_closed_without_downstream_processing(self):
        failure_cases = (
            ("first-check", self._result("unknown"), ("en",)),
            ("after-safe-check", self._result("uncertain", candidate_language="vi"), ("en", "vi")),
        )
        for name, detection, expected_languages in failure_cases:
            with self.subTest(case=name):
                self.detector.detect.return_value = detection
                calls = []

                def fail_layer0(message, lang):
                    self.assertEqual(message, self.original_message)
                    calls.append(lang)
                    if len(calls) == len(expected_languages):
                        raise RuntimeError("private Layer 0 error")
                    return True, ""

                with self.assertLogs("app.core.orchestrator", level="ERROR") as logs:
                    yielded, mocks = self._run_with_pipeline_mocks(fail_layer0)

                self.assertEqual(yielded, [SAFE_FALLBACK_RESPONSE])
                self.assertEqual(calls, list(expected_languages))
                self.assertIn("RuntimeError", " ".join(logs.output))
                self.assertNotIn("private Layer 0 error", " ".join(logs.output))
                self.assertNotIn(self.original_message, " ".join(logs.output))
                for mock_name in (
                    "translate", "layer1_text", "layer1_audio", "layer2_get", "layer3",
                    "layer4", "layer5", "layer2_save",
                ):
                    mocks[mock_name].assert_not_called()

    def test_supported_regional_languages_normalize_for_layer0_and_translation(self):
        cases = (
            ("vi-VN", "vi", "vi"),
            ("en-US", "en", None),
            ("en-GB", "en", None),
        )
        for raw_language, expected_layer0_language, expected_translation_language in cases:
            with self.subTest(language=raw_language):
                self.detector.detect.return_value = self._result(
                    "detected", raw_language, raw_language
                )
                yielded, mocks = self._run_with_pipeline_mocks(
                    lambda message, lang: (True, "")
                )

                self.assertEqual(
                    yielded,
                    ["translated" if expected_translation_language else "complete response"],
                )
                self.detector.detect.assert_called_once_with(self.original_message)
                mocks["layer0"].assert_called_once_with(
                    self.original_message, lang=expected_layer0_language
                )
                if expected_translation_language:
                    self.assertEqual(
                        mocks["translate"].call_args_list,
                        [
                            call(
                                self.original_message,
                                source_lang=expected_translation_language,
                                target_lang="en",
                            ),
                            call(
                                "complete response",
                                source_lang="en",
                                target_lang="vi",
                            ),
                        ],
                    )
                    mocks["layer1_text"].assert_called_once_with("translated")
                else:
                    mocks["translate"].assert_not_called()
                    mocks["layer1_text"].assert_called_once_with(self.original_message)
                self.detector.detect.reset_mock()

    def test_unsupported_languages_never_translate_or_enter_qwen_pipeline(self):
        for language in ("ja", "ko", "zh", "zh-Hant", "fr", "ja-JP", "ko-KR", "zh-CN", "fr-CA"):
            with self.subTest(language=language):
                self.detector.detect.return_value = self._result("detected", language, language)
                yielded, mocks = self._run_with_pipeline_mocks(
                    lambda message, lang: (True, "")
                )

                self.assertEqual(yielded, [_UNSUPPORTED_LANGUAGE_RESPONSE])
                self.assertEqual(
                    mocks["layer0"].call_args_list,
                    [
                        call(self.original_message, lang="en"),
                        call(self.original_message, lang="vi"),
                    ],
                )
                for name in (
                    "translate", "layer1_text", "layer1_audio", "layer2_get", "layer3",
                    "layer4", "layer5", "layer2_save",
                ):
                    mocks[name].assert_not_called()

    def test_none_detection_runs_full_scan_and_preserves_input(self):
        self.detector.detect.return_value = None

        yielded, mocks = self._run_with_pipeline_mocks(lambda message, lang: (True, ""))

        self.assertEqual(yielded, [_LANGUAGE_CLARIFICATION_RESPONSE])
        self.assertEqual(
            mocks["layer0"].call_args_list,
            [
                call(self.original_message, lang="en"),
                call(self.original_message, lang="vi"),
            ],
        )
        mocks["translate"].assert_not_called()
        mocks["layer1_text"].assert_not_called()
        mocks["layer2_save"].assert_not_called()

    @patch("app.core.orchestrator.layer2_dialogue.save_chat_turn")
    @patch("app.core.orchestrator.layer5_filter.filter_output")
    @patch("app.core.orchestrator.layer4_llm.generate_response_stream")
    @patch("app.core.orchestrator.layer3_rag.retrieve_interventions")
    @patch("app.core.orchestrator.layer2_dialogue.get_chat_history")
    @patch("app.core.orchestrator.layer1_nlu.analyze_text_emotion")
    @patch("app.core.orchestrator.layer0_safety.check_input_risk")
    def test_detection_precedes_layer0_and_safe_path_keeps_original_message(
        self, mock_l0, mock_l1, mock_l2_get, mock_l3, mock_l4, mock_l5, mock_l2_save
    ):
        self.detector.detect.return_value = self._result("detected", "en", "en")
        call_order = []
        self.detector.detect.side_effect = lambda message: (
            call_order.append(("detect", message))
            or self._result("detected", "en", "en")
        )
        mock_l0.side_effect = lambda message, *, lang: (
            call_order.append(("layer0", message, lang)) or (True, "")
        )
        mock_l1.return_value = "neutral"
        mock_l2_get.return_value = ""
        mock_l3.return_value = "context"
        mock_l4.return_value = iter(["complete response"])
        mock_l5.return_value = "complete response"

        yielded = list(chat_orchestrator_stream(self.session_id, self.original_message))

        self.assertEqual(call_order[:2], [
            ("detect", self.original_message),
            ("layer0", self.original_message, "en"),
        ])
        mock_l1.assert_called_once_with(self.original_message)
        mock_l3.assert_called_once_with(self.original_message, current_emotion="neutral")
        mock_l4.assert_called_once_with(f"\nUser: {self.original_message}", "context")
        mock_l2_save.assert_called_once_with(
            self.session_id, self.original_message, "complete response"
        )
        self.assertEqual(yielded, ["complete response"])

    @patch("app.core.orchestrator.layer2_dialogue.save_chat_turn")
    @patch("app.core.orchestrator.layer5_filter.filter_output")
    @patch("app.core.orchestrator.layer4_llm.generate_response_stream")
    @patch("app.core.orchestrator.layer3_rag.retrieve_interventions")
    @patch("app.core.orchestrator.layer2_dialogue.get_chat_history")
    @patch("app.core.orchestrator.layer1_nlu.analyze_audio_emotion")
    @patch("app.core.orchestrator.layer1_nlu.analyze_text_emotion")
    @patch("app.core.orchestrator.layer0_safety.check_input_risk")
    def test_safe_audio_path_preserves_existing_audio_emotion_override(
        self, mock_l0, mock_l1_text, mock_l1_audio, mock_l2_get, mock_l3, mock_l4,
        mock_l5, mock_l2_save,
    ):
        self.detector.detect.return_value = self._result("detected", "en", "en")
        mock_l0.return_value = (True, "")
        mock_l1_text.return_value = "neutral"
        mock_l1_audio.return_value = "anxiety"
        mock_l2_get.return_value = ""
        mock_l3.return_value = "context"
        mock_l4.return_value = iter(["complete response"])
        mock_l5.return_value = "complete response"

        list(chat_orchestrator_stream(
            self.session_id, self.original_message, audio_file="voice.wav"
        ))

        mock_l0.assert_called_once_with(self.original_message, lang="en")
        mock_l1_text.assert_called_once_with(self.original_message)
        mock_l1_audio.assert_called_once_with("voice.wav")
        mock_l3.assert_called_once_with(self.original_message, current_emotion="anxiety")

    @patch("app.core.orchestrator.layer2_dialogue.save_chat_turn")
    @patch("app.core.orchestrator.layer5_filter.filter_output")
    @patch("app.core.orchestrator.layer4_llm.generate_response_stream")
    @patch("app.core.orchestrator.layer3_rag.retrieve_interventions")
    @patch("app.core.orchestrator.layer2_dialogue.get_chat_history")
    @patch("app.core.orchestrator.layer1_nlu.analyze_audio_emotion")
    @patch("app.core.orchestrator.layer1_nlu.analyze_text_emotion")
    @patch("app.core.orchestrator.layer0_safety.check_input_risk")
    def test_vietnamese_crisis_stops_downstream_layers(
        self, mock_l0, mock_l1_text, mock_l1_audio, mock_l2_get, mock_l3, mock_l4,
        mock_l5, mock_l2_save,
    ):
        self.detector.detect.return_value = self._result("detected", "vi", "vi")
        mock_l0.return_value = (False, "Vietnamese emergency")

        yielded = list(chat_orchestrator_stream(
            self.session_id, self.original_message, audio_file="voice.wav"
        ))

        self.assertEqual(yielded, ["Vietnamese emergency"])
        mock_l0.assert_called_once_with(self.original_message, lang="vi")
        mock_l1_text.assert_not_called()
        mock_l1_audio.assert_not_called()
        mock_l2_get.assert_not_called()
        mock_l3.assert_not_called()
        mock_l4.assert_not_called()
        mock_l5.assert_not_called()
        mock_l2_save.assert_not_called()


class TestOrchestratorInputTranslation(unittest.TestCase):
    """Input-only translation tests; all FastText and M2M100 calls are mocked."""

    def setUp(self):
        self.session_id = "translation-test-session"
        self.detector = MagicMock()
        self.get_detector = patch(
            "app.core.orchestrator.get_language_detector", return_value=self.detector
        )
        self.get_detector.start()
        self.addCleanup(self.get_detector.stop)

    @staticmethod
    def _result(status, language=None, candidate_language=None):
        return SimpleNamespace(
            status=status,
            language=language,
            candidate_language=candidate_language,
        )

    @patch("app.core.orchestrator.layer2_dialogue.save_chat_turn")
    @patch("app.core.orchestrator.layer5_filter.filter_output")
    @patch("app.core.orchestrator.layer4_llm.generate_response_stream")
    @patch("app.core.orchestrator.layer3_rag.retrieve_interventions")
    @patch("app.core.orchestrator.layer2_dialogue.get_chat_history")
    @patch("app.core.orchestrator.layer1_nlu.analyze_text_emotion")
    @patch("app.core.orchestrator.layer0_safety.check_input_risk")
    @patch("app.core.orchestrator.translate")
    def test_vietnamese_input_and_output_are_translated_and_history_stays_canonical(
        self, mock_translate, mock_l0, mock_l1, mock_l2_get, mock_l3, mock_l4,
        mock_l5, mock_l2_save,
    ):
        original = "Tôi cần hỗ trợ"
        pipeline_message = "I need support"
        translated_response = "Tôi luôn sẵn sàng lắng nghe bạn."
        output_response = "I am here to listen."
        mock_l0.return_value = (True, "")
        mock_l1.return_value = "neutral"
        mock_l2_get.return_value = ""
        mock_l3.return_value = "context"
        mock_l4.return_value = iter(["complete English response"])
        mock_l5.side_effect = lambda text, lang: (
            "filtered English response" if lang == "en" else text
        )
        mock_translate.side_effect = [
            pipeline_message,
            translated_response,
        ]
        self.detector.detect.return_value = self._result("detected", "vi", "vi")

        yielded = list(chat_orchestrator_stream(self.session_id, original))

        self.detector.detect.assert_called_once_with(original)
        mock_l0.assert_called_once_with(original, lang="vi")
        self.assertEqual(
            mock_translate.call_args_list,
            [
                call(original, source_lang="vi", target_lang="en"),
                call("filtered English response", source_lang="en", target_lang="vi"),
            ],
        )
        mock_l1.assert_called_once_with(pipeline_message)
        mock_l3.assert_called_once_with(pipeline_message, current_emotion="neutral")
        mock_l4.assert_called_once_with(
            f"\nUser: {pipeline_message}", "context"
        )
        self.assertEqual(
            mock_l5.call_args_list,
            [
                call("complete English response", lang="en"),
                call(translated_response, lang="vi"),
            ],
        )
        mock_l2_save.assert_called_once_with(
            self.session_id, pipeline_message, "filtered English response"
        )
        self.assertEqual(yielded, [translated_response])

    @patch("app.core.orchestrator.layer2_dialogue.save_chat_turn")
    @patch("app.core.orchestrator.layer5_filter.filter_output")
    @patch("app.core.orchestrator.layer4_llm.generate_response_stream")
    @patch("app.core.orchestrator.layer3_rag.retrieve_interventions")
    @patch("app.core.orchestrator.layer2_dialogue.get_chat_history")
    @patch("app.core.orchestrator.layer1_nlu.analyze_text_emotion")
    @patch("app.core.orchestrator.layer0_safety.check_input_risk")
    @patch("app.core.orchestrator.translate")
    def test_confident_english_passthrough_does_not_translate(
        self, mock_translate, mock_l0, mock_l1, mock_l2_get, mock_l3, mock_l4,
        mock_l5, mock_l2_save,
    ):
        original = "I need support"
        self.detector.detect.return_value = self._result("detected", "en", "en")
        mock_l0.return_value = (True, "")
        mock_l1.return_value = "neutral"
        mock_l2_get.return_value = ""
        mock_l3.return_value = "context"
        mock_l4.return_value = iter(["complete response"])
        mock_l5.return_value = "filtered response"

        list(chat_orchestrator_stream(self.session_id, original))

        mock_translate.assert_not_called()
        mock_l1.assert_called_once_with(original)
        mock_l3.assert_called_once_with(original, current_emotion="neutral")
        mock_l4.assert_called_once_with(f"\nUser: {original}", "context")
        mock_l2_save.assert_called_once_with(self.session_id, original, "filtered response")

    @patch("app.core.orchestrator.layer2_dialogue.save_chat_turn")
    @patch("app.core.orchestrator.layer5_filter.filter_output")
    @patch("app.core.orchestrator.layer4_llm.generate_response_stream")
    @patch("app.core.orchestrator.layer3_rag.retrieve_interventions")
    @patch("app.core.orchestrator.layer2_dialogue.get_chat_history")
    @patch("app.core.orchestrator.layer1_nlu.analyze_text_emotion")
    @patch("app.core.orchestrator.layer0_safety.check_input_risk")
    @patch("app.core.orchestrator.translate")
    def test_unresolved_languages_do_not_translate_or_use_candidate_as_source(
        self, mock_translate, mock_l0, mock_l1, mock_l2_get, mock_l3, mock_l4,
        mock_l5, mock_l2_save,
    ):
        original = "Short or unsupported input"
        cases = (
            self._result("uncertain", candidate_language="vi"),
            self._result("uncertain", candidate_language="ja"),
            self._result("uncertain"),
            self._result("unknown"),
        )
        mock_l0.return_value = (True, "")
        mock_l1.return_value = "neutral"
        mock_l2_get.return_value = ""
        mock_l3.return_value = "context"
        mock_l4.return_value = iter(["complete response"])
        mock_l5.return_value = "filtered response"

        for detection in cases:
            with self.subTest(status=detection.status, language=detection.language):
                self.detector.detect.return_value = detection
                mock_l4.return_value = iter(["complete response"])
                yielded = list(chat_orchestrator_stream(self.session_id, original))

                self.assertEqual(
                    mock_l0.call_args_list,
                    [
                        call(original, lang="en"),
                        call(original, lang="vi"),
                    ],
                )
                self.assertEqual(yielded, [_LANGUAGE_CLARIFICATION_RESPONSE])
                self.detector.detect.assert_called_once_with(original)
                mock_translate.assert_not_called()
                mock_l1.assert_not_called()
                mock_l2_get.assert_not_called()
                mock_l2_save.assert_not_called()
                for mock in (
                    self.detector.detect, mock_l0, mock_l1, mock_l2_get, mock_l3,
                    mock_l4, mock_l5, mock_l2_save,
                ):
                    mock.reset_mock()

    @patch("app.core.orchestrator.layer2_dialogue.save_chat_turn")
    @patch("app.core.orchestrator.layer5_filter.filter_output")
    @patch("app.core.orchestrator.layer4_llm.generate_response_stream")
    @patch("app.core.orchestrator.layer3_rag.retrieve_interventions")
    @patch("app.core.orchestrator.layer2_dialogue.get_chat_history")
    @patch("app.core.orchestrator.layer1_nlu.analyze_text_emotion")
    @patch("app.core.orchestrator.layer0_safety.check_input_risk")
    @patch("app.core.orchestrator.translate")
    def test_translation_failures_stop_pipeline_with_localized_response(
        self, mock_translate, mock_l0, mock_l1, mock_l2_get, mock_l3, mock_l4,
        mock_l5, mock_l2_save,
    ):
        original = "Tôi cần hỗ trợ"
        self.detector.detect.return_value = self._result("detected", "vi", "vi")
        mock_l0.return_value = (True, "")

        for outcome in (TranslationError("model unavailable"), RuntimeError("inference failed"), "   "):
            with self.subTest(outcome=type(outcome).__name__):
                if isinstance(outcome, BaseException):
                    mock_translate.side_effect = outcome
                else:
                    mock_translate.side_effect = None
                    mock_translate.return_value = outcome

                yielded = list(chat_orchestrator_stream(self.session_id, original))

                self.assertEqual(yielded, [
                    "Xin lỗi, chúng tôi chưa thể xử lý tin nhắn của bạn lúc này. Vui lòng thử lại sau."
                ])
                mock_l0.assert_called_once_with(original, lang="vi")
                mock_l1.assert_not_called()
                mock_l2_get.assert_not_called()
                mock_l3.assert_not_called()
                mock_l4.assert_not_called()
                mock_l5.assert_not_called()
                mock_l2_save.assert_not_called()
                for mock in (mock_l0, mock_translate, mock_l1, mock_l2_get, mock_l3, mock_l4, mock_l5, mock_l2_save):
                    mock.reset_mock()

    @patch("app.core.orchestrator.layer2_dialogue.save_chat_turn")
    @patch("app.core.orchestrator.layer5_filter.filter_output")
    @patch("app.core.orchestrator.layer4_llm.generate_response_stream")
    @patch("app.core.orchestrator.layer3_rag.retrieve_interventions")
    @patch("app.core.orchestrator.layer2_dialogue.get_chat_history")
    @patch("app.core.orchestrator.layer1_nlu.analyze_text_emotion")
    @patch("app.core.orchestrator.layer0_safety.check_input_risk")
    @patch("app.core.orchestrator.translate")
    def test_unsupported_japanese_does_not_translate_or_call_pipeline(
        self, mock_translate, mock_l0, mock_l1, mock_l2_get, mock_l3, mock_l4,
        mock_l5, mock_l2_save,
    ):
        original = "私は死にたい"
        self.detector.detect.return_value = self._result("detected", "ja", "ja")
        mock_l0.return_value = (True, "")

        yielded = list(chat_orchestrator_stream(self.session_id, original))

        self.assertEqual(yielded, [_UNSUPPORTED_LANGUAGE_RESPONSE])
        self.assertEqual(
            mock_l0.call_args_list,
            [call(original, lang="en"), call(original, lang="vi")],
        )
        mock_translate.assert_not_called()
        mock_l1.assert_not_called()
        mock_l2_get.assert_not_called()
        mock_l3.assert_not_called()
        mock_l4.assert_not_called()
        mock_l5.assert_not_called()
        mock_l2_save.assert_not_called()


class TestOrchestratorOutputTranslation(unittest.TestCase):
    def _run_request(
        self,
        *,
        language="vi",
        qwen_chunks=("Complete English response.",),
        generation_exception=None,
        input_translation="English pipeline message",
        output_translation="Câu trả lời bằng tiếng Việt.",
        output_translation_exception=None,
        english_filter_result=None,
        english_filter_exception=None,
        vietnamese_filter_result=None,
        vietnamese_filter_exception=None,
        save_exception=None,
    ):
        events = []
        detector = MagicMock()
        detector.detect.return_value = SimpleNamespace(
            status="detected", language=language, candidate_language=language
        )

        with ExitStack() as stack:
            mocks = {
                "get_detector": stack.enter_context(
                    patch("app.core.orchestrator.get_language_detector", return_value=detector)
                ),
                "layer0": stack.enter_context(
                    patch("app.core.orchestrator.layer0_safety.check_input_risk", return_value=(True, ""))
                ),
                "translate": stack.enter_context(
                    patch("app.core.orchestrator.translate")
                ),
                "layer1_text": stack.enter_context(
                    patch("app.core.orchestrator.layer1_nlu.analyze_text_emotion", return_value="neutral")
                ),
                "layer2_get": stack.enter_context(
                    patch("app.core.orchestrator.layer2_dialogue.get_chat_history", return_value="")
                ),
                "layer2_save": stack.enter_context(
                    patch("app.core.orchestrator.layer2_dialogue.save_chat_turn")
                ),
                "layer3": stack.enter_context(
                    patch("app.core.orchestrator.layer3_rag.retrieve_interventions", return_value="context")
                ),
                "layer4": stack.enter_context(
                    patch("app.core.orchestrator.layer4_llm.generate_response_stream")
                ),
                "layer5": stack.enter_context(
                    patch("app.core.orchestrator.layer5_filter.filter_output")
                ),
            }

            def translation_side_effect(text, source_lang, target_lang):
                events.append(f"translate_{source_lang}_{target_lang}")
                if source_lang == "vi":
                    return input_translation
                if output_translation_exception is not None:
                    raise output_translation_exception
                return output_translation

            def qwen_side_effect(*args):
                events.append("qwen_start")

                def generate():
                    for chunk in qwen_chunks:
                        yield chunk
                    if generation_exception is not None:
                        raise generation_exception
                    events.append("qwen_complete")

                return generate()

            def filter_side_effect(response, lang):
                events.append(f"filter_{lang}")
                if lang == "en":
                    if english_filter_exception is not None:
                        raise english_filter_exception
                    return (
                        english_filter_result
                        if english_filter_result is not None
                        else response
                    )
                if vietnamese_filter_exception is not None:
                    raise vietnamese_filter_exception
                return (
                    vietnamese_filter_result
                    if vietnamese_filter_result is not None
                    else response
                )

            def save_side_effect(*args):
                events.append("save")
                if save_exception is not None:
                    raise save_exception

            mocks["translate"].side_effect = translation_side_effect
            mocks["layer4"].side_effect = qwen_side_effect
            mocks["layer5"].side_effect = filter_side_effect
            mocks["layer2_save"].side_effect = save_side_effect

            original_message = (
                "I need support." if language == "en" else "Tôi cần hỗ trợ."
            )
            response_stream = chat_orchestrator_stream(
                "output-translation-session", original_message
            )
            yielded = [next(response_stream)]
            events.append("yield")
            with self.assertRaises(StopIteration):
                next(response_stream)

        detector.detect.assert_called_once_with(original_message)
        return yielded, events, mocks, original_message

    def test_english_response_is_filtered_and_saved_without_translation(self):
        yielded, events, mocks, original = self._run_request(
            language="en",
            english_filter_result="Filtered English response.",
        )

        self.assertEqual(yielded, ["Filtered English response."])
        mocks["translate"].assert_not_called()
        mocks["layer5"].assert_called_once_with(
            "Complete English response.", lang="en"
        )
        mocks["layer2_save"].assert_called_once_with(
            "output-translation-session", original, "Filtered English response."
        )
        self.assertEqual(
            events,
            ["qwen_start", "qwen_complete", "filter_en", "save", "yield"],
        )

    def test_vietnamese_output_order_and_english_canonical_history(self):
        yielded, events, mocks, _ = self._run_request(
            qwen_chunks=("Complete English ", "response."),
            english_filter_result="English Layer 5 response.",
        )

        self.assertEqual(yielded, ["Câu trả lời bằng tiếng Việt."])
        self.assertEqual(
            mocks["translate"].call_args_list,
            [
                call("Tôi cần hỗ trợ.", source_lang="vi", target_lang="en"),
                call(
                    "English Layer 5 response.",
                    source_lang="en",
                    target_lang="vi",
                ),
            ],
        )
        self.assertEqual(
            mocks["layer5"].call_args_list,
            [
                call("Complete English response.", lang="en"),
                call("Câu trả lời bằng tiếng Việt.", lang="vi"),
            ],
        )
        mocks["layer2_save"].assert_called_once_with(
            "output-translation-session",
            "English pipeline message",
            "English Layer 5 response.",
        )
        self.assertEqual(
            events,
            [
                "translate_vi_en",
                "qwen_start",
                "qwen_complete",
                "filter_en",
                "translate_en_vi",
                "filter_vi",
                "save",
                "yield",
            ],
        )

    def test_english_filter_failure_is_fail_closed_without_translation_or_history(self):
        for language, expected in (
            ("en", SAFE_FALLBACK_RESPONSE),
            ("vi", VIETNAMESE_SAFE_FALLBACK_RESPONSE),
        ):
            with self.subTest(language=language):
                yielded, _, mocks, _ = self._run_request(
                    language=language,
                    english_filter_exception=RuntimeError("private filter details"),
                )
                self.assertEqual(yielded, [expected])
                self.assertEqual(len(yielded), 1)
                mocks["layer5"].assert_called_once_with(
                    "Complete English response.", lang="en"
                )
                if language == "en":
                    mocks["translate"].assert_not_called()
                else:
                    mocks["translate"].assert_called_once_with(
                        "Tôi cần hỗ trợ.", source_lang="vi", target_lang="en"
                    )
                mocks["layer2_save"].assert_not_called()

    def test_output_translation_exceptions_and_empty_results_fail_closed(self):
        cases = (
            (TranslationError("private translation details"), "unused"),
            (RuntimeError("private translation details"), "unused"),
            (None, ""),
            (None, "   "),
            (None, "Complete English response."),
        )
        for error, output in cases:
            with self.subTest(error=type(error).__name__, output=repr(output)):
                yielded, _, mocks, _ = self._run_request(
                    output_translation=output,
                    output_translation_exception=error,
                )
                self.assertEqual(
                    yielded, [VIETNAMESE_OUTPUT_TRANSLATION_FAILURE_RESPONSE]
                )
                self.assertEqual(len(yielded), 1)
                self.assertNotIn("Complete English response.", yielded[0])
                self.assertNotIn("private translation details", yielded[0])
                mocks["layer5"].assert_called_once_with(
                    "Complete English response.", lang="en"
                )
                mocks["layer2_save"].assert_not_called()

    def test_vietnamese_validation_exception_fails_closed_without_history(self):
        yielded, _, mocks, _ = self._run_request(
            vietnamese_filter_exception=RuntimeError("private validation details")
        )

        self.assertEqual(
            yielded, [VIETNAMESE_OUTPUT_VALIDATION_FAILURE_RESPONSE]
        )
        self.assertNotIn("Complete English response.", yielded[0])
        mocks["layer2_save"].assert_not_called()

    def test_vietnamese_filter_replacement_is_returned_without_normal_history(self):
        replacement = "Xin lỗi, tôi không thể chẩn đoán hoặc kê đơn điều trị."
        yielded, _, mocks, _ = self._run_request(
            vietnamese_filter_result=replacement,
        )

        self.assertEqual(yielded, [replacement])
        mocks["layer2_save"].assert_not_called()

    def test_generation_failure_before_or_after_chunks_returns_localized_fallback(self):
        for chunks in ((), ("Partial Qwen content.",)):
            with self.subTest(chunks=chunks):
                yielded, _, mocks, _ = self._run_request(
                    qwen_chunks=chunks,
                    generation_exception=ConnectionError("private model details"),
                )
                self.assertEqual(yielded, [VIETNAMESE_SAFE_FALLBACK_RESPONSE])
                self.assertEqual(len(yielded), 1)
                self.assertNotIn("Partial Qwen content.", yielded[0])
                self.assertNotIn("private model details", yielded[0])
                mocks["layer5"].assert_not_called()
                mocks["translate"].assert_called_once_with(
                    "Tôi cần hỗ trợ.", source_lang="vi", target_lang="en"
                )
                mocks["layer2_save"].assert_called_once_with(
                    "output-translation-session",
                    "English pipeline message",
                    SAFE_FALLBACK_RESPONSE,
                )

    def test_empty_qwen_response_uses_localized_fallback_without_output_translation(self):
        yielded, _, mocks, _ = self._run_request(qwen_chunks=())

        self.assertEqual(yielded, [VIETNAMESE_SAFE_FALLBACK_RESPONSE])
        mocks["layer5"].assert_not_called()
        mocks["translate"].assert_called_once_with(
            "Tôi cần hỗ trợ.", source_lang="vi", target_lang="en"
        )
        mocks["layer2_save"].assert_called_once_with(
            "output-translation-session",
            "English pipeline message",
            SAFE_FALLBACK_RESPONSE,
        )

    def test_history_save_failure_does_not_suppress_validated_vietnamese_output(self):
        yielded, _, mocks, _ = self._run_request(
            save_exception=RuntimeError("private history details")
        )

        self.assertEqual(yielded, ["Câu trả lời bằng tiếng Việt."])
        mocks["layer2_save"].assert_called_once()


if __name__ == "__main__":
    unittest.main()
