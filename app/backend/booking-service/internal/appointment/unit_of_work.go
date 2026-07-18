package appointment

import (
	"errors"

	appappointment "booking-service/internal/booking/application/appointment"
)

var errTxRecordNotFound = errors.New("record not found")

type UnitOfWork = appappointment.UnitOfWork

type Tx = appappointment.Tx
