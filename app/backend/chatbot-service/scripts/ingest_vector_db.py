# scripts/ingest_vector_db.py
import json
import os
# pyrefly: ignore [missing-import]
from langchain_core.documents import Document
# pyrefly: ignore [missing-import]
from langchain_community.embeddings import OllamaEmbeddings
# pyrefly: ignore [missing-import]
from langchain_postgres import PGVector
# pyrefly: ignore [missing-import]
from langchain_postgres.vectorstores import PGVector

# Cấu hình kết nối tới Docker PostgreSQL
CONNECTION_STRING = "postgresql+psycopg2://admin:secretpassword@localhost:5432/wellness_bot_db"
COLLECTION_NAME = "mental_health_techniques"
PROCESSED_FILE = "data/rag_docs/processed/rag_knowledge_base.json" # File JSON bạn vừa tạo ra

def ingest_data_to_pgvector():
    print("Khởi tạo mô hình Embedding (Ollama: nomic-embed-text)...")
    embeddings = OllamaEmbeddings(
        model="nomic-embed-text",
        base_url="http://localhost:11434"
    )

    print("Đang đọc dữ liệu từ file JSON...")
    with open(PROCESSED_FILE, "r", encoding="utf-8") as f:
        json_data = json.load(f)

    # Chuyển đổi JSON thành định dạng Document của Langchain
    docs = []
    for item in json_data:
        doc = Document(
            page_content=item["content"],
            metadata={
                "chunk_id": item["chunk_id"],
                "source_file": item["metadata"]["source_file"],
                "page": item["metadata"]["page"]
            }
        )
        docs.append(doc)

    print(f"Bắt đầu nạp {len(docs)} chunks vào PostgreSQL (pgvector)... Quá trình này có thể mất vài phút.")
    
    # Kết nối và nạp dữ liệu vào DB
    db = PGVector.from_documents(
        embedding=embeddings,
        documents=docs,
        collection_name=COLLECTION_NAME,
        connection=CONNECTION_STRING,
        use_jsonb=True,
    )
    
    print("✅ HOÀN TẤT! Dữ liệu đã được nạp thành công vào Vector Database.")

if __name__ == "__main__":
    ingest_data_to_pgvector()