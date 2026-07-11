from typing import List
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from ..database import get_db
from .. import models, types
from .dependencies import get_template_by_slug, check_admin_role, generate_unique_slug

router = APIRouter(prefix="/api/v1/assessments/templates", tags=["Templates"])


@router.post("")
def create_template(
    template: types.TemplateCreate,
    db: Session = Depends(get_db)
):
    db_template = (
        db.query(models.AssessTemplate)
        .filter(
            models.AssessTemplate.code == template.code,
            models.AssessTemplate.is_active.is_(True)
        )
        .first()
    )
    if db_template:
        raise HTTPException(
            status_code=400,
            detail="Lỗi: Mã bài test này đã tồn tại!"
        )

    # Generate unique 10-character slug
    slug = generate_unique_slug(models.AssessTemplate, db)

    new_template = models.AssessTemplate(
        code=template.code,
        title=template.title,
        description=template.description,
        instruction=template.instruction,
        certification=template.certification,
        slug=slug
    )

    db.add(new_template)
    db.commit()
    db.refresh(new_template)

    return {
        "message": "Tạo bài test thành công!",
        "data": types.TemplateResponse.model_validate(new_template)
    }


@router.get(
    "",
    response_model=List[types.TemplateResponse]
)
def get_all_templates(db: Session = Depends(get_db)):
    templates = (
        db.query(models.AssessTemplate)
        .filter(models.AssessTemplate.is_active.is_(True))
        .all()
    )
    return templates


@router.get(
    "/{slug}",
    response_model=types.TemplateResponse
)
def get_template(slug: str, db: Session = Depends(get_db)):
    template = get_template_by_slug(slug, db)
    if not template:
        raise HTTPException(
            status_code=404,
            detail="Không tìm thấy bài test này!"
        )
    return template


@router.patch(
    "/{slug}",
    response_model=types.TemplateResponse
)
def update_template(
    slug: str,
    payload: types.TemplateUpdate,
    db: Session = Depends(get_db)
):
    template = get_template_by_slug(slug, db)
    if not template:
        raise HTTPException(
            status_code=404,
            detail="Không tìm thấy bài test để cập nhật!"
        )
    
    update_data = payload.model_dump(exclude_unset=True)
    
    if "code" in update_data and update_data["code"] != template.code:
        duplicate = db.query(models.AssessTemplate).filter(
            models.AssessTemplate.code == update_data["code"],
            models.AssessTemplate.is_active.is_(True)
        ).first()
        if duplicate:
            raise HTTPException(status_code=400, detail="Mã bài test này đã tồn tại!")

    for key, value in update_data.items():
        setattr(template, key, value)

    db.commit()
    db.refresh(template)
    return template


@router.delete("/{slug}")
def delete_template(
    slug: str,
    db: Session = Depends(get_db),
    is_admin: bool = Depends(check_admin_role)
):
    template = get_template_by_slug(slug, db)
    if not template:
        raise HTTPException(
            status_code=404,
            detail="Không tìm thấy bài test để xóa!"
        )
    
    # Soft delete the template
    template.is_active = False
    db.commit()
    return {
        "message": "Xóa bài test thành công (Soft Delete)!"
    }
