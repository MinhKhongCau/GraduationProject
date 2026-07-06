package postgres

import (
	"booking-service/internal/domain"
	"errors"
	"time"

	"gorm.io/gorm"
)

// AppointmentRepository xử lý các truy vấn liên quan đến khóa slot, đặt lịch, thanh toán
type AppointmentRepository struct {
	db *gorm.DB
}

func NewAppointmentRepository(db *gorm.DB) *AppointmentRepository {
	return &AppointmentRepository{db: db}
}

// Lấy Appointment theo SlotID
func (r *AppointmentRepository) GetAppointmentBySlotID(slotID string) (*domain.Appointment, error) {
	var appt domain.Appointment
	err := r.db.Where("slot_id = ?", slotID).First(&appt).Error
	if err != nil {
		return nil, err
	}
	return &appt, nil
}

// Hủy Appointment do bác sĩ nghỉ phép (TimeOff)
func (r *AppointmentRepository) CancelAppointmentByExpert(appointmentID string, reason string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var appt domain.Appointment
		if err := tx.Where("appointment_id = ?", appointmentID).First(&appt).Error; err != nil {
			return err
		}

		canceledBy := "EXPERT"
		if err := tx.Model(&appt).Updates(map[string]interface{}{
			"status":              domain.AppointmentStatusCancelled,
			"cancellation_reason": reason,
			"cancelled_by":        &canceledBy,
			"updated_at":          time.Now().UnixMilli(),
		}).Error; err != nil {
			return err
		}

		// Trả Slot về AVAILABLE
		if err := tx.Model(&domain.ExpertSlot{}).
			Where("slot_id = ?", appt.SlotID).
			Updates(map[string]interface{}{
				"status":            domain.SlotStatusAvailable,
				"locked_expires_at": nil,
				"locked_by":         nil,
			}).Error; err != nil {
			return err
		}

		return nil
	})
}

// =====================================================================
// 1. KHÓA SLOT TẠM THỜI (Atomic Update - Chống Race Condition)
// Chỉ khóa nếu slot đang AVAILABLE
// Trả về lỗi nếu không có dòng nào được cập nhật (slot đã bị người khác lấy)
// =====================================================================
func (r *AppointmentRepository) LockSlot(slotID string, patientID string) error {
	lockedExpiresAt := time.Now().UnixMilli() + 900_000 // Khoá 15 phút = 900,000 ms

	result := r.db.Model(&domain.ExpertSlot{}).
		Where("slot_id = ? AND status = ?", slotID, domain.SlotStatusAvailable).
		Updates(map[string]interface{}{
			"status":            domain.SlotStatusLocked,
			"locked_expires_at": lockedExpiresAt,
			"locked_by":         patientID,
		})

	if result.Error != nil {
		return result.Error
	}
	// Nếu không có dòng nào bị ảnh hưởng => slot đã bị người khác giữ hoặc đã được đặt rồi
	if result.RowsAffected == 0 {
		return errors.New("slot đã bị người khác giữ hoặc đã được đặt, vui lòng chọn giờ khác")
	}
	return nil
}

// =====================================================================
// 2. TẠO CUỘC HẸN MỚI (Transaction: Kiểm tra lại lock + Insert appointment)
// Đảm bảo PatientID phải khớp với người đang giữ chỗ (LockedBy)
// =====================================================================
func (r *AppointmentRepository) CreateAppointment(appointment *domain.Appointment) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Lấy thông tin slot và kiểm tra lại trạng thái (double-check)
		var slot domain.ExpertSlot
		if err := tx.Where("slot_id = ?", appointment.SlotID).First(&slot).Error; err != nil {
			return errors.New("không tìm thấy slot: " + err.Error())
		}

		// Đảm bảo slot đang được giữ bởi đúng bệnh nhân này
		if slot.Status != domain.SlotStatusLocked || slot.LockedBy == nil || *slot.LockedBy != appointment.PatientID {
			return errors.New("slot không được giữ bởi bạn, vui lòng thực hiện lại từ đầu")
		}

		// Đảm bảo lock chưa hết hạn
		nowMs := time.Now().UnixMilli()
		if slot.LockedExpiresAt != nil && *slot.LockedExpiresAt < nowMs {
			return errors.New("phiên giữ chỗ đã hết hạn 15 phút, vui lòng chọn lại")
		}

		// Tạo cuộc hẹn với trạng thái PENDING_PAYMENT
		appointment.CreatedAt = nowMs
		appointment.UpdatedAt = nowMs
		if err := tx.Create(appointment).Error; err != nil {
			return errors.New("lỗi khi tạo cuộc hẹn: " + err.Error())
		}

		return nil
	})
}

