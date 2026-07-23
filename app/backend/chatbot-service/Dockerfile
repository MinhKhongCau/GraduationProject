FROM python:3.10-slim

WORKDIR /workspace

# Install system dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    libpq-dev \
    ffmpeg \
    libsndfile1 \
    && rm -rf /var/lib/apt/lists/*

# Cấu hình pip timeout cực dài
ENV PIP_DEFAULT_TIMEOUT=600
ENV PIP_RETRIES=20

COPY requirements.txt .

# Cài đặt với retry và timeout cao
RUN pip install --no-cache-dir \
    --timeout 600 \
    --retries 20 \
    --default-timeout 600 \
    -r requirements.txt

# Tải Whisper model (sau khi đã có packages)
RUN pip install --timeout 600 openai-whisper && \
    python -c "import whisper; whisper.load_model('tiny')"

# Copy source code
COPY app/ app/
COPY scripts/ scripts/

# Tạo thư mục
RUN mkdir -p temp_audio data/rag_docs/processed models

EXPOSE 8086

CMD ["sh", "-c", "python scripts/ingest_vector_db.py || true && uvicorn app.main:app --host 0.0.0.0 --port 8086"]