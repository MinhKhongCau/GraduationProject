from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from ..database import get_db
from .. import models, types
from .dependencies import get_option_group_by_slug, check_admin_role, generate_unique_slug

router = APIRouter(prefix="/api/v1/assessments/option-groups", tags=["Option Groups"])


@router.post("")
def create_option_group(
    group: types.OptionGroupCreate,
    db: Session = Depends(get_db)
):
    db_group = (
        db.query(models.AssessOptionGroup)
        .filter(
            models.AssessOptionGroup.group_code == group.group_code,
            models.AssessOptionGroup.is_active.is_(True)
        )
        .first()
    )

    if db_group:
        raise HTTPException(
            status_code=400,
            detail="Lỗi: Mã nhóm phương án này đã tồn tại!"
        )

    # Generate unique 10-character group slug
    group_slug = generate_unique_slug(models.AssessOptionGroup, db)

    new_group = models.AssessOptionGroup(
        group_code=group.group_code,
        group_name=group.group_name,
        description=group.description,
        slug=group_slug
    )
    db.add(new_group)
    db.commit()
    db.refresh(new_group)

    for opt in group.options:
        # Generate unique 10-character option slug
        opt_slug = generate_unique_slug(models.AssessOption, db)
        new_option = models.AssessOption(
            group_id=new_group.group_id,
            label=opt.label,
            value=opt.value,
            score_value=opt.score_value,
            order_index=opt.order_index,
            slug=opt_slug
        )
        db.add(new_option)

    db.commit()

    return {
        "message": "Tạo nhóm phương án và các tùy chọn thành công!",
        "group_id": new_group.group_id
    }


@router.get(
    "/{slug}",
    response_model=types.OptionGroupResponse
)
def get_option_group(slug: str, db: Session = Depends(get_db)):
    group = get_option_group_by_slug(slug, db)
    if not group:
        raise HTTPException(
            status_code=404,
            detail="Không tìm thấy nhóm phương án này!"
        )
    options = (
        db.query(models.AssessOption)
        .filter(
            models.AssessOption.group_id == group.group_id,
            models.AssessOption.is_active.is_(True)
        )
        .order_by(models.AssessOption.order_index.asc())
        .all()
    )
    return types.OptionGroupResponse(
        group_code=group.group_code,
        group_name=group.group_name,
        description=group.description,
        slug=group.slug,
        options=[types.OptionResponse.model_validate(opt) for opt in options]
    )


@router.patch(
    "/{slug}",
    response_model=types.OptionGroupResponse
)
def update_option_group(
    slug: str,
    payload: types.OptionGroupUpdate,
    db: Session = Depends(get_db)
):
    group = get_option_group_by_slug(slug, db)
    if not group:
        raise HTTPException(
            status_code=404,
            detail="Không tìm thấy nhóm phương án để cập nhật!"
        )
    
    update_data = payload.model_dump(exclude_unset=True)

    if "group_code" in update_data and update_data["group_code"] != group.group_code:
        duplicate = db.query(models.AssessOptionGroup).filter(
            models.AssessOptionGroup.group_code == update_data["group_code"],
            models.AssessOptionGroup.is_active.is_(True)
        ).first()
        if duplicate:
            raise HTTPException(status_code=400, detail="Mã nhóm phương án này đã tồn tại!")

    for key, value in update_data.items():
        setattr(group, key, value)

    db.commit()
    db.refresh(group)
    
    options = (
        db.query(models.AssessOption)
        .filter(
            models.AssessOption.group_id == group.group_id,
            models.AssessOption.is_active.is_(True)
        )
        .order_by(models.AssessOption.order_index.asc())
        .all()
    )
    return types.OptionGroupResponse(
        group_code=group.group_code,
        group_name=group.group_name,
        description=group.description,
        slug=group.slug,
        options=[types.OptionResponse.model_validate(opt) for opt in options]
    )


@router.delete("/{slug}")
def delete_option_group(
    slug: str,
    db: Session = Depends(get_db),
    is_admin: bool = Depends(check_admin_role)
):
    group = get_option_group_by_slug(slug, db)
    if not group:
        raise HTTPException(
            status_code=404,
            detail="Không tìm thấy nhóm phương án để xóa!"
        )
    
    # Soft delete group and all associated options
    group.is_active = False
    db.query(models.AssessOption).filter(
        models.AssessOption.group_id == group.group_id
    ).update({"is_active": False}, synchronize_session=False)
    db.commit()
    return {
        "message": "Xóa nhóm phương án thành công (Soft Delete)!"
    }
