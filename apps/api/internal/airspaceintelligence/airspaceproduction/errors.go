package airspaceproduction

import "errors"

var (
	ErrObservationReaderRequired = errors.New(
		"airspace production observation reader is required",
	)
	ErrRegionResolverRequired = errors.New(
		"airspace production region resolver is required",
	)
	ErrPostgresPoolRequired = errors.New(
		"airspace production PostgreSQL pool is required",
	)
	ErrInvalidRequest = errors.New(
		"airspace production request is invalid",
	)
	ErrObservationCapacityExceeded = errors.New(
		"airspace production observation capacity was exceeded",
	)
	ErrProductionCompositionFailed = errors.New(
		"airspace production composition failed",
	)
)
