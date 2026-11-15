# app/core/orchestrator.py
import logging
import re
from typing import NamedTuple
from app.language import get_language_detector
from app.layers import layer0_safety, layer1_nlu, layer2_dialogue, layer3_rag, layer4_llm, layer5_filter
from app.translation.translation_service import translate

logger = logging.getLogger(__name__)

SAFE_FALLBACK_RESPONSE = "I'm sorry, I'm experiencing a technical issue right now."
VIETNAMESE_SAFE_FALLBACK_RESPONSE = (
    "Xin lỗi, hiện tại tôi gặp sự cố kỹ thuật. Vui lòng thử lại sau."
)
VIETNAMESE_OUTPUT_TRANSLATION_FAILURE_RESPONSE = (
    "Xin lỗi, tôi chưa thể chuyển câu trả lời sang tiếng Việt lúc này. Vui lòng thử lại sau."
)
VIETNAMESE_OUTPUT_VALIDATION_FAILURE_RESPONSE = (
    "Xin lỗi, tôi chưa thể xác nhận câu trả lời an toàn bằng tiếng Việt lúc này. Vui lòng thử lại sau."
)
_LAYER0_SUPPORTED_LANGUAGES = {"en", "vi"}
_INPUT_TRANSLATION_LANGUAGES = {"vi"}
_UNSUPPORTED_LANGUAGE_RESPONSE = (
    "This chatbot currently supports only English and Vietnamese. Please use English or Vietnamese. "
    "/ Hiện tại chatbot chỉ hỗ trợ tiếng Anh và tiếng Việt. Vui lòng sử dụng tiếng Anh hoặc tiếng Việt."
)
_LANGUAGE_CLARIFICATION_RESPONSE = (
    "I couldn't confidently identify the language. Please write a longer message in English or Vietnamese. "
    "/ Tôi chưa thể xác định ngôn ngữ một cách chắc chắn. Vui lòng viết tin nhắn dài hơn bằng tiếng Anh hoặc tiếng Việt."
)
_INPUT_TRANSLATION_FAILURE_RESPONSES = {
    "vi": "Xin lỗi, chúng tôi chưa thể xử lý tin nhắn của bạn lúc này. Vui lòng thử lại sau.",
}


class _DetectionMetadata(NamedTuple):
    status: str | None = None
    language: str | None = None
    layer0_language: str | None = None
    candidate_language: str | None = None


def _detect_language(user_message: str):
    """Detect language once, returning None if the detector is unavailable."""
    try:
        return get_language_detector().detect(user_message)
    except Exception as exc:
        logger.warning(
            "Language detection failed; using English safety fallback: %s",
            type(exc).__name__,
        )
        return None


def _normalise_language_code(language_value: object) -> str | None:
    if language_value is None:
        return None
    if not isinstance(language_value, str) or not re.fullmatch(
        r"[A-Za-z]{2,3}(?:[-_][A-Za-z0-9]{2,8})*", language_value
    ):
        return None
    return language_value.replace("_", "-").lower()


def _read_detection_metadata(detection) -> _DetectionMetadata:
    """Read detector fields once; malformed metadata becomes unresolved."""
    try:
        status = getattr(detection, "status", None)
        raw_language = getattr(detection, "language", None)
        raw_candidate = getattr(detection, "candidate_language", None)
    except Exception as exc:
        logger.warning("Language detection metadata failed: %s", type(exc).__name__)
        return _DetectionMetadata()

    if not isinstance(status, str) or status not in {"detected", "uncertain", "unknown"}:
        return _DetectionMetadata()

    language = _normalise_language_code(raw_language)
    candidate_language = _normalise_language_code(raw_candidate)
    if (
        (raw_language is not None and language is None)
        or (raw_candidate is not None and candidate_language is None)
        or (status == "detected" and language is None)
    ):
        return _DetectionMetadata()

    layer0_language = _primary_language(language)

    return _DetectionMetadata(
        status=status,
        language=language,
        layer0_language=layer0_language,
        candidate_language=candidate_language,
    )


def _primary_language(language: str | None) -> str | None:
    return language.split("-", 1)[0] if language is not None else None


def _get_layer0_scan_order(detection: _DetectionMetadata) -> list[str]:
    """Return the ordered Layer 0 language checks for the raw user message."""
    detected_language = _primary_language(detection.language)
    if detection.status == "detected" and detected_language in _LAYER0_SUPPORTED_LANGUAGES:
        return [detection.layer0_language or detected_language]

    return ["en", "vi"]


def _run_layer0_safety_checks(
    user_message: str, detection: _DetectionMetadata
) -> tuple[bool, str]:
    """Check the original message across the required Layer 0 language order."""
    for language in _get_layer0_scan_order(detection):
        try:
            is_safe, emergency_response = layer0_safety.check_input_risk(
                user_message, lang=language
            )
        except Exception as exc:
            logger.error("Layer 0 safety check failed: %s", type(exc).__name__)
            return False, SAFE_FALLBACK_RESPONSE

        if not is_safe:
            return False, emergency_response

    return True, ""


def _translate_input_to_english(
    user_message: str, detection: _DetectionMetadata
) -> tuple[str | None, str | None]:
    """Return English pipeline input or a localized failure response.

    Translation is permitted only for a confident, supported non-English
    detection. Uncertain, unknown, and unsupported results preserve the
    existing untranslated pipeline behavior.
    """
    detected_language = _primary_language(detection.language)
    if detection.status == "detected" and detected_language == "en":
        return user_message, None
    if (
        detection.status != "detected"
        or detected_language not in _INPUT_TRANSLATION_LANGUAGES
    ):
        logger.info("Language detection is unresolved or unsupported; skipping input translation")
        return user_message, None

    try:
        translated_message = translate(
            user_message, source_lang=detected_language, target_lang="en"
        )
        if not translated_message or not translated_message.strip():
            raise ValueError("Translation returned an empty result")
        return translated_message, None
    except Exception as exc:
        logger.error("Input translation failed: %s", type(exc).__name__)
        return None, _INPUT_TRANSLATION_FAILURE_RESPONSES[detected_language]


def chat_orchestrator_stream(session_id: str, user_message: str, audio_file: str = None):
    # 1. Detect language, then check the original input for crisis signals.
    detection = _read_detection_metadata(_detect_language(user_message))
    is_safe, emergency_response = _run_layer0_safety_checks(user_message, detection)
    if not is_safe:
        yield emergency_response
        return

    detected_language = _primary_language(detection.language)
    if detection.status != "detected":
        yield _LANGUAGE_CLARIFICATION_RESPONSE
        return
    if detected_language not in _LAYER0_SUPPORTED_LANGUAGES:
        yield _UNSUPPORTED_LANGUAGE_RESPONSE
        return

    # 2. Translate only confidently detected supported non-English input.
    pipeline_message, translation_failure_response = _translate_input_to_english(
        user_message, detection
    )
    if translation_failure_response is not None:
        yield translation_failure_response
        return

    # 3. Layer 1: Nhận diện cảm xúc
    emotion = layer1_nlu.analyze_text_emotion(pipeline_message)
    if audio_file:
        emotion = layer1_nlu.analyze_audio_emotion(audio_file)
    
    # 4. Layer 2: Lấy lịch sử hội thoại
    history = layer2_dialogue.get_chat_history(session_id)
    
    # 5. Layer 3: RAG - Truy xuất tài liệu
    context = layer3_rag.retrieve_interventions(pipeline_message, current_emotion=emotion)
    
    # 6. Layer 4: Aggregate the complete response before filtering or yielding.
    try:
        full_response = ""
        for chunk in layer4_llm.generate_response_stream(f"{history}\nUser: {pipeline_message}", context):
            full_response += chunk
    except Exception as exc:
        logger.error("Error during LLM response generation: %s", type(exc).__name__)
        final_safe_response = (
            VIETNAMESE_SAFE_FALLBACK_RESPONSE
            if detected_language == "vi"
            else SAFE_FALLBACK_RESPONSE
        )
        canonical_response = SAFE_FALLBACK_RESPONSE
    else:
        if not full_response or not full_response.strip():
            logger.warning("Empty or whitespace-only response received from LLM")
            final_safe_response = (
                VIETNAMESE_SAFE_FALLBACK_RESPONSE
                if detected_language == "vi"
                else SAFE_FALLBACK_RESPONSE
            )
            canonical_response = SAFE_FALLBACK_RESPONSE
        else:
            # Layer 5 must filter the complete English Qwen output before translation.
            try:
                english_safe_response = layer5_filter.filter_output(
                    full_response, lang="en"
                )
            except Exception as exc:
                logger.error("English output safety filtering failed: %s", type(exc).__name__)
                yield (
                    VIETNAMESE_SAFE_FALLBACK_RESPONSE
                    if detected_language == "vi"
                    else SAFE_FALLBACK_RESPONSE
                )
                return

            canonical_response = english_safe_response
            if detected_language == "vi":
                try:
                    translated_response = translate(
                        english_safe_response, source_lang="en", target_lang="vi"
                    )
                except Exception as exc:
                    logger.error("Output translation failed: %s", type(exc).__name__)
                    yield VIETNAMESE_OUTPUT_TRANSLATION_FAILURE_RESPONSE
                    return

                if (
                    not isinstance(translated_response, str)
                    or not translated_response.strip()
                    or translated_response.strip() == english_safe_response.strip()
                ):
                    logger.warning("Output translation returned an empty result")
                    yield VIETNAMESE_OUTPUT_TRANSLATION_FAILURE_RESPONSE
                    return

                try:
                    final_safe_response = layer5_filter.filter_output(
                        translated_response, lang="vi"
                    )
                except Exception as exc:
                    logger.error(
                        "Vietnamese output safety filtering failed: %s",
                        type(exc).__name__,
                    )
                    yield VIETNAMESE_OUTPUT_VALIDATION_FAILURE_RESPONSE
                    return

                if final_safe_response != translated_response:
                    # A Layer 5 replacement is safe to return but not a normal turn.
                    yield final_safe_response
                    return
            else:
                final_safe_response = english_safe_response

    # 7. Save only canonical English content after a successful user-facing validation.
    try:
        layer2_dialogue.save_chat_turn(
            session_id, pipeline_message, canonical_response
        )
    except Exception as exc:
        logger.error("Error saving chat turn to dialogue history: %s", type(exc).__name__)

    # 8. Yield exactly one complete, filtered response.
    yield final_safe_response
