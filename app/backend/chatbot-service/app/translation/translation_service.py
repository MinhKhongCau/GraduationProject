"""Translation service – lazy, thread-safe, token-aware M2M100 inference.

Implements M2M100 418M translation module for chatbot-service:
- Lazy model loading (loaded once per process on demand).
- Thread-safe loading and inference via separate _load_lock and _inference_lock.
- Local model validation ensuring all required files exist and are non-empty.
- local_files_only=True during tokenizer and model loading.
- CPU inference wrapped in torch.inference_mode() with do_sample=False.
- Token-aware chunking based on actual tokenizer encoding and model max position limits.
- Preservation of paragraphs and language-specific sentence boundaries.
- No raw mental-health or user text logged.
"""

from __future__ import annotations

import os
import re
import threading
from typing import List

import torch
from transformers import M2M100ForConditionalGeneration, M2M100Tokenizer

# ----------------------------------------------------------------------
# Public exception
# ----------------------------------------------------------------------
class TranslationError(RuntimeError):
    """Base exception for all translation service errors."""


# ----------------------------------------------------------------------
# Constants
# ----------------------------------------------------------------------
_SUPPORTED_LANGUAGES = {"vi", "en", "ja", "ko", "zh"}

_MODEL_PATH = os.getenv(
    "TRANSLATION_MODEL_PATH", "/models/translation/m2m100_418M"
)

_REQUIRED_FILES = {
    "config.json",
    "generation_config.json",
    "pytorch_model.bin",
    "sentencepiece.bpe.model",
    "special_tokens_map.json",
    "tokenizer_config.json",
    "vocab.json",
}

# ----------------------------------------------------------------------
# Global state (lazy-loaded)
# ----------------------------------------------------------------------
_model: M2M100ForConditionalGeneration | None = None
_tokenizer: M2M100Tokenizer | None = None
_load_lock = threading.Lock()
_inference_lock = threading.Lock()


# ----------------------------------------------------------------------
# Validation helper
# ----------------------------------------------------------------------
def _model_is_ready() -> bool:
    """Check if the translation model directory contains all required non-empty files."""
    if not os.path.isdir(_MODEL_PATH):
        return False
    for filename in _REQUIRED_FILES:
        filepath = os.path.join(_MODEL_PATH, filename)
        if not os.path.isfile(filepath) or os.path.getsize(filepath) <= 0:
            return False
    return True


# ----------------------------------------------------------------------
# Lazy loader
# ----------------------------------------------------------------------
def _load() -> None:
    """Load model and tokenizer lazily and thread-safely.

    Assigns to global variables only after both have successfully loaded
    to prevent partially initialized states.
    """
    global _model, _tokenizer
    with _load_lock:
        if _model is not None and _tokenizer is not None:
            return

        if not _model_is_ready():
            raise TranslationError(
                f"Translation model not found or incomplete at {_MODEL_PATH}. "
                "Run scripts/setup_translation_model.sh to populate the volume."
            )

        try:
            tokenizer_tmp = M2M100Tokenizer.from_pretrained(
                _MODEL_PATH, local_files_only=True
            )
            model_tmp = M2M100ForConditionalGeneration.from_pretrained(
                _MODEL_PATH, local_files_only=True
            )
            model_tmp.eval()
        except Exception as exc:
            raise TranslationError(f"Failed to initialise translation model: {exc}") from exc

        _tokenizer = tokenizer_tmp
        _model = model_tmp


# ----------------------------------------------------------------------
# Language-specific punctuation splitting
# ----------------------------------------------------------------------
_SENTENCE_PATTERNS = {
    "en": r"[.!?]+",
    "vi": r"[.!?]+",
    "ja": r"[。！？]+",
    "ko": r"[.!?。！？]+",
    "zh": r"[。！？]+",
}


def _split_into_sentences(text: str, lang: str) -> List[str]:
    """Split text into sentences while preserving sentence-ending punctuation."""
    pattern = _SENTENCE_PATTERNS.get(lang, r"[.!?]+")
    parts = re.split(f"({pattern})", text)
    sentences: List[str] = []
    buffer = ""
    for part in parts:
        buffer += part
        if re.search(pattern, part):
            stripped = buffer.strip()
            if stripped:
                sentences.append(stripped)
            buffer = ""
    remainder = buffer.strip()
    if remainder:
        sentences.append(remainder)
    return sentences if sentences else ([text.strip()] if text.strip() else [])


