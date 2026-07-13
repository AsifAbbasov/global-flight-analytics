package trends

import "errors"

var (
	ErrInvalidHistory      = errors.New("invalid airport trend history")
	ErrInsufficientHistory = errors.New("airport trend requires at least two history entries")
	ErrIncomparableWindows = errors.New("airport trend requires equal window durations")
	ErrInvalidTime         = errors.New("invalid airport trend time")
)
