from fastapi import FastAPI
from fastapi.openapi.utils import get_openapi
from .middleware import ResponseEnvelopeMiddleware
from .api import (
    templates_router,
    dimensions_router,
    option_groups_router,
    options_router,
    questions_router,
    assessments_router,
)

OPENAPI_TAGS = [
    {
        "name": "Templates",
        "description": "Quản lý bài test (mã, tiêu đề, mô tả, hướng dẫn, chứng nhận).",
    },
    {
        "name": "Dimensions",
        "description": (
            "Quản lý khía cạnh (thang đo) để gán cho câu hỏi trong bài "
            "test, ví dụ: Trầm cảm, Lo âu, Căng thẳng."
        ),
    },
    {
        "name": "Option Groups",
        "description": "Quản lý nhóm phương án trả lời dùng chung cho câu hỏi.",
    },
    {
        "name": "Options",
        "description": "Quản lý từng phương án trả lời trong một nhóm.",
    },
    {
        "name": "Questions",
        "description": "Quản lý câu hỏi của từng bài test và khía cạnh được gán.",
    },
    {
        "name": "Assessments",
        "description": "Nộp bài làm, chấm điểm và xem kết quả kèm nhận xét từ AI.",
    },
]

app = FastAPI(
    title="MindCare Assessment Service API",
    version="1.0",
    description=(
        "Dịch vụ quản lý bài đánh giá tâm lý (Assessments) của MindCare: "
        "bài test, khía cạnh, nhóm phương án trả lời, câu hỏi, và kết quả "
        "làm bài kèm nhận xét từ AI."
    ),
    docs_url="/swagger-ui",
    openapi_tags=OPENAPI_TAGS,
)


def custom_openapi():
    if app.openapi_schema:
        return app.openapi_schema

    schema = get_openapi(
        title=app.title,
        version=app.version,
        description=app.description,
        routes=app.routes,
        tags=app.openapi_tags,
    )
    schema.setdefault("components", {})["securitySchemes"] = {
        "BearerAuth": {
            "type": "http",
            "scheme": "bearer",
            "description": "Gửi token dạng `Authorization: Bearer <token>`.",
        }
    }
    schema["security"] = [{"BearerAuth": []}]
    app.openapi_schema = schema
    return app.openapi_schema


app.openapi = custom_openapi

app.add_middleware(ResponseEnvelopeMiddleware)


@app.get("/")
def read_root():
    return {
        "message": "Assessment Service is running!"
    }


@app.get("/health")
def health():
    return {
        "service": "assessment-service",
        "status": "up and running"
    }


# Register all routers
app.include_router(templates_router)
app.include_router(dimensions_router)
app.include_router(option_groups_router)
app.include_router(options_router)
app.include_router(questions_router)
app.include_router(assessments_router)
