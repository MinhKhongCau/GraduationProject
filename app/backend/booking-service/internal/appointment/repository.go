package appointment

import (
	appappointment "booking-service/internal/booking/application/appointment"

	"gorm.io/gorm"
)

type Repository = appappointment.Repository

type pgRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &pgRepository{db: db}
}
