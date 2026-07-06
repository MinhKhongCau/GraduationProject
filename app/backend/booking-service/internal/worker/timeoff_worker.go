package worker

import (
	"booking-service/internal/repository/postgres"
	"log"
	"time"
)

// StartTimeOffWorker khởi chạy worker để quét các lịch nghỉ phép (TimeOff) chưa được xử lý
// và dọn dẹp các lịch bị ảnh hưởng (cả Slot và Appointment).
func StartTimeOffWorker(
	timeoffRepo *postgres.TimeOffRepository,
	slotRepo *postgres.SlotRepository,
	appointmentRepo *postgres.AppointmentRepository,
) {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		log.Println("🔄 TimeOff Worker đã khởi động, quét mỗi 30 giây...")

		for range ticker.C {
			timeOffs, err := timeoffRepo.GetUnprocessedTimeOffs()
			if err != nil {
				log.Printf("⚠️ TimeOff Worker lỗi khi lấy danh sách: %v", err)
				continue
			}

			for _, to := range timeOffs {
				log.Printf("⚙️ Đang xử lý TimeOff ID: %s (Expert: %s)", to.TimeOffID, to.ExpertID)

				// 1. Lấy tất cả Slot bị đè
				slots, err := timeoffRepo.GetOverlappingSlots(to.ExpertID, to.StartDatetime, to.EndDatetime)
				if err != nil {
					log.Printf("Lỗi lấy slots đè: %v", err)
					continue
				}

				for _, slot := range slots {
					// 2. Nếu slot chưa ai đặt (AVAILABLE hoặc bị khoá tạm) -> Xoá luôn
					if slot.Status != 2 { // 2 = OCCUPIED
						_ = timeoffRepo.DeleteAvailableSlots(to.ExpertID, to.StartDatetime, to.EndDatetime)
					} else {
						// 3. Nếu slot đã có người đặt (OCCUPIED) -> Huỷ Appointment
						appt, err := appointmentRepo.GetAppointmentBySlotID(slot.SlotID)
						if err == nil && appt != nil {
							_ = appointmentRepo.CancelAppointmentByExpert(appt.AppointmentID, to.Reason)
							
							// TODO: Tích hợp Notification Service
							// Ghi log để nhắc nhở tích hợp sau này
							log.Printf("[TODO - NOTIFICATION SERVICE] 🚨 Cần bắn sự kiện/gửi email báo cho Patient %s về việc huỷ lịch hẹn %s do Bác sĩ nghỉ đột xuất. Lý do: %s", appt.PatientID, appt.AppointmentID, to.Reason)
						}
						// Và phải ẩn slot đó đi hoặc xoá đi (tuỳ business, ở đây ta có thể xoá luôn để ko ai thấy nữa)
						// Nhưng vì slot đã có appointment nên ta cứ để đó với trạng thái nào đó, hoặc xoá. 
						// Nếu xoá slot, sẽ mất foreign key ở Appointment. Thay vào đó, ta chuyển nó về AVAILABLE và sẽ bị hàm DeleteAvailableSlots ở dưới/trên xoá.
						// Hoặc tốt nhất là cứ kệ slot đó, vì appointment đã bị CANCELLED. Nhưng nếu để đó thì nó lại hiển thị lên màn hình đặt lịch?
						// Khi Appointment CANCELLED, Slot sẽ về lại AVAILABLE, rồi TimeOffWorker sẽ xoá nó! (Logic xử lý ở hàm CancelAppointmentByExpert sẽ trả Slot về AVAILABLE, sau đó hàm DeleteAvailableSlots xoá)
					}
				}

				// Lặp lại xoá một lần nữa cho chắc ăn tất cả Slot AVAILABLE trong vùng này
				_ = timeoffRepo.DeleteAvailableSlots(to.ExpertID, to.StartDatetime, to.EndDatetime)

				// 4. Đánh dấu đã xử lý
				if err := timeoffRepo.MarkAsProcessed(to.TimeOffID); err != nil {
					log.Printf("Lỗi mark processed timeoff %s: %v", to.TimeOffID, err)
				} else {
					log.Printf("✅ Đã xử lý xong TimeOff ID: %s", to.TimeOffID)
				}
			}
		}
	}()
}
