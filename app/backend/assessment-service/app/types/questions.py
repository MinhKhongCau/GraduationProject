from pydantic import BaseModel, Field
from typing import Optional, List
from .options import OptionResponse

class QuestionCreate(BaseModel):
    content: str
    dimension: str
    question_order: int


class QuestionBulkCreate(BaseModel):
    template_id: str
    group_id: str
    questions: List[QuestionCreate]


class QuestionResponse(BaseModel):
    content: str
    dimension: Optional[str] = None
    question_order: int
    is_required: bool
    slug: Optional[str] = Field(default=None, min_length=10, max_length=10)

    class Config:
        from_attributes = True


class QuestionWithOptionsResponse(BaseModel):
    content: str
    dimension: Optional[str] = None
    question_order: int
    slug: Optional[str] = Field(default=None, min_length=10, max_length=10)
    options: List[OptionResponse]

    class Config:
        from_attributes = True


class QuestionUpdate(BaseModel):
    content: Optional[str] = None
    dimension: Optional[str] = None
    question_order: Optional[int] = None
    is_required: Optional[bool] = None
