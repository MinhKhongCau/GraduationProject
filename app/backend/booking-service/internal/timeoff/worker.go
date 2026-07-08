package timeoff

import (
	"log"
	"time"
)

// StartWorker khởi chạy worker chạy ngầm định kỳ kiểm tra các bản ghi TimeOff
func StartWorker(usecase Usecase) {
	go func() {
		ticker := time.NewTicker(2 * time.Minute) // Chạy mỗi 2 phút
		defer ticker.Stop()

		log.Println("🔄 Time-off Worker đã khởi động...")

		for range ticker.C {
			usecase.ProcessTimeOffs()
		}
	}()
}
