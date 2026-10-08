package appointment

import (
	bookingquery "booking-service/internal/application/query"
	appointmentdomain "booking-service/internal/domain/appointment"
	"context"
	"fmt"
	"log"
	"slices"
	"strings"
)

type AppointmentListQuery struct {
	ActorID   string
	ActorRole string
	FromMs    int64
	ToMs      int64
	Status    *appointmentdomain.AppointmentStatus
	ExpertID  string
	PatientID string
	// ScopeExperts = true giới hạn trong ExpertIDs (phạm vi Admin quản lý, do usecase điền).
	ScopeExperts bool
	ExpertIDs    []string
	Page         bookingquery.PageRequest
}

type AppointmentReader interface {
	ListAppointments(query AppointmentListQuery) ([]appointmentdomain.Appointment, int64, error)
}

func ParseAppointmentStatus(value string) (*appointmentdomain.AppointmentStatus, error) {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" {
		return nil, nil
	}
	statuses := map[string]appointmentdomain.AppointmentStatus{
		"0":               appointmentdomain.AppointmentStatusPendingPayment,
		"PENDING_PAYMENT": appointmentdomain.AppointmentStatusPendingPayment,
		"1":               appointmentdomain.AppointmentStatusConfirmed,
		"CONFIRMED":       appointmentdomain.AppointmentStatusConfirmed,
		"2":               appointmentdomain.AppointmentStatusCancelled,
		"CANCELLED":       appointmentdomain.AppointmentStatusCancelled,
		"3":               appointmentdomain.AppointmentStatusCompleted,
		"COMPLETED":       appointmentdomain.AppointmentStatusCompleted,
	}
	status, ok := statuses[value]
	if !ok {
		return nil, fmt.Errorf("%w: unsupported appointment status", bookingquery.ErrInvalidQuery)
	}
	return &status, nil
}

func (u *appointmentUsecase) ListAppointments(ctx context.Context, filter AppointmentListQuery) (bookingquery.Page[appointmentdomain.Appointment], error) {
	if filter.ActorID == "" || (filter.ActorRole != "PATIENT" && filter.ActorRole != "EXPERT") {
		return bookingquery.Page[appointmentdomain.Appointment]{}, ErrUnauthorized
	}
	if filter.ActorRole == "PATIENT" {
		filter.PatientID = filter.ActorID
	} else {
		filter.ExpertID = filter.ActorID
	}
	filter.ScopeExperts, filter.ExpertIDs = false, nil
	return u.listDetails(ctx, filter)
}

func (u *appointmentUsecase) ListAdminAppointments(ctx context.Context, filter AppointmentListQuery) (bookingquery.Page[appointmentdomain.Appointment], error) {
	if filter.ActorID == "" || filter.ActorRole != "ADMIN" {
		return bookingquery.Page[appointmentdomain.Appointment]{}, ErrUnauthorized
	}
	managed, err := u.managedExperts(ctx, filter.ActorID)
	if err != nil {
		return bookingquery.Page[appointmentdomain.Appointment]{}, err
	}
	if filter.ExpertID != "" {
		if !slices.Contains(managed, filter.ExpertID) {
			return bookingquery.Page[appointmentdomain.Appointment]{}, ErrUnauthorized
		}
		managed = []string{filter.ExpertID}
	}
	filter.ExpertID = ""
	filter.ScopeExperts, filter.ExpertIDs = true, managed
	if len(managed) == 0 {
		return bookingquery.NewPage([]appointmentdomain.Appointment{}, filter.Page, 0), nil
	}
	return u.listDetails(ctx, filter)
}

