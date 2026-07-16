# app/layers/layer5_filter.py
import re

# Các cụm từ mà một AI hỗ trợ tâm lý tuyệt đối không được dùng
DIAGNOSIS_KEYWORDS = [
    r"\byou have\b.*\b(depression|anxiety|bipolar|schizophrenia|ptsd)\b",
    r"\bi diagnose you\b",
    r"\bprescribe\b",
    r"\bdosage\b",
    r"\btake.*(mg|pill|medication)\b"
]

def filter_output(ai_response: str) -> str:
    """
    Quét câu trả lời của LLM trước khi trả về cho người dùng cuối.
    """
    response_lower = ai_response.lower()
    
    for pattern in DIAGNOSIS_KEYWORDS:
        if re.search(pattern, response_lower):
            print(f"[Layer 5 - Filter] ⚠️ Phát hiện câu trả lời vi phạm an toàn y khoa! Đang thay thế...")
            return (
                "I'm here to support you, but please note that I cannot provide medical diagnoses or prescribe treatments. "
                "I highly recommend consulting with a licensed healthcare professional for a proper evaluation."
            )
            
    return ai_response

# ==========================================
# TEST NHANH
# ==========================================
if __name__ == "__main__":
    test_ai_msg = "Based on what you said, I think you have clinical depression. You should take 20mg of Fluoxetine."
    filtered_msg = filter_output(test_ai_msg)
    print(f"🤖 Output cuối cùng:\n{filtered_msg}")