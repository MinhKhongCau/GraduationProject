from pydantic import BaseModel, Field
from typing import Optional

class OptionCreate(BaseModel):
    label: str
    value: str
    score_value: int
    order_index: int


class OptionResponse(BaseModel):
    label: str
    value: str
    score_value: int
    order_index: int
    slug: Optional[str] = Field(default=None, min_length=10, max_length=10)

    class Config:
        from_attributes = True


class OptionUpdate(BaseModel):
    label: Optional[str] = None
    value: Optional[str] = None
    score_value: Optional[int] = None
    order_index: Optional[int] = None
