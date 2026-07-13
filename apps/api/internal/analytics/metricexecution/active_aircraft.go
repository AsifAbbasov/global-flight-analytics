package metricexecution

import (
	"context"
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/analyticalresult"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/metrics"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func (service *Service) ActiveAircraft(
	ctx context.Context,
	request ActiveAircraftRequest,
) (Execution[int], error) {
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
						"%d additional trajectories for already counted aircraft were removed before calculating active aircraft.",
						removed,
					),
				},
			},
		)
	}

	return executeTrajectoryMetric(
		ctx,
		service,
		MetricIDActiveAircraft,
		trajectoryeligibility.CapabilityTrafficMetrics,
		unique,
		metadata,
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
