package appointment

import (
	"booking-service/internal/domain"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	GetAppointmentByID(appointmentID string) (*domain.Appointment, error)
	GetAppointmentBySlotID(slotID string) (*domain.Appointment, error)
	CancelAppointmentByExpert(appointmentID string, reason string) error
	CancelAppointmentByPatient(appointmentID string, patientID string, reason string) error
	GetAppointmentsByPatient(patientID string) ([]domain.Appointment, error)
	GetAppointmentsByExpert(expertID string, fromDate, toDate int64, status *domain.AppointmentStatus) ([]domain.Appointment, error)
	LockSlot(slotID string, patientID string) error
	CreateAppointment(appointment *domain.Appointment) error
	ConfirmPayment(appointmentID string) error
	HandlePaymentResult(command HandlePaymentResultCommand) error
	CancelExpiredLocks() (int64, error)
}

type pgRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &pgRepository{db: db}
}

// Lấy Appointment theo SlotID (ưu tiên lấy cuộc hẹn chưa bị huỷ, nếu không thì lấy cuộc hẹn mới nhất)
func (r *pgRepository) GetAppointmentBySlotID(slotID string) (*domain.Appointment, error) {
	var appt domain.Appointment
	err := r.db.Where("slot_id = ? AND status != ?", slotID, domain.AppointmentStatusCancelled).Order("created_at DESC").First(&appt).Error
	if err != nil {
		err = r.db.Where("slot_id = ?", slotID).Order("created_at DESC").First(&appt).Error
		if err != nil {
			return nil, err
		}
	}
	return &appt, nil
}

// Lấy Appointment theo AppointmentID
func (r *pgRepository) GetAppointmentByID(appointmentID string) (*domain.Appointment, error) {
	var appt domain.Appointment
	err := r.db.Table("Booking_Appointments").
		Select("\"Booking_Appointments\".*, \"Booking_Expert_Slots\".price").
		Joins("JOIN \"Booking_Expert_Slots\" ON \"Booking_Appointments\".slot_id = \"Booking_Expert_Slots\".slot_id").
		Where("\"Booking_Appointments\".appointment_id = ?", appointmentID).
		First(&appt).Error
	if err != nil {
		return nil, err
	}
	return &appt, nil
}

// Hủy Appointment do bác sĩ nghỉ phép (TimeOff)
func (r *pgRepository) CancelAppointmentByExpert(appointmentID string, reason string) error {
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

// Hủy Appointment do bệnh nhân (Patient)
func (r *pgRepository) CancelAppointmentByPatient(appointmentID string, patientID string, reason string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var appt domain.Appointment
		if err := tx.Where("appointment_id = ? AND patient_id = ?", appointmentID, patientID).First(&appt).Error; err != nil {
			return errors.New("không tìm thấy cuộc hẹn hoặc bạn không có quyền hủy")
		}

		if appt.Status == domain.AppointmentStatusCancelled {
			return errors.New("cuộc hẹn đã bị hủy trước đó")
		}

		canceledBy := "PATIENT"
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

// Lấy danh sách cuộc hẹn của Patient
func (r *pgRepository) GetAppointmentsByPatient(patientID string) ([]domain.Appointment, error) {
	var appts []domain.Appointment
	err := r.db.Where("patient_id = ?", patientID).Order("created_at desc").Find(&appts).Error
	return appts, err
}

// Lấy danh sách cuộc hẹn của Expert
func (r *pgRepository) GetAppointmentsByExpert(expertID string, fromDate, toDate int64, status *domain.AppointmentStatus) ([]domain.Appointment, error) {
	var appts []domain.Appointment
	query := r.db.Where("expert_id = ?", expertID)

	if fromDate > 0 {
		query = query.Where("created_at >= ?", fromDate)
	}
	if toDate > 0 {
		query = query.Where("created_at <= ?", toDate)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	err := query.Order("created_at desc").Find(&appts).Error
	return appts, err
}

// =====================================================================
// 1. KHÓA SLOT TẠM THỜI (Atomic Update - Chống Race Condition)
// Chỉ khóa nếu slot đang AVAILABLE
// Trả về lỗi nếu không có dòng nào được cập nhật (slot đã bị người khác lấy)
// =====================================================================
func (r *pgRepository) LockSlot(slotID string, patientID string) error {
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
func (r *pgRepository) CreateAppointment(appointment *domain.Appointment) error {
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
func (r *pgRepository) ConfirmPayment(appointmentID string) error {
	return r.HandlePaymentResult(HandlePaymentResultCommand{
		AppointmentID: appointmentID,
		Status:        PaymentResultSuccess,
	})
}

// HandlePaymentResult applies payment SUCCESS/FAILED state transitions atomically.
func (r *pgRepository) HandlePaymentResult(command HandlePaymentResultCommand) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var appt domain.Appointment
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("appointment_id = ?", command.AppointmentID).
			First(&appt).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}

		var slot domain.ExpertSlot
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("slot_id = ?", appt.SlotID).
			First(&slot).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: slot not found", ErrPaymentResultConflict)
			}
			return err
		}

		nowMs := time.Now().UnixMilli()
		switch command.Status {
		case PaymentResultSuccess:
			return r.applyPaymentSuccess(tx, &appt, &slot, nowMs)
		case PaymentResultFailed:
			return r.applyPaymentFailure(tx, &appt, &slot, nowMs)
		default:
			return ErrInvalidPaymentResultStatus
		}
	})
}

