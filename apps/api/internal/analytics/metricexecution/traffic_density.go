package metricexecution

import (
	"context"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/metrics"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/snapshot"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func (service *Service) TrafficDensity(
	ctx context.Context,
	request TrafficDensityRequest,
) (Execution[float64], error) {
	return executeTrajectoryMetric(
		ctx,
		service,
		metrics.TrafficDensityMetricID,
		trajectoryeligibility.CapabilityTrafficMetrics,
		request.Trajectories,
		request.PublicationMetadata,
		prepareUniqueAircraftContributors(
			"%d additional eligible trajectories for already counted aircraft were removed before calculating traffic density.",
		),
		func(
			ctx context.Context,
			allowed []trajectory.FlightTrajectory,
			evaluatedAt time.Time,
		) (metricCalculation[float64], error) {
			if err := ctx.Err(); err != nil {
				return metricCalculation[float64]{}, err
			}

			value, err := (metrics.TrafficDensityMetric{}).Calculate(
				snapshot.Snapshot{
					ActiveAircraft:       len(allowed),
					AreaSquareKilometers: request.AreaSquareKilometers,
				},
			)
			if err != nil {
				return metricCalculation[float64]{}, err
			}

			return metricCalculation[float64]{Value: value}, nil
		},
	)
}
