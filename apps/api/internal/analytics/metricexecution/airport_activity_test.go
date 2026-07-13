package metricexecution

import (
	"context"
	"errors"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/analyticalresult"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func TestAirportActivityCountsEligibleArrivalsAndDepartures(
	t *testing.T,
) {
	service := metricTestService(
		t,
		allowUnlessDeniedICAO,
	)

	execution, err := service.AirportActivity(
		context.Background(),
		AirportActivityRequest{
			Arrivals: []trajectory.FlightTrajectory{
				healthyMetricTrajectory(
					"a",
					"ARRIVAL",
				),
			},
			Departures: []trajectory.FlightTrajectory{
				healthyMetricTrajectory(
					"b",
					"DEPARTURE",
				),
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"expected airport activity execution, got %v",
			err,
		)
	}

	if execution.Result.Value != 2 ||
		execution.Result.Status !=
			analyticalresult.StatusComplete {
		t.Fatalf(
			"expected complete airport activity two, got %#v",
			execution.Result,
		)
	}

	if execution.Scope.AllowedCount != 2 ||
		execution.Scope.DeniedCount != 0 {
		t.Fatalf(
			"unexpected scope summary: %#v",
			execution.Scope,
		)
	}
}

func TestAirportActivityRejectsConflictingMovementClassification(
	t *testing.T,
) {
	service := metricTestService(
		t,
		allowUnlessDeniedICAO,
	)
	item := healthyMetricTrajectory(
		"a",
		"SAME",
	)

	execution, err := service.AirportActivity(
		context.Background(),
		AirportActivityRequest{
			Arrivals: []trajectory.FlightTrajectory{
				item,
			},
			Departures: []trajectory.FlightTrajectory{
				item,
			},
		},
	)

	if execution.MetricID != "" {
		t.Fatalf(
			"expected empty execution, got %#v",
			execution,
		)
	}

	if !errors.Is(
		err,
		ErrAirportMovementConflict,
	) {
		t.Fatalf(
			"expected movement conflict, got %v",
			err,
		)
	}
}
