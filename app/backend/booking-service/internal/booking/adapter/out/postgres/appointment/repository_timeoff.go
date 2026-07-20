package appointmentpostgres

import (
	"booking-service/internal/booking/domain"
	"gorm.io/gorm"
)

func isSlotCoveredByTimeOff(db *gorm.DB, slot domain.ExpertSlot) (bool, error) {
	var count int64
	err := db.Model(&domain.ExpertTimeOff{}).
		Where("expert_id = ? AND start_datetime < ? AND end_datetime > ?", slot.ExpertID, slot.EndTime, slot.StartTime).
		Count(&count).Error
	return count > 0, err
}

func withoutActiveTimeOffForSlot(query *gorm.DB) *gorm.DB {
	return query.Where(`NOT EXISTS (SELECT 1 FROM "Booking_Expert_Time_Off" time_off WHERE time_off.expert_id = "Booking_Expert_Slots".expert_id AND time_off.start_datetime < "Booking_Expert_Slots".end_time AND time_off.end_datetime > "Booking_Expert_Slots".start_time)`)
}

func releasedSlotUpdatesForCoverage(db *gorm.DB, slot domain.ExpertSlot) (map[string]interface{}, error) {
	covered, err := isSlotCoveredByTimeOff(db, slot)
	if err != nil {
		return nil, err
	}
	updates := domain.ReleasedSlotUpdates()
	if covered {
		updates["status"] = domain.SlotStatusUnavailable
	}
	return updates, nil
}
