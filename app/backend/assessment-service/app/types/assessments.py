from pydantic import BaseModel
from typing import List, Optional
from datetime import datetime
import uuid

class AnswerSubmit(BaseModel):
    question_id: str
    option_id: str


class AssessmentSubmit(BaseModel):
    template_id: str
    user_id: str
    answers: List[AnswerSubmit]


class AssessmentResultResponse(BaseModel):
    result_id: uuid.UUID
    template_code: str
    template_title: str
    total_score: int
    dimension_scores: dict
    ai_evaluation: Optional[str] = None
    created_at: datetime

    class Config:
        from_attributes = True
