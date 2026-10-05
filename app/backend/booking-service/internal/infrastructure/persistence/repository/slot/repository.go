package repository

import (
	appslot "booking-service/internal/application/slot"
	slotdomain "booking-service/internal/domain/slot"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type pgRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) appslot.Repository {
	return &pgRepository{db: db}
}

// 1. Lấy danh sách NGÀY có lịch trống trong khoảng thời gian
// Điều kiện: status=AVAILABLE, is_locked=false, và start_time > thời điểm hiện tại (không hiển thị slot quá khứ)
func (r *pgRepository) GetAvailableDates(expertID string, startDate, endDate time.Time) ([]string, error) {
	var dates []string
	nowMs := time.Now().UnixMilli()

	err := withoutActiveTimeOff(r.db.Table(`"Booking_Expert_Slots"`)).
		Where(`expert_id = ? AND date_slot >= ? AND date_slot <= ? AND status = ? AND start_time > ?`,
			expertID, startDate, endDate, slotdomain.SlotStatusAvailable, nowMs).
		Group("TO_CHAR(date_slot, 'YYYY-MM-DD')").
		Order("TO_CHAR(date_slot, 'YYYY-MM-DD') ASC").
		Pluck("TO_CHAR(date_slot, 'YYYY-MM-DD')", &dates).Error

	return dates, err
}

// 2. Lấy danh sách KHUNG GIỜ trống của một ngày cụ thể (theo expert_id)
// Điều kiện: status=AVAILABLE, is_locked=false, start_time > now (loại slot quá khứ)
func (r *pgRepository) GetAvailableTimes(date string, expertID string) ([]appslot.SlotTimeResult, error) {
	var times []appslot.SlotTimeResult
	nowMs := time.Now().UnixMilli()

	err := withoutActiveTimeOff(r.db.Table(`"Booking_Expert_Slots"`)).
		Where(`TO_CHAR(date_slot, 'YYYY-MM-DD') = ? AND expert_id = ? AND status = ? AND start_time > ?`,
			date, expertID, slotdomain.SlotStatusAvailable, nowMs).
		Select("slot_id, start_time, end_time").
		Order("start_time ASC").
		Scan(&times).Error

	return times, err
}

func withoutActiveTimeOff(query *gorm.DB) *gorm.DB {
	return query.Where(`NOT EXISTS (SELECT 1 FROM "Booking_Expert_Time_Off" time_off WHERE time_off.expert_id = "Booking_Expert_Slots".expert_id AND time_off.start_datetime < "Booking_Expert_Slots".end_time AND time_off.end_datetime > "Booking_Expert_Slots".start_time)`)
}

// 3. Lấy toàn bộ danh sách Slot của Expert (bao gồm AVAILABLE, LOCKED, OCCUPIED)
func (r *pgRepository) GetSlotsByExpert(expertID string, fromDate, toDate int64) ([]slotdomain.ExpertSlot, error) {
	var slots []slotdomain.ExpertSlot
	query := r.db.Where("expert_id = ?", expertID).
		Where(`NOT (status = ? AND EXISTS (SELECT 1 FROM "Booking_Expert_Time_Off" time_off WHERE time_off.expert_id = "Booking_Expert_Slots".expert_id AND time_off.start_datetime < "Booking_Expert_Slots".end_time AND time_off.end_datetime > "Booking_Expert_Slots".start_time))`, slotdomain.SlotStatusAvailable)

	if fromDate > 0 {
		query = query.Where("start_time >= ?", fromDate)
	}
	if toDate > 0 {
		query = query.Where("start_time <= ?", toDate)
	}

	err := query.Order("start_time ASC").Find(&slots).Error
	return slots, err
}

// 4. LƯU HÀNG LOẠT (BULK INSERT) với cơ chế chống trùng lịch (Idempotency)
// Sử dụng ON CONFLICT DO NOTHING: nếu slot (expert_id, start_time) đã tồn tại thì bỏ qua, không báo lỗi
func (r *pgRepository) GetOverlappingSlots(expertID string, startMs, endMs int64) ([]slotdomain.ExpertSlot, error) {
	var slots []slotdomain.ExpertSlot
	err := r.db.Where("expert_id = ? AND start_time < ? AND end_time > ?", expertID, endMs, startMs).Order("start_time ASC").Find(&slots).Error
	return slots, err
}

func (r *pgRepository) BulkInsertSlots(slots []slotdomain.ExpertSlot) (int64, error) {
	if len(slots) == 0 {
		return 0, nil
	}

	// CreateInBatches + OnConflict DoNothing: chia nhỏ 100 dòng/lần và bỏ qua khi trùng
	result := r.db.
		Clauses(clause.OnConflict{DoNothing: true}).
		CreateInBatches(slots, 100)
	return result.RowsAffected, result.Error
}

func (r *pgRepository) ListAvailableDates(filter appslot.AvailableDateQuery) ([]string, int64, error) {
	baseQuery := withoutActiveTimeOff(r.db.Table(`"Booking_Expert_Slots"`)).
		Where("expert_id = ? AND start_time >= ? AND start_time < ? AND status = ? AND start_time > ?",
			filter.ExpertID, filter.FromMs, filter.ToMs, slotdomain.SlotStatusAvailable, time.Now().UnixMilli())

	var total int64
	if err := baseQuery.Session(&gorm.Session{}).Select("COUNT(DISTINCT TO_CHAR(date_slot, 'YYYY-MM-DD'))").Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	var dates []string
	err := baseQuery.Session(&gorm.Session{}).
		Select("TO_CHAR(date_slot, 'YYYY-MM-DD') AS date_value").
		Group("TO_CHAR(date_slot, 'YYYY-MM-DD')").
		Order("date_value ASC").
		Limit(filter.Page.Size).
		Offset(filter.Page.Offset()).
		Pluck("date_value", &dates).Error

	return dates, total, err
}

func (r *pgRepository) ListAvailableTimes(filter appslot.AvailableTimeQuery) ([]appslot.SlotTimeResult, int64, error) {
	query := withoutActiveTimeOff(r.db.Table(`"Booking_Expert_Slots"`)).Where(`TO_CHAR(date_slot, 'YYYY-MM-DD') = ? AND expert_id = ? AND status = ? AND start_time > ?`, filter.Date, filter.ExpertID, slotdomain.SlotStatusAvailable, time.Now().UnixMilli())
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []appslot.SlotTimeResult
	err := query.Select("slot_id, start_time, end_time, price").Order("start_time ASC, slot_id ASC").Limit(filter.Page.Size).Offset(filter.Page.Offset()).Scan(&items).Error
	return items, total, err
}

func (r *pgRepository) ListExpertSlots(filter appslot.ExpertSlotQuery) ([]slotdomain.ExpertSlot, int64, error) {
	query := r.db.Model(&slotdomain.ExpertSlot{}).Where("expert_id = ? AND start_time >= ? AND start_time < ?", filter.ExpertID, filter.FromMs, filter.ToMs)
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.AvailabilityID != "" {
		query = query.Where("availability_id = ?", filter.AvailabilityID)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []slotdomain.ExpertSlot
	err := query.Order("start_time ASC, slot_id ASC").Limit(filter.Page.Size).Offset(filter.Page.Offset()).Find(&items).Error
	return items, total, err
}
