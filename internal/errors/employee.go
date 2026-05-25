package errors

import "errors"

var (
	ErrInvalidEmployeeFullName = errors.New("invalid employee full name")
	ErrInvalidEmployeePosition = errors.New("invalid employee position")
)