func (r *pgRepository) applyPaymentSuccess(tx *gorm.DB, appt *domain.Appointment, slot *domain.ExpertSlot, nowMs int64) error {
	transition, err := planPaymentResultTransition(*appt, *slot, PaymentResultSuccess, nowMs)
	if err != nil {
		return err
	}
	return r.persistPaymentResultTransition(tx, appt, transition)
}

func (r *pgRepository) applyPaymentFailure(tx *gorm.DB, appt *domain.Appointment, slot *domain.ExpertSlot, nowMs int64) error {
	transition, err := planPaymentResultTransition(*appt, *slot, PaymentResultFailed, nowMs)
	if err != nil {
		return err
	}
	return r.persistPaymentResultTransition(tx, appt, transition)
}

type paymentResultTransition struct {
	appointmentUpdates map[string]interface{}
	slotUpdates        map[string]interface{}
	expectedSlotStatus domain.SlotStatus
	noop               bool
}

func planPaymentResultTransition(appt domain.Appointment, slot domain.ExpertSlot, status PaymentResultStatus, nowMs int64) (*paymentResultTransition, error) {
	switch status {
	case PaymentResultSuccess:
		return planPaymentSuccess(appt, slot, nowMs)
	case PaymentResultFailed:
		return planPaymentFailure(appt, slot, nowMs)
	default:
		return nil, ErrInvalidPaymentResultStatus
	}
}

func planPaymentSuccess(appt domain.Appointment, slot domain.ExpertSlot, nowMs int64) (*paymentResultTransition, error) {
	switch appt.Status {
	case domain.AppointmentStatusPendingPayment:
		if slot.Status != domain.SlotStatusLocked {
			return nil, fmt.Errorf("%w: success requires LOCKED slot, got %s", ErrPaymentResultConflict, slot.Status.String())
		}
		confirmedAt := nowMs
		return &paymentResultTransition{
			appointmentUpdates: map[string]interface{}{
				"status":       domain.AppointmentStatusConfirmed,
				"updated_at":   nowMs,
				"confirmed_at": confirmedAt,
			},
			slotUpdates: map[string]interface{}{
				"status":            domain.SlotStatusOccupied,
				"locked_expires_at": nil,
				"locked_by":         nil,
			},
			expectedSlotStatus: domain.SlotStatusLocked,
		}, nil
	case domain.AppointmentStatusConfirmed:
		return &paymentResultTransition{noop: true}, nil
	case domain.AppointmentStatusCancelled:
		return nil, fmt.Errorf("%w: cancelled appointment cannot be confirmed", ErrPaymentResultConflict)
	default:
		return nil, fmt.Errorf("%w: unsupported appointment status %d", ErrPaymentResultConflict, appt.Status)
	}
}

func planPaymentFailure(appt domain.Appointment, slot domain.ExpertSlot, nowMs int64) (*paymentResultTransition, error) {
	switch appt.Status {
	case domain.AppointmentStatusPendingPayment:
		if slot.Status != domain.SlotStatusLocked {
			return nil, fmt.Errorf("%w: failure requires LOCKED slot, got %s", ErrPaymentResultConflict, slot.Status.String())
		}
		cancelledBy := "PAYMENT"
		return &paymentResultTransition{
			appointmentUpdates: map[string]interface{}{
				"status":              domain.AppointmentStatusCancelled,
				"cancellation_reason": "Payment failed",
				"cancelled_by":        &cancelledBy,
				"updated_at":          nowMs,
			},
			slotUpdates: map[string]interface{}{
				"status":            domain.SlotStatusAvailable,
				"locked_expires_at": nil,
				"locked_by":         nil,
			},
			expectedSlotStatus: domain.SlotStatusLocked,
		}, nil
	case domain.AppointmentStatusCancelled:
		return &paymentResultTransition{noop: true}, nil
	case domain.AppointmentStatusConfirmed:
		return nil, fmt.Errorf("%w: confirmed appointment cannot be cancelled by payment failure", ErrPaymentResultConflict)
	default:
		return nil, fmt.Errorf("%w: unsupported appointment status %d", ErrPaymentResultConflict, appt.Status)
	}
}

func (r *pgRepository) persistPaymentResultTransition(tx *gorm.DB, appt *domain.Appointment, transition *paymentResultTransition) error {
	if transition.noop {
		return nil
	}
	if err := tx.Model(appt).Updates(transition.appointmentUpdates).Error; err != nil {
		return err
	}

	result := tx.Model(&domain.ExpertSlot{}).
		Where("slot_id = ? AND status = ?", appt.SlotID, transition.expectedSlotStatus).
		Updates(transition.slotUpdates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: slot is no longer %s", ErrPaymentResultConflict, transition.expectedSlotStatus.String())
	}
	return nil
}

// =====================================================================
// 4. DON DEP CAC LOCK HET HAN (Goi boi Background Worker moi 60 giay)
// Tim cac slot bi khoa nhung da qua 15 phut -> mo khoa + huy appointment lien quan
// =====================================================================
func (r *pgRepository) CancelExpiredLocks() (int64, error) {
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
