# app/core/orchestrator.py
import logging
from app.layers import layer0_safety, layer1_nlu, layer2_dialogue, layer3_rag, layer4_llm, layer5_filter

logger = logging.getLogger(__name__)

SAFE_FALLBACK_RESPONSE = "I'm sorry, I'm experiencing a technical issue right now."

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
    
    # 5. Layer 4 & Layer 5: Thu thập toàn bộ phản hồi và lọc an toàn trước khi gửi cho client
    try:
        full_response = ""
        for chunk in layer4_llm.generate_response_stream(f"{history}\nUser: {user_message}", context):
            full_response += chunk
        
        if not full_response or not full_response.strip():
            logger.warning("Empty or whitespace-only response received from LLM")
            final_safe_response = SAFE_FALLBACK_RESPONSE
        else:
            # 6. Layer 5: Lọc kết quả ĐÃ GỘP trước khi yield
            final_safe_response = layer5_filter.filter_output(full_response)
    except Exception as exc:
        logger.error("Error during LLM response generation or safety filtering: %s", type(exc).__name__)
        final_safe_response = SAFE_FALLBACK_RESPONSE

    # 7. Lưu lại lịch sử an toàn vào Layer 2 trước khi yield (bảo toàn lịch sử kể cả khi client ngắt kết nối khi nhận yield)
    try:
        layer2_dialogue.save_chat_turn(session_id, user_message, final_safe_response)
    except Exception as exc:
        logger.error("Error saving chat turn to dialogue history: %s", type(exc).__name__)

    # 8. Chỉ yield sau khi đã lọc an toàn và lưu lịch sử hoàn tất
    yield final_safe_response
