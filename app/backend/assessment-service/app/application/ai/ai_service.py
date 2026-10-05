import os
from typing import Mapping, Sequence

from dotenv import load_dotenv
from google import genai

load_dotenv()

GEMINI_API_KEY = os.getenv("GEMINI_API_KEY")
GEMINI_MODEL = "gemini-2.5-flash"
MAINTENANCE_MESSAGE = (
    "Hệ thống AI đang bảo trì, vui lòng liên hệ chuyên gia để được tư vấn thêm."
)
FALLBACK_MESSAGE = (
    "Đã ghi nhận kết quả bài test. Hãy đặt lịch với bác sĩ để được tư vấn "
    "chi tiết hơn nhé."
)


def _format_dimension_scores(dimension_scores: Mapping[str, int]) -> str:
    return "\n".join(
        f"- {dimension}: {score}"
        for dimension, score in sorted(dimension_scores.items())
    )


def _format_answers(answers: Sequence[Mapping[str, str]]) -> str:
    if not answers:
        return "N/A"
    return "\n".join(
        f"- Q: {answer.get('question_content', '')} | A: {answer.get('option_label', '')}"
        for answer in answers
    )


def _build_prompt(
    formatted_scores: str,
    formatted_answers: str,
    title: str | None,
    description: str | None,
    instruction: str | None,
    certification: str | None,
) -> str:
    return f"""
        You are an empathetic clinical psychology consultant.

        Assessment information:
        - Title: {title or "N/A"}
        - Description: {description or "N/A"}
        - Instruction: {instruction or "N/A"}
        - Certification: {certification or "N/A"}

        The user just completed this assessment with the following dimension scores:
        {formatted_scores}

        The user's individual questions and chosen answers were:
        {formatted_answers}

        Response requirements:
        - First, detect whether the assessment information above is written in English or Vietnamese, then reply entirely in that same language.
        - Use the individual questions and answers to make the advice specific and personal, not just a generic summary of the scores.
        - Keep the response to approximately 100 words.
        - This is for reference only, it is not a medical diagnosis.
        - If any score indicates a high severity level, gently recommend the user consult a doctor or psychologist on the MindCare platform.
        - Tone should be supportive, clear, and non-alarming.
        """.strip()


def generate_psychological_advice(
    dimension_scores: Mapping[str, int] | None,
    answers: Sequence[Mapping[str, str]] | None = None,
    title: str | None = None,
    description: str | None = None,
    instruction: str | None = None,
    certification: str | None = None,
) -> str:
    """
    Tạo lời khuyên tâm lý tham khảo từ điểm số và câu trả lời bài test bằng Gemini.
    """
    if not dimension_scores:
        return FALLBACK_MESSAGE

    if not GEMINI_API_KEY:
        return MAINTENANCE_MESSAGE

    prompt = _build_prompt(
        _format_dimension_scores(dimension_scores),
        _format_answers(answers or []),
        title,
        description,
        instruction,
        certification,
    )

    try:
        client = genai.Client(api_key=GEMINI_API_KEY)
        response = client.models.generate_content(
            model=GEMINI_MODEL,
            contents=prompt
        )

        response_text = getattr(response, "text", "") or ""
        response_text = response_text.strip()
        return response_text or FALLBACK_MESSAGE
    except Exception as e:
        print(f"Lỗi khi gọi AI: {e}")
        return FALLBACK_MESSAGE
