package metrics

import "context"

type Repository interface {
	CountActiveAircraft(
		ctx context.Context,
		query ActiveAircraftQuery,
	) (ActiveAircraftObservationSummary, error)
}
