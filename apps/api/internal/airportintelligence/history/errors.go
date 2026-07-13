package history

import "errors"

var (
	ErrInvalidIdentity   = errors.New("invalid airport identity")
	ErrEmptyHistory      = errors.New("airport history is empty")
	ErrInvalidEntry      = errors.New("invalid airport history entry")
	ErrAirportMismatch   = errors.New("airport history entry belongs to another airport")
	ErrDuplicateWindow   = errors.New("duplicate airport history window")
	ErrOverlappingWindow = errors.New("overlapping airport history windows")
	ErrInvalidTime       = errors.New("invalid airport history time")
)
