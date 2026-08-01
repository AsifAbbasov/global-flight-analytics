package airspaceproduction

import "errors"

var (
	ErrObservationContextRequired = errors.New(
		"airspace observation context is required",
	)
	ErrProductionContextRequired = errors.New(
		"airspace production context is required",
	)
)
