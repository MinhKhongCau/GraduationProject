import os
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker
from dotenv import load_dotenv

load_dotenv()

DATABASE_URL = os.getenv("DATABASE_URL")

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
