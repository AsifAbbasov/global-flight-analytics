package validator

import "errors"

var (
	ErrInvalidMinimumCompleteness = errors.New(
		"minimum valid completeness score must be between zero and one",
	)
	ErrInvalidMinimumInputQuality = errors.New(
		"minimum valid input quality score must be between zero and one",
	)
	ErrInvalidNumericTolerance = errors.New(
		"numeric tolerance must be finite and greater than zero",
	)
)