// =====================================================================
// 3. XÁC NHẬN THANH TOÁN THÀNH CÔNG (Webhook xử lý sau khi cổng TT gọi vào)
// Cập nhật appointment -> CONFIRMED, slot -> OCCUPIED và giải phóng lock
// =====================================================================
func (r *AppointmentRepository) ConfirmPayment(appointmentID string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Lấy thông tin appointment
		var appt domain.Appointment
		if err := tx.Where("appointment_id = ?", appointmentID).First(&appt).Error; err != nil {
			return errors.New("không tìm thấy cuộc hẹn")
		}

		// Cập nhật trạng thái cuộc hẹn thành CONFIRMED
		nowMs := time.Now().UnixMilli()
		if err := tx.Model(&appt).Updates(map[string]interface{}{
			"status":       domain.AppointmentStatusConfirmed,
			"updated_at":   nowMs,
			"confirmed_at": nowMs,
		}).Error; err != nil {
			return err
		}

		// Cập nhật slot: OCCUPIED + giải phóng lock
		if err := tx.Model(&domain.ExpertSlot{}).
			Where("slot_id = ?", appt.SlotID).
			Updates(map[string]interface{}{
				"status":            domain.SlotStatusOccupied,
				"locked_expires_at": nil,
				"locked_by":         nil,
			}).Error; err != nil {
			return err
		}

		return nil
	})
}

// =====================================================================
// 4. DỌN DẸP CÁC LOCK HẾT HẠN (Gọi bởi Background Worker mỗi 60 giây)
// Tìm các slot bị khóa nhưng đã quá 15 phút -> Mở khóa + Hủy appointment liên quan
// =====================================================================
func (r *AppointmentRepository) CancelExpiredLocks() (int64, error) {
	nowMs := time.Now().UnixMilli()
	var totalCleaned int64

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Tìm tất cả các slot đã hết hạn lock
		var expiredSlots []domain.ExpertSlot
		if err := tx.
			Where("status = ? AND locked_expires_at < ?",
				domain.SlotStatusLocked, nowMs).
			Find(&expiredSlots).Error; err != nil {
			return err
		}

		if len(expiredSlots) == 0 {
			return nil // Không có gì để dọn
		}

		// Lấy danh sách slot_id để xử lý hàng loạt
		slotIDs := make([]string, 0, len(expiredSlots))
		for _, s := range expiredSlots {
			slotIDs = append(slotIDs, s.SlotID)
		}

		// 2. Hủy các Appointment PENDING_PAYMENT liên quan đến slot hết hạn
		canceledBy := "SYSTEM"
		result := tx.Model(&domain.Appointment{}).
			Where("slot_id IN ? AND status = ?", slotIDs, domain.AppointmentStatusPendingPayment).
			Updates(map[string]interface{}{
				"status":              domain.AppointmentStatusCancelled,
				"cancellation_reason": "Quá hạn thanh toán 15 phút",
				"cancelled_by":        &canceledBy,
				"updated_at":          nowMs,
			})
		if result.Error != nil {
			return result.Error
		}

		// 3. Mở khóa tất cả slot hết hạn, trả về trạng thái AVAILABLE
		result = tx.Model(&domain.ExpertSlot{}).
			Where("slot_id IN ?", slotIDs).
			Updates(map[string]interface{}{
				"status":            domain.SlotStatusAvailable,
				"locked_expires_at": nil,
				"locked_by":         nil,
			})
		if result.Error != nil {
			return result.Error
		}

		totalCleaned = result.RowsAffected
		return nil
	})

	return totalCleaned, err
}
