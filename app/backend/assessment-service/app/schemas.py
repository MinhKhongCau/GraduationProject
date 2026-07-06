# File: app/schemas.py
from pydantic import BaseModel, Field
from typing import Optional, List
import uuid


class TemplateDBSchema(BaseModel):
    template_id: uuid.UUID
    code: str
    title: str
    description: Optional[str] = None
    slug: Optional[str] = Field(default=None, min_length=10, max_length=10)
    is_active: bool

    class Config:
        from_attributes = True


class OptionDBSchema(BaseModel):
    option_id: uuid.UUID
    group_id: uuid.UUID
    label: str
    value: str
    score_value: int
    order_index: int
    slug: Optional[str] = Field(default=None, min_length=10, max_length=10)

    class Config:
        from_attributes = True


class OptionGroupDBSchema(BaseModel):
    group_id: uuid.UUID
    group_code: str
    group_name: str
    description: Optional[str] = None
    slug: Optional[str] = Field(default=None, min_length=10, max_length=10)
    options: List[OptionDBSchema] = []

    class Config:
        from_attributes = True


class QuestionDBSchema(BaseModel):
    question_id: uuid.UUID
    template_id: uuid.UUID
    group_id: uuid.UUID
    content: str
    dimension: Optional[str] = None
    question_order: int
    is_required: bool
    slug: Optional[str] = Field(default=None, min_length=10, max_length=10)

    class Config:
        from_attributes = True