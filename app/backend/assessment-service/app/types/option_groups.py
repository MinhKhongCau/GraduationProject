from pydantic import BaseModel, Field
from typing import Optional, List
from .options import OptionCreate, OptionResponse

class OptionGroupCreate(BaseModel):
    group_code: str
    group_name: str
    description: Optional[str] = None
    options: List[OptionCreate]


class OptionGroupResponse(BaseModel):
    group_code: str
    group_name: str
    description: Optional[str]
    slug: Optional[str] = Field(default=None, min_length=10, max_length=10)
    options: List[OptionResponse] = []

    class Config:
        from_attributes = True


class OptionGroupUpdate(BaseModel):
    group_code: Optional[str] = None
    group_name: Optional[str] = None
    description: Optional[str] = None
