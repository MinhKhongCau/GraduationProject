# scripts/process_rag_docs.py
import os
import json
# pyrefly: ignore [missing-import]
from langchain_community.document_loaders import PyPDFLoader
# pyrefly: ignore [missing-import]
from langchain_text_splitters import RecursiveCharacterTextSplitter

RAW_DIR = "data/rag_docs/raw/"
PROCESSED_DIR = "data/rag_docs/processed/"

def process_pdfs_to_json():
    os.makedirs(PROCESSED_DIR, exist_ok=True)
    
    # Cấu hình bộ cắt chữ (Chunking)
    # chunk_size: Số lượng ký tự tối đa trong 1 đoạn (đảm bảo LLM đọc vừa đủ)
    # chunk_overlap: Số ký tự trùng lặp giữa 2 đoạn liên tiếp (để không bị mất ngữ cảnh ở giữa câu)
    text_splitter = RecursiveCharacterTextSplitter(
        chunk_size=1000,
        chunk_overlap=200,
        length_function=len,
    )

    processed_data = []
    
    print("Bắt đầu xử lý các file PDF cho RAG (Layer 3)...")
    
    # Quét tất cả các file PDF trong thư mục raw
    for filename in os.listdir(RAW_DIR):
        if filename.endswith(".pdf"):
            file_path = os.path.join(RAW_DIR, filename)
            print(f"\nĐang đọc file: {filename}")
            
            try:
                # Đọc file PDF
                loader = PyPDFLoader(file_path)
                pages = loader.load()
                
                # Băm nhỏ các trang thành nhiều chunk
                chunks = text_splitter.split_documents(pages)
                
                print(f"-> Đã cắt thành {len(chunks)} chunks.")
                
                for i, chunk in enumerate(chunks):
                    # Làm sạch text cơ bản (xóa ký tự newline thừa)
                    clean_text = chunk.page_content.replace("\n", " ").strip()
                    
                    # Tạo cấu trúc lưu trữ chuẩn
                    chunk_data = {
                        "chunk_id": f"{filename}_chunk_{i}",
                        "content": clean_text,
                        "metadata": {
                            "source_file": filename,
                            "page": chunk.metadata.get("page", 0) + 1, # page index bắt đầu từ 0
                        }
                    }
                    processed_data.append(chunk_data)
                    
            except Exception as e:
                print(f"❌ Lỗi khi xử lý file {filename}: {e}")

    # Lưu toàn bộ dữ liệu đã xử lý ra file JSON
    output_file = os.path.join(PROCESSED_DIR, "rag_knowledge_base.json")
    with open(output_file, "w", encoding="utf-8") as f:
        json.dump(processed_data, f, ensure_ascii=False, indent=4)
        
    print(f"\n✅ HOÀN TẤT! Đã lưu tổng cộng {len(processed_data)} chunks vào {output_file}")

if __name__ == "__main__":
    process_pdfs_to_json()