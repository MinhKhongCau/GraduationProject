"""FastAPI router for multilingual translation service.

Endpoint: POST /api/v1/translation/translate
Uses synchronous CPU inference handled automatically by FastAPI threadpool.
"""

from __future__ import annotations

import logging
from fastapi import APIRouter, HTTPException, status
from pydantic import BaseModel, Field, field_validator

from app.translation.translation_service import (
    TranslationError,
    translate,
)

logger = logging.getLogger(__name__)

router = APIRouter()

SUPPORTED_LANGUAGES = {"vi", "en", "ja", "ko", "zh"}
MODEL_IDENTIFIER = "facebook/m2m100_418M"


class TranslationRequest(BaseModel):
    text: str = Field(..., description="Text to translate")
    source_lang: str = Field(..., description="Source language code (vi, en, ja, ko, zh)")
    target_lang: str = Field(..., description="Target language code (vi, en, ja, ko, zh)")

    @field_validator("source_lang", "target_lang")
    @classmethod
    def validate_language_code(cls, v: str) -> str:
        cleaned = (v or "").strip().lower()
        if cleaned not in SUPPORTED_LANGUAGES:
            raise ValueError(
                f"Unsupported language code '{v}'. Supported: {sorted(SUPPORTED_LANGUAGES)}"
            )
        return cleaned


class TranslationResponse(BaseModel):
    source_text: str = Field(..., description="Original input text")
    translated_text: str = Field(..., description="Translated output text")
    source_language: str = Field(..., description="Source language code")
    target_language: str = Field(..., description="Target language code")
    model_name: str = Field(..., description="Model identifier")
    status: str = Field(..., description="Operation status, e.g. success")


@router.post(
    "/translate",
    response_model=TranslationResponse,
    status_code=status.HTTP_200_OK,
    responses={
        422: {"description": "Validation error or unsupported language"},
        503: {"description": "Translation service unavailable or model missing"},
    },
)
def translate_endpoint(request: TranslationRequest) -> TranslationResponse:
    """Synchronous route handler for translation inference.

    Synchronous def allows FastAPI / AnyIO to schedule CPU-bound inference
    in the external thread pool without blocking the async event loop.
    """
    try:
        translated = translate(
            text=request.text,
            source_lang=request.source_lang,
            target_lang=request.target_lang,
        )
        return TranslationResponse(
            source_text=request.text,
            translated_text=translated,
            source_language=request.source_lang,
            target_language=request.target_lang,
            model_name=MODEL_IDENTIFIER,
            status="success",
        )
    except TranslationError as err:
        err_msg = str(err).lower()
        logger.error("TranslationError encountered in translation_router")
        if "unsupported" in err_msg or "token limit" in err_msg or "oversized" in err_msg:
            raise HTTPException(
                status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
                detail="Invalid translation request or input exceeds token limits",
            )
        raise HTTPException(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
            detail="Translation service unavailable",
        )
    except Exception:
        logger.error("Unexpected error encountered during translation")
        raise HTTPException(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
            detail="Translation service unavailable",
        )
