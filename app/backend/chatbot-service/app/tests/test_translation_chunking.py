# app/tests/test_translation_chunking.py
"""Tests for token-aware chunking and punctuation handling in translation service.

Verifies:
- Sentence-aware splitting on language-specific punctuation (vi, en, ja, zh, ko).
- Dynamic token counting using tokenizer.encode() rather than character counts.
- Enforcement of model max position embeddings limit with safety margin.
- Rejection of oversized single sentences (raises TranslationError).
- Paragraph boundary preservation across multiple lines.
- Preservation of input order and no silent truncation.

Runs cleanly with:
    python3 -m unittest app/tests/test_translation_chunking.py
    pytest app/tests/test_translation_chunking.py
"""

from __future__ import annotations

import unittest
from unittest.mock import MagicMock, patch

import app.translation.translation_service as svc
from app.translation.translation_service import (
    TranslationError,
    _chunk_paragraph,
    _split_into_sentences,
    translate,
)


class TestTranslationChunking(unittest.TestCase):

    def setUp(self):
        svc._model = None
        svc._tokenizer = None

    def tearDown(self):
        svc._model = None
        svc._tokenizer = None

    def test_split_into_sentences_english_vietnamese(self):
        """English and Vietnamese sentence punctuation splitting (. ! ?)."""
        text = "Hello doctor! I feel anxious. Can we talk?"
        sentences = _split_into_sentences(text, "en")
        self.assertEqual(sentences, ["Hello doctor!", "I feel anxious.", "Can we talk?"])

        vi_text = "Chào bác sĩ! Tôi thấy mệt mỏi. Liệu tôi có ổn không?"
        vi_sentences = _split_into_sentences(vi_text, "vi")
        self.assertEqual(vi_sentences, ["Chào bác sĩ!", "Tôi thấy mệt mỏi.", "Liệu tôi có ổn không?"])

    def test_split_into_sentences_japanese(self):
        """Japanese punctuation splitting (。 ！？)."""
        ja_text = "こんにちは。最近よく眠れません！どうすればいいですか？"
        ja_sentences = _split_into_sentences(ja_text, "ja")
        self.assertEqual(ja_sentences, ["こんにちは。", "最近よく眠れません！", "どうすればいいですか？"])

    def test_split_into_sentences_chinese(self):
        """Chinese punctuation splitting (。 ！？)."""
        zh_text = "您好。我感到很有压力！你能帮助我吗？"
        zh_sentences = _split_into_sentences(zh_text, "zh")
        self.assertEqual(zh_sentences, ["您好。", "我感到很有压力！", "你能帮助我吗？"])

    def test_split_into_sentences_korean(self):
        """Korean punctuation splitting (. ! ? 。 ！？)."""
        ko_text = "안녕하세요. 요즘 너무 불안해요! 어떻게 해야 할까요?"
        ko_sentences = _split_into_sentences(ko_text, "ko")
        self.assertEqual(ko_sentences, ["안녕하세요.", "요즘 너무 불안해요!", "어떻게 해야 할까요?"])

    def test_token_aware_chunking_merging(self):
        """Sentences are merged into chunks until the token limit is reached."""
        mock_tok = MagicMock()
        mock_model = MagicMock()
        mock_model.config.max_position_embeddings = 100
        # limit will be 100 - 16 = 84

        def fake_encode(text, add_special_tokens=True):
            # 1 word ~ 10 tokens
            word_count = len(text.split())
            return [1] * (word_count * 10)

        mock_tok.encode = MagicMock(side_effect=fake_encode)
        svc._model = mock_model
        svc._tokenizer = mock_tok

        # sentence 1: 3 words -> 30 tokens
        # sentence 2: 3 words -> combined 6 words -> 60 tokens (fits <= 84)
        # sentence 3: 4 words -> combined 10 words -> 100 tokens (> 84, starts new chunk)
        para = "One two three. Four five six. Seven eight nine ten."
        chunks = _chunk_paragraph(para, "en")

        self.assertEqual(len(chunks), 2)
        self.assertEqual(chunks[0], "One two three. Four five six.")
        self.assertEqual(chunks[1], "Seven eight nine ten.")

    def test_oversized_single_sentence_rejected(self):
        """A single sentence exceeding the token limit must raise TranslationError."""
        mock_tok = MagicMock()
        mock_model = MagicMock()
        mock_model.config.max_position_embeddings = 50
        # limit is 50 - 16 = 34

        mock_tok.encode = MagicMock(return_value=[1] * 100)
        svc._model = mock_model
        svc._tokenizer = mock_tok

        with self.assertRaises(TranslationError) as ctx:
            _chunk_paragraph("This is an excessively long single sentence without any punctuation.", "en")
        self.assertIn("exceeds the maximum token limit", str(ctx.exception))

    def test_paragraph_preservation(self):
        """Multiple paragraphs separated by newlines must preserve structure."""
        mock_tok = MagicMock()
        mock_tok.src_lang = None
        mock_tok.get_lang_id = MagicMock(return_value=128001)
        mock_tok.encode = MagicMock(return_value=[1, 2, 3])

        captured_chunks = []
        def fake_call(text, **kw):
            captured_chunks.append(text)
            return {"input_ids": MagicMock()}

        mock_tok.side_effect = fake_call
        mock_tok.batch_decode = MagicMock(side_effect=lambda toks, **kw: [captured_chunks[-1].upper()])

        mock_model = MagicMock()
        mock_model.config.max_position_embeddings = 1024
        mock_model.eval = MagicMock(return_value=mock_model)
        mock_model.generate = MagicMock(return_value=MagicMock())

        mock_tok_cls = MagicMock()
        mock_tok_cls.from_pretrained.return_value = mock_tok
        mock_model_cls = MagicMock()
        mock_model_cls.from_pretrained.return_value = mock_model

        multiline_text = "Paragraph one is here.\n\nParagraph two is here."

        with patch.object(svc, "_model_is_ready", return_value=True), \
             patch.object(svc, "M2M100Tokenizer", mock_tok_cls), \
             patch.object(svc, "M2M100ForConditionalGeneration", mock_model_cls):

            result = translate(multiline_text, "en", "vi")

            lines = result.split("\n")
            self.assertEqual(len(lines), 3)
            self.assertEqual(lines[0], "PARAGRAPH ONE IS HERE.")
            self.assertEqual(lines[1], "")  # preserved blank line
            self.assertEqual(lines[2], "PARAGRAPH TWO IS HERE.")

    def test_no_silent_truncation_all_sentences_processed(self):
        """Ensure all sentences in a paragraph are accounted for in generated chunks."""
        mock_tok = MagicMock()
        mock_model = MagicMock()
        mock_model.config.max_position_embeddings = 50

        mock_tok.encode = MagicMock(side_effect=lambda text, **kw: [1] * (20 if len(text.split()) <= 2 else 40))
        svc._model = mock_model
        svc._tokenizer = mock_tok

        para = "Sentence one. Sentence two. Sentence three."
        chunks = _chunk_paragraph(para, "en")

        combined_chunks = " ".join(chunks)
        self.assertIn("Sentence one.", combined_chunks)
        self.assertIn("Sentence two.", combined_chunks)
        self.assertIn("Sentence three.", combined_chunks)


if __name__ == "__main__":
    unittest.main()
