package branch

import "errors"

var (
	ErrNotFound          = errors.New("branch not found")
	ErrNameTaken         = errors.New("branch name already exists")
	ErrNameRequired      = errors.New("branch name is required")
	ErrNameTooLong       = errors.New("branch name must be at most 100 characters")
	ErrInvalidType       = errors.New("invalid branch type")
	ErrInvalidTime       = errors.New("official start time must be HH:MM")
	ErrLateRateRequired  = errors.New("late rate is required")
	ErrInvalidLateRate   = errors.New("late rate must be zero or greater")
	ErrLateRateTooLarge  = errors.New("late rate exceeds the maximum allowed")
	ErrLateRatePrecision = errors.New("late rate must have at most 2 decimal places")
)
