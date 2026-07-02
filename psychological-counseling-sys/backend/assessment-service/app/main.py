from fastapi import FastAPI, Depends, HTTPException, Header
from sqlalchemy.orm import Session
from .database import engine, get_db
from . import models, schemas
from .services.ai_service import generate_psychological_advice

app = FastAPI(
    title="MindCare Assessment Service API",
    description="Tài liệu API cho hệ thống Đánh giá Tâm lý (Assessment Service)",
    version="1.0.0",
    docs_url="/docs",
    redoc_url="/redoc"
)

# Lệnh này tương tự như ddl-auto: update bên Spring Boot
models.Base.metadata.create_all(bind=engine)


def get_current_user_id(authorization: str = Header(None)):
    if not authorization or not authorization.startswith("Bearer "):
        raise HTTPException(status_code=401, detail="Bạn chưa đăng nhập (Thiếu Token)!")
    # 1. Lấy token cắt bỏ chữ Bearer
    token = authorization.split(" ")[1]
    extracted_user_id = token 
    return extracted_user_id


@app.get("/")
def read_root():
    return {
        "message": "Assessment Service is running!"
    }


@app.post("/api/v1/assessments/templates")
def create_template(
    template: schemas.TemplateCreate,
    db: Session = Depends(get_db)
):
    # 1. Kiểm tra xem mã Code này đã tồn tại chưa
    db_template = (
        db.query(models.AssessTemplate)
        .filter(models.AssessTemplate.code == template.code)
        .first()
    )
    if db_template:
        raise HTTPException(
            status_code=400,
            detail="Lỗi: Mã bài test này đã tồn tại!"
        )

    # 2. Tạo Entity mới để lưu xuống DB
    new_template = models.AssessTemplate(
        code=template.code,
        title=template.title,
        description=template.description
    )

    # 3. Lưu và Commit
    db.add(new_template)
    db.commit()
    # Lấy lại data sau khi lưu (để có được template_id tự sinh)
    db.refresh(new_template)

    return {
        "message": "Tạo bài test thành công!",
        "data": new_template
    }


@app.post("/api/v1/assessments/option-groups")
def create_option_group(
    group: schemas.OptionGroupCreate,
    db: Session = Depends(get_db)
):
    # 1. Kiểm tra xem mã Group này đã tồn tại chưa
    db_group = (
        db.query(models.AssessOptionGroup)
        .filter(models.AssessOptionGroup.group_code == group.group_code)
        .first()
    )

    if db_group:
        raise HTTPException(
            status_code=400,
            detail="Lỗi: Mã nhóm phương án này đã tồn tại!"
        )

    # 2. Tạo Nhóm phương án (Group)
    new_group = models.AssessOptionGroup(
        group_code=group.group_code,
        group_name=group.group_name,
        description=group.description
    )
    db.add(new_group)
    db.commit()
    db.refresh(new_group)

    # 3. Quét danh sách các Options gửi kèm và lưu vào DB
    for opt in group.options:
        new_option = models.AssessOption(
            group_id=new_group.group_id,
            label=opt.label,
            value=opt.value,
            score_value=opt.score_value,
            order_index=opt.order_index
        )
        db.add(new_option)

    db.commit()

    return {
        "message": "Tạo nhóm phương án và các tùy chọn thành công!",
        "group_id": new_group.group_id
    }


@app.post("/api/v1/assessments/questions/bulk")
def create_bulk_questions(
    payload: schemas.QuestionBulkCreate,
    db: Session = Depends(get_db)
):
    added_questions = []

    # Duyệt qua từng câu hỏi gửi lên và tạo Entity
    for q in payload.questions:
        new_question = models.AssessQuestion(
            template_id=payload.template_id,
            group_id=payload.group_id,
            content=q.content,
            dimension=q.dimension,
            question_order=q.question_order,
            is_required=True
        )
        db.add(new_question)
        added_questions.append(new_question)

    db.commit()

    return {
        "message": f"Đã thêm thành công {len(added_questions)} "
                   f"câu hỏi vào bài test!"
    }


