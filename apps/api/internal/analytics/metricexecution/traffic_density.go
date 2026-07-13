package metricexecution

import (
	"context"
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/analyticalresult"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/metrics"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/snapshot"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func (service *Service) TrafficDensity(
	ctx context.Context,
	request TrafficDensityRequest,
) (Execution[float64], error) {
	unique, removed :=
		uniqueAircraftTrajectories(
			request.Trajectories,
		)

	metadata := request.PublicationMetadata
	if removed > 0 {
		metadata.Warnings = mergeNotices(
			metadata.Warnings,
			[]analyticalresult.Notice{
				{
					Code: NoticeCodeDuplicateTrajectoriesRemoved,
					Message: fmt.Sprintf(
						"%d additional trajectories for already counted aircraft were removed before calculating traffic density.",
						removed,
					),
				},
			},
		)
	}

	return executeTrajectoryMetric(
		ctx,
		service,
		metrics.TrafficDensityMetricID,
		trajectoryeligibility.CapabilityTrafficMetrics,
		unique,
		metadata,
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
