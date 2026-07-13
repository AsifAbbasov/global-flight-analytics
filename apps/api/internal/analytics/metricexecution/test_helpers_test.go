package metricexecution

import (
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/confidencereport"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/executor"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/scopeguard"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

type metricEvaluatorFunction func(
	item trajectory.FlightTrajectory,
	now time.Time,
) trajectoryeligibility.Evaluation

func (
	function metricEvaluatorFunction,
) Evaluate(
	item trajectory.FlightTrajectory,
	now time.Time,
) trajectoryeligibility.Evaluation {
	return function(
		item,
		now,
	)
}

func metricTestService(
	t *testing.T,
	decision func(
		item trajectory.FlightTrajectory,
		capability trajectoryeligibility.Capability,
	) trajectoryeligibility.Decision,
) *Service {
	t.Helper()

	guard, err := scopeguard.New(
		scopeguard.Config{
			Evaluator: metricEvaluatorFunction(
				func(
					item trajectory.FlightTrajectory,
					now time.Time,
				) trajectoryeligibility.Evaluation {
					decisions := make(
						[]trajectoryeligibility.Decision,
						0,
						len(trajectoryeligibility.Capabilities()),
					)

					for _, capability := range trajectoryeligibility.Capabilities() {
						decisions = append(
							decisions,
							decision(
								item,
								capability,
							),
						)
					}

					return trajectoryeligibility.Evaluation{
						Decisions: decisions,
					}
				},
			),
			Now: func() time.Time {
				return metricTestTime()
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"expected scope guard, got %v",
			err,
		)
	}

	service, err := New(
		executor.NewWithDependencies(
			nil,
			guard,
			confidencereport.NewDefault(),
		),
	)
	if err != nil {
		t.Fatalf(
			"expected metric service, got %v",
			err,
		)
	}

	return service
}

func allowUnlessDeniedICAO(
	item trajectory.FlightTrajectory,
	capability trajectoryeligibility.Capability,
) trajectoryeligibility.Decision {
	allowed := item.ICAO24 != "DENY"
	reasons := []trajectoryeligibility.ReasonCode(nil)

	if !allowed {
		reasons = []trajectoryeligibility.ReasonCode{
			trajectoryeligibility.
				ReasonLowQualityScore,
		}
	}

	return trajectoryeligibility.Decision{
		Capability: capability,
		Allowed:    allowed,
		Reasons:    reasons,
	}
}

func healthyMetricTrajectory(
	identityCharacter string,
	icao24 string,
) trajectory.FlightTrajectory {
	return trajectory.FlightTrajectory{
		IdentityKey: "flight-identity-" +
			strings.Repeat(
				identityCharacter,
				64,
			),
		IdentityBasis: trajectory.
			FlightIdentityBasisSourceFlightID,
		FlightID: "11111111-1111-4111-8111-" +
			strings.Repeat(
				identityCharacter,
				12,
			),
		ICAO24:       icao24,
		QualityScore: 0.95,
		PointCount:   6,
		StartTime: metricTestTime().
			Add(-4 * time.Minute),
		EndTime: metricTestTime().
			Add(-30 * time.Second),
	}
}

func metricTestTime() time.Time {
	return time.Date(
		2026,
		time.July,
		13,
		18,
		30,
		0,
		0,
		time.UTC,
	)
}
