package appointment

import (
	bookingquery "booking-service/internal/booking/application/query"
	"booking-service/internal/booking/domain"
	"fmt"
	"strings"
)

type AppointmentListQuery struct {
	ActorID   string
	ActorRole string
	FromMs    int64
	ToMs      int64
	Status    *domain.AppointmentStatus
	ExpertID  string
	PatientID string
	Page      bookingquery.PageRequest
}

type AppointmentReader interface {
	ListAppointments(query AppointmentListQuery) ([]domain.Appointment, int64, error)
}

func ParseAppointmentStatus(value string) (*domain.AppointmentStatus, error) {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" {
		return nil, nil
	}
	statuses := map[string]domain.AppointmentStatus{
		"0":               domain.AppointmentStatusPendingPayment,
		"PENDING_PAYMENT": domain.AppointmentStatusPendingPayment,
		"1":               domain.AppointmentStatusConfirmed,
		"CONFIRMED":       domain.AppointmentStatusConfirmed,
		"2":               domain.AppointmentStatusCancelled,
		"CANCELLED":       domain.AppointmentStatusCancelled,
		"3":               domain.AppointmentStatusCompleted,
		"COMPLETED":       domain.AppointmentStatusCompleted,
	}
	status, ok := statuses[value]
	if !ok {
		return nil, fmt.Errorf("%w: unsupported appointment status", bookingquery.ErrInvalidQuery)
	}
	return &status, nil
}

func (u *appointmentUsecase) ListAppointments(filter AppointmentListQuery) (bookingquery.Page[domain.Appointment], error) {
	if filter.ActorID == "" || (filter.ActorRole != "PATIENT" && filter.ActorRole != "EXPERT") {
		return bookingquery.Page[domain.Appointment]{}, ErrUnauthorized
	}
	if filter.ActorRole == "PATIENT" {
		filter.PatientID = filter.ActorID
	} else {
		filter.ExpertID = filter.ActorID
	}
	reader, ok := u.repo.(AppointmentReader)
	if !ok {
		return bookingquery.Page[domain.Appointment]{}, ErrReadRepositoryUnavailable
	}
	items, total, err := reader.ListAppointments(filter)
	if err != nil {
		return bookingquery.Page[domain.Appointment]{}, err
	}
	return bookingquery.NewPage(items, filter.Page, total), nil
}

func (u *appointmentUsecase) GetAppointmentDetail(actorID, actorRole, appointmentID string) (*domain.Appointment, error) {
	if actorID == "" || (actorRole != "PATIENT" && actorRole != "EXPERT" && actorRole != "ADMIN") {
		return nil, ErrUnauthorized
	}
	result, err := u.repo.GetAppointmentByID(appointmentID)
	if err != nil {
		return nil, err
	}
	if actorRole == "PATIENT" && result.PatientID != actorID {
		return nil, ErrUnauthorized
	}
	if actorRole == "EXPERT" && result.ExpertID != actorID {
		return nil, ErrUnauthorized
	}
	return result, nil
}
