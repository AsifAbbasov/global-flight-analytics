package dataqualityintegration

import "errors"

var (
	ErrEvaluatedAtRequired = errors.New(
		"data quality integration evaluation time is required",
	)
	ErrNoUsableObservationTimes = errors.New(
		"data quality integration requires at least one usable observation time",
	)
)
