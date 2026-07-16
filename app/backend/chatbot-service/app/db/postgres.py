# app/db/postgres.py
import os
# pyrefly: ignore [missing-import]
from langchain_community.embeddings import OllamaEmbeddings
# pyrefly: ignore [missing-import]
from langchain_postgres.vectorstores import PGVector

# Đọc cấu hình từ biến môi trường (Environment Variables)
DB_USER = os.getenv("POSTGRES_USER", "admin")
DB_PASSWORD = os.getenv("POSTGRES_PASSWORD", "secretpassword")
DB_HOST = os.getenv("POSTGRES_HOST", "localhost")
DB_PORT = os.getenv("POSTGRES_PORT", "5432")
DB_NAME = os.getenv("POSTGRES_DB", "wellness_bot_db")

# Chuỗi kết nối Database
CONNECTION_STRING = f"postgresql+psycopg2://{DB_USER}:{DB_PASSWORD}@{DB_HOST}:{DB_PORT}/{DB_NAME}"
COLLECTION_NAME = "mental_health_techniques"

# Khởi tạo Embedding Model (Giống với lúc Ingest data)
def get_embeddings_model():
    return OllamaEmbeddings(
        model="nomic-embed-text",
        base_url="http://localhost:11434"
    )

# Hàm khởi tạo và trả về kết nối Vector Database
def get_vector_db() -> PGVector:
    """
    Tạo kết nối tới bảng vector trong PostgreSQL để truy vấn RAG.
    """
    embeddings = get_embeddings_model()
    
    vector_db = PGVector(
        embeddings=embeddings,
        collection_name=COLLECTION_NAME,
        connection=CONNECTION_STRING,
        use_jsonb=True,
    )
    return vector_db

# (Tùy chọn tương lai) Khởi tạo Connection Pool cho SQLAlchemy 
# nếu muốn lưu trữ lịch sử chat (Session State) ở dạng bảng SQL truyền thống.
# def get_session_db():
#     pass