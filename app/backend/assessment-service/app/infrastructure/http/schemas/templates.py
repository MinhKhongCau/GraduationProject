from pydantic import BaseModel, Field, field_validator
from typing import Optional

class TemplateCreate(BaseModel):
    code: str
    title: str
    description: Optional[str] = None
    instruction: Optional[str] = None
    certification: Optional[str] = None

    @field_validator("instruction", "certification")
    @classmethod
    def validate_word_count(cls, v: Optional[str]) -> Optional[str]:
        if v:
            words = v.split()
            if len(words) > 3000:
                raise ValueError("Không được vượt quá 3000 từ")
        return v


class TemplateResponse(BaseModel):
    code: str
    title: str
    description: Optional[str]
    instruction: Optional[str] = None
    certification: Optional[str] = None
    slug: Optional[str] = Field(default=None, min_length=10, max_length=10)
    is_active: bool

    class Config:
        from_attributes = True


class TemplateUpdate(BaseModel):
    code: Optional[str] = None
    title: Optional[str] = None
    description: Optional[str] = None
    instruction: Optional[str] = None
    certification: Optional[str] = None
    is_active: Optional[bool] = None

    @field_validator("instruction", "certification")
    @classmethod
    def validate_word_count(cls, v: Optional[str]) -> Optional[str]:
        if v:
            words = v.split()
            if len(words) > 3000:
                raise ValueError("Không được vượt quá 3000 từ")
        return v
