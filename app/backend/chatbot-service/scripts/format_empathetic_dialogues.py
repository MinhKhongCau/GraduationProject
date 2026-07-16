# scripts/format_empathetic_dialogues.py
import json
import os
from collections import defaultdict

RAW_FILE = "data/finetune_sets/raw_conversations/empathetic_dialogues_raw_train.jsonl"
OUTPUT_FILE = "data/finetune_sets/formatted/empathetic_dialogues_chatml_train.jsonl"

# System Prompt tập trung vào sự thấu cảm (nhẹ nhàng hơn, không mang tính y khoa)
SYSTEM_PROMPT = (
    "You are a good listener and a highly empathetic friend. "
    "Your task is to share and empathize with the user's everyday stories, "
    "helping them feel understood and unjudged."
)

def format_empathetic_dialogues():
    os.makedirs(os.path.dirname(OUTPUT_FILE), exist_ok=True)
    
    # Dùng dictionary để gom nhóm các câu nói theo ID hội thoại
    conversations = defaultdict(list)
    
    print("Đang đọc và gom nhóm dữ liệu Empathetic Dialogues...")
    with open(RAW_FILE, "r", encoding="utf-8") as f:
        for line in f:
            data = json.loads(line)
            conv_id = data.get("conv_id")
            utterance_idx = data.get("utterance_idx")
            prompt = data.get("prompt") # Ngữ cảnh cảm xúc ban đầu
            utterance = data.get("utterance")
            
            # Bỏ qua các dòng bị lỗi string thay thế của tập dataset này
            if "_comma_" in utterance:
                utterance = utterance.replace("_comma_", ",")
                
            conversations[conv_id].append({
                "idx": utterance_idx,
                "text": utterance
            })

    formatted_count = 0
    print("Đang chuyển đổi sang định dạng ChatML...")
    
    with open(OUTPUT_FILE, "w", encoding="utf-8") as out_f:
        for conv_id, turns in conversations.items():
            # Sắp xếp lại hội thoại theo đúng thứ tự thời gian
            turns = sorted(turns, key=lambda x: x["idx"])
            
            messages = [{"role": "system", "content": SYSTEM_PROMPT}]
            
            # Người nói đầu tiên luôn là user
            for i, turn in enumerate(turns):
                role = "user" if i % 2 == 0 else "assistant"
                messages.append({"role": role, "content": turn["text"]})
            
            # Chỉ lưu những hội thoại có cả hỏi và đáp
            if len(messages) >= 3:
                out_f.write(json.dumps({"messages": messages}, ensure_ascii=False) + "\n")
                formatted_count += 1

    print(f"✅ Hoàn tất! Đã tạo ra {formatted_count} cặp hội thoại chuẩn ChatML.")
    print(f"File được lưu tại: {OUTPUT_FILE}")

if __name__ == "__main__":
    format_empathetic_dialogues()