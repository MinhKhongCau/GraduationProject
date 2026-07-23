# app/main.py
# pyrefly: ignore [missing-import]
from fastapi import FastAPI, File, UploadFile, Form
# pyrefly: ignore [missing-import]
from fastapi.responses import StreamingResponse
# pyrefly: ignore [missing-import]
from pydantic import BaseModel
# pyrefly: ignore [missing-import]
import uuid
# pyrefly: ignore [missing-import]
import uvicorn
import os
# pyrefly: ignore [missing-import]
import shutil
# pyrefly: ignore [missing-import]
import whisper
import tempfile
# pyrefly: ignore [missing-import]
import numpy as np

from app.core.orchestrator import chat_orchestrator_stream

app = FastAPI(
    title="Clinical AI Mental Wellness API",
    description="API cho hệ thống chatbot hỗ trợ sức khỏe tinh thần có tích hợp Voice & Speech-to-Text",
    version="2.0.0"
)

# Đảm bảo thư mục lưu file tạm tồn tại
os.makedirs("temp_audio", exist_ok=True)

# Load Whisper model một lần khi startup (để không phải load lại mỗi request)
print("🎤 [Whisper] Loading speech-to-text model...")
try:
    whisper_model = whisper.load_model("tiny")  # Options: tiny, base, small, medium, large
    print(f"✅ [Whisper] Model 'tiny' loaded successfully!")
except Exception as e:
    print(f"⚠️ [Whisper] Could not load model: {e}")
    whisper_model = None

class ChatRequest(BaseModel):
    message: str
    session_id: str = None

@app.post("/api/v1/chat/stream")
async def chat_stream_endpoint(req: ChatRequest):
    """Endpoint chat văn bản thông thường."""
    session_id = req.session_id or str(uuid.uuid4())
    generator = chat_orchestrator_stream(
        session_id=session_id, 
        user_message=req.message
    )
    return StreamingResponse(generator, media_type="text/plain")

@app.post("/api/v1/chat/voice")
async def chat_voice_endpoint(
    message: str = Form(None),  # KHÔNG bắt buộc nữa
    session_id: str = Form(None),
    audio_file: UploadFile = File(...)
):
    """
    Endpoint nhận file âm thanh (.wav, .mp3, .m4a) và tin nhắn văn bản (tùy chọn).
    
    - Nếu KHÔNG có message: Tự động dùng Whisper để chuyển giọng nói thành văn bản.
    - Nếu CÓ message: Dùng message được cung cấp.
    """
    session_id = session_id or str(uuid.uuid4())
    
    # 1. Lưu file âm thanh tạm thời
    temp_file_path = f"temp_audio/{session_id}_{audio_file.filename}"
    with open(temp_file_path, "wb") as buffer:
        shutil.copyfileobj(audio_file.file, buffer)
    
    print(f"📁 [Voice] Audio saved: {temp_file_path}")
    print(f"📝 [Voice] User message: {message if message else '(empty - will use Whisper)'}")
    
    # 2. Nếu không có text message, dùng Whisper để transcribe
    if not message or message.strip() == "":
        if whisper_model is not None:
            try:
                print(f"🎤 [Whisper] Transcribing audio...")
                result = whisper_model.transcribe(temp_file_path)
                message = result["text"].strip()
                detected_language = result.get("language", "unknown")
                print(f"✅ [Whisper] Detected language: {detected_language}")
                print(f"✅ [Whisper] Transcribed text: \"{message}\"")
                
                if not message:
                    message = "I need someone to talk to."  # Fallback nếu không nhận diện được
                    print("⚠️ [Whisper] No text detected, using fallback message")
            except Exception as e:
                print(f"❌ [Whisper] Transcription failed: {e}")
                message = "I need someone to talk to."
        else:
            print("⚠️ [Whisper] Model not loaded. Please provide a text message.")
            return StreamingResponse(
                iter(["Error: Speech-to-text model is not available. Please provide a text message with your audio."]), 
                media_type="text/plain"
            )
    
    # 3. Truyền vào Orchestrator
    generator = chat_orchestrator_stream(
        session_id=session_id, 
        user_message=message,
        audio_file=temp_file_path  # Vẫn gửi audio để phân tích cảm xúc giọng nói
    )
    
    # 4. Trả về StreamingResponse
    return StreamingResponse(generator, media_type="text/plain")

@app.post("/api/v1/chat/voice-only")
async def chat_voice_only_endpoint(
    audio_file: UploadFile = File(...),
    session_id: str = Form(None)
):
    """
    Endpoint CHỈ NHẬN AUDIO (không cần text).
    Tự động chuyển giọng nói thành text và phân tích cảm xúc.
    """
    session_id = session_id or str(uuid.uuid4())
    
    # Lưu file
    temp_file_path = f"temp_audio/{session_id}_{audio_file.filename}"
    with open(temp_file_path, "wb") as buffer:
        shutil.copyfileobj(audio_file.file, buffer)
    
    print(f"🎤 [Voice-Only] Audio saved: {temp_file_path}")
    
    # Whisper transcribe
    if whisper_model is not None:
        try:
            print(f"🎤 [Whisper] Transcribing audio...")
            result = whisper_model.transcribe(temp_file_path)
            message = result["text"].strip()
            detected_language = result.get("language", "unknown")
            print(f"✅ [Whisper] Language: {detected_language}, Text: \"{message}\"")
            
            if not message:
                message = "I need someone to talk to."
        except Exception as e:
            print(f"❌ [Whisper] Failed: {e}")
            message = "I need someone to talk to."
    else:
        return StreamingResponse(
            iter(["Error: Speech-to-text model not available."]), 
            media_type="text/plain"
        )
    
    # Orchestrator
    generator = chat_orchestrator_stream(
        session_id=session_id,
        user_message=message,
        audio_file=temp_file_path
    )
    
    return StreamingResponse(generator, media_type="text/plain")

@app.get("/health")
def health_check():
    return {
        "status": "ok",
        "whisper_model": "loaded" if whisper_model is not None else "not loaded",
        "version": "2.0.0"
    }

@app.get("/api/v1/models/info")
def get_model_info():
    """Thông tin về các models đang sử dụng."""
    return {
        "speech_to_text": {
            "model": "whisper-tiny" if whisper_model else None,
            "status": "loaded" if whisper_model else "not loaded",
            "description": "OpenAI Whisper for speech-to-text"
        },
        "text_llm": {
            "model": os.getenv("LLM_MODEL", "qwen2.5:3b"),
            "provider": "Ollama",
            "url": os.getenv("OLLAMA_BASE_URL", "http://ollama:11434")
        },
        "embeddings": {
            "model": "nomic-embed-text",
            "provider": "Ollama"
        },
        "voice_emotion": {
            "model": "wav2vec2-emotion",
            "status": "loaded"
        }
    }

if __name__ == "__main__":
    uvicorn.run("app.main:app", host="0.0.0.0", port=8086, reload=True)