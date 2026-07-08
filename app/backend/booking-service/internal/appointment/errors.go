package appointment

import "errors"

var (
	ErrNotFound      = errors.New("appointment not found")
	ErrInvalidStatus = errors.New("invalid appointment status for this operation")
	ErrCannotCancel  = errors.New("appointment cannot be cancelled")
	ErrUnauthorized  = errors.New("unauthorized access to appointment")
)
