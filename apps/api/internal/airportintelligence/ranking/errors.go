package ranking

import "errors"

var (
	ErrInvalidConfiguration = errors.New("invalid airport ranking configuration")
	ErrInvalidInput         = errors.New("invalid airport ranking input")
	ErrDuplicateAirport     = errors.New("duplicate airport in ranking input")
	ErrIncomparableWindow   = errors.New("incomparable airport statistics window")
)
