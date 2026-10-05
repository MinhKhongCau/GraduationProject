from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from app.config.database import get_db
from app.infrastructure.persistence import models
from app.infrastructure.http import schemas
from app.infrastructure.http.dependencies import get_option_by_slug, check_admin_role

router = APIRouter(prefix="/api/v1/assessments/options", tags=["Options"])


@router.get(
    "/{slug}",
    response_model=schemas.OptionResponse
)
def get_option(slug: str, db: Session = Depends(get_db)):
    option = get_option_by_slug(slug, db)
    if not option:
        raise HTTPException(
            status_code=404,
            detail="Không tìm thấy phương án trả lời này!"
        )
    return option


@router.patch(
    "/{slug}",
    response_model=schemas.OptionResponse
)
def update_option(
    slug: str,
    payload: schemas.OptionUpdate,
    db: Session = Depends(get_db)
):
    option = get_option_by_slug(slug, db)
    if not option:
        raise HTTPException(
            status_code=404,
            detail="Không tìm thấy phương án trả lời để cập nhật!"
        )
    
    update_data = payload.model_dump(exclude_unset=True)

    for key, value in update_data.items():
        setattr(option, key, value)

    db.commit()
    db.refresh(option)
    return option


@router.delete("/{slug}")
def delete_option(
    slug: str,
    db: Session = Depends(get_db),
    is_admin: bool = Depends(check_admin_role)
):
    option = get_option_by_slug(slug, db)
    if not option:
        raise HTTPException(
            status_code=404,
            detail="Không tìm thấy phương án trả lời để xóa!"
        )
    
    option.is_active = False
    db.commit()
    return {
        "message": "Xóa phương án trả lời thành công (Soft Delete)!"
    }
