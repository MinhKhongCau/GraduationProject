# app/layers/layer4_llm.py
import os
# pyrefly: ignore [missing-import]
from openai import OpenAI
# pyrefly: ignore [missing-import]
import httpx

# 1. CẬP NHẬT URL NGROK (Làm giá trị mặc định nếu biến môi trường bị trống)
ollama_base_url = "https://encroach-granular-ridden.ngrok-free.dev"
# Thêm timeout để không bị treo vĩnh viễn
http_client = httpx.Client(timeout=90.0)  # 90 giây timeout

# 2. BỔ SUNG HEADER ĐỂ VƯỢT LỚP CHẶN CỦA NGROK
client = OpenAI(
    base_url=f"{ollama_base_url}/v1",
    api_key="ollama",
    default_headers={"ngrok-skip-browser-warning": "true"}, # <--- Rất quan trọng
    http_client=http_client
)

# 3. ĐỔI TÊN MODEL KHỚP VỚI BẢN GGUF BẠN VỪA TẠO
MODEL_NAME = "qwen2.5-7b-custom:latest" 

# System Prompt giữ nguyên...
SYSTEM_PROMPT = """You are a highly empathetic, safe, and professional mental wellness support chatbot.
Your goal is to listen, validate the user's feelings, and suggest coping strategies based ONLY on the provided context.

STRICT RULES:
1. NEVER diagnose a medical or mental health condition.
2. NEVER prescribe or recommend medications.
3. If you don't know the answer or the context doesn't provide enough information, simply empathize and suggest seeking professional help. Do not make up facts.
4. Reply entirely in ENGLISH. Use a warm, supportive, and conversational tone.
5. Keep responses concise and brief (2-4 sentences max).
"""

def generate_response_stream(user_message: str, rag_context: str):
    """Gọi Local LLM bằng cơ chế Streaming"""
    user_prompt = f"""
    Based on the following reference materials, please help and respond to the user.
    If the reference materials do not contain relevant techniques, focus on empathizing with the user.

    --- REFERENCE MATERIALS ---
    {rag_context}
    ---------------------------

    User says: "{user_message}"
    
    Provide a brief, warm, and helpful response in 2-4 sentences:
    """

    try:
        print(f"[Layer 4 - LLM] Generating response with {MODEL_NAME}...")
        response = client.chat.completions.create(
            model=MODEL_NAME,
            messages=[
                {"role": "system", "content": SYSTEM_PROMPT},
                {"role": "user", "content": user_prompt}
            ],
            temperature=0.4,
            max_tokens=200, 
            stop=["User:", "\n\n\n", "<|im_end|>"],
            stream=True,
            timeout=60  
        )
        
        for chunk in response:
            if chunk.choices[0].delta.content is not None:
                yield chunk.choices[0].delta.content
                
    except Exception as e:
        print(f"\n[Layer 4 - LLM] ❌ Lỗi khi gọi LLM: {e}")
        yield "I'm sorry, I'm experiencing a technical issue right now."

def generate_response(user_message: str, rag_context: str) -> str:
    """Non-streaming version cho test"""
    return "".join(generate_response_stream(user_message, rag_context))