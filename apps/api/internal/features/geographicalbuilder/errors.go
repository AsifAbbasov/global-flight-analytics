package geographicalbuilder

import "errors"

var (
	ErrInvalidGeographicCellPrecision = errors.New(
		"geographic cell precision must be between zero and six",
	)
)
