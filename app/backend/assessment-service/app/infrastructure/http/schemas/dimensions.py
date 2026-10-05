from pydantic import BaseModel, Field
from typing import Optional, List

class DimensionCreate(BaseModel):
    code: str
    name: str
    description: Optional[str] = None


class DimensionResponse(BaseModel):
    code: str
    name: str
    description: Optional[str] = None
    slug: Optional[str] = Field(default=None, min_length=10, max_length=10)
    is_active: bool

    class Config:
        from_attributes = True


class DimensionUpdate(BaseModel):
    code: Optional[str] = None
    name: Optional[str] = None
    description: Optional[str] = None
    is_active: Optional[bool] = None


class DimensionQuestionsBulkAssign(BaseModel):
    question_ids: List[str]  # holds question slugs
