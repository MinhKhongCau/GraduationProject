package appointment

import appappointment "booking-service/internal/booking/application/appointment"

var errTxRecordNotFound = appappointment.ErrTxRecordNotFound

type UnitOfWork = appappointment.UnitOfWork

type Tx = appappointment.Tx