# ----------------------------------------------------------------------
# Token-aware chunking
# ----------------------------------------------------------------------
def _chunk_paragraph(paragraph: str, src_lang: str) -> List[str]:
    """Chunk a paragraph into token-safe slices using the loaded tokenizer."""
    if _model is None or _tokenizer is None:
        raise TranslationError("Model and tokenizer must be loaded before chunking.")

    max_position = getattr(_model.config, "max_position_embeddings", 1024)
    # Reserve safety allowance for BOS/EOS, language tokens, and generation headroom
    safety_margin = 16
    limit = max_position - safety_margin

    sentences = _split_into_sentences(paragraph, src_lang)
    if not sentences:
        return []

    chunks: List[str] = []
    current_chunk = ""

    for sentence in sentences:
        candidate = f"{current_chunk} {sentence}".strip() if current_chunk else sentence
        candidate_tokens = len(_tokenizer.encode(candidate, add_special_tokens=True))

        if candidate_tokens <= limit:
            current_chunk = candidate
        else:
            if not current_chunk:
                # Single sentence itself exceeds the token limit
                raise TranslationError(
                    f"A single sentence exceeds the maximum token limit ({limit} tokens)."
                )
            chunks.append(current_chunk)

            # Check if the sentence alone fits into the next chunk
            single_tokens = len(_tokenizer.encode(sentence, add_special_tokens=True))
            if single_tokens > limit:
                raise TranslationError(
                    f"A single sentence exceeds the maximum token limit ({limit} tokens)."
                )
            current_chunk = sentence

    if current_chunk:
        chunks.append(current_chunk)

    return chunks


# ----------------------------------------------------------------------
# Public API
# ----------------------------------------------------------------------
def translate(text: str, source_lang: str, target_lang: str) -> str:
    """Translate text from source_lang to target_lang.

    - Thread-safe for multi-threaded inference.
    - Validates source and target languages.
    - Passthrough without model load when source == target.
    - Preserves paragraphs and handles empty/whitespace inputs.
    - Raises TranslationError on failure without logging raw user text.
    """
    src = (source_lang or "").strip().lower()
    tgt = (target_lang or "").strip().lower()

    if src not in _SUPPORTED_LANGUAGES:
        raise TranslationError(f"Unsupported source language: {src}")
    if tgt not in _SUPPORTED_LANGUAGES:
        raise TranslationError(f"Unsupported target language: {tgt}")

    # Passthrough for identical languages (avoids unnecessary model loading)
    if src == tgt:
        return text

    # Safely handle empty or whitespace-only text
    if not text or not text.strip():
        return text

    # Ensure model is ready
    if _model is None or _tokenizer is None:
        _load()

    # The entire inference pipeline is locked to ensure src_lang state is not contaminated
    with _inference_lock:
        _tokenizer.src_lang = src

        # Preserve paragraph structure by splitting on newlines
        paragraphs = text.split("\n")
        translated_paragraphs: List[str] = []

        for para in paragraphs:
            if not para.strip():
                # Preserve empty or whitespace line
                translated_paragraphs.append(para)
                continue

            chunks = _chunk_paragraph(para, src)
            translated_chunks: List[str] = []

            for chunk in chunks:
                try:
                    inputs = _tokenizer(chunk, return_tensors="pt", padding=True)
                    target_lang_id = _tokenizer.get_lang_id(tgt)

                    with torch.inference_mode():
                        generated_tokens = _model.generate(
                            **inputs,
                            forced_bos_token_id=target_lang_id,
                            do_sample=False,
                        )

                    decoded = _tokenizer.batch_decode(
                        generated_tokens, skip_special_tokens=True
                    )[0]
                    translated_chunks.append(decoded.strip())
                except Exception as exc:
                    raise TranslationError(f"Inference failure: {exc}") from exc

            translated_paragraphs.append(" ".join(translated_chunks))

        return "\n".join(translated_paragraphs)
