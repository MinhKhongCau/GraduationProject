import unittest
from types import SimpleNamespace
from unittest.mock import patch

from app.layers import layer4_llm


def _chunk(content):
    return SimpleNamespace(
        choices=[SimpleNamespace(delta=SimpleNamespace(content=content))]
    )


class TestLayer4GenerationFailures(unittest.TestCase):
    def test_successful_generation_yields_all_chunks(self):
        response = [_chunk("A complete "), _chunk("response.")]

        with patch.object(
            layer4_llm.client.chat.completions, "create", return_value=response
        ):
            self.assertEqual(
                list(layer4_llm.generate_response_stream("User text", "Context")),
                ["A complete ", "response."],
            )

    def test_failure_before_first_chunk_propagates_without_content(self):
        def failing_response():
            raise RuntimeError("sensitive prompt details")
            yield "unreachable"

        with patch.object(
            layer4_llm.client.chat.completions,
            "create",
            return_value=failing_response(),
        ):
            with self.assertRaisesRegex(RuntimeError, "sensitive prompt details"):
                list(layer4_llm.generate_response_stream("Sensitive user text", "Context"))

    def test_failure_after_chunks_propagates_without_appending_error_text(self):
        def failing_response():
            yield _chunk("Partial response")
            raise ConnectionError("sensitive prompt details")

        with patch.object(
            layer4_llm.client.chat.completions,
            "create",
            return_value=failing_response(),
        ):
            stream = layer4_llm.generate_response_stream("Sensitive user text", "Context")
            self.assertEqual(next(stream), "Partial response")
            with self.assertRaisesRegex(ConnectionError, "sensitive prompt details"):
                next(stream)

    def test_error_log_contains_type_but_not_prompt_or_exception_message(self):
        sensitive_prompt = "SECRET_USER_PROMPT_123"

        with patch.object(
            layer4_llm.client.chat.completions,
            "create",
            side_effect=RuntimeError(sensitive_prompt),
        ):
            with self.assertLogs("app.layers.layer4_llm", level="ERROR") as captured:
                with self.assertRaises(RuntimeError):
                    list(layer4_llm.generate_response_stream(sensitive_prompt, "Context"))

        log_text = " ".join(captured.output)
        self.assertIn("RuntimeError", log_text)
        self.assertNotIn(sensitive_prompt, log_text)
        self.assertNotIn("sensitive", log_text.lower())


if __name__ == "__main__":
    unittest.main()
