from typing import List
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from app.config.database import get_db
from app.infrastructure.persistence import models
from app.infrastructure.http import schemas
from app.infrastructure.http.dependencies import (
    get_dimension_by_slug,
    get_question_by_slug,
    check_admin_role,
    generate_unique_slug,
)

router = APIRouter(prefix="/api/v1/assessments/dimensions", tags=["Dimensions"])


@router.post("")
def create_dimension(
    payload: schemas.DimensionCreate,
    db: Session = Depends(get_db),
    is_admin: bool = Depends(check_admin_role)
):
    db_dimension = (
        db.query(models.AssessDimension)
        .filter(
            models.AssessDimension.code == payload.code,
            models.AssessDimension.is_active.is_(True)
        )
        .first()
    )
    if db_dimension:
        raise HTTPException(
            status_code=400,
            detail="Lỗi: Mã khía cạnh (dimension) này đã tồn tại!"
        )

    # Generate unique 10-character slug
    slug = generate_unique_slug(models.AssessDimension, db)

    new_dimension = models.AssessDimension(
        code=payload.code,
        name=payload.name,
        description=payload.description,
        slug=slug
    )

    db.add(new_dimension)
    db.commit()
    db.refresh(new_dimension)

    return {
        "message": "Tạo khía cạnh (dimension) thành công!",
        "data": schemas.DimensionResponse.model_validate(new_dimension)
    }


@router.get(
    "",
    response_model=List[schemas.DimensionResponse]
)
def get_all_dimensions(db: Session = Depends(get_db)):
    dimensions = (
        db.query(models.AssessDimension)
        .filter(models.AssessDimension.is_active.is_(True))
        .all()
    )
    return dimensions


@router.get(
    "/{slug}",
    response_model=schemas.DimensionResponse
)
def get_dimension(slug: str, db: Session = Depends(get_db)):
    dimension = get_dimension_by_slug(slug, db)
    if not dimension:
        raise HTTPException(
            status_code=404,
            detail="Không tìm thấy khía cạnh (dimension) này!"
        )
    return dimension


@router.patch(
    "/{slug}",
    response_model=schemas.DimensionResponse
)
def update_dimension(
    slug: str,
    payload: schemas.DimensionUpdate,
    db: Session = Depends(get_db),
    is_admin: bool = Depends(check_admin_role)
):
    dimension = get_dimension_by_slug(slug, db)
    if not dimension:
        raise HTTPException(
            status_code=404,
            detail="Không tìm thấy khía cạnh (dimension) để cập nhật!"
        )

    update_data = payload.model_dump(exclude_unset=True)

    if "code" in update_data and update_data["code"] != dimension.code:
        duplicate = db.query(models.AssessDimension).filter(
            models.AssessDimension.code == update_data["code"],
            models.AssessDimension.is_active.is_(True)
        ).first()
        if duplicate:
            raise HTTPException(status_code=400, detail="Mã khía cạnh (dimension) này đã tồn tại!")

    for key, value in update_data.items():
        setattr(dimension, key, value)

    db.commit()
    db.refresh(dimension)
    return dimension


@router.patch("/{slug}/questions/bulk")
def bulk_assign_questions_to_dimension(
    slug: str,
    payload: schemas.DimensionQuestionsBulkAssign,
    db: Session = Depends(get_db),
    is_admin: bool = Depends(check_admin_role)
):
    dimension = get_dimension_by_slug(slug, db)
    if not dimension:
        raise HTTPException(
            status_code=404,
            detail="Không tìm thấy khía cạnh (dimension) này!"
        )

    if not payload.question_ids:
        raise HTTPException(
            status_code=400,
            detail="Danh sách câu hỏi không được để trống!"
        )

    questions = []
    missing_slugs = []
    for question_slug in payload.question_ids:
        question = get_question_by_slug(question_slug, db)
        if not question:
            missing_slugs.append(question_slug)
            continue
        questions.append(question)

    if missing_slugs:
        raise HTTPException(
            status_code=404,
            detail=f"Không tìm thấy câu hỏi tương ứng với: {', '.join(missing_slugs)}"
        )

    for question in questions:
        question.dimension_id = dimension.dimension_id

    db.commit()

    return {
        "message": f"Đã gán {len(questions)} câu hỏi vào khía cạnh '{dimension.name}'!",
        "data": {
            "dimension": schemas.DimensionResponse.model_validate(dimension),
            "questions": [schemas.QuestionResponse.model_validate(q) for q in questions]
        }
    }


@router.delete("/{slug}")
def delete_dimension(
    slug: str,
    db: Session = Depends(get_db),
    is_admin: bool = Depends(check_admin_role)
):
    dimension = get_dimension_by_slug(slug, db)
    if not dimension:
        raise HTTPException(
            status_code=404,
            detail="Không tìm thấy khía cạnh (dimension) để xóa!"
        )

    # Soft delete the dimension
    dimension.is_active = False
    db.commit()
    return {
        "message": "Xóa khía cạnh (dimension) thành công (Soft Delete)!"
    }
