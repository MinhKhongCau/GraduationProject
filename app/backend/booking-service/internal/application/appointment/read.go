package appointment

import (
	bookingquery "booking-service/internal/application/query"
	appointmentdomain "booking-service/internal/domain/appointment"
	"fmt"
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
	Page      bookingquery.PageRequest
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

func (u *appointmentUsecase) ListAppointments(filter AppointmentListQuery) (bookingquery.Page[appointmentdomain.Appointment], error) {
	if filter.ActorID == "" || (filter.ActorRole != "PATIENT" && filter.ActorRole != "EXPERT") {
		return bookingquery.Page[appointmentdomain.Appointment]{}, ErrUnauthorized
	}
	if filter.ActorRole == "PATIENT" {
		filter.PatientID = filter.ActorID
	} else {
		filter.ExpertID = filter.ActorID
	}
	reader, ok := u.repo.(AppointmentReader)
	if !ok {
		return bookingquery.Page[appointmentdomain.Appointment]{}, ErrReadRepositoryUnavailable
	}
	items, total, err := reader.ListAppointments(filter)
	if err != nil {
		return bookingquery.Page[appointmentdomain.Appointment]{}, err
	}
	return bookingquery.NewPage(items, filter.Page, total), nil
}

func (u *appointmentUsecase) GetAppointmentDetail(actorID, actorRole, appointmentID string) (*appointmentdomain.Appointment, error) {
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
