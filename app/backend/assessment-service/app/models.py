from sqlalchemy import (
    Column, String, Integer, Boolean, ForeignKey, Text, DateTime
)
from sqlalchemy.dialects.postgresql import UUID, JSONB
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.sql import func
import uuid

Base = declarative_base()


# =============================================
# 1. BÀI TEST / TEMPLATE
# =============================================
class AssessTemplate(Base):
    __tablename__ = "assess_templates"
    template_id = Column(
        UUID(as_uuid=True), primary_key=True, default=uuid.uuid4
    )
    code = Column(String, unique=True, index=True)
    title = Column(String)
    description = Column(Text)
    is_active = Column(Boolean, default=True)
    slug = Column(String, unique=True, index=True, nullable=True)
    created_at = Column(
        DateTime(timezone=True), server_default=func.now()
    )


# =============================================
# 2. NHÓM PHƯƠNG ÁN TRẢ LỜI
# =============================================
class AssessOptionGroup(Base):
    __tablename__ = "assess_option_groups"
    group_id = Column(
        UUID(as_uuid=True), primary_key=True, default=uuid.uuid4
    )
    group_code = Column(String, unique=True)
    group_name = Column(String)
    description = Column(Text)
    is_active = Column(Boolean, default=True)
    slug = Column(String, unique=True, index=True, nullable=True)


# =============================================
# 3. PHƯƠNG ÁN TRẢ LỜI
# =============================================
class AssessOption(Base):
    __tablename__ = "assess_options"
    option_id = Column(
        UUID(as_uuid=True), primary_key=True, default=uuid.uuid4
    )
    group_id = Column(
        UUID(as_uuid=True),
        ForeignKey("assess_option_groups.group_id")
    )
    label = Column(String)
    value = Column(String)
    score_value = Column(Integer)
    order_index = Column(Integer)
    is_active = Column(Boolean, default=True)
    slug = Column(String, unique=True, index=True, nullable=True)


# =============================================
# 4. CÂU HỎI
# =============================================
class AssessQuestion(Base):
    __tablename__ = "assess_questions"
    question_id = Column(
        UUID(as_uuid=True), primary_key=True, default=uuid.uuid4
    )
    template_id = Column(
        UUID(as_uuid=True),
        ForeignKey("assess_templates.template_id")
    )
    group_id = Column(
        UUID(as_uuid=True),
        ForeignKey("assess_option_groups.group_id")
    )
    content = Column(Text)
    # Ví dụ: DEPRESSION, ANXIETY
    dimension = Column(String)
    question_order = Column(Integer)
    is_required = Column(Boolean, default=True)
    is_active = Column(Boolean, default=True)
    slug = Column(String, unique=True, index=True, nullable=True)


# =============================================
# 5. LUẬT TÍNH ĐIỂM (MỚI)
# =============================================
class AssessScoreRule(Base):
    __tablename__ = "assess_score_rules"
    rule_id = Column(
        UUID(as_uuid=True), primary_key=True, default=uuid.uuid4
    )
    template_id = Column(
        UUID(as_uuid=True),
        ForeignKey("assess_templates.template_id")
    )
    # Áp dụng cho nhóm bệnh nào
    dimension = Column(String, nullable=True)
    min_score = Column(Integer)
    max_score = Column(Integer)
    # Nhẹ, Nặng, Bình thường
    severity_level = Column(String)
    short_advice = Column(Text)


# =============================================
# 6. KẾT QUẢ BÀI TEST
# =============================================
class AssessResult(Base):
    __tablename__ = "assess_results"
    result_id = Column(
        UUID(as_uuid=True), primary_key=True, default=uuid.uuid4
    )
    template_id = Column(
        UUID(as_uuid=True),
        ForeignKey("assess_templates.template_id")
    )
    # Lấy từ Token của Auth Service
    user_id = Column(UUID(as_uuid=True))
    total_score = Column(Integer)
    # Lưu trữ dạng JSON {"DEPRESSION": 15, "ANXIETY": 8}
    dimension_scores = Column(JSONB)
    ai_evaluation = Column(Text)
    created_at = Column(
        DateTime(timezone=True), server_default=func.now()
    )


# =============================================
# 7. CÂU TRẢ LỜI CHI TIẾT CỦA NGƯỜI DÙNG
# =============================================
class AssessUserAnswer(Base):
    __tablename__ = "assess_user_answers"
    answer_id = Column(
        UUID(as_uuid=True), primary_key=True, default=uuid.uuid4
    )
    result_id = Column(
        UUID(as_uuid=True),
        ForeignKey("assess_results.result_id")
    )
    question_id = Column(
        UUID(as_uuid=True),
        ForeignKey("assess_questions.question_id")
    )
    option_id = Column(
        UUID(as_uuid=True),
        ForeignKey("assess_options.option_id")
    )
