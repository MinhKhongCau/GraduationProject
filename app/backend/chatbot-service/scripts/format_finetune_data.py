# scripts/format_finetune_data.py
import json
import os

RAW_FILE = "data/finetune_sets/raw_conversations/esconv_raw_train.jsonl"
OUTPUT_FILE = "data/finetune_sets/formatted/esconv_chatml_train.jsonl"

# SYSTEM PROMPT (Layer 4 Guardrail)
SYSTEM_PROMPT = (
    "You are an empathetic and safe mental wellness support chatbot. "
    "Your task is to listen, empathize, and suggest emotional management techniques. "
    "Absolutely do not diagnose illnesses, prescribe medication, or replace a medical professional. "
    "Always respond in a concise, warm, and supportive tone."
)

def format_esconv_to_chatml():
    # Đảm bảo thư mục đầu ra tồn tại
    os.makedirs(os.path.dirname(OUTPUT_FILE), exist_ok=True)
    
    formatted_count = 0
    
    print("Đang xử lý và định dạng dữ liệu ESConv...")
    with open(RAW_FILE, "r", encoding="utf-8") as infile, \
         open(OUTPUT_FILE, "w", encoding="utf-8") as outfile:
        
        for line in infile:
            if not line.strip():
                continue
            data = json.loads(line)
            
            # Nếu dữ liệu thực tế bị bọc trong chuỗi JSON ở key "text"
            if "text" in data and isinstance(data["text"], str):
                try:
                    data = json.loads(data["text"])
                except Exception:
                    pass
            
            dialog = data.get("dialog", [])
            
            # Khởi tạo một phiên hội thoại mới với System Prompt
            messages = [{"role": "system", "content": SYSTEM_PROMPT}]
            
            # ESConv lưu từng câu nói trong mảng 'dialog'
            for turn in dialog:
                speaker = turn.get("speaker")
                content = turn.get("text") or turn.get("content")
                
                # Bỏ qua các câu rỗng
                if not content or content.strip() == "":
                    continue
                    
                # Ánh xạ vai trò: seeker/usr -> user, supporter/sys -> assistant
                if speaker in ("seeker", "usr"):
                    messages.append({"role": "user", "content": content})
                elif speaker in ("supporter", "sys"):
                    messages.append({"role": "assistant", "content": content})
            
            # Chỉ lưu những hội thoại có ít nhất 1 câu hỏi và 1 câu trả lời
            if len(messages) >= 3: 
                # Ghi ra file định dạng jsonl chuẩn
                outfile.write(json.dumps({"messages": messages}, ensure_ascii=False) + "\n")
                formatted_count += 1

    print(f"✅ Hoàn tất! Đã tạo ra {formatted_count} cặp hội thoại chuẩn ChatML.")
    print(f"File được lưu tại: {OUTPUT_FILE}")

if __name__ == "__main__":
    format_esconv_to_chatml()