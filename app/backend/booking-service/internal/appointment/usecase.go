package appointment

import appappointment "booking-service/internal/booking/application/appointment"

type Usecase = appappointment.Usecase

type appointmentUsecase struct {
	repo Repository
	uow  UnitOfWork
}

func NewUsecase(repo Repository) Usecase {
	usecase := &appointmentUsecase{repo: repo}
	if uow, ok := repo.(UnitOfWork); ok {
		usecase.uow = uow
	}
	return usecase
}
