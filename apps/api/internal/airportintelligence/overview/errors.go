package overview

import "errors"

var (
	ErrInvalidInput       = errors.New("invalid airport overview input")
	ErrAirportMismatch    = errors.New("airport overview ICAO codes do not match")
	ErrRankingNotFound    = errors.New("airport ranking entry not found")
	ErrIncomparableWindow = errors.New("airport overview windows do not match")
	ErrInvalidTime        = errors.New("invalid airport overview time")
)
