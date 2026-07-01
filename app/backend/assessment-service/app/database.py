import os
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker
from dotenv import load_dotenv

load_dotenv()

DB_URL = os.getenv("DB_URL")
DB_USER = os.getenv("DB_USER")
DB_PASSWORD = os.getenv("DB_PASSWORD")
DB_NAME = os.getenv("DB_NAME")
DB_SSLMODE = os.getenv("DB_SSLMODE")
DB_PORT = os.getenv("DB_PORT")

DATABASE_URL = f"postgresql://{DB_USER}:{DB_PASSWORD}@{DB_URL}:{DB_PORT}/{DB_NAME}?sslmode={DB_SSLMODE}"

# 1. Tạo engine kết nối
engine = create_engine(DATABASE_URL)

# 2. Tạo Session để thao tác với DB
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)


# 3. Hàm cung cấp session cho các API (Dependency Injection)
def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()
