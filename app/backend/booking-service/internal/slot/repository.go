package slot

import (
	"booking-service/internal/booking/domain"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	GetAvailableDates(expertID string, startDate, endDate time.Time) ([]string, error)
	GetAvailableTimes(date string, expertID string) ([]SlotTimeResult, error)
	GetSlotsByExpert(expertID string, fromDate, toDate int64) ([]domain.ExpertSlot, error)
	GetOverlappingSlots(expertID string, startMs, endMs int64) ([]domain.ExpertSlot, error)
	BulkInsertSlots(slots []domain.ExpertSlot) (int64, error)
}

type pgRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &pgRepository{db: db}
}

// AvailableDateResult - Kết quả ngày có lịch trống trả về cho bệnh nhân
type AvailableDateResult struct {
	DateSlot string `json:"date_slot"` // Định dạng "YYYY-MM-DD"
}

// 1. Lấy danh sách NGÀY có lịch trống trong khoảng thời gian
// Điều kiện: status=AVAILABLE, is_locked=false, và start_time > thời điểm hiện tại (không hiển thị slot quá khứ)
func (r *pgRepository) GetAvailableDates(expertID string, startDate, endDate time.Time) ([]string, error) {
	var dates []string
	nowMs := time.Now().UnixMilli()

	// DISTINCT DATE: gom tất cả các slot trong cùng ngày thành 1 dòng
	// Lọc: chỉ lấy ngày có ít nhất 1 slot AVAILABLE, chưa bị khóa, và chưa qua (start_time > now)
	err := withoutActiveTimeOff(r.db.Table(`"Booking_Expert_Slots"`)).
		Where(`expert_id = ? AND date_slot >= ? AND date_slot <= ? AND status = ? AND start_time > ?`,
			expertID, startDate, endDate, domain.SlotStatusAvailable, nowMs).
		Select(`DISTINCT TO_CHAR(date_slot, 'YYYY-MM-DD')`).
		Pluck(`TO_CHAR(date_slot, 'YYYY-MM-DD')`, &dates).Error

	return dates, err
}

// SlotTimeResult - Kết quả khung giờ trả về cho bệnh nhân (dùng Unix ms 13 số)
type SlotTimeResult struct {
	SlotID    string  `json:"slot_id"`
	Price     float64 `json:"price"`
	StartTime int64   `json:"start_time"` // Unix timestamp 13 số (ms)
	EndTime   int64   `json:"end_time"`   // Unix timestamp 13 số (ms)
}

// 2. Lấy danh sách KHUNG GIỜ trống của một ngày cụ thể (theo expert_id)
// Điều kiện: status=AVAILABLE, is_locked=false, start_time > now (loại slot quá khứ)
func (r *pgRepository) GetAvailableTimes(date string, expertID string) ([]SlotTimeResult, error) {
	var times []SlotTimeResult
	nowMs := time.Now().UnixMilli()

	err := withoutActiveTimeOff(r.db.Table(`"Booking_Expert_Slots"`)).
		Where(`TO_CHAR(date_slot, 'YYYY-MM-DD') = ? AND expert_id = ? AND status = ? AND start_time > ?`,
			date, expertID, domain.SlotStatusAvailable, nowMs).
		Select("slot_id, start_time, end_time").
		Order("start_time ASC").
		Scan(&times).Error

	return times, err
}

func withoutActiveTimeOff(query *gorm.DB) *gorm.DB {
	return query.Where(`NOT EXISTS (SELECT 1 FROM "Booking_Expert_Time_Off" time_off WHERE time_off.expert_id = "Booking_Expert_Slots".expert_id AND time_off.start_datetime < "Booking_Expert_Slots".end_time AND time_off.end_datetime > "Booking_Expert_Slots".start_time)`)
}

// 3. Lấy toàn bộ danh sách Slot của Expert (bao gồm AVAILABLE, LOCKED, OCCUPIED)
func (r *pgRepository) GetSlotsByExpert(expertID string, fromDate, toDate int64) ([]domain.ExpertSlot, error) {
	var slots []domain.ExpertSlot
	query := r.db.Where("expert_id = ?", expertID).
		Where(`NOT (status = ? AND EXISTS (SELECT 1 FROM "Booking_Expert_Time_Off" time_off WHERE time_off.expert_id = "Booking_Expert_Slots".expert_id AND time_off.start_datetime < "Booking_Expert_Slots".end_time AND time_off.end_datetime > "Booking_Expert_Slots".start_time))`, domain.SlotStatusAvailable)

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
func (r *pgRepository) GetOverlappingSlots(expertID string, startMs, endMs int64) ([]domain.ExpertSlot, error) {
	var slots []domain.ExpertSlot
	err := r.db.Where("expert_id = ? AND start_time < ? AND end_time > ?", expertID, endMs, startMs).Order("start_time ASC").Find(&slots).Error
	return slots, err
}

func (r *pgRepository) BulkInsertSlots(slots []domain.ExpertSlot) (int64, error) {
	if len(slots) == 0 {
		return 0, nil
	}

	// CreateInBatches + OnConflict DoNothing: chia nhỏ 100 dòng/lần và bỏ qua khi trùng
	result := r.db.
		Clauses(clause.OnConflict{DoNothing: true}).
		CreateInBatches(slots, 100)
	return result.RowsAffected, result.Error
}

func (r *pgRepository) ListAvailableDates(filter AvailableDateQuery) ([]string, int64, error) {
	query := withoutActiveTimeOff(r.db.Table(`"Booking_Expert_Slots"`)).Where("expert_id = ? AND start_time >= ? AND start_time < ? AND status = ? AND start_time > ?", filter.ExpertID, filter.FromMs, filter.ToMs, domain.SlotStatusAvailable, time.Now().UnixMilli())
	var total int64
	if err := query.Distinct("date_slot").Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var dates []string
	err := query.Select(`DISTINCT TO_CHAR(date_slot, 'YYYY-MM-DD') AS date_value`).Order("date_value ASC").Limit(filter.Page.Size).Offset(filter.Page.Offset()).Pluck("date_value", &dates).Error
	return dates, total, err
}

func (r *pgRepository) ListAvailableTimes(filter AvailableTimeQuery) ([]SlotTimeResult, int64, error) {
	query := withoutActiveTimeOff(r.db.Table(`"Booking_Expert_Slots"`)).Where(`TO_CHAR(date_slot, 'YYYY-MM-DD') = ? AND expert_id = ? AND status = ? AND start_time > ?`, filter.Date, filter.ExpertID, domain.SlotStatusAvailable, time.Now().UnixMilli())
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []SlotTimeResult
	err := query.Select("slot_id, start_time, end_time, price").Order("start_time ASC, slot_id ASC").Limit(filter.Page.Size).Offset(filter.Page.Offset()).Scan(&items).Error
	return items, total, err
}

func (r *pgRepository) ListExpertSlots(filter ExpertSlotQuery) ([]domain.ExpertSlot, int64, error) {
	query := r.db.Model(&domain.ExpertSlot{}).Where("expert_id = ? AND start_time >= ? AND start_time < ?", filter.ExpertID, filter.FromMs, filter.ToMs)
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
	var items []domain.ExpertSlot
	err := query.Order("start_time ASC, slot_id ASC").Limit(filter.Page.Size).Offset(filter.Page.Offset()).Find(&items).Error
	return items, total, err
}