// GetAppointmentDetail đọc một lịch hẹn (kèm giá/giờ khám của slot), kiểm tra quyền xem rồi gọi
// profile-service qua gRPC để gắn hồ sơ chuyên gia và tài khoản đã đặt lịch.
func (u *appointmentUsecase) GetAppointmentDetail(ctx context.Context, actorID, actorRole, appointmentID string) (*appointmentdomain.Appointment, error) {
	if actorID == "" || (actorRole != "PATIENT" && actorRole != "EXPERT" && actorRole != "ADMIN") {
		return nil, ErrUnauthorized
	}
	result, err := u.repo.GetAppointmentByID(appointmentID)
	if err != nil {
		return nil, err
	}
	switch actorRole {
	case "PATIENT":
		if result.PatientID != actorID {
			return nil, ErrUnauthorized
		}
	case "EXPERT":
		if result.ExpertID != actorID {
			return nil, ErrUnauthorized
		}
	case "ADMIN":
		managed, err := u.managedExperts(ctx, actorID)
		if err != nil {
			return nil, err
		}
		if !slices.Contains(managed, result.ExpertID) {
			return nil, ErrUnauthorized
		}
	}
	u.attachProfiles(ctx, []*appointmentdomain.Appointment{result})
	return result, nil
}

// GetAppointmentSummaries trả lịch hẹn theo danh sách id (kèm giờ khám, giá) cho payment-service.
func (u *appointmentUsecase) GetAppointmentSummaries(ctx context.Context, appointmentIDs []string) ([]appointmentdomain.Appointment, error) {
	if len(appointmentIDs) == 0 {
		return []appointmentdomain.Appointment{}, nil
	}
	reader, ok := u.repo.(AppointmentBatchReader)
	if !ok {
		return nil, ErrReadRepositoryUnavailable
	}
	return reader.GetAppointmentsByIDs(appointmentIDs)
}

// AppointmentBatchReader đọc nhiều lịch hẹn theo id (JOIN slot để có giờ khám và giá).
type AppointmentBatchReader interface {
	GetAppointmentsByIDs(appointmentIDs []string) ([]appointmentdomain.Appointment, error)
}

func (u *appointmentUsecase) listDetails(ctx context.Context, filter AppointmentListQuery) (bookingquery.Page[appointmentdomain.Appointment], error) {
	reader, ok := u.repo.(AppointmentReader)
	if !ok {
		return bookingquery.Page[appointmentdomain.Appointment]{}, ErrReadRepositoryUnavailable
	}
	items, total, err := reader.ListAppointments(filter)
	if err != nil {
		return bookingquery.Page[appointmentdomain.Appointment]{}, err
	}
	refs := make([]*appointmentdomain.Appointment, len(items))
	for i := range items {
		refs[i] = &items[i]
	}
	u.attachProfiles(ctx, refs)
	return bookingquery.NewPage(items, filter.Page, total), nil
}

func (u *appointmentUsecase) managedExperts(ctx context.Context, adminID string) ([]string, error) {
	if u.directory == nil {
		return nil, ErrProfileServiceUnavailable
	}
	managed, err := u.directory.ListManagedExpertIDs(ctx, adminID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrProfileServiceUnavailable, err)
	}
	return managed, nil
}

// attachProfiles gắn hồ sơ chuyên gia và tài khoản đặt lịch bằng một lần gọi profile-service.
// Lỗi tra cứu chỉ được log: lịch hẹn vẫn trả về, chỉ thiếu hồ sơ.
func (u *appointmentUsecase) attachProfiles(ctx context.Context, appointments []*appointmentdomain.Appointment) {
	if u.directory == nil || len(appointments) == 0 {
		return
	}
	ids := make([]string, 0, len(appointments)*2)
	for _, a := range appointments {
		ids = appendUnique(ids, a.ExpertID)
		ids = appendUnique(ids, a.PatientID)
	}
	profiles, err := u.directory.GetProfileSummaries(ctx, ids)
	if err != nil {
		log.Printf("⚠️  appointment profiles lookup failed: %v", err)
		return
	}
	for _, a := range appointments {
		a.AttachProfiles(profiles)
	}
}

func appendUnique(ids []string, id string) []string {
	if id == "" || slices.Contains(ids, id) {
		return ids
	}
	return append(ids, id)
}
