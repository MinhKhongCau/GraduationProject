# pyrefly: ignore [missing-import]
from huggingface_hub import snapshot_download

print("Bắt đầu tải dữ liệu model thực sự (file nhị phân)...")
snapshot_download(
    repo_id="r-f/wav2vec-english-speech-emotion-recognition",
    local_dir="models/wav2vec-emotion"
)
print("Hoàn tất tải model!")