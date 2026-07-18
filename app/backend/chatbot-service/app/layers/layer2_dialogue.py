# app/layers/layer2_dialogue.py

# Sử dụng in-memory dictionary cho quá trình phát triển nhanh
session_storage = {}

def get_chat_history(session_id: str, limit: int = 4) -> str:
    """
    Lấy lịch sử chat gần nhất để đưa vào ngữ cảnh cho LLM.
    Trả về chuỗi văn bản định dạng rõ ràng các lượt hội thoại.
    """
    history = session_storage.get(session_id, [])
    if not history:
        return ""
        
    # Lấy 'limit' lượt hội thoại gần nhất (tránh nhồi quá nhiều làm tràn token của LLM)
    recent_history = history[-limit:]
    
    formatted_history = []
    for turn in recent_history:
        formatted_history.append(f"User: {turn['user']}")
        formatted_history.append(f"AI: {turn['ai']}")
        
    return "\n".join(formatted_history)

def save_chat_turn(session_id: str, user_msg: str, ai_response: str):
    """
    Lưu lượt hội thoại mới vào bộ nhớ của phiên (session).
    """
    if session_id not in session_storage:
        session_storage[session_id] = []
        
    session_storage[session_id].append({
        "user": user_msg,
        "ai": ai_response
    })
    
def clear_session(session_id: str):
    """
    Xóa bộ nhớ khi người dùng kết thúc phiên hoặc muốn bắt đầu lại.
    """
    if session_id in session_storage:
        del session_storage[session_id]

# ==========================================
# TEST NHANH 
# ==========================================
if __name__ == "__main__":
    test_session = "user_123"
    
    print("Mô phỏng 2 lượt chat...")
    save_chat_turn(test_session, "Hi, I feel anxious.", "I am here for you. Tell me more.")
    save_chat_turn(test_session, "I have a big exam tomorrow.", "It's normal to feel nervous before an exam.")
    
    print("\nLịch sử hội thoại hiện tại đưa vào Prompt sẽ là:")
    print("-" * 40)
    print(get_chat_history(test_session))
    print("-" * 40)