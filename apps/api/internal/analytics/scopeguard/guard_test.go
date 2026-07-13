package scopeguard

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

type evaluatorFunction func(
	item trajectory.FlightTrajectory,
	now time.Time,
) trajectoryeligibility.Evaluation

func (function evaluatorFunction) Evaluate(
	item trajectory.FlightTrajectory,
	now time.Time,
) trajectoryeligibility.Evaluation {
	return function(item, now)
}

type pointerEvaluator struct{}

func (*pointerEvaluator) Evaluate(
	item trajectory.FlightTrajectory,
	now time.Time,
) trajectoryeligibility.Evaluation {
	return allowedEvaluation(trajectoryeligibility.CapabilityTrafficMetrics)
}

func TestNewRejectsNilEvaluator(t *testing.T) {
	guard, err := New(Config{})
	if !errors.Is(err, ErrEvaluatorRequired) {
		t.Fatalf("expected evaluator required error, got %v", err)
	}
	if guard != nil {
		t.Fatal("expected nil guard")
	}
}

func TestNewRejectsTypedNilEvaluator(t *testing.T) {
	var evaluator *pointerEvaluator
	guard, err := New(Config{Evaluator: evaluator})
	if !errors.Is(err, ErrEvaluatorRequired) {
		t.Fatalf("expected evaluator required error, got %v", err)
	}
	if guard != nil {
		t.Fatal("expected nil guard")
	}
}

func TestCheckUsesOneUTCClockValueAndReturnsIndependentReasons(t *testing.T) {
	localTime := time.Date(2026, time.July, 13, 19, 30, 0, 0, time.FixedZone("test", 4*60*60))
	calls := 0
	evaluator := evaluatorFunction(func(
		item trajectory.FlightTrajectory,
		now time.Time,
	) trajectoryeligibility.Evaluation {
		calls++
		if !now.Equal(localTime.UTC()) || now.Location() != time.UTC {
			t.Fatalf("expected UTC evaluation time %s, got %s", localTime.UTC(), now)
		}
		return deniedEvaluation(
			trajectoryeligibility.CapabilityRouteInference,
			trajectoryeligibility.ReasonLowQualityScore,
		)
	})

	guard, err := New(Config{
		Evaluator: evaluator,
		Now:       func() time.Time { return localTime },
	})
	if err != nil {
		t.Fatalf("expected guard creation, got %v", err)
	}

	decision, err := guard.Check(
		trajectory.FlightTrajectory{ICAO24: "ABC123"},
		trajectoryeligibility.CapabilityRouteInference,
	)
	if err != nil {
		t.Fatalf("expected no check error, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected one evaluator call, got %d", calls)
	}
	if decision.Allowed {
		t.Fatal("expected denied decision")
	}
	if !decision.HasReason(trajectoryeligibility.ReasonLowQualityScore) {
		t.Fatalf("expected low quality reason, got %v", decision.Reasons)
	}
	if !decision.EvaluatedAt.Equal(localTime.UTC()) {
		t.Fatalf("expected evaluated time %s, got %s", localTime.UTC(), decision.EvaluatedAt)
	}

	decision.Reasons[0] = trajectoryeligibility.ReasonMissingIdentity
	second, err := guard.Check(
		trajectory.FlightTrajectory{ICAO24: "ABC123"},
		trajectoryeligibility.CapabilityRouteInference,
	)
	if err != nil {
		t.Fatalf("expected no second check error, got %v", err)
	}
	if !second.HasReason(trajectoryeligibility.ReasonLowQualityScore) {
		t.Fatal("expected evaluator reason slices to remain independent")
	}
}

func TestCheckRejectsUnknownCapabilityWithoutEvaluation(t *testing.T) {
	calls := 0
	guard, err := New(Config{
		Evaluator: evaluatorFunction(func(
			item trajectory.FlightTrajectory,
			now time.Time,
		) trajectoryeligibility.Evaluation {
			calls++
			return trajectoryeligibility.Evaluation{}
		}),
	})
	if err != nil {
		t.Fatalf("expected guard creation, got %v", err)
	}

	_, err = guard.Check(
		trajectory.FlightTrajectory{},
		trajectoryeligibility.Capability("unknown"),
	)
	if !errors.Is(err, ErrCapabilityUnknown) {
		t.Fatalf("expected unknown capability error, got %v", err)
	}
	if calls != 0 {
		t.Fatalf("expected no evaluator calls, got %d", calls)
	}
}

func TestCheckReturnsMissingDecisionError(t *testing.T) {
	guard, err := New(Config{
		Evaluator: evaluatorFunction(func(
			item trajectory.FlightTrajectory,
			now time.Time,
		) trajectoryeligibility.Evaluation {
			return trajectoryeligibility.Evaluation{}
		}),
	})
	if err != nil {
		t.Fatalf("expected guard creation, got %v", err)
	}

	_, err = guard.Check(
		trajectory.FlightTrajectory{},
		trajectoryeligibility.CapabilityTrafficMetrics,
	)
	if !errors.Is(err, ErrDecisionMissing) {
		t.Fatalf("expected missing decision error, got %v", err)
	}
}

func TestRequireReturnsTypedDeniedError(t *testing.T) {
	evaluatedAt := time.Date(2026, time.July, 13, 15, 0, 0, 0, time.UTC)
	guard := mustGuard(t, Config{
		Evaluator: evaluatorFunction(func(
			item trajectory.FlightTrajectory,
			now time.Time,
		) trajectoryeligibility.Evaluation {
			return deniedEvaluation(
				trajectoryeligibility.CapabilityProjection,
				trajectoryeligibility.ReasonMissingAltitude,
			)
		}),
		Now: func() time.Time { return evaluatedAt },
	})

	item := trajectory.FlightTrajectory{
		IdentityKey: "flight-identity-example",
		ICAO24:      "ABC123",
	}
	decision, err := guard.Require(
		item,
		trajectoryeligibility.CapabilityProjection,
	)
	if !errors.Is(err, ErrDenied) {
		t.Fatalf("expected denied error, got %v", err)
	}

	var deniedErr *DeniedError
	if !errors.As(err, &deniedErr) {
		t.Fatalf("expected typed denied error, got %T", err)
	}
	if deniedErr.IdentityKey != item.IdentityKey || deniedErr.ICAO24 != item.ICAO24 {
		t.Fatalf("unexpected denied identity fields: %#v", deniedErr)
	}
	if !decision.HasReason(trajectoryeligibility.ReasonMissingAltitude) {
		t.Fatalf("expected missing altitude decision, got %v", decision.Reasons)
	}

	deniedErr.Reasons[0] = trajectoryeligibility.ReasonLowQualityScore
	if !decision.HasReason(trajectoryeligibility.ReasonMissingAltitude) {
		t.Fatal("expected denied error reasons not to alias decision reasons")
	}
}

func TestRunBlocksDeniedOperation(t *testing.T) {
	guard := mustGuard(t, Config{
		Evaluator: evaluatorFunction(func(
			item trajectory.FlightTrajectory,
			now time.Time,
		) trajectoryeligibility.Evaluation {
			return deniedEvaluation(
				trajectoryeligibility.CapabilityRouteInference,
				trajectoryeligibility.ReasonMissingIdentity,
			)
		}),
	})

	called := false
	decision, err := guard.Run(
		context.Background(),
		trajectory.FlightTrajectory{ICAO24: "ABC123"},
		trajectoryeligibility.CapabilityRouteInference,
		func(ctx context.Context, item trajectory.FlightTrajectory) error {
			called = true
			return nil
		},
	)
	if !errors.Is(err, ErrDenied) {
		t.Fatalf("expected denied error, got %v", err)
	}
	if called {
		t.Fatal("expected denied operation not to execute")
	}
	if decision.Allowed {
		t.Fatal("expected denied decision")
	}
}

func TestRunExecutesAllowedOperationWithNonNilContext(t *testing.T) {
	guard := mustGuard(t, Config{
		Evaluator: evaluatorFunction(func(
			item trajectory.FlightTrajectory,
			now time.Time,
		) trajectoryeligibility.Evaluation {
			return allowedEvaluation(
				trajectoryeligibility.CapabilityTrafficMetrics,
			)
		}),
	})

	called := 0
	item := trajectory.FlightTrajectory{ICAO24: "ABC123"}
	decision, err := guard.Run(
		nil,
		item,
		trajectoryeligibility.CapabilityTrafficMetrics,
		func(ctx context.Context, actual trajectory.FlightTrajectory) error {
			called++
			if ctx == nil {
				t.Fatal("expected non-nil context")
			}
			if actual.ICAO24 != item.ICAO24 {
				t.Fatalf("expected ICAO24 %s, got %s", item.ICAO24, actual.ICAO24)
			}
			return nil
		},
	)
	if err != nil {
		t.Fatalf("expected allowed operation, got %v", err)
	}
	if !decision.Allowed || called != 1 {
		t.Fatalf("expected one allowed operation, decision=%#v calls=%d", decision, called)
	}
}

func TestRunValidatesOperationBeforeEvaluation(t *testing.T) {
	calls := 0
	guard := mustGuard(t, Config{
		Evaluator: evaluatorFunction(func(
			item trajectory.FlightTrajectory,
			now time.Time,
		) trajectoryeligibility.Evaluation {
			calls++
			return allowedEvaluation(trajectoryeligibility.CapabilityTrafficMetrics)
		}),
	})

	_, err := guard.Run(
		context.Background(),
		trajectory.FlightTrajectory{},
		trajectoryeligibility.CapabilityTrafficMetrics,
		nil,
	)
	if !errors.Is(err, ErrOperationRequired) {
		t.Fatalf("expected operation required error, got %v", err)
	}
	if calls != 0 {
		t.Fatalf("expected no evaluation, got %d calls", calls)
	}
}

func TestRunWrapsOperationError(t *testing.T) {
	expected := errors.New("calculation failed")
	guard := mustGuard(t, Config{
		Evaluator: evaluatorFunction(func(
			item trajectory.FlightTrajectory,
			now time.Time,
		) trajectoryeligibility.Evaluation {
			return allowedEvaluation(trajectoryeligibility.CapabilityTrafficMetrics)
		}),
	})

	_, err := guard.Run(
		context.Background(),
		trajectory.FlightTrajectory{},
		trajectoryeligibility.CapabilityTrafficMetrics,
		func(ctx context.Context, item trajectory.FlightTrajectory) error {
			return expected
		},
	)
	if !errors.Is(err, expected) {
		t.Fatalf("expected wrapped operation error, got %v", err)
	}
}

func TestFilterUsesOneTimestampPreservesOrderAndDoesNotMutateInput(t *testing.T) {
	evaluatedAt := time.Date(2026, time.July, 13, 16, 0, 0, 0, time.UTC)
	times := make([]time.Time, 0, 3)
	guard := mustGuard(t, Config{
		Evaluator: evaluatorFunction(func(
			item trajectory.FlightTrajectory,
			now time.Time,
		) trajectoryeligibility.Evaluation {
			times = append(times, now)
			if item.ICAO24 == "DENY" {
				return deniedEvaluation(
					trajectoryeligibility.CapabilityHistoricalAggregation,
					trajectoryeligibility.ReasonLowQualityScore,
				)
			}
			return allowedEvaluation(
				trajectoryeligibility.CapabilityHistoricalAggregation,
			)
		}),
		Now: func() time.Time { return evaluatedAt },
	})

	items := []trajectory.FlightTrajectory{
		{ICAO24: "ALLOW-1"},
		{ICAO24: "DENY"},
		{ICAO24: "ALLOW-2"},
	}
	original := append([]trajectory.FlightTrajectory(nil), items...)

	result, err := guard.Filter(
		items,
		trajectoryeligibility.CapabilityHistoricalAggregation,
	)
	if err != nil {
		t.Fatalf("expected successful filtering, got %v", err)
	}
	if result.AllowedCount() != 2 || result.DeniedCount() != 1 {
		t.Fatalf("unexpected filter counts: allowed=%d denied=%d", result.AllowedCount(), result.DeniedCount())
	}
	if result.Allowed[0].ICAO24 != "ALLOW-1" || result.Allowed[1].ICAO24 != "ALLOW-2" {
		t.Fatalf("expected allowed order to be preserved, got %#v", result.Allowed)
	}
	if result.Denied[0].Trajectory.ICAO24 != "DENY" {
		t.Fatalf("expected denied item, got %#v", result.Denied)
	}
	if !result.Denied[0].Decision.HasReason(trajectoryeligibility.ReasonLowQualityScore) {
		t.Fatalf("expected denied reason, got %v", result.Denied[0].Decision.Reasons)
	}
	for _, current := range times {
		if !current.Equal(evaluatedAt) {
			t.Fatalf("expected shared evaluation time %s, got %s", evaluatedAt, current)
		}
	}
	if !reflect.DeepEqual(items, original) {
		t.Fatal("expected filter not to mutate input")
	}
}

func mustGuard(t *testing.T, config Config) *Guard {
	t.Helper()
	guard, err := New(config)
	if err != nil {
		t.Fatalf("expected guard creation, got %v", err)
	}
	return guard
}

func allowedEvaluation(capability trajectoryeligibility.Capability) trajectoryeligibility.Evaluation {
	return trajectoryeligibility.Evaluation{
		Decisions: []trajectoryeligibility.Decision{{
			Capability: capability,
			Allowed:    true,
		}},
	}
}

func deniedEvaluation(
	capability trajectoryeligibility.Capability,
	reasons ...trajectoryeligibility.ReasonCode,
) trajectoryeligibility.Evaluation {
	return trajectoryeligibility.Evaluation{
		Decisions: []trajectoryeligibility.Decision{{
			Capability: capability,
			Allowed:    false,
			Reasons: append(
				[]trajectoryeligibility.ReasonCode(nil),
				reasons...,
			),
		}},
	}
}
