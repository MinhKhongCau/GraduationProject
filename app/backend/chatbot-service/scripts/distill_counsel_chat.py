# scripts/distill_counsel_chat.py
import json
import os
# pyrefly: ignore [missing-import]
from openai import OpenAI

# 1. CẤU HÌNH KẾT NỐI ĐẾN OLLAMA LOCAL
# Đảm bảo bạn đang mở app Ollama hoặc đã chạy lệnh: ollama run qwen2.5
client = OpenAI(
    base_url="http://localhost:11434/v1",
    api_key="ollama" # Bắt buộc có chuỗi bất kỳ, nhưng Ollama không kiểm tra
)

# Tên mô hình chính xác bạn vừa tải
MODEL_NAME = "qwen2.5"

RAW_FILE = "data/finetune_sets/raw_conversations/counsel_chat_raw_train.jsonl"
OUTPUT_FILE = "data/finetune_sets/formatted/counsel_chat_safe_train.jsonl"

# Prompt hướng dẫn LLM cách "tẩy rửa" dữ liệu
REWRITE_PROMPT = """
Below is a real response from a psychologist to a patient. 
Your task is to ACT AS AN AI ASSISTANT and REWRITE this response so that it strictly follows these rules:
1. Maintain a high level of empathy and understanding.
2. ABSOLUTELY NO medical diagnosis (e.g., do not say "you have depression" or "you suffer from anxiety").
3. ABSOLUTELY NO medication prescriptions or medical treatments.
4. If the original response contains a diagnosis or prescription, replace it with gentle advice to seek a licensed medical professional or suggest basic stress management techniques (like deep breathing).
5. Respond in English using a warm, supportive, and conversational tone.

Original response from the psychologist:
"{original_answer}"
"""

SYSTEM_PROMPT = (
    "You are an empathetic and safe mental wellness support chatbot. "
    "You must never diagnose illnesses or prescribe medication."
)

def distill_counsel_chat(limit=None):
    os.makedirs(os.path.dirname(OUTPUT_FILE), exist_ok=True)
    
    print(f"Bắt đầu quá trình tẩy rửa bằng Local LLM ({MODEL_NAME})...")
    if limit:
        print(f"Sẽ xử lý tối đa: {limit} mẫu.")
    else:
        print("Sẽ xử lý TOÀN BỘ dữ liệu trong file.")
        
    processed_count = 0
    
    with open(RAW_FILE, "r", encoding="utf-8") as in_f, \
         open(OUTPUT_FILE, "a", encoding="utf-8") as out_f:
        
        for line in in_f:
            if limit and processed_count >= limit:
                break
                
            data = json.loads(line)
            question = data.get("questionText", "")
            original_answer = data.get("answerText", "")
            
            if not question or not original_answer:
                continue
                
            try:
                # Gọi trực tiếp Qwen-2.5 qua Ollama API
                response = client.chat.completions.create(
                    model=MODEL_NAME,
                    messages=[
                        {"role": "user", "content": REWRITE_PROMPT.format(original_answer=original_answer)}
                    ],
                    temperature=0.3
                )
                
                safe_answer = response.choices[0].message.content.strip()
                
                # Tạo format ChatML lưu vào file
                chatml_entry = {
                    "messages": [
                        {"role": "system", "content": SYSTEM_PROMPT},
                        {"role": "user", "content": question},
                        {"role": "assistant", "content": safe_answer}
                    ]
                }
                
                out_f.write(json.dumps(chatml_entry, ensure_ascii=False) + "\n")
                out_f.flush() # Lưu ngay xuống đĩa
                
                processed_count += 1
                print(f"✅ Đã xử lý an toàn mẫu {processed_count}")
                
                # KHÔNG CẦN time.sleep() VÌ CHẠY LOCAL KHÔNG BỊ GIỚI HẠN RATE LIMIT NỮA!
                
            except Exception as e:
                print(f"❌ Lỗi ở mẫu {processed_count + 1}: {e}")
                print("💡 Đảm bảo rằng Ollama đang chạy ngầm (thử gõ 'ollama ps' trên terminal khác).")
                break # Dừng lại ngay nếu lỗi mất kết nối với Ollama

    print(f"🎉 Hoàn tất! Dữ liệu sạch được lưu tại: {OUTPUT_FILE}")

if __name__ == "__main__":
    # Hiện tại tôi set limit=None để nó chạy HẾT TOÀN BỘ file. 
    # Nếu muốn test nhanh 10 dòng, bạn sửa thành: distill_counsel_chat(limit=10)
    distill_counsel_chat(limit=None)