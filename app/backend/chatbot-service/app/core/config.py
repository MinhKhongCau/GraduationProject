from pydantic_settings import BaseSettings
from typing import Optional

class Settings(BaseSettings):
    PROJECT_NAME: str = "Mental Wellness Bot"
    DATABASE_URL: str = "postgresql://postgres:postgres@localhost:5432/mental_wellness"
    
    # Model Paths
    LLM_MODEL_PATH: str = "models/llm/Llama-3-8B.gguf"
    EMBEDDING_MODEL_PATH: str = "models/embedding"
    
    # API Keys
    OPENAI_API_KEY: Optional[str] = None
    
    class Config:
        env_file = ".env"
        case_sensitive = True

settings = Settings()
