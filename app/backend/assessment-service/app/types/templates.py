from pydantic import BaseModel, Field
from typing import Optional

class TemplateCreate(BaseModel):
    code: str
    title: str
    description: Optional[str] = None


class TemplateResponse(BaseModel):
    code: str
    title: str
    description: Optional[str]
    slug: Optional[str] = Field(default=None, min_length=10, max_length=10)
    is_active: bool
    
    class Config:
        from_attributes = True


class TemplateUpdate(BaseModel):
    code: Optional[str] = None
    title: Optional[str] = None
    description: Optional[str] = None
    is_active: Optional[bool] = None
