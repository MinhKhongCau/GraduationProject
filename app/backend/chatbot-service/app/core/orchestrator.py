# app/core/orchestrator.py
from app.layers import layer0_safety, layer1_nlu, layer2_dialogue, layer3_rag, layer4_llm, layer5_filter

def chat_orchestrator_stream(session_id: str, user_message: str, audio_file: str = None):
    # 1. Layer 0: Kiểm tra rủi ro khẩn cấp
    is_safe, emergency_response = layer0_safety.check_input_risk(user_message)
    if not is_safe:
        yield emergency_response
        return

    # 2. Layer 1: Nhận diện cảm xúc
    emotion = layer1_nlu.analyze_text_emotion(user_message)
    if audio_file:
        emotion = layer1_nlu.analyze_audio_emotion(audio_file)
    
    # 3. Layer 2: Lấy lịch sử hội thoại
    history = layer2_dialogue.get_chat_history(session_id)
    
    # 4. Layer 3: RAG - Truy xuất tài liệu
    context = layer3_rag.retrieve_interventions(user_message, current_emotion=emotion)
    
    # 5. Layer 4: Đẩy từng chữ ra ngoài (Streaming)
    full_response = ""
    for chunk in layer4_llm.generate_response_stream(f"{history}\nUser: {user_message}", context):
        full_response += chunk
        yield chunk  # Đẩy chữ ra UI ngay lập tức
    
    # 6. Layer 5: Lọc kết quả ĐÃ GỘP để lưu vào lịch sử an toàn
    final_safe_response = layer5_filter.filter_output(full_response)
    
    # Lưu lại lịch sử vào Layer 2
    layer2_dialogue.save_chat_turn(session_id, user_message, final_safe_response)