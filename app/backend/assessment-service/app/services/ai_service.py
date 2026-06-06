import os
from typing import Mapping

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


def _build_prompt(formatted_scores: str) -> str:
    return f"""
        Bạn là một chuyên gia tham vấn tâm lý, giao tiếp nhẹ nhàng và giàu đồng cảm.

        Người dùng vừa hoàn thành bài đánh giá DASS-21 với điểm số:
        {formatted_scores}

        Yêu cầu trả lời:
        - Viết bằng tiếng Việt.
        - Độ dài khoảng 3 đến 4 câu.
        - Chỉ mang tính tham khảo, không chẩn đoán bệnh.
        - Nếu có thang điểm cao, khuyên người dùng nên gặp bác sĩ hoặc chuyên gia tâm lý trên hệ thống MindCare.
        - Giọng điệu hỗ trợ, rõ ràng, không gây hoang mang.
        """.strip()


def generate_psychological_advice(
    dimension_scores: Mapping[str, int] | None
) -> str:
    """
    Tạo lời khuyên tâm lý tham khảo từ điểm số DASS-21 bằng Gemini.
    """
    if not dimension_scores:
        return FALLBACK_MESSAGE

    if not GEMINI_API_KEY:
        return MAINTENANCE_MESSAGE

    prompt = _build_prompt(_format_dimension_scores(dimension_scores))

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
