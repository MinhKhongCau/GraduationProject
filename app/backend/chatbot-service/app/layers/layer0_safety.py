# app/layers/layer0_safety.py
import re

# Danh sách các từ khóa rủi ro cao (Khủng hoảng, tự hại)
CRISIS_KEYWORDS = [
    r"\bsuicide\b", r"\bkill myself\b", r"\bwant to die\b", 
    r"\bend my life\b", r"\bhurt myself\b", r"\bno reason to live\b"
]

def check_input_risk(user_message: str) -> tuple[bool, str]:
    """
    Quét tin nhắn đầu vào xem có dấu hiệu khủng hoảng không.
    Trả về: (is_safe, emergency_response_if_any)
    """
    message_lower = user_message.lower()
    
    for pattern in CRISIS_KEYWORDS:
        if re.search(pattern, message_lower):
            # Kịch bản can thiệp khẩn cấp (Hardcoded)
            emergency_msg = (
                "I am so sorry you are feeling this way, but please know you are not alone. "
                "If you are in immediate danger or experiencing a crisis, please reach out for help immediately. "
                "In Vietnam, you can contact the Psychiatric Hospital Hotline at 1900 1267, "
                "or go to the nearest emergency room. Your life is extremely valuable."
            )
            return False, emergency_msg
            
    return True, ""

# ==========================================
# TEST NHANH
# ==========================================
if __name__ == "__main__":
    test_msg = "I am so tired of everything, I just want to die."
    is_safe, response = check_input_risk(test_msg)
    print(f"Message safe: {is_safe}")
    if not is_safe:
        print(f"🚨 Kích hoạt Crisis Protocol:\n{response}")