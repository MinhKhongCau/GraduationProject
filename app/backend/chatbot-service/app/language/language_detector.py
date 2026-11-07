# app/language/language_detector.py
"""
Language Detection Module using FastText (lid.176.ftz).

Provides self-contained, offline language identification across 176 languages
with confidence scoring, multi-candidate alternatives, and clear status
categorization ('detected', 'uncertain', 'unknown').
"""

from __future__ import annotations

import logging
import os
import threading
import unicodedata
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any, List, Optional, Union

logger = logging.getLogger(__name__)

# Default configuration constants
DEFAULT_MIN_CONFIDENCE: float = 0.70
DEFAULT_TOP_K: int = 3
DEFAULT_MODEL_REL_PATH: Path = Path("models") / "language" / "lid.176.ftz"

# Detection status categories
STATUS_DETECTED: str = "detected"
STATUS_UNCERTAIN: str = "uncertain"
STATUS_UNKNOWN: str = "unknown"

# Project root directory: resolved relative to this file (<project_root>/app/language/language_detector.py)
PROJECT_ROOT: Path = Path(__file__).resolve().parent.parent.parent

# Common Vietnamese diacritics character set for supporting verification
VIETNAMESE_DIACRITICS_CHARS = frozenset(
    "àáảãạăằắẳẵặâầấẩẫậèéẻẽẹêềếểễệìíỉĩịòóỏõọôồốổỗộơờớởỡợùúủũụưừứửữựỳýỷỹỵđ"
    "ÀÁẢÃẠĂẰẮẲẴẶÂẦẤẨẪẬÈÉẺẼẸÊỀẾỂỄỆÌÍỈĨỊÒÓỎÕỌÔỒỐỔỖỘƠỜỚỞỠỢÙÚỦŨỤƯỪỨỬỮỰỲÝỶỸỴĐ"
)


class ModelNotFoundError(FileNotFoundError):
    """Raised when the FastText pretrained language identification model is missing."""
    pass


@dataclass(frozen=True)
class LanguageAlternative:
    """Alternative language candidate predicted by FastText."""
    language: str
    confidence: float

    def to_dict(self) -> dict[str, Any]:
        return {
            "language": self.language,
            "confidence": self.confidence,
        }


@dataclass(frozen=True)
class LanguageDetectionResult:
    """
    Structured result of language detection.

    Attributes:
        language: Standard ISO language code (e.g. 'vi', 'en', 'ja'), or None if
                  the language cannot be determined reliably (uncertain/unknown).
        confidence: Prediction confidence score between 0.0 and 1.0.
        status: One of 'detected', 'uncertain', or 'unknown'.
        alternatives: List of top alternative language predictions.
        candidate_language: The raw predicted language code, populated even when
                            confidence is below the threshold (status == 'uncertain').
        has_diacritics: Supporting flag indicating if Vietnamese diacritics were detected.
        raw_text_length: Length of the input text before cleaning.
    """
    language: Optional[str]
    confidence: float
    status: str
    alternatives: List[LanguageAlternative] = field(default_factory=list)
    candidate_language: Optional[str] = None
    has_diacritics: bool = False
    raw_text_length: int = 0

    @property
    def is_reliable(self) -> bool:
        """Convenience property: True if the prediction meets the confidence threshold."""
        return self.status == STATUS_DETECTED

    def to_dict(self) -> dict[str, Any]:
        """Serializes the result to a dictionary representation."""
        return {
            "language": self.language,
            "confidence": self.confidence,
            "status": self.status,
            "candidate_language": self.candidate_language,
            "has_diacritics": self.has_diacritics,
            "raw_text_length": self.raw_text_length,
            "alternatives": [alt.to_dict() for alt in self.alternatives],
        }


