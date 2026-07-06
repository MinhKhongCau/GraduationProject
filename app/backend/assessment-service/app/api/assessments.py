import uuid
from typing import List
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from ..database import get_db
from .. import models, types
from ..services.ai_service import generate_psychological_advice
from .dependencies import get_template_by_slug, get_current_user_id

router = APIRouter(prefix="/api/v1/assessments", tags=["Assessments"])


@router.post("/submit")
def submit_assessment(
    payload: types.AssessmentSubmit,
    db: Session = Depends(get_db),
    user_id: str = Depends(get_current_user_id)
):
    template = get_template_by_slug(payload.template_id, db)
    if not template:
        raise HTTPException(
            status_code=404,
            detail="Không tìm thấy bài test tương ứng!"
        )

    dimension_scores = {}
    total_score = 0
    valid_answers = []

    for ans in payload.answers:
        question = (
            db.query(models.AssessQuestion)
            .filter(
                models.AssessQuestion.slug == ans.question_id,
                models.AssessQuestion.is_active.is_(True)
            )
            .first()
        )
        option = (
            db.query(models.AssessOption)
            .filter(
                models.AssessOption.slug == ans.option_id,
                models.AssessOption.is_active.is_(True)
            )
            .first()
        )

        if not question or not option:
            continue

        dim = question.dimension
        score = option.score_value

        if dim not in dimension_scores:
            dimension_scores[dim] = 0
        dimension_scores[dim] += score
        total_score += score

        valid_answers.append({"question": question, "option": option})

    ai_eval = generate_psychological_advice(dimension_scores)

    new_result = models.AssessResult(
        template_id=template.template_id,
        user_id=uuid.UUID(payload.user_id),
        total_score=total_score,
        dimension_scores=dimension_scores,
        ai_evaluation=ai_eval
    )
    db.add(new_result)
    db.commit()
    db.refresh(new_result)

    for item in valid_answers:
        new_user_ans = models.AssessUserAnswer(
            result_id=new_result.result_id,
            question_id=item["question"].question_id,
            option_id=item["option"].option_id
        )
        db.add(new_user_ans)
    db.commit()

    return {
        "message": "Nộp bài và chấm điểm thành công!",
        "result_id": new_result.result_id,
        "total_score": total_score,
        "dimension_scores": dimension_scores,
        "ai_evaluation": ai_eval
    }


@router.get("/self", response_model=List[types.AssessmentResultResponse])
def get_my_assessments(
    db: Session = Depends(get_db),
    user_id: str = Depends(get_current_user_id)
):
    try:
        user_uuid = uuid.UUID(user_id)
    except ValueError:
        raise HTTPException(
            status_code=400,
            detail="User ID extract from token is not a valid UUID!"
        )

    results = (
        db.query(
            models.AssessResult.result_id,
            models.AssessTemplate.code.label("template_code"),
            models.AssessTemplate.title.label("template_title"),
            models.AssessResult.total_score,
            models.AssessResult.dimension_scores,
            models.AssessResult.ai_evaluation,
            models.AssessResult.created_at
        )
        .join(models.AssessTemplate, models.AssessResult.template_id == models.AssessTemplate.template_id)
        .filter(models.AssessResult.user_id == user_uuid)
        .order_by(models.AssessResult.created_at.desc())
        .all()
    )

    return results
