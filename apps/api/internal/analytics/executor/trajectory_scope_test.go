package executor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/calculator"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/registry"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/scopeguard"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

type executorEvaluatorFunction func(
	item trajectory.FlightTrajectory,
	now time.Time,
) trajectoryeligibility.Evaluation

func (function executorEvaluatorFunction) Evaluate(
	item trajectory.FlightTrajectory,
	now time.Time,
) trajectoryeligibility.Evaluation {
	return function(item, now)
}

func TestNewStoresDefaultScopeGuard(t *testing.T) {
	calc := calculator.New(registry.New())
	executor := New(calc)

	if executor.ScopeGuard() == nil {
		t.Fatal("expected default analytical scope guard")
	}
	if executor.Calculator() != calc {
		t.Fatal("expected calculator to remain stored")
	}
}

func TestNewWithScopeGuardStoresProvidedGuard(t *testing.T) {
	guard := mustExecutorScopeGuard(t, executorEvaluatorFunction(func(
		item trajectory.FlightTrajectory,
		now time.Time,
	) trajectoryeligibility.Evaluation {
		return executorAllowedEvaluation(
			trajectoryeligibility.CapabilityTrafficMetrics,
		)
	}))

	executor := NewWithScopeGuard(nil, guard)
	if executor.ScopeGuard() != guard {
		t.Fatal("expected provided analytical scope guard")
	}
}

func TestNewWithScopeGuardReplacesNilGuardWithDefault(t *testing.T) {
	executor := NewWithScopeGuard(nil, nil)
	if executor.ScopeGuard() == nil {
		t.Fatal("expected nil guard to be replaced with default guard")
	}
}

func TestExecuteTrajectoryBlocksDeniedOperation(t *testing.T) {
	guard := mustExecutorScopeGuard(t, executorEvaluatorFunction(func(
		item trajectory.FlightTrajectory,
		now time.Time,
	) trajectoryeligibility.Evaluation {
		return trajectoryeligibility.Evaluation{
			Decisions: []trajectoryeligibility.Decision{{
				Capability: trajectoryeligibility.CapabilityRouteInference,
				Allowed:    false,
				Reasons: []trajectoryeligibility.ReasonCode{
					trajectoryeligibility.ReasonMissingIdentity,
				},
			}},
		}
	}))
	executor := NewWithScopeGuard(nil, guard)

	called := false
	decision, err := executor.ExecuteTrajectory(
		context.Background(),
		trajectory.FlightTrajectory{ICAO24: "ABC123"},
		trajectoryeligibility.CapabilityRouteInference,
		func(ctx context.Context, item trajectory.FlightTrajectory) error {
			called = true
			return nil
		},
	)
	if !errors.Is(err, scopeguard.ErrDenied) {
		t.Fatalf("expected analytical scope denial, got %v", err)
	}
	if called {
		t.Fatal("expected denied executor operation not to run")
	}
	if decision.Allowed {
		t.Fatal("expected denied executor decision")
	}
}

func TestExecuteTrajectoryRunsAllowedOperation(t *testing.T) {
	guard := mustExecutorScopeGuard(t, executorEvaluatorFunction(func(
		item trajectory.FlightTrajectory,
		now time.Time,
	) trajectoryeligibility.Evaluation {
		return executorAllowedEvaluation(
			trajectoryeligibility.CapabilityTrafficMetrics,
		)
	}))
	executor := NewWithScopeGuard(nil, guard)

	called := 0
	decision, err := executor.ExecuteTrajectory(
		context.Background(),
		trajectory.FlightTrajectory{ICAO24: "ABC123"},
		trajectoryeligibility.CapabilityTrafficMetrics,
		func(ctx context.Context, item trajectory.FlightTrajectory) error {
			called++
			return nil
		},
	)
	if err != nil {
		t.Fatalf("expected allowed executor operation, got %v", err)
	}
	if !decision.Allowed || called != 1 {
		t.Fatalf("expected one allowed operation, decision=%#v calls=%d", decision, called)
	}
}

func TestFilterTrajectoriesDelegatesToScopeGuard(t *testing.T) {
	guard := mustExecutorScopeGuard(t, executorEvaluatorFunction(func(
		item trajectory.FlightTrajectory,
		now time.Time,
	) trajectoryeligibility.Evaluation {
		allowed := item.ICAO24 != "DENY"
		decision := trajectoryeligibility.Decision{
			Capability: trajectoryeligibility.CapabilityHistoricalAggregation,
			Allowed:    allowed,
		}
		if !allowed {
			decision.Reasons = []trajectoryeligibility.ReasonCode{
				trajectoryeligibility.ReasonLowQualityScore,
			}
		}
		return trajectoryeligibility.Evaluation{
			Decisions: []trajectoryeligibility.Decision{decision},
		}
	}))
	executor := NewWithScopeGuard(nil, guard)

	result, err := executor.FilterTrajectories(
		[]trajectory.FlightTrajectory{
			{ICAO24: "ALLOW"},
			{ICAO24: "DENY"},
		},
		trajectoryeligibility.CapabilityHistoricalAggregation,
	)
	if err != nil {
		t.Fatalf("expected successful filtering, got %v", err)
	}
	if result.AllowedCount() != 1 || result.DeniedCount() != 1 {
		t.Fatalf("unexpected filter counts: allowed=%d denied=%d", result.AllowedCount(), result.DeniedCount())
	}
}

func mustExecutorScopeGuard(
	t *testing.T,
	evaluator scopeguard.Evaluator,
) *scopeguard.Guard {
	t.Helper()
	guard, err := scopeguard.New(scopeguard.Config{
		Evaluator: evaluator,
		Now: func() time.Time {
			return time.Date(2026, time.July, 13, 17, 0, 0, 0, time.UTC)
		},
	})
	if err != nil {
		t.Fatalf("expected scope guard creation, got %v", err)
	}
	return guard
}

func executorAllowedEvaluation(
	capability trajectoryeligibility.Capability,
) trajectoryeligibility.Evaluation {
	return trajectoryeligibility.Evaluation{
		Decisions: []trajectoryeligibility.Decision{{
			Capability: capability,
			Allowed:    true,
		}},
	}
}
