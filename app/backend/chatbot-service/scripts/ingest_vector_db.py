# scripts/ingest_vector_db.py
import json
import os
# pyrefly: ignore [missing-import]
from langchain_core.documents import Document
# pyrefly: ignore [missing-import]
from langchain_ollama import OllamaEmbeddings
# pyrefly: ignore [missing-import]
from langchain_postgres import PGVector
# pyrefly: ignore [missing-import]
from langchain_postgres.vectorstores import PGVector

# Cấu hình kết nối tới Docker PostgreSQL
DB_USER = os.getenv("POSTGRES_USER", "admin")
DB_PASSWORD = os.getenv("POSTGRES_PASSWORD", "secretpassword")
DB_HOST = os.getenv("POSTGRES_HOST", "localhost")
DB_PORT = os.getenv("POSTGRES_PORT", "5432")
DB_NAME = os.getenv("POSTGRES_DB", "wellness_bot_db")

CONNECTION_STRING = f"postgresql+psycopg2://{DB_USER}:{DB_PASSWORD}@{DB_HOST}:{DB_PORT}/{DB_NAME}"
COLLECTION_NAME = "mental_health_techniques"
PROCESSED_FILE = "data/rag_docs/processed/rag_knowledge_base.json" # File JSON bạn vừa tạo ra

def ingest_data_to_pgvector():
    import time
    import psycopg2

    print("Kiểm tra kết nối PostgreSQL...")
    retries = 15
    conn = None
    while retries > 0:
        try:
            conn = psycopg2.connect(
                dbname=DB_NAME,
                user=DB_USER,
                password=DB_PASSWORD,
                host=DB_HOST,
                port=DB_PORT
            )
            print("Kết nối PostgreSQL thành công!")
            break
        except Exception as e:
            print(f"Chưa kết nối được DB (Đang thử lại... Còn {retries} lần): {e}")
            retries -= 1
            time.sleep(3)

    if not conn:
        print("❌ Lỗi: Không thể kết nối tới PostgreSQL sau nhiều lần thử.")
        return

    # Kiểm tra xem dữ liệu đã được nạp trước đó chưa
    try:
        with conn.cursor() as cur:
            cur.execute("""
                SELECT EXISTS (
                    SELECT FROM information_schema.tables 
                    WHERE table_name = 'langchain_pg_embedding'
                );
            """)
            table_exists = cur.fetchone()[0]
            if table_exists:
                cur.execute("SELECT COUNT(*) FROM langchain_pg_embedding;")
                count = cur.fetchone()[0]
                if count > 0:
                    print(f"Dữ liệu RAG đã được nạp trước đó (Hiện có {count} records). Bỏ qua bước nạp.")
                    return
    except Exception as e:
        print(f"Kiểm tra dữ liệu cũ gặp lỗi (sẽ tiến hành nạp mới): {e}")
    finally:
        if conn:
            conn.close()

    print("Khởi tạo mô hình Embedding (Ollama: nomic-embed-text)...")
    ollama_base_url = os.getenv("OLLAMA_BASE_URL", "http://localhost:11434")
    embeddings = OllamaEmbeddings(
        model="nomic-embed-text",
        base_url=ollama_base_url
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