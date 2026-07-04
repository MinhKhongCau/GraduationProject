package usecase

import (
	"booking-service/internal/domain"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// nowMs trả về Unix timestamp hiện tại theo milliseconds (13 chữ số)
func nowMs() int64 {
	return time.Now().UnixMilli()
}

// =====================================================================
// HÀM 1: CẮT CA LÀM VIỆC
// Đọc chuỗi "HH:MM" từ TimeTemplate, kết hợp ngày mục tiêu để tạo
// ra StartTime/EndTime dưới dạng Unix timestamp 13 số (milliseconds).
// =====================================================================
func SliceShiftIntoSlots(expertID string, targetDate time.Time, tplStartStr, tplEndStr string, durationMinutes int) ([]domain.ExpertSlot, error) {
	var slots []domain.ExpertSlot

	// 1. Lấy Ngày-Tháng-Năm từ targetDate (ví dụ: 2026-07-05)
	year, month, day := targetDate.Date()
	loc := targetDate.Location()

	// 2. Parse chuỗi "HH:MM" sang giờ/phút
	var startHour, startMin, endHour, endMin int
	if _, err := fmt.Sscanf(tplStartStr, "%d:%d", &startHour, &startMin); err != nil {
		return nil, fmt.Errorf("định dạng start_time không hợp lệ '%s': %w", tplStartStr, err)
	}
	if _, err := fmt.Sscanf(tplEndStr, "%d:%d", &endHour, &endMin); err != nil {
		return nil, fmt.Errorf("định dạng end_time không hợp lệ '%s': %w", tplEndStr, err)
	}

	// 3. LẮP GHÉP: Ép giờ của Template vào Ngày của vòng lặp
	// Kết quả: 2026-07-05 08:00:00 +07:00
	shiftStart := time.Date(year, month, day, startHour, startMin, 0, 0, loc)
	shiftEnd := time.Date(year, month, day, endHour, endMin, 0, 0, loc)

	currentTime := shiftStart

	for currentTime.Before(shiftEnd) {
		slotEndTime := currentTime.Add(time.Duration(durationMinutes) * time.Minute)
		if slotEndTime.After(shiftEnd) {
			break
		}

		// Chỉ lưu ngày (không giờ) vào DateSlot - GORM sẽ lưu kiểu DATE
		dateOnly := time.Date(year, month, day, 0, 0, 0, 0, loc)

		slot := domain.ExpertSlot{
			SlotID:    uuid.New().String(),
			ExpertID:  expertID,
			DateSlot:  dateOnly,
			StartTime: currentTime.UnixMilli(),   // Unix timestamp 13 số (ms)
			EndTime:   slotEndTime.UnixMilli(),   // Unix timestamp 13 số (ms)
			Status:    domain.SlotStatusAvailable,
			IsLocked:  false,
		}

		slots = append(slots, slot)
		currentTime = slotEndTime
	}

	return slots, nil
}

// =====================================================================
// HÀM 2: GOROUTINES SINH LỊCH SONG SONG CHO NHIỀU NGÀY
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

			// Bỏ qua ngày nghỉ đột xuất
			if isTimeOff(date, timeOffs) {
				return
			}

			dayOfWeek := getDayOfWeek(date)

			for _, avail := range availabilities {
				if avail.DayOfWeek == dayOfWeek && avail.IsEnabled {
					template := findTemplateByID(avail.TemplateID, timeTemplates)
					if template != nil && template.IsActive {

						// Cắt lịch (đọc chuỗi "HH:MM" từ template)
						slots, err := SliceShiftIntoSlots(expertID, date, template.StartTime, template.EndTime, 60)

						if err == nil {
							mu.Lock()
							allGeneratedSlots = append(allGeneratedSlots, slots...)
							mu.Unlock()
						} else {
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

// =====================================================================
// HÀM HELPERS
// =====================================================================

// getDayOfWeek trả về thứ trong tuần: 1=Thứ 2, ..., 7=Chủ Nhật
func getDayOfWeek(date time.Time) int {
	d := int(date.Weekday())
	if d == 0 {
		return 7 // Chủ Nhật
	}
	return d
}

// isTimeOff kiểm tra một ngày có nằm trong khoảng thời gian nghỉ không
// So sánh trực tiếp với Unix timestamp 13 số (ms) đã lưu trong DB
func isTimeOff(date time.Time, timeOffs []domain.ExpertTimeOff) bool {
	// Tính mốc bắt đầu và kết thúc của ngày đó (ms)
	year, month, day := date.Date()
	loc := date.Location()
	dayStartMs := time.Date(year, month, day, 0, 0, 0, 0, loc).UnixMilli()
	dayEndMs := time.Date(year, month, day, 23, 59, 59, 999_000_000, loc).UnixMilli()

	for _, to := range timeOffs {
		// Kiểm tra xem ngày có overlap với khoảng thời gian nghỉ không
		if to.StartDatetime <= dayEndMs && to.EndDatetime >= dayStartMs {
			return true
		}
	}
	return false
}

// findTemplateByID tìm TimeTemplate theo ID
func findTemplateByID(id string, templates []domain.TimeTemplate) *domain.TimeTemplate {
	for _, t := range templates {
		if t.TemplateID == id {
			return &t
		}
	}
	return nil
}
