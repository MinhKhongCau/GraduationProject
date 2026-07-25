import uuid
import pytest
from unittest.mock import patch
from app import models

USER_UUID = "11111111-1111-1111-1111-111111111111"


def setup_assessment_data(db_session):
    template = models.AssessTemplate(
        template_id=uuid.uuid4(),
        code="DASS21",
        title="Trắc nghiệm DASS-21",
        description="Đánh giá Trầm cảm, Lo âu, Căng thẳng",
        instruction="Hãy chọn đáp án phù hợp nhất",
        slug="dass-21",
        is_active=True
    )
    db_session.add(template)

    dimension = models.AssessDimension(
        dimension_id=uuid.uuid4(),
        code="DEPRESSION",
        name="Trầm cảm",
        slug="dep-dim",
        is_active=True
    )
    db_session.add(dimension)

    group = models.AssessOptionGroup(
        group_id=uuid.uuid4(),
        group_code="SCALE_4",
        group_name="Thang 4 mức độ",
        description="0-3 points",
        slug="scale-4",
        is_active=True
    )
    db_session.add(group)

    option = models.AssessOption(
        option_id=uuid.uuid4(),
        group_id=group.group_id,
        label="Rất đúng",
        value="3",
        score_value=3,
        order_index=3,
        slug="opt-3",
        is_active=True
    )
    db_session.add(option)

    question = models.AssessQuestion(
        question_id=uuid.uuid4(),
        template_id=template.template_id,
        group_id=group.group_id,
        dimension_id=dimension.dimension_id,
        content="Tôi cảm thấy khó mà làm mình dịu đi",
        question_order=1,
        slug="q1-slug",
        is_active=True
    )
    db_session.add(question)
    db_session.commit()

    return template, question, option


def test_submit_assessment_success(client, db_session):
    """TC-ASM-SUB-01 - Nộp bài trắc nghiệm và chấm điểm thành công (Happy Case)"""
    template, question, option = setup_assessment_data(db_session)

    payload = {
        "template_id": "dass-21",
        "answers": [
            {
                "question_id": "q1-slug",
                "option_id": "opt-3",
                "question_content": "Tôi cảm thấy khó mà làm mình dịu đi",
                "option_label": "Rất đúng"
            }
        ]
    }

    with patch("app.api.assessments.generate_psychological_advice", return_value="Nhận xét từ AI: Bạn có mức độ trầm cảm nhẹ."):
        response = client.post(
            "/api/v1/assessments/submit",
            json=payload,
            headers={"X-User-Id": USER_UUID}
        )

    assert response.status_code == 200
    data = response.json()
    assert data["message"] == "Nộp bài và chấm điểm thành công!"
    result = data["result"]
    assert result["total_score"] == 3
    assert result["dimension_scores"] == {"DEPRESSION": 3}
    assert "result_id" in result


def test_submit_assessment_template_not_found(client):
    """TC-ASM-SUB-02 - Nộp bài thất bại do mã bài trắc nghiệm không tồn tại"""
    payload = {
        "template_id": "non-existent-slug",
        "answers": []
    }

    response = client.post(
        "/api/v1/assessments/submit",
        json=payload,
        headers={"X-User-Id": USER_UUID}
    )

    assert response.status_code == 404
    assert response.json()["message"] == "Không tìm thấy bài test tương ứng!"


def test_submit_assessment_unauthorized(client):
    """TC-ASM-SUB-04 - Nộp bài thất bại do thiếu header định danh người dùng"""
    payload = {
        "template_id": "dass-21",
        "answers": []
    }

    response = client.post(
        "/api/v1/assessments/submit",
        json=payload
    )

    assert response.status_code == 401


def test_get_my_assessments_success(client, db_session):
    """TC-ASM-HIS-01 - Xem danh sách kết quả trắc nghiệm cá nhân thành công"""
    template, _, _ = setup_assessment_data(db_session)

    result = models.AssessResult(
        result_id=uuid.uuid4(),
        template_id=template.template_id,
        user_id=uuid.UUID(USER_UUID),
        total_score=15,
        dimension_scores={"DEPRESSION": 15},
        ai_evaluation="Tâm lý bình thường"
    )
    db_session.add(result)
    db_session.commit()

    response = client.get(
        "/api/v1/assessments/self",
        headers={"X-User-Id": USER_UUID}
    )

    assert response.status_code == 200
    envelope = response.json()
    results = envelope["result"]
    assert len(results) >= 1
    assert results[0]["total_score"] == 15
    assert results[0]["template_code"] == "DASS21"


def test_get_my_assessments_invalid_uuid(client):
    """TC-ASM-HIS-02 - Xem danh sách kết quả thất bại do User ID không hợp lệ"""
    response = client.get(
        "/api/v1/assessments/self",
        headers={"X-User-Id": "invalid-uuid"}
    )

    assert response.status_code == 400
    assert response.json()["message"] == "User ID extract from token is not a valid UUID!"
