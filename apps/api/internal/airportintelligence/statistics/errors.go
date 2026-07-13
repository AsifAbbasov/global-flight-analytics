package statistics

import "errors"

var (
	ErrInvalidConfiguration = errors.New("invalid airport statistics configuration")
	ErrInvalidIdentity      = errors.New("invalid airport statistics identity")
	ErrInvalidWindow        = errors.New("invalid airport statistics window")
	ErrInvalidCounters      = errors.New("invalid airport statistics counters")
	ErrInvalidTime          = errors.New("invalid airport statistics time")
)
