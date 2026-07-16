# app/main.py
# pyrefly: ignore [missing-import]
from fastapi import FastAPI, File, UploadFile, Form
# pyrefly: ignore [missing-import]
from fastapi.responses import StreamingResponse
# pyrefly: ignore [missing-import]
from pydantic import BaseModel
import uuid
# pyrefly: ignore [missing-import]
import uvicorn
import os
import shutil

from app.core.orchestrator import chat_orchestrator_stream

app = FastAPI(
    title="Clinical AI Mental Wellness API",
    description="API cho hệ thống chatbot hỗ trợ sức khỏe tinh thần có tích hợp Voice",
    version="1.0.0"
)

# Đảm bảo thư mục lưu file tạm tồn tại
os.makedirs("temp_audio", exist_ok=True)

class ChatRequest(BaseModel):
    message: str
    session_id: str = None

@app.post("/api/chat/stream")
async def chat_stream_endpoint(req: ChatRequest):
    """Endpoint chat văn bản thông thường."""
    session_id = req.session_id or str(uuid.uuid4())
    generator = chat_orchestrator_stream(session_id=session_id, user_message=req.message)
    return StreamingResponse(generator, media_type="text/plain")

@app.post("/api/chat/voice")
async def chat_voice_endpoint(
    message: str = Form(...), 
    session_id: str = Form(None),
    audio_file: UploadFile = File(...)
):
    """
    Endpoint nhận file âm thanh (.wav) và tin nhắn văn bản.
    """
    session_id = session_id or str(uuid.uuid4())
    
    # 1. Lưu file âm thanh tạm thời vào ổ cứng để Layer 1 đọc
    temp_file_path = f"temp_audio/{session_id}_{audio_file.filename}"
    with open(temp_file_path, "wb") as buffer:
        shutil.copyfileobj(audio_file.file, buffer)
        
    # 2. Truyền đường dẫn file audio vào Orchestrator
    generator = chat_orchestrator_stream(
        session_id=session_id, 
        user_message=message,
        audio_file=temp_file_path
    )
    
    # Trả về StreamingResponse giống như Text
    return StreamingResponse(generator, media_type="text/plain")

@app.get("/health")
def health_check():
    return {"status": "ok"}

if __name__ == "__main__":
    uvicorn.run("app.main:app", host="0.0.0.0", port=8000, reload=True)