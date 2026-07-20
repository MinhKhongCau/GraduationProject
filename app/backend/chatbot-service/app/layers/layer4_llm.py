# app/layers/layer4_llm.py
import os
# pyrefly: ignore [missing-import]
from openai import OpenAI

# Khởi tạo kết nối tới Local LLM (Ollama)
ollama_base_url = os.getenv("OLLAMA_BASE_URL", "http://localhost:11434")
client = OpenAI(
    base_url=f"{ollama_base_url}/v1",
    api_key="ollama"
)
MODEL_NAME = "qwen2.5"

# System Prompt kết hợp Rào chắn An toàn (Safety Guardrails)
SYSTEM_PROMPT = """You are a highly empathetic, safe, and professional mental wellness support chatbot.
Your goal is to listen, validate the user's feelings, and suggest coping strategies based ONLY on the provided context.

STRICT RULES:
1. NEVER diagnose a medical or mental health condition.
2. NEVER prescribe or recommend medications.
3. If you don't know the answer or the context doesn't provide enough information, simply empathize and suggest seeking professional help. Do not make up facts.
4. Reply entirely in ENGLISH. Use a warm, supportive, and conversational tone.
"""

def generate_response(user_message: str, rag_context: str) -> str:
    """
    Gọi Local LLM để sinh câu trả lời dựa trên câu hỏi và tài liệu RAG.
    """
    # Lắp ghép tài liệu RAG vào Prompt
    user_prompt = f"""
    Based on the following reference materials, please help and respond to the user.
    If the reference materials do not contain relevant techniques, focus on empathizing with the user.

    --- REFERENCE MATERIALS ---
    {rag_context}
    ---------------------------

    User says: "{user_message}"
    """

    try:
        print("[Layer 4 - LLM] Generating response...")
        response = client.chat.completions.create(
            model=MODEL_NAME,
            messages=[
                {"role": "system", "content": SYSTEM_PROMPT},
                {"role": "user", "content": user_prompt}
            ],
            temperature=0.4, # Nhiệt độ thấp để câu trả lời bám sát RAG
            max_tokens=500,
            presence_penalty=0.6, # Thêm tham số này: Phạt mô hình nếu nó lặp lại các từ đã dùng
            stop=["User:", "\n\n\n", "<|im_end|>"] # Thêm tham số này: Ép mô hình ngắt câu ngay lập tức khi gặp các dấu hiệu kết thúc
        )
        return response.choices[0].message.content.strip()
    except Exception as e:
        print(f"[Layer 4 - LLM] ❌ Lỗi khi gọi LLM: {e}")
        return "I'm sorry, I'm experiencing a technical issue right now. Could you please repeat that?"

# Bổ sung hàm này vào DƯỚI CÙNG của file app/layers/layer4_llm.py

def generate_response_stream(user_message: str, rag_context: str):
    """
    Gọi Local LLM bằng cơ chế Streaming (trả về từng chữ - Generator).
    """
    user_prompt = f"""
    Based on the following reference materials, please help and respond to the user.
    If the reference materials do not contain relevant techniques, focus on empathizing with the user.

    --- REFERENCE MATERIALS ---
    {rag_context}
    ---------------------------

    User says: "{user_message}"
    """

    try:
        response = client.chat.completions.create(
            model=MODEL_NAME,
            messages=[
                {"role": "system", "content": SYSTEM_PROMPT},
                {"role": "user", "content": user_prompt}
            ],
            temperature=0.4,
            max_tokens=500,
            stop=["User:", "\n\n\n", "<|im_end|>"],
            stream=True  # BẬT CHẾ ĐỘ STREAMING
        )
        
        # Dùng 'yield' để trả về từng mảnh chữ (chunk) ngay khi LLM vừa nghĩ ra
        for chunk in response:
            if chunk.choices[0].delta.content is not None:
                yield chunk.choices[0].delta.content
                
    except Exception as e:
        print(f"\n[Layer 4 - LLM] ❌ Lỗi khi gọi LLM: {e}")
        yield "I'm sorry, I'm experiencing a technical issue right now."
        
# ==========================================
# TEST NHANH 
# ==========================================
if __name__ == "__main__":
    # Đã đổi câu hỏi test sang tiếng Anh cho đồng bộ
    test_user_message = "Lately I've been procrastinating a lot because I feel so anxious. What should I do?"
    
    dummy_rag_context = """
    --- TECHNIQUE 1 (Source: Healthy vs. Unhealthy Coping Strategies.pdf) ---
    Coping strategies are actions we take to deal with stress or uncomfortable emotions. 
    Unhealthy coping strategies: Procrastination, Social withdrawal, Overeating.
    Healthy coping strategies: Exercise, Talking about your problem, Relaxation techniques (e.g. deep breathing), Problem-solving techniques.
    """
    
    print("Bắt đầu test Layer 4...")
    ai_response = generate_response(test_user_message, dummy_rag_context)
    
    print("\n" + "="*50)
    print("🤖 KẾT QUẢ TỪ AI (QWEN-2.5):")
    print("="*50)
    print(ai_response)