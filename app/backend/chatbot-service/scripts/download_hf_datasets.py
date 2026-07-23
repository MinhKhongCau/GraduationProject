# scripts/download_hf_datasets.py
import os
# pyrefly: ignore [missing-import]
from datasets import load_dataset

# 1. Đảm bảo thư mục lưu trữ đã tồn tại
SAVE_DIR = "data/finetune_sets/raw_conversations"
os.makedirs(SAVE_DIR, exist_ok=True)

def download_and_save(dataset_path: str, split_name: str, output_filename: str, revision: str = None):
    print(f"Đang tải {dataset_path} ({split_name})...")
    try:
        # Kéo dữ liệu từ máy chủ Hugging Face về RAM
        kwargs = {}
        if revision:
            kwargs["revision"] = revision
        else:
            kwargs["trust_remote_code"] = True
            
        dataset = load_dataset(dataset_path, split=split_name, **kwargs)
        
        # Đường dẫn lưu file
        save_path = os.path.join(SAVE_DIR, output_filename)
        
        # Xuất trực tiếp ra định dạng JSON Lines
        dataset.to_json(save_path, force_ascii=False)
        print(f"✅ Đã lưu thành công tại: {save_path}\n")
    except Exception as e:
        print(f"❌ Lỗi khi tải {dataset_path}: {e}\n")

if __name__ == "__main__":
    print("BẮT ĐẦU QUÁ TRÌNH TẢI DATASET FINE-TUNE...\n" + "-"*40)
    
    # 1. Tải ESConv (Tập dữ liệu hỗ trợ cảm xúc tốt nhất)
    download_and_save(
        dataset_path="thu-coai/esconv", 
        split_name="train", 
        output_filename="esconv_raw_train.jsonl"
    )
    
    # 2. Tải Empathetic Dialogues (Giúp AI học cách thấu cảm)
    download_and_save(
        dataset_path="facebook/empathetic_dialogues", 
        split_name="train", 
        output_filename="empathetic_dialogues_raw_train.jsonl",
        revision="refs/convert/parquet"
    )
    
    # 3. Tải Counsel Chat (Tham vấn tâm lý thực tế)
    download_and_save(
        dataset_path="nbertagnolli/counsel-chat", 
        split_name="train", 
        output_filename="counsel_chat_raw_train.jsonl"
    )
    
    print("-" * 40 + "\nHOÀN TẤT!")