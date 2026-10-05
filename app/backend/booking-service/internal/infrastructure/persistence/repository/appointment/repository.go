package repository

import (
	appappointment "booking-service/internal/application/appointment"

	"gorm.io/gorm"
)

type pgRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) appappointment.Repository {
	return &pgRepository{db: db}
}
