package metricexecution

import (
	"context"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/analyticalresult"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func TestActiveAircraftFiltersDeniedAndDuplicateContributors(
	t *testing.T,
) {
	service := metricTestService(
		t,
		allowUnlessDeniedICAO,
	)

	allowed := healthyMetricTrajectory(
		"a",
		"ALLOW",
	)
	duplicate := allowed
	denied := healthyMetricTrajectory(
		"b",
		"DENY",
	)

	execution, err := service.ActiveAircraft(
		context.Background(),
		ActiveAircraftRequest{
			Trajectories: []trajectory.FlightTrajectory{
				allowed,
				duplicate,
				denied,
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"expected active aircraft execution, got %v",
			err,
		)
	}

	if execution.MetricID !=
		MetricIDActiveAircraft {
		t.Fatalf(
			"unexpected metric id %s",
			execution.MetricID,
		)
	}

	if execution.Result.Value != 1 ||
		execution.Result.Status !=
			analyticalresult.StatusLimited {
		t.Fatalf(
			"expected limited active count one, got %#v",
			execution.Result,
		)
	}

	if execution.Scope.InputCount != 2 ||
		execution.Scope.AllowedCount != 1 ||
		execution.Scope.DeniedCount != 1 {
		t.Fatalf(
			"unexpected scope summary: %#v",
			execution.Scope,
		)
	}

	if !containsNotice(
		execution.Result.Warnings,
		nil,
		NoticeCodeDuplicateTrajectoriesRemoved,
	) {
		t.Fatalf(
			"expected duplicate warning, got %#v",
			execution.Result.Warnings,
		)
	}

	if !containsNotice(
		nil,
		execution.Result.Limitations,
		NoticeCodeIneligibleTrajectoriesExcluded,
	) {
		t.Fatalf(
			"expected excluded contributor limitation, got %#v",
			execution.Result.Limitations,
		)
	}

	if execution.ConfidenceReport == nil {
		t.Fatal("expected confidence report")
	}
}

func TestActiveAircraftReturnsDeniedWhenEveryContributorIsDenied(
	t *testing.T,
) {
	service := metricTestService(
		t,
		func(
			item trajectory.FlightTrajectory,
			capability trajectoryeligibility.Capability,
		) trajectoryeligibility.Decision {
			return trajectoryeligibility.Decision{
				Capability: capability,
				Allowed:    false,
				Reasons: []trajectoryeligibility.ReasonCode{
					trajectoryeligibility.
						ReasonMissingIdentity,
					trajectoryeligibility.
						ReasonLowQualityScore,
				},
			}
		},
	)

	execution, err := service.ActiveAircraft(
		context.Background(),
		ActiveAircraftRequest{
			Trajectories: []trajectory.FlightTrajectory{
				healthyMetricTrajectory(
					"a",
					"DENY-1",
				),
				healthyMetricTrajectory(
					"b",
					"DENY-2",
				),
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"expected denied active aircraft result, got %v",
			err,
		)
	}

	if !execution.IsDenied() ||
		execution.Result.HasValue {
		t.Fatalf(
			"expected denied result without value, got %#v",
			execution.Result,
		)
	}

	if execution.ConfidenceReport != nil {
		t.Fatal("expected no confidence report for denial")
	}

	if len(execution.Scope.Reasons) != 2 ||
		execution.Scope.Reasons[0].Reason !=
			trajectoryeligibility.ReasonLowQualityScore ||
		execution.Scope.Reasons[1].Reason !=
			trajectoryeligibility.ReasonMissingIdentity ||
		execution.Scope.Reasons[0].Count != 2 ||
		execution.Scope.Reasons[1].Count != 2 {
		t.Fatalf(
			"expected deterministic reason counts, got %#v",
			execution.Scope.Reasons,
		)
	}
}

func TestActiveAircraftPublishesEmptyObservationSetAsLimited(
	t *testing.T,
) {
	service := metricTestService(
		t,
		allowUnlessDeniedICAO,
	)

	execution, err := service.ActiveAircraft(
		context.Background(),
		ActiveAircraftRequest{},
	)
	if err != nil {
		t.Fatalf(
			"expected empty active aircraft result, got %v",
			err,
		)
	}

	if execution.Result.Value != 0 ||
		execution.Result.Status !=
			analyticalresult.StatusLimited ||
		execution.Result.Confidence.Level !=
			analyticalresult.ConfidenceLevelLow {
		t.Fatalf(
			"unexpected empty observation result: %#v",
			execution.Result,
		)
	}

	if !containsNotice(
		nil,
		execution.Result.Limitations,
		NoticeCodeNoTrajectoryObservations,
	) {
		t.Fatal("expected no-observations limitation")
	}
}

func TestActiveAircraftLimitsFutureTrajectoryTimestamp(
	t *testing.T,
) {
	service := metricTestService(
		t,
		allowUnlessDeniedICAO,
	)

	item := healthyMetricTrajectory(
		"a",
		"ALLOW",
	)
	item.EndTime = metricTestTime().
		Add(time.Minute)

	execution, err := service.ActiveAircraft(
		context.Background(),
		ActiveAircraftRequest{
			Trajectories: []trajectory.FlightTrajectory{
				item,
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"expected future timestamp execution, got %v",
			err,
		)
	}

	if execution.Result.Status !=
		analyticalresult.StatusLimited {
		t.Fatalf(
			"expected limited result, got %s",
			execution.Result.Status,
		)
	}

	if !containsNotice(
		nil,
		execution.Result.Limitations,
		NoticeCodeFutureObservationTime,
	) {
		t.Fatalf(
			"expected future observation limitation, got %#v",
			execution.Result.Limitations,
		)
	}
}

func TestActiveAircraftFiltersBeforeAircraftDeduplication(
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

	execution, err := service.ActiveAircraft(
		context.Background(),
		ActiveAircraftRequest{
			Trajectories: []trajectory.FlightTrajectory{
				olderEligible,
				newerDenied,
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"expected active aircraft execution, got %v",
			err,
		)
	}

	if execution.Result.Value != 1 {
		t.Fatalf(
			"expected eligible older trajectory to preserve aircraft, got %d",
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
