from fastapi import APIRouter, Depends
from app.schemas.chat import ChatRequest, ChatResponse
from app.core.orchestrator import main_pipeline
from app.api.dependencies import get_db

router = APIRouter()

@router.post("/chat", response_model=ChatResponse)
async def chat(request: ChatRequest, db=Depends(get_db)):
    """
    Endpoint for sending messages to the mental wellness bot.
    Processes request through Layer 0 -> Layer 5.
    """
    response_text = await main_pipeline(request.message, request.session_id, db)
    return ChatResponse(response=response_text, session_id=request.session_id)

@router.get("/health")
async def health_check():
    return {"status": "healthy"}
