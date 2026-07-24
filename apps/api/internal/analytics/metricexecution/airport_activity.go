package metricexecution

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/analyticalresult"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/metrics"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

const (
	DefaultAirportActivityRadiusKilometers = 15.0
	MaximumAirportActivityRadiusKilometers = 100.0
)

func (
	service *Service,
) AirportActivity(
	ctx context.Context,
	request AirportActivityRequest,
) (Execution[int], error) {
	if err := request.Airport.Validate(); err != nil {
		return Execution[int]{},
			fmt.Errorf(
				"%w: %v",
				ErrAirportActivityAirportInvalid,
				err,
			)
	}

	radius := request.RadiusKilometers
	if radius == 0 {
		radius = DefaultAirportActivityRadiusKilometers
	}
	if math.IsNaN(radius) ||
		math.IsInf(radius, 0) ||
		radius <= 0 ||
		radius > MaximumAirportActivityRadiusKilometers {
		return Execution[int]{},
			fmt.Errorf(
				"%w: %f",
				ErrAirportActivityRadiusInvalid,
				radius,
			)
	}

	return executeTrajectoryMetric(
		ctx,
		service,
		MetricIDAirportActivity,
		trajectoryeligibility.CapabilityAirportActivity,
		request.Trajectories,
		request.PublicationMetadata,
		prepareUniqueTrajectoryContributors(
			"%d duplicate eligible airport movement trajectories were removed before classification.",
		),
		func(
			ctx context.Context,
			allowed []trajectory.FlightTrajectory,
			evaluatedAt time.Time,
		) (metricCalculation[int], error) {
			if err := ctx.Err(); err != nil {
				return metricCalculation[int]{}, err
			}

			arrivalCount := 0
			departureCount := 0
			unrelatedCount := 0
			ambiguousCount := 0

			for _, item := range allowed {
				switch classifyAirportMovement(
					item,
					request.Airport.Latitude,
					request.Airport.Longitude,
					radius,
				) {
				case movementRoleArrival:
					arrivalCount++
				case movementRoleDeparture:
					departureCount++
				case movementRoleUnrelated:
					unrelatedCount++
				default:
					ambiguousCount++
				}
			}

			limitations := make(
				[]analyticalresult.Notice,
				0,
				2,
			)
			if unrelatedCount > 0 {
				limitations = append(
					limitations,
					analyticalresult.Notice{
						Code: NoticeCodeUnrelatedAirportTrajectoriesExcluded,
						Message: fmt.Sprintf(
							"%d eligible trajectories did not cross the airport activity geofence and were excluded.",
							unrelatedCount,
						),
					},
				)
			}
			if ambiguousCount > 0 {
				limitations = append(
					limitations,
					analyticalresult.Notice{
						Code: NoticeCodeAmbiguousAirportMovementsExcluded,
						Message: fmt.Sprintf(
							"%d eligible trajectories could not be classified as an arrival or departure and were excluded.",
							ambiguousCount,
						),
					},
				)
			}

			value := (metrics.AirportActivity{}).Calculate(
				arrivalCount,
				departureCount,
			)

			return metricCalculation[int]{
				Value:       value,
				Limitations: limitations,
			}, nil
		},
	)
}