def resolve_model_path(custom_path: Optional[Union[str, Path]] = None) -> Path:
    """
    Resolves the FastText model path consistently regardless of current working directory.

    Resolution order:
      1. Explicit custom_path argument (if provided)
      2. Environment variable FASTTEXT_LID_MODEL_PATH
      3. Default location: <project_root>/models/language/lid.176.ftz
    """
    if custom_path:
        p = Path(custom_path)
        return p if p.is_absolute() else (PROJECT_ROOT / p).resolve()

    env_val = os.getenv("FASTTEXT_LID_MODEL_PATH")
    if env_val:
        p = Path(env_val)
        return p if p.is_absolute() else (PROJECT_ROOT / p).resolve()

    return (PROJECT_ROOT / DEFAULT_MODEL_REL_PATH).resolve()


def check_vietnamese_diacritics(text: str) -> bool:
    """
    Checks if the text contains Vietnamese-specific diacritical characters.
    Used strictly as supporting information, never as the sole determinant of
    language prediction or confidence score.
    """
    return any(c in VIETNAMESE_DIACRITICS_CHARS for c in text)


class LanguageDetector:
    """
    Production-grade, thread-safe Language Detector using FastText lid.176.ftz.

    Loads the binary model lazily once per process and provides structured,
    confidence-gated predictions.
    """

    def __init__(
        self,
        model_path: Optional[Union[str, Path]] = None,
        min_confidence: Optional[float] = None,
        top_k: Optional[int] = None,
        model: Optional[Any] = None,
    ) -> None:
        """
        Initializes the LanguageDetector.

        Args:
            model_path: Path to lid.176.ftz model file. If None, resolved from env/default.
            min_confidence: Confidence threshold for 'detected' status (default 0.70).
            top_k: Number of alternative predictions to return (default 3).
            model: Optional pre-loaded or mock FastText model instance (for dependency injection & testing).
        """
        self.model_path = resolve_model_path(model_path)
        
        # Load configurable thresholds from arguments or environment
        if min_confidence is not None:
            self.min_confidence = float(min_confidence)
        else:
            env_conf = os.getenv("LANGUAGE_DETECTION_MIN_CONFIDENCE")
            self.min_confidence = float(env_conf) if env_conf else DEFAULT_MIN_CONFIDENCE

        if top_k is not None:
            self.top_k = int(top_k)
        else:
            env_top_k = os.getenv("LANGUAGE_DETECTION_TOP_K")
            self.top_k = int(env_top_k) if env_top_k else DEFAULT_TOP_K

        self._model = model
        self._lock = threading.Lock()

    def _load_model(self) -> Any:
        """Loads the FastText model from disk."""
        if not self.model_path.is_file():
            raise ModelNotFoundError(
                f"FastText language identification model not found at '{self.model_path}'.\n"
                f"Please download the official model 'lid.176.ftz':\n"
                f"  curl -L -o '{self.model_path}' "
                f"https://dl.fbaipublicfiles.com/fasttext/supervised-models/lid.176.ftz\n"
                f"Or set FASTTEXT_LID_MODEL_PATH to point to your model file."
            )

        try:
            import fasttext
        except ImportError as exc:
            raise ImportError(
                "The 'fasttext' module is required for LanguageDetector. "
                "Please add 'fasttext-wheel' to requirements.txt and install it."
            ) from exc

        logger.info("Loading FastText language model from: %s", self.model_path)
        # Suppress fasttext verbose warning on newlines if library allows
        loaded_model = fasttext.load_model(str(self.model_path))
        logger.info("FastText language model loaded successfully.")
        return loaded_model

    def get_model(self) -> Any:
        """Returns the loaded FastText model instance, loading it on first access (thread-safe)."""
        if self._model is None:
            with self._lock:
                if self._model is None:
                    self._model = self._load_model()
        return self._model

    def detect(
        self,
        text: str,
        top_k: Optional[int] = None,
        min_confidence: Optional[float] = None,
        allow_uncertain_language: bool = False,
    ) -> LanguageDetectionResult:
        """
        Detects the language of the provided user text.

        The original text string is NEVER modified or overwritten.

        Args:
            text: Input string to identify.
            top_k: Number of alternative predictions to return (defaults to self.top_k).
            min_confidence: Confidence threshold (defaults to self.min_confidence).
            allow_uncertain_language: If True, populates .language with candidate_language
                                      even when status is 'uncertain'. If False (default),
                                      .language is None unless status is 'detected'.

        Returns:
            LanguageDetectionResult with language, confidence, status, and alternatives.
        """
        if text is None:
            text = ""

        raw_length = len(text)
        has_diacritics = check_vietnamese_diacritics(text)

        # 1. Edge Case: Verify usable alphabetic content
        # If the input contains no letters at all (e.g., pure numbers, pure emojis, pure punctuation),
        # there is no linguistic signal to identify.
        has_letters = any(c.isalpha() for c in text)
        if not has_letters:
            return LanguageDetectionResult(
                language=None,
                confidence=0.0,
                status=STATUS_UNKNOWN,
                alternatives=[],
                candidate_language=None,
                has_diacritics=has_diacritics,
                raw_text_length=raw_length,
            )

        # 2. Input normalization for FastText (on a temporary copy only)
        # - Unicode NFC normalization ensures consistent representation of accented letters.
        # - FastText expects single-line text without newline characters.
        normalized = unicodedata.normalize("NFC", text)
        clean_for_model = " ".join(normalized.split())

        if not clean_for_model:
            return LanguageDetectionResult(
                language=None,
                confidence=0.0,
                status=STATUS_UNKNOWN,
                alternatives=[],
                candidate_language=None,
                has_diacritics=has_diacritics,
                raw_text_length=raw_length,
            )

        # 3. Model prediction
        k = top_k if top_k is not None else self.top_k
        threshold = min_confidence if min_confidence is not None else self.min_confidence

        model = self.get_model()
        try:
            labels, probs = model.predict(clean_for_model, k=k)
        except Exception as exc:
            logger.error("FastText prediction failed for input of length %d: %s", raw_length, exc)
            return LanguageDetectionResult(
                language=None,
                confidence=0.0,
                status=STATUS_UNKNOWN,
                alternatives=[],
                candidate_language=None,
                has_diacritics=has_diacritics,
                raw_text_length=raw_length,
            )

        if not labels:
            return LanguageDetectionResult(
                language=None,
                confidence=0.0,
                status=STATUS_UNKNOWN,
                alternatives=[],
                candidate_language=None,
                has_diacritics=has_diacritics,
                raw_text_length=raw_length,
            )

        # 4. Parse FastText predictions and normalize labels
        alternatives: List[LanguageAlternative] = []
        for label, prob in zip(labels, probs):
            # Normalization: '__label__vi' -> 'vi', '__label__en' -> 'en'
            clean_code = label.replace("__label__", "").strip().lower()
            alternatives.append(
                LanguageAlternative(
                    language=clean_code,
                    confidence=round(float(prob), 4),
                )
            )

        top_candidate = alternatives[0].language
        top_confidence = alternatives[0].confidence

        # 5. Evaluate confidence threshold
        if top_confidence >= threshold:
            status = STATUS_DETECTED
            language = top_candidate
        else:
            status = STATUS_UNCERTAIN
            language = top_candidate if allow_uncertain_language else None

        return LanguageDetectionResult(
            language=language,
            confidence=top_confidence,
            status=status,
            alternatives=alternatives,
            candidate_language=top_candidate,
            has_diacritics=has_diacritics,
            raw_text_length=raw_length,
        )


# Global singleton instance cache
_detector_instance: Optional[LanguageDetector] = None
_instance_lock = threading.Lock()


def get_language_detector(
    model_path: Optional[Union[str, Path]] = None,
    min_confidence: Optional[float] = None,
    top_k: Optional[int] = None,
    force_reload: bool = False,
) -> LanguageDetector:
    """
    Returns the application-wide singleton instance of LanguageDetector.
    Thread-safe and lazily instantiated.
    """
    global _detector_instance
    if _detector_instance is None or force_reload:
        with _instance_lock:
            if _detector_instance is None or force_reload:
                _detector_instance = LanguageDetector(
                    model_path=model_path,
                    min_confidence=min_confidence,
                    top_k=top_k,
                )
    return _detector_instance

