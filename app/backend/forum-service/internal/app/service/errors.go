package service

import "errors"

var (
	ErrNotFound          = errors.New("service: resource not found")
	ErrForbidden         = errors.New("service: forbidden")
	ErrCategoryHasPosts  = errors.New("service: category still has posts")
	ErrInvalidTransition = errors.New("service: invalid status transition")
)
