package appointment

import appappointment "booking-service/internal/booking/application/appointment"

type Usecase = appappointment.Usecase

func NewUsecase(repo Repository) Usecase {
	return appappointment.NewUsecase(repo)
}
