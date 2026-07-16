# app/layers/layer3_rag.py
from app.db.postgres import get_vector_db

def retrieve_interventions(user_message: str, current_emotion: str = None, top_k: int = 3) -> str:
    """
    Truy xuất các kỹ thuật can thiệp tâm lý an toàn từ pgvector.
    
    Args:
        user_message (str): Câu nói hiện tại của người dùng.
        current_emotion (str): Cảm xúc nhận diện được từ Layer 1 (vd: 'anxiety', 'sadness').
        top_k (int): Số lượng chunk tài liệu muốn lấy lên.
        
    Returns:
        str: Chuỗi văn bản chứa các kỹ thuật đã được định dạng để nạp vào Prompt.
    """
    try:
        db = get_vector_db()
        
        # 1. TỐI ƯU HÓA CÂU TRUY VẤN (Query Reformulation)
        # Nếu chỉ dùng câu nói của user (vd: "Tôi mệt quá"), vector search có thể trả về sai.
        # Chúng ta ghép thêm cảm xúc để định hướng vector tìm đúng bài tập đối phó.
        search_query = user_message
        if current_emotion:
            search_query = f"Techniques and coping strategies for feeling {current_emotion}. User says: {user_message}"
            
        print(f"[Layer 3 - RAG] Đang tìm kiếm tài liệu cho query: '{search_query}'")
        
        # 2. TÌM KIẾM VECTOR (Similarity Search)
        results = db.similarity_search(search_query, k=top_k)
        
        if not results:
            return "Không có hướng dẫn can thiệp cụ thể nào từ cơ sở dữ liệu."
            
        # 3. TỔNG HỢP VÀ ĐỊNH DẠNG KẾT QUẢ
        formatted_docs = []
        for i, doc in enumerate(results):
            source = doc.metadata.get('source_file', 'Unknown Source')
            content = doc.page_content
            # Định dạng rõ ràng để LLM (Layer 4) dễ dàng đọc và hiểu đây là tài liệu tham khảo
            formatted_docs.append(f"--- KỸ THUẬT {i+1} (Trích từ: {source}) ---\n{content}\n")
            
        # Nối tất cả thành một khối text duy nhất
        rag_context = "\n".join(formatted_docs)
        return rag_context
        
    except Exception as e:
        print(f"[Layer 3 - RAG] ❌ Lỗi khi truy xuất dữ liệu: {e}")
        return "Hệ thống kiến thức tạm thời không khả dụng."

# ==========================================
# TEST NHANH (Chỉ chạy khi gọi trực tiếp file này)
# ==========================================
if __name__ == "__main__":
    # Giả lập luồng dữ liệu đi từ Layer 1 xuống
    test_message = "I can't stop thinking about my mistakes. I feel like a failure."
    test_emotion = "anxiety" # Giả sử wav2vec nhận diện được sự lo âu trong giọng nói
    
    print("Bắt đầu test Layer 3...")
    retrieved_context = retrieve_interventions(test_message, test_emotion, top_k=2)
    
    print("\n" + "="*50)
    print("📝 KẾT QUẢ TRẢ VỀ TỪ DATABASE:")
    print("="*50)
    print(retrieved_context)