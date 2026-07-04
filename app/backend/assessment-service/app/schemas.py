# File: app/schemas.py
from pydantic import BaseModel
from typing import Optional
import uuid


# DTO để nhận dữ liệu khi tạo Template mới
class TemplateCreate(BaseModel):
    code: str
    title: str
    description: Optional[str] = None


# DTO cho từng Phương án trả lời
class OptionCreate(BaseModel):
    label: str           # Ví dụ: "Không bao giờ"
    value: str           # Ví dụ: "NEVER"
    score_value: int     # Điểm số: 0
    order_index: int     # Thứ tự hiển thị: 1


# DTO cho Nhóm phương án (chứa nhiều phương án bên trong)
class OptionGroupCreate(BaseModel):
    group_code: str
    group_name: str
    description: Optional[str] = None
    options: list[OptionCreate]


# DTO cho một câu hỏi lẻ
class QuestionCreate(BaseModel):
    content: str
    dimension: str       # Mảng bệnh (Ví dụ: DEPRESSION, ANXIETY, STRESS)
    question_order: int


# DTO tổng để bọc danh sách câu hỏi gửi lên
class QuestionBulkCreate(BaseModel):
    template_id: str
    group_id: str
    questions: list[QuestionCreate]


# DTO cho một câu trả lời lẻ (Câu 1 chọn option_id gì)
class AnswerSubmit(BaseModel):
    question_id: str
    option_id: str


# DTO cho toàn bộ bài nộp của User
class AssessmentSubmit(BaseModel):
    template_id: str
    user_id: str
    answers: list[AnswerSubmit]


# DTO trả về thông tin cơ bản của bài test
class TemplateResponse(BaseModel):
    template_id: uuid.UUID
    code: str
    title: str
    description: Optional[str]
    
    class Config:
        from_attributes = True


# DTO trả về chi tiết 1 phương án
class OptionResponse(BaseModel):
    option_id: uuid.UUID
    label: str
    score_value: int
    order_index: int

    class Config:
        from_attributes = True


# DTO trả về Câu hỏi kèm danh sách các phương án của nó
class QuestionWithOptionsResponse(BaseModel):
    question_id: uuid.UUID
    content: str
    dimension: Optional[str] = None
    question_order: int
    options: list[OptionResponse]

    class Config:
        from_attributes = True