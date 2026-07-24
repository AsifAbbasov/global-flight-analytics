package metricexecution

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/analyticalresult"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func TestTrafficDensityUsesOnlyEligibleUniqueAircraft(
	t *testing.T,
) {
	service := metricTestService(
		t,
		allowUnlessDeniedICAO,
	)

	first := healthyMetricTrajectory(
		"a",
		"ALLOW-1",
	)
	second := healthyMetricTrajectory(
		"b",
		"ALLOW-2",
	)

	execution, err := service.TrafficDensity(
		context.Background(),
		TrafficDensityRequest{
			Trajectories: []trajectory.FlightTrajectory{
				first,
				first,
				second,
			},
			AreaSquareKilometers: 100,
		},
	)
	if err != nil {
		t.Fatalf(
			"expected traffic density execution, got %v",
			err,
		)
	}

	if math.Abs(
		execution.Result.Value-0.02,
	) > 0.000001 {
		t.Fatalf(
			"expected density 0.02, got %f",
			execution.Result.Value,
		)
	}

	if execution.Result.Status !=
		analyticalresult.StatusLimited {
		t.Fatalf(
			"expected duplicate warning to limit result, got %s",
			execution.Result.Status,
		)
	}

	if execution.Scope.InputCount != 2 ||
		execution.Scope.AllowedCount != 2 {
		t.Fatalf(
			"unexpected scope summary: %#v",
			execution.Scope,
		)
	}
}

func TestTrafficDensityMapsMetricValidationFailure(
	t *testing.T,
) {
	service := metricTestService(
		t,
		allowUnlessDeniedICAO,
	)

	execution, err := service.TrafficDensity(
		context.Background(),
		TrafficDensityRequest{
			Trajectories: []trajectory.FlightTrajectory{
				healthyMetricTrajectory(
					"a",
					"ALLOW",
				),
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"expected typed failed result, got %v",
			err,
		)
	}

	if !execution.IsFailed() ||
		execution.Result.Failure == nil {
		t.Fatalf(
			"expected failed density result, got %#v",
			execution.Result,
		)
	}

	if execution.Result.Failure.Code !=
		"analytical_operation_failed" {
		t.Fatalf(
			"unexpected failure: %#v",
			execution.Result.Failure,
		)
	}
}

func TestTrafficDensityFiltersBeforeAircraftDeduplication(
	t *testing.T,
) {
	olderEligible := healthyMetricTrajectory(
		"a",
		"SAME",
	)
	newerDenied := healthyMetricTrajectory(
		"b",
		"SAME",
	)
	olderEligible.EndTime = metricTestTime().
		Add(-2 * time.Minute)
	newerDenied.EndTime = metricTestTime().
		Add(-time.Minute)

	service := metricTestService(
		t,
		func(
			item trajectory.FlightTrajectory,
			capability trajectoryeligibility.Capability,
		) trajectoryeligibility.Decision {
			allowed :=
				item.IdentityKey ==
					olderEligible.IdentityKey
			reasons := []trajectoryeligibility.ReasonCode(nil)
			if !allowed {
				reasons = []trajectoryeligibility.ReasonCode{
					trajectoryeligibility.ReasonLowQualityScore,
				}
			}

			return trajectoryeligibility.Decision{
				Capability: capability,
				Allowed:    allowed,
				Reasons:    reasons,
			}
		},
	)

	execution, err := service.TrafficDensity(
		context.Background(),
		TrafficDensityRequest{
			Trajectories: []trajectory.FlightTrajectory{
				olderEligible,
				newerDenied,
			},
			AreaSquareKilometers: 100,
		},
	)
	if err != nil {
		t.Fatalf(
			"expected traffic density execution, got %v",
			err,
		)
	}

	if math.Abs(execution.Result.Value-0.01) >
		0.000001 {
		t.Fatalf(
			"expected density 0.01, got %f",
			execution.Result.Value,
		)
	}

	if execution.Scope.InputCount != 2 ||
		execution.Scope.AllowedCount != 1 ||
		execution.Scope.DeniedCount != 1 {
		t.Fatalf(
			"unexpected contributor scope: %#v",
			execution.Scope,
		)
	}
}
