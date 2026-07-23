import json
import random

# Danh sách 3 file gốc của bạn
input_files = [
    "data/finetune_sets/formatted/counsel_chat_safe_train.jsonl",
    "data/finetune_sets/formatted/empathetic_dialogues_chatml_train.jsonl",
    "data/finetune_sets/formatted/esconv_chatml_train.jsonl"
]

all_data = []

print("Đang đọc và gộp dữ liệu từ các file...")
# Đọc và gộp tất cả các dòng (mỗi dòng là 1 JSON object)
for file_name in input_files:
    try:
        with open(file_name, 'r', encoding='utf-8') as f:
            for line in f:
                if line.strip():
                    all_data.append(json.loads(line))
    except FileNotFoundError:
        print(f"Cảnh báo: Không tìm thấy file {file_name}")

# Xáo trộn dữ liệu (Shuffle) để các mẫu hội thoại trộn đều với nhau
random.seed(42) # Cố định seed để nếu chạy lại vẫn ra kết quả như cũ
random.shuffle(all_data)

# Tính toán mốc cắt cho tỷ lệ 80 - 10 - 10
total_len = len(all_data)
train_end = int(total_len * 0.8)
val_end = int(total_len * 0.9)

train_data = all_data[:train_end]
val_data = all_data[train_end:val_end]
test_data = all_data[val_end:]

# Hàm lưu dữ liệu ra file jsonl
def save_jsonl(data, filename):
    with open(filename, 'w', encoding='utf-8') as f:
        for item in data:
            # ensure_ascii=False giúp giữ nguyên font chữ tiếng Việt
            f.write(json.dumps(item, ensure_ascii=False) + '\n')

print("Đang lưu ra 3 file mới...")
save_jsonl(train_data, "data/finetune_sets/split_data/train.jsonl")
save_jsonl(val_data, "data/finetune_sets/split_data/val.jsonl")
save_jsonl(test_data, "data/finetune_sets/split_data/test.jsonl")

# Báo cáo kết quả
print("="*30)
print("HOÀN TẤT CHIA DỮ LIỆU!")
print(f"Tổng số mẫu hội thoại: {total_len}")
print(f" - Tập huấn luyện (train.jsonl): {len(train_data)} mẫu")
print(f" - Tập xác thực (val.jsonl): {len(val_data)} mẫu")
print(f" - Tập kiểm thử (test.jsonl): {len(test_data)} mẫu")