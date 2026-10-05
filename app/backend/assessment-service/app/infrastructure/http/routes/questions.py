from typing import List
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from app.config.database import get_db
from app.infrastructure.persistence import models
from app.infrastructure.http import schemas
from app.infrastructure.http.dependencies import (
    get_template_by_slug,
    get_option_group_by_slug,
    get_question_by_slug,
    get_dimension_by_slug,
    check_admin_role,
    generate_unique_slug,
)

router = APIRouter(prefix="/api/v1/assessments", tags=["Questions"])


@router.post("/questions/bulk")
def create_bulk_questions(
    payload: schemas.QuestionBulkCreate,
    db: Session = Depends(get_db)
):
    added_questions = []

    # Verify template exists by slug
    template = get_template_by_slug(payload.template_id, db)
    if not template:
        raise HTTPException(status_code=404, detail="Không tìm thấy bài test tương ứng!")

    # Verify group exists by slug
    group = get_option_group_by_slug(payload.group_id, db)
    if not group:
        raise HTTPException(status_code=404, detail="Không tìm thấy nhóm phương án tương ứng!")

    for q in payload.questions:
        # Verify dimension exists by slug
        dimension = get_dimension_by_slug(q.dimension_id, db)
        if not dimension:
            raise HTTPException(
                status_code=404,
                detail=f"Không tìm thấy khía cạnh (dimension) tương ứng cho câu hỏi thứ {q.question_order}!"
            )

        # Generate unique 10-character question slug
        q_slug = generate_unique_slug(models.AssessQuestion, db)
        new_question = models.AssessQuestion(
            template_id=template.template_id,
            group_id=group.group_id,
            content=q.content,
            dimension_id=dimension.dimension_id,
            question_order=q.question_order,
            is_required=True,
            slug=q_slug
        )
        db.add(new_question)
        added_questions.append(new_question)

    db.commit()
    for q in added_questions:
        db.refresh(q)

    return {
        "message": f"Đã thêm thành công {len(added_questions)} câu hỏi vào bài test!",
        "data": [schemas.QuestionResponse.model_validate(q) for q in added_questions]
    }


@router.get(
    "/questions/{slug}",
    response_model=schemas.QuestionResponse
)
def get_question(slug: str, db: Session = Depends(get_db)):
    question = get_question_by_slug(slug, db)
    if not question:
        raise HTTPException(
            status_code=404,
            detail="Không tìm thấy câu hỏi này!"
        )
    return question


@router.patch(
    "/questions/{slug}",
    response_model=schemas.QuestionResponse
)
def update_question(
    slug: str,
    payload: schemas.QuestionUpdate,
    db: Session = Depends(get_db),
    is_admin: bool = Depends(check_admin_role)
):
    question = get_question_by_slug(slug, db)
    if not question:
        raise HTTPException(
            status_code=404,
            detail="Không tìm thấy câu hỏi để cập nhật!"
        )

    update_data = payload.model_dump(exclude_unset=True)

    if "dimension_id" in update_data:
        dimension_slug = update_data.pop("dimension_id")
        dimension = get_dimension_by_slug(dimension_slug, db)
        if not dimension:
            raise HTTPException(
                status_code=404,
                detail="Không tìm thấy khía cạnh (dimension) tương ứng!"
            )
        question.dimension_id = dimension.dimension_id

    for key, value in update_data.items():
        setattr(question, key, value)

    db.commit()
    db.refresh(question)
    return question


@router.delete("/questions/{slug}")
def delete_question(
    slug: str,
    db: Session = Depends(get_db),
    is_admin: bool = Depends(check_admin_role)
):
    question = get_question_by_slug(slug, db)
    if not question:
        raise HTTPException(
            status_code=404,
            detail="Không tìm thấy câu hỏi để xóa!"
        )
    
    question.is_active = False
    db.commit()
    return {
        "message": "Xóa câu hỏi thành công (Soft Delete)!"
    }


@router.get(
    "/templates/{slug}/questions",
    response_model=List[schemas.QuestionWithOptionsResponse]
)
def get_questions_by_template(
    slug: str,
    db: Session = Depends(get_db)
):
    db_template = get_template_by_slug(slug, db)
    if not db_template:
        raise HTTPException(
            status_code=404,
            detail="Lỗi: Không tìm thấy bài test này!"
        )

    questions = (
        db.query(models.AssessQuestion)
        .filter(
            models.AssessQuestion.template_id == db_template.template_id,
            models.AssessQuestion.is_active.is_(True)
        )
        .order_by(models.AssessQuestion.question_order.asc())
        .all()
    )

    result = []
    for q in questions:
        options = (
            db.query(models.AssessOption)
            .filter(
                models.AssessOption.group_id == q.group_id,
                models.AssessOption.is_active.is_(True)
            )
            .order_by(models.AssessOption.order_index.asc())
            .all()
        )

        q_data = schemas.QuestionWithOptionsResponse(
            content=q.content,
            dimension=q.dimension,
            question_order=q.question_order,
            slug=q.slug,
            options=[
                schemas.OptionResponse.model_validate(opt)
                for opt in options
            ]
        )
        result.append(q_data)

    return result
