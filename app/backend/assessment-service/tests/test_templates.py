import uuid
from app.infrastructure.persistence import models


def test_create_template_success(client):
    """TC-ASM-TPL-01 - Tạo bài trắc nghiệm mới thành công (Happy Case)"""
    payload = {
        "code": "MBTI_TEST",
        "title": "Trắc nghiệm tính cách MBTI",
        "description": "Phân loại 16 nhóm tính cách",
        "instruction": "Hãy trả lời trung thực",
        "certification": "Tiêu chuẩn quốc tế"
    }

    response = client.post("/api/v1/assessments/templates", json=payload)
    assert response.status_code == 200
    data = response.json()
    assert data["message"] == "Tạo bài test thành công!"
    assert data["result"]["code"] == "MBTI_TEST"
    assert "slug" in data["result"]
    assert len(data["result"]["slug"]) <= 10


def test_create_template_duplicate_code(client, db_session):
    """TC-ASM-TPL-02 - Tạo bài trắc nghiệm thất bại do trùng lặp mã bài test"""
    template = models.AssessTemplate(
        template_id=uuid.uuid4(),
        code="DUP_TEST",
        title="Bài test mẫu",
        slug="dup1234567",
        is_active=True
    )
    db_session.add(template)
    db_session.commit()

    payload = {
        "code": "DUP_TEST",
        "title": "Bài test trùng mã",
        "description": "Mô tả bài test"
    }

    response = client.post("/api/v1/assessments/templates", json=payload)
    assert response.status_code == 400
    assert response.json()["message"] == "Lỗi: Mã bài test này đã tồn tại!"


def test_get_template_success(client, db_session):
    """TC-ASM-TPL-03 - Xem thông tin chi tiết bài trắc nghiệm theo slug thành công"""
    template = models.AssessTemplate(
        template_id=uuid.uuid4(),
        code="GET_TEST",
        title="Bài test lấy thông tin",
        description="Mô tả chi tiết",
        slug="get1234567",
        is_active=True
    )
    db_session.add(template)
    db_session.commit()

    response = client.get("/api/v1/assessments/templates/get1234567")
    assert response.status_code == 200
    assert response.json()["result"]["title"] == "Bài test lấy thông tin"


def test_update_template_success(client, db_session):
    """TC-ASM-TPL-04 - Cập nhật thông tin bài trắc nghiệm thành công"""
    template = models.AssessTemplate(
        template_id=uuid.uuid4(),
        code="UPDATE_TEST",
        title="Tiêu đề cũ",
        slug="upd1234567",
        is_active=True
    )
    db_session.add(template)
    db_session.commit()

    payload = {"title": "Tiêu đề đã được cập nhật"}
    response = client.patch("/api/v1/assessments/templates/upd1234567", json=payload)
    assert response.status_code == 200
    assert response.json()["result"]["title"] == "Tiêu đề đã được cập nhật"


def test_delete_template_soft_delete(client, db_session):
    """TC-ASM-TPL-05 - Xóa bài trắc nghiệm thành công bằng phương pháp Soft Delete"""
    template = models.AssessTemplate(
        template_id=uuid.uuid4(),
        code="DELETE_TEST",
        title="Bài test bị xóa",
        slug="del1234567",
        is_active=True
    )
    db_session.add(template)
    db_session.commit()

    response = client.delete(
        "/api/v1/assessments/templates/del1234567",
        headers={"X-User-Role": "ADMIN"}
    )
    assert response.status_code == 200
    assert response.json()["message"] == "Xóa bài test thành công (Soft Delete)!"

    db_session.refresh(template)
    assert template.is_active is False


def test_create_bulk_questions_success(client, db_session):
    """TC-ASM-QST-01 - Thêm hàng loạt câu hỏi vào bài trắc nghiệm thành công"""
    template = models.AssessTemplate(
        template_id=uuid.uuid4(), code="Q_TEST", title="Test Q", slug="qtpl123456", is_active=True
    )
    dimension = models.AssessDimension(
        dimension_id=uuid.uuid4(), code="ANXIETY", name="Lo âu", slug="anx1234567", is_active=True
    )
    group = models.AssessOptionGroup(
        group_id=uuid.uuid4(), group_code="SCALE_Q", group_name="Group Q", slug="grpq123456", is_active=True
    )
    db_session.add_all([template, dimension, group])
    db_session.commit()

    payload = {
        "template_id": "qtpl123456",
        "group_id": "grpq123456",
        "questions": [
            {
                "content": "Bạn có cảm thấy lo âu không?",
                "dimension_id": "anx1234567",
                "question_order": 1
            }
        ]
    }

    response = client.post("/api/v1/assessments/questions/bulk", json=payload)
    assert response.status_code == 200
    data = response.json()
    assert "Đã thêm thành công 1 câu hỏi" in data["message"]


def test_create_bulk_questions_dimension_not_found(client, db_session):
    """TC-ASM-QST-02 - Thêm hàng loạt câu hỏi thất bại do khía cạnh không tồn tại"""
    template = models.AssessTemplate(
        template_id=uuid.uuid4(), code="Q_TEST2", title="Test Q2", slug="qtpl789012", is_active=True
    )
    group = models.AssessOptionGroup(
        group_id=uuid.uuid4(), group_code="SCALE_Q2", group_name="Group Q2", slug="grpq789012", is_active=True
    )
    db_session.add_all([template, group])
    db_session.commit()

    payload = {
        "template_id": "qtpl789012",
        "group_id": "grpq789012",
        "questions": [
            {
                "content": "Câu hỏi test",
                "dimension_id": "nodim12345",
                "question_order": 1
            }
        ]
    }

    response = client.post("/api/v1/assessments/questions/bulk", json=payload)
    assert response.status_code == 404
    assert "Không tìm thấy khía cạnh" in response.json()["message"]


def test_delete_question_soft_delete(client, db_session):
    """TC-ASM-QST-04 - Xóa câu hỏi thành công bằng Soft Delete"""
    question = models.AssessQuestion(
        question_id=uuid.uuid4(),
        content="Câu hỏi sắp bị xóa",
        question_order=1,
        slug="delq123456",
        is_active=True
    )
    db_session.add(question)
    db_session.commit()

    response = client.delete(
        "/api/v1/assessments/questions/delq123456",
        headers={"X-User-Role": "ADMIN"}
    )
    assert response.status_code == 200
    assert response.json()["message"] == "Xóa câu hỏi thành công (Soft Delete)!"

    db_session.refresh(question)
    assert question.is_active is False
