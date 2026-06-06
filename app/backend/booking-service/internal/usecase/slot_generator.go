package usecase

import (
	"booking-service/internal/domain"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// =====================================================================
// HÀM 1: CẮT CA LÀM VIỆC (Đã fix lỗi định dạng Giây của PostgreSQL)
// =====================================================================
// Truyền thẳng time.Time vào thay vì string
func SliceShiftIntoSlots(expertID string, targetDate time.Time, tplStartTime, tplEndTime time.Time, durationMinutes int) ([]domain.ExpertSlot, error) {
	var slots []domain.ExpertSlot

	// 1. Lấy Ngày-Tháng-Năm từ targetDate (ví dụ: ngày đang chạy là 25/04/2026)
	year, month, day := targetDate.Date()
	loc := targetDate.Location()

	// 2. Lấy Giờ-Phút từ cấu hình Template dưới DB (ví dụ: 08:00)
	startHour, startMin, _ := tplStartTime.Clock()
	endHour, endMin, _ := tplEndTime.Clock()

	// 3. THUẬT TOÁN LẮP GHÉP: Ép Giờ của Template vào Ngày của vòng lặp
	// Kết quả: 2026-04-25 08:00:00+07
	shiftStart := time.Date(year, month, day, startHour, startMin, 0, 0, loc)
	shiftEnd := time.Date(year, month, day, endHour, endMin, 0, 0, loc)

	currentTime := shiftStart

	for currentTime.Before(shiftEnd) {
		slotEndTime := currentTime.Add(time.Duration(durationMinutes) * time.Minute)
		if slotEndTime.After(shiftEnd) {
			break
		}

		slot := domain.ExpertSlot{
			SlotID:    uuid.New().String(),
			ExpertID:  expertID,
			DateSlot:  targetDate,
			StartTime: currentTime, // Gán thẳng Object thời gian, sạch sẽ tuyệt đối!
			EndTime:   slotEndTime,
			Status:    "AVAILABLE",
			IsLocked:  false,
		}

		slots = append(slots, slot)
		currentTime = slotEndTime // Nhích lên slot tiếp theo
	}

	return slots, nil
}

// =====================================================================
// HÀM 2: GOROUTINES SINH LỊCH (Đã thêm Log in ra lỗi)
// =====================================================================
func GenerateSlotsForNextDays(
	expertID string,
	daysToGenerate int,
	availabilities []domain.Availability,
	timeTemplates []domain.TimeTemplate,
	timeOffs []domain.ExpertTimeOff,
) ([]domain.ExpertSlot, error) {

	var allGeneratedSlots []domain.ExpertSlot
	var mu sync.Mutex
	var wg sync.WaitGroup

	now := time.Now()

	for i := 0; i < daysToGenerate; i++ {
		targetDate := now.AddDate(0, 0, i)
		wg.Add(1)

		go func(date time.Time) {
			defer wg.Done()

			if isTimeOff(date, timeOffs) {
				return
			}

			dayOfWeek := getDayOfWeek(date)

			for _, avail := range availabilities {
				if avail.DayOfWeek == dayOfWeek && avail.IsEnabled {
					template := findTemplateByID(avail.TemplateID, timeTemplates)
					if template != nil && template.IsActive {

						// Cắt lịch
						slots, err := SliceShiftIntoSlots(expertID, date, template.StartTime, template.EndTime, 60)

						if err == nil {
							mu.Lock()
							allGeneratedSlots = append(allGeneratedSlots, slots...)
							mu.Unlock()
						} else {
							// SỬA Ở ĐÂY: In lỗi ra Terminal màu đỏ để chúng ta dễ bắt bệnh
							fmt.Printf("❌ LỖI CẮT LỊCH NGÀY %v: %v\n", date.Format("2006-01-02"), err)
						}
					}
				}
			}
		}(targetDate)
	}

	wg.Wait()
	return allGeneratedSlots, nil
}

// Các hàm Helpers giữ nguyên
func getDayOfWeek(date time.Time) int {
	d := int(date.Weekday())
	if d == 0 {
		return 7
	}
	return d
}

func isTimeOff(date time.Time, timeOffs []domain.ExpertTimeOff) bool {
	for _, to := range timeOffs {
		if date.Year() == to.StartDatetime.Year() && date.YearDay() == to.StartDatetime.YearDay() {
			return true
		}
	}
	return false
}

func findTemplateByID(id string, templates []domain.TimeTemplate) *domain.TimeTemplate {
	for _, t := range templates {
		if t.TemplateID == id {
			return &t
		}
	}
	return nil
}
