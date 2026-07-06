import uuid
import secrets
import string
from typing import Optional
from fastapi import Header, HTTPException, Depends
from sqlalchemy.orm import Session
from ..database import get_db
from .. import models

def get_current_user_id(authorization: str = Header(None)):
    if not authorization or not authorization.startswith("Bearer "):
        raise HTTPException(status_code=401, detail="Bạn chưa đăng nhập (Thiếu Token)!")
    token = authorization.split(" ")[1]
    extracted_user_id = token 
    return extracted_user_id


def check_admin_role(
    x_user_role: Optional[str] = Header(None, alias="X-User-Role"),
    authorization: str = Header(None)
):
    if x_user_role and x_user_role.upper() == "ADMIN":
        return True
    if authorization and authorization.startswith("Bearer "):
        token = authorization.split(" ")[1]
        if token.lower() == "admin":
            return True
    raise HTTPException(status_code=403, detail="Chỉ ADMIN mới có quyền thực hiện thao tác này!")


def get_template_by_slug(slug: str, db: Session) -> Optional[models.AssessTemplate]:
    return db.query(models.AssessTemplate).filter(
        models.AssessTemplate.slug == slug,
        models.AssessTemplate.is_active.is_(True)
    ).first()


def get_option_group_by_slug(slug: str, db: Session) -> Optional[models.AssessOptionGroup]:
    return db.query(models.AssessOptionGroup).filter(
        models.AssessOptionGroup.slug == slug,
        models.AssessOptionGroup.is_active.is_(True)
    ).first()


def get_option_by_slug(slug: str, db: Session) -> Optional[models.AssessOption]:
    return db.query(models.AssessOption).filter(
        models.AssessOption.slug == slug,
        models.AssessOption.is_active.is_(True)
    ).first()


def get_question_by_slug(slug: str, db: Session) -> Optional[models.AssessQuestion]:
    return db.query(models.AssessQuestion).filter(
        models.AssessQuestion.slug == slug,
        models.AssessQuestion.is_active.is_(True)
    ).first()


def generate_unique_slug(model_class, db: Session) -> str:
    alphabet = string.ascii_lowercase + string.digits
    while True:
        slug = "".join(secrets.choice(alphabet) for _ in range(10))
        exists = db.query(model_class).filter(model_class.slug == slug).first()
        if not exists:
            return slug
