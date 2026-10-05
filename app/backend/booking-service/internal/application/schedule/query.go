package schedule

import (
	bookingquery "booking-service/internal/application/query"
	scheduledomain "booking-service/internal/domain/schedule"
	"errors"
)

type AvailabilityListQuery struct {
	ExpertID        string
	Active          *bool
	EffectiveFromMs int64
	EffectiveToMs   int64
	Page            bookingquery.PageRequest
}

type TemplateListQuery struct {
	Active *bool
	Page   bookingquery.PageRequest
}

type scheduleReader interface {
	ListAvailabilities(query AvailabilityListQuery) ([]scheduledomain.Availability, int64, error)
	ListTimeTemplates(query TemplateListQuery) ([]scheduledomain.TimeTemplate, int64, error)
}

func (u *scheduleUsecase) ListAvailabilities(filter AvailabilityListQuery) (bookingquery.Page[scheduledomain.Availability], error) {
	reader, ok := u.repo.(scheduleReader)
	if !ok {
		return bookingquery.Page[scheduledomain.Availability]{}, errors.New("schedule reader unavailable")
	}
	items, total, err := reader.ListAvailabilities(filter)
	if err != nil {
		return bookingquery.Page[scheduledomain.Availability]{}, err
	}
	return bookingquery.NewPage(items, filter.Page, total), nil
}

func (u *scheduleUsecase) ListTimeTemplates(filter TemplateListQuery) (bookingquery.Page[scheduledomain.TimeTemplate], error) {
	reader, ok := u.repo.(scheduleReader)
	if !ok {
		return bookingquery.Page[scheduledomain.TimeTemplate]{}, errors.New("schedule reader unavailable")
	}
	items, total, err := reader.ListTimeTemplates(filter)
	if err != nil {
		return bookingquery.Page[scheduledomain.TimeTemplate]{}, err
	}
	return bookingquery.NewPage(items, filter.Page, total), nil
}
