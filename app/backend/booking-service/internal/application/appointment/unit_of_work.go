package appointment

import "errors"

var (
	ErrTxRecordNotFound = errors.New("record not found")
	errTxRecordNotFound = ErrTxRecordNotFound
)
