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

type movementRole string

const (
	movementRoleArrival   movementRole = "arrival"
	movementRoleDeparture movementRole = "departure"
)

func (
	service *Service,
) AirportActivity(
	ctx context.Context,
	request AirportActivityRequest,
) (Execution[int], error) {
	arrivals, removedArrivals :=
		uniqueTrajectories(
			request.Arrivals,
		)
	departures, removedDepartures :=
		uniqueTrajectories(
			request.Departures,
		)

	roles := make(
		map[string]movementRole,
		len(arrivals)+len(departures),
	)
	combined := make(
		[]trajectory.FlightTrajectory,
		0,
		len(arrivals)+len(departures),
	)

	for index, item := range arrivals {
		key, stable :=
			trajectoryContributorKey(
				item,
			)
		if !stable {
			key = fmt.Sprintf(
				"arrival-unkeyed:%d",
				index,
			)
		}

		roles[key] = movementRoleArrival
		combined = append(
			combined,
			item,
		)
	}

	for index, item := range departures {
		key, stable :=
			trajectoryContributorKey(
				item,
			)
		if !stable {
			key = fmt.Sprintf(
				"departure-unkeyed:%d",
				index,
			)
		}

		if existing, exists := roles[key]; exists &&
			existing != movementRoleDeparture {
			return Execution[int]{},
				fmt.Errorf(
					"%w: contributor=%s",
					ErrAirportMovementConflict,
					key,
				)
		}

		roles[key] = movementRoleDeparture
		combined = append(
			combined,
			item,
		)
	}

	removed :=
		removedArrivals +
			removedDepartures
	metadata := request.
		PublicationMetadata

	if removed > 0 {
		metadata.Warnings = mergeNotices(
			metadata.Warnings,
			[]analyticalresult.Notice{
				{
					Code: NoticeCodeDuplicateTrajectoriesRemoved,
					Message: fmt.Sprintf(
						"%d duplicate airport movement contributors were removed before calculation.",
						removed,
					),
				},
			},
		)
	}

	return executeTrajectoryMetric(
		ctx,
		service,
		MetricIDAirportActivity,
		trajectoryeligibility.
			CapabilityAirportActivity,
		combined,
		metadata,
		nil,
		func(
			ctx context.Context,
			allowed []trajectory.FlightTrajectory,
			evaluatedAt time.Time,
		) (metricCalculation[int], error) {
			if err := ctx.Err(); err != nil {
				return metricCalculation[int]{},
					err
			}

			arrivalCount := 0
			departureCount := 0

			for index, item := range allowed {
				key, stable :=
					trajectoryContributorKey(
						item,
					)
				if !stable {
					key = fmt.Sprintf(
						"allowed-unkeyed:%d",
						index,
					)
				}

				role, exists := roles[key]
				if !exists {
					return metricCalculation[int]{},
						fmt.Errorf(
							"%w: contributor=%s",
							ErrAirportMovementRoleMissing,
							key,
						)
				}

				switch role {
				case movementRoleArrival:
					arrivalCount++

				case movementRoleDeparture:
					departureCount++
				}
			}

			value :=
				(metrics.AirportActivity{}).
					Calculate(
						arrivalCount,
						departureCount,
					)

			return metricCalculation[int]{
				Value: value,
			}, nil
		},
	)
}
