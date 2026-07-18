package appointment

import (
	appointmentpostgres "booking-service/internal/booking/adapter/out/postgres/appointment"
	appappointment "booking-service/internal/booking/application/appointment"

	"gorm.io/gorm"
)

type Repository = appappointment.Repository

func NewRepository(db *gorm.DB) Repository {
	return appointmentpostgres.NewRepository(db)
}
