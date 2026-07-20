package schedule

import (
	bookingquery "booking-service/internal/booking/application/query"
	"booking-service/internal/booking/domain"
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
	ListAvailabilities(query AvailabilityListQuery) ([]domain.Availability, int64, error)
	ListTimeTemplates(query TemplateListQuery) ([]domain.TimeTemplate, int64, error)
}

func (u *scheduleUsecase) ListAvailabilities(filter AvailabilityListQuery) (bookingquery.Page[domain.Availability], error) {
	reader, ok := u.repo.(scheduleReader)
	if !ok {
		return bookingquery.Page[domain.Availability]{}, errors.New("schedule reader unavailable")
	}
	items, total, err := reader.ListAvailabilities(filter)
	if err != nil {
		return bookingquery.Page[domain.Availability]{}, err
	}
	return bookingquery.NewPage(items, filter.Page, total), nil
}

func (u *scheduleUsecase) ListTimeTemplates(filter TemplateListQuery) (bookingquery.Page[domain.TimeTemplate], error) {
	reader, ok := u.repo.(scheduleReader)
	if !ok {
		return bookingquery.Page[domain.TimeTemplate]{}, errors.New("schedule reader unavailable")
	}
	items, total, err := reader.ListTimeTemplates(filter)
	if err != nil {
		return bookingquery.Page[domain.TimeTemplate]{}, err
	}
	return bookingquery.NewPage(items, filter.Page, total), nil
}
