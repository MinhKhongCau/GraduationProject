# app/layers/layer1_nlu.py
# pyrefly: ignore [missing-import]
import torch
# pyrefly: ignore [missing-import]
import torch.nn.functional as F
# pyrefly: ignore [missing-import]
import librosa
# pyrefly: ignore [missing-import]
from transformers import AutoFeatureExtractor, AutoModelForAudioClassification

# --- BẮT ĐẦU VÁ LỖI BẢO MẬT PYTORCH 2.6 ---
# Lưu lại hàm gốc của PyTorch
original_torch_load = torch.load

# Viết đè một hàm mới, tự động chèn weights_only=False
def safe_torch_load(*args, **kwargs):
    kwargs['weights_only'] = False
    return original_torch_load(*args, **kwargs)

# Tráo đổi hàm gốc bằng hàm đã vá của chúng ta
torch.load = safe_torch_load
# --- KẾT THÚC VÁ LỖI ---

import os

# Bắt buộc sử dụng mô hình trực tiếp từ Hugging Face Hub (hoặc tùy chỉnh qua biến môi trường EMOTION_MODEL_NAME)
MODEL_NAME = os.getenv("EMOTION_MODEL_NAME", "r-f/wav2vec-english-speech-emotion-recognition")

print(f"[Layer 1] Loading Speech Emotion model from Hugging Face Hub ({MODEL_NAME})...")
try:
    processor = AutoFeatureExtractor.from_pretrained(MODEL_NAME)
    model = AutoModelForAudioClassification.from_pretrained(MODEL_NAME)
    print("[Layer 1] Model loaded successfully.")
except Exception as e:
    print(f"[Layer 1] Warning: Could not load model. Error: {e}")
    processor, model = None, None

import warnings

def analyze_audio_emotion(file_path: str) -> str:
    """
    Read an audio file (.wav, .m4a, .mp3, etc.) and predict the user's emotion.
    """
    if model is None or processor is None:
        return "neutral"

    try:
        # Bỏ qua các cảnh báo fallback audioread không cần thiết khi đọc file .m4a / .mp3
        with warnings.catch_warnings():
            warnings.simplefilter("ignore")
            speech, sr = librosa.load(file_path, sr=16000)
        
        inputs = processor(speech, sampling_rate=16000, return_tensors="pt", padding=True)
        
        with torch.no_grad():
            outputs = model(**inputs)
        
        scores = F.softmax(outputs.logits, dim=1)
        predicted_class_id = torch.argmax(scores, dim=-1).item()
        emotion = model.config.id2label[predicted_class_id]
        
        print(f"[Layer 1] Detected Audio Emotion: {emotion}")
        return emotion
    except Exception as e:
        print(f"[Layer 1] Audio processing error: {e}")
        return "neutral"

def analyze_text_emotion(text: str) -> str:
    """
    Analyze emotion via Text as a fallback if Voice is not used.
    """
    text_lower = text.lower()
    if any(word in text_lower for word in ["anxious", "worry", "panic", "stress", "overwhelmed"]):
        return "anxiety"
    if any(word in text_lower for word in ["sad", "depressed", "cry", "hopeless", "failure"]):
        return "sadness"
    if any(word in text_lower for word in ["angry", "mad", "frustrated", "annoyed"]):
        return "anger"
    
    return "neutral"

# ==========================================
# QUICK TEST
# ==========================================
if __name__ == "__main__":
    test_text = "I am feeling very overwhelmed and panic right now."
    print(f"Test Text: '{test_text}'")
    print(f"Detected Text Emotion: {analyze_text_emotion(test_text)}")