from fastapi import FastAPI
from .api import (
    templates_router,
    option_groups_router,
    options_router,
    questions_router,
    assessments_router,
)

app = FastAPI(title="MindCare Assessment Service", docs_url="/swagger-ui")


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
app.include_router(option_groups_router)
app.include_router(options_router)
app.include_router(questions_router)
app.include_router(assessments_router)
