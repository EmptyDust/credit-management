package apperrors

import "errors"

var (
	ErrConflict     = errors.New("resource conflict")
	ErrUserNotFound = errors.New("user not found")
)