@app.post("/api/v1/assessments/submit")
def submit_assessment(
    payload: schemas.AssessmentSubmit,
    db: Session = Depends(get_db),
    user_id: str = Depends(get_current_user_id)
):
    dimension_scores = {}
    total_score = 0
    valid_answers = []

    # 1. Duyệt qua từng câu trả lời gửi lên để tính toán
    for ans in payload.answers:
        # Kéo câu hỏi và đáp án từ DB lên để soi điểm
        question = (
            db.query(models.AssessQuestion)
            .filter(models.AssessQuestion.question_id == ans.question_id)
            .first()
        )
        option = (
            db.query(models.AssessOption)
            .filter(models.AssessOption.option_id == ans.option_id)
            .first()
        )

        if not question or not option:
            continue

        dim = question.dimension  # Ví dụ: "STRESS"
        score = option.score_value  # Ví dụ: 3 điểm

        # 2. Cộng điểm vào đúng "rổ" mảng bệnh
        if dim not in dimension_scores:
            dimension_scores[dim] = 0
        dimension_scores[dim] += score
        total_score += score

        # Lưu trữ lại để tí nữa insert db
        valid_answers.append({"question": question, "option": option})

    # Tạo nhận xét AI dựa trên điểm số đã chấm.
    ai_eval = generate_psychological_advice(dimension_scores)

    # 4. Lưu Bản ghi Kết quả (AssessResult)
    new_result = models.AssessResult(
        template_id=payload.template_id,
        user_id=payload.user_id,
        total_score=total_score,
        dimension_scores=dimension_scores,
        ai_evaluation=ai_eval
    )
    db.add(new_result)
    db.commit()
    db.refresh(new_result)

    # 5. Lưu Lịch sử Từng câu trả lời (AssessUserAnswer)
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


# ==========================================
# 1. API LẤY TẤT CẢ BÀI TEST (CHO MÀN HÌNH DANH SÁCH)
# ==========================================
@app.get(
    "/api/v1/assessments/templates",
    response_model=list[schemas.TemplateResponse]
)
def get_all_templates(db: Session = Depends(get_db)):
    templates = (
        db.query(models.AssessTemplate)
        .filter(models.AssessTemplate.is_active is True)
        .all()
    )
    return templates


# ==========================================
# 2. API LẤY TOÀN BỘ CÂU HỎI & OPTION (CHO MÀN HÌNH LÀM BÀI)
# ==========================================
@app.get(
    "/api/v1/assessments/templates/{template_id}/questions",
    response_model=list[schemas.QuestionWithOptionsResponse]
)
def get_questions_by_template(
    template_id: str,
    db: Session = Depends(get_db)
):
    # 1. Lấy tất cả câu hỏi thuộc bài test này, sắp xếp theo thứ tự
    questions = (
        db.query(models.AssessQuestion)
        .filter(models.AssessQuestion.template_id == template_id)
        .order_by(models.AssessQuestion.question_order.asc())
        .all()  # ĐỔI order_index THÀNH order_by
    )

    result = []
    for q in questions:
        # 2. Với mỗi câu hỏi, lấy các options thuộc group của nó
        options = (
            db.query(models.AssessOption)
            .filter(models.AssessOption.group_id == q.group_id)
            .order_by(models.AssessOption.order_index.asc())
            .all()
        )

        # 3. Đóng gói dữ liệu
        q_data = schemas.QuestionWithOptionsResponse(
            question_id=q.question_id,
            content=q.content,
            dimension=q.dimension,
            question_order=q.question_order,
            # Đổi chữ 'from_attributes' thành 'model_validate'
            options=[
                schemas.OptionResponse.model_validate(opt)
                for opt in options
            ]
        )
        result.append(q_data)

    return result
