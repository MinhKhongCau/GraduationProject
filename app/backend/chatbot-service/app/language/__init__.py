# app/language/__init__.py
"""
Language processing package for the chatbot-service.

Provides independent language identification capabilities using FastText (lid.176.ftz).
"""

from app.language.language_detector import (
    DEFAULT_MIN_CONFIDENCE,
    DEFAULT_TOP_K,
    STATUS_DETECTED,
    STATUS_UNCERTAIN,
    STATUS_UNKNOWN,
    LanguageAlternative,
    LanguageDetectionResult,
    LanguageDetector,
    ModelNotFoundError,
    get_language_detector,
)

__all__ = [
    "LanguageDetector",
    "LanguageDetectionResult",
    "LanguageAlternative",
    "ModelNotFoundError",
    "get_language_detector",
    "DEFAULT_MIN_CONFIDENCE",
    "DEFAULT_TOP_K",
    "STATUS_DETECTED",
    "STATUS_UNCERTAIN",
    "STATUS_UNKNOWN",
]

