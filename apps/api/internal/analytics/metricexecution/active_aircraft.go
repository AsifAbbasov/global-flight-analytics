package metricexecution

import (
	"context"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/metrics"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func (service *Service) ActiveAircraft(
	ctx context.Context,
	request ActiveAircraftRequest,
) (Execution[int], error) {
	return executeTrajectoryMetric(
		ctx,
		service,
		MetricIDActiveAircraft,
		trajectoryeligibility.CapabilityTrafficMetrics,
		request.Trajectories,
		request.PublicationMetadata,
		prepareUniqueAircraftContributors(
			"%d additional eligible trajectories for already counted aircraft were removed before calculating active aircraft.",
		),
		func(
			ctx context.Context,
			allowed []trajectory.FlightTrajectory,
			evaluatedAt time.Time,
		) (metricCalculation[int], error) {
			if err := ctx.Err(); err != nil {
				return metricCalculation[int]{}, err
			}

			value := (metrics.ActiveAircraft{}).Calculate(
				len(allowed),
			)

			return metricCalculation[int]{Value: value}, nil
		},
	)
}
