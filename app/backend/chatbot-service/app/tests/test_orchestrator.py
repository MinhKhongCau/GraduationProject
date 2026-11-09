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
from unittest.mock import MagicMock, patch

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
    from app.core.orchestrator import SAFE_FALLBACK_RESPONSE, chat_orchestrator_stream


class TestOrchestratorStreamingSafety(unittest.TestCase):
    """Test suite for orchestrator streaming safety and Layer 5 enforcement."""

    def setUp(self):
        self.session_id = "test-session-456"
        self.user_message = "I have been feeling stressed about exams."

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
        mock_l5_filter.side_effect = lambda text: text  # Identity filter

        generator = chat_orchestrator_stream(self.session_id, self.user_message)
        yielded_chunks = list(generator)

        # Must yield exactly 1 consolidated block, not 3 individual chunks
        self.assertEqual(len(yielded_chunks), 1)
        expected_full = "It is normal to feel exam stress. Try deep breathing."
        self.assertEqual(yielded_chunks[0], expected_full)

        # Layer 5 must have received the full consolidated text
        mock_l5_filter.assert_called_once_with(expected_full)

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

        def tracked_filter(text):
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

        # Verify history saved the fallback, NOT the partial chunk
        mock_l2_save.assert_called_once_with(self.session_id, self.user_message, SAFE_FALLBACK_RESPONSE)

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

        # Verify raw output is not in history
        mock_l2_save.assert_called_once_with(self.session_id, self.user_message, SAFE_FALLBACK_RESPONSE)

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


if __name__ == "__main__":
    unittest.main()
