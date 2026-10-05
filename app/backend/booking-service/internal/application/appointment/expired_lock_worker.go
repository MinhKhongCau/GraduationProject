package appointment

import (
	"log"
	"time"
)

// StartExpiredLockWorker khởi chạy worker chạy ngầm để dọn dẹp các slot giữ chỗ quá hạn 15 phút.
func StartExpiredLockWorker(repo Repository) {
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		log.Println("🔄 Expired Lock Worker đã khởi động, quét mỗi 60 giây...")

		for range ticker.C {
			cleaned, err := repo.CancelExpiredLocks()
			if err != nil {
				log.Printf("⚠️  Worker lỗi khi dọn expired locks: %v", err)
			} else if cleaned > 0 {
				log.Printf("🧹 Worker đã dọn %d slot hết hạn, mở lại cho bệnh nhân khác.", cleaned)
			}
		}
	}()
}
