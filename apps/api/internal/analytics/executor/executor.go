package executor

import (
	"context"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/calculator"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/confidencereport"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/scopeguard"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

type Executor struct {
	calculator          *calculator.Calculator
	scopeGuard          *scopeguard.Guard
	confidenceEvaluator *confidencereport.Evaluator
}

func New(
	calculator *calculator.Calculator,
) *Executor {
	return NewWithDependencies(
		calculator,
		scopeguard.NewDefault(),
		confidencereport.NewDefault(),
	)
}

func NewWithScopeGuard(
	calculator *calculator.Calculator,
	guard *scopeguard.Guard,
) *Executor {
	return NewWithDependencies(
		calculator,
		guard,
		confidencereport.NewDefault(),
	)
}

func NewWithDependencies(
	calculator *calculator.Calculator,
	guard *scopeguard.Guard,
	confidenceEvaluator *confidencereport.Evaluator,
) *Executor {
	if guard == nil {
		guard = scopeguard.NewDefault()
	}

	if confidenceEvaluator == nil {
		confidenceEvaluator =
			confidencereport.NewDefault()
	}

	return &Executor{
		calculator:          calculator,
		scopeGuard:          guard,
		confidenceEvaluator: confidenceEvaluator,
	}
}

func (
	executor *Executor,
) Calculator() *calculator.Calculator {
	return executor.calculator
}

func (
	executor *Executor,
) ScopeGuard() *scopeguard.Guard {
	return executor.scopeGuard
}

func (
	executor *Executor,
) ConfidenceEvaluator() *confidencereport.Evaluator {
	return executor.confidenceEvaluator
}

func (
	executor *Executor,
) ExecuteTrajectory(
	ctx context.Context,
	item trajectory.FlightTrajectory,
	capability trajectoryeligibility.Capability,
	operation scopeguard.Operation,
) (scopeguard.Decision, error) {
	return executor.scopeGuard.Run(
		ctx,
		item,
		capability,
		operation,
	)
}

func (
	executor *Executor,
) FilterTrajectories(
	items []trajectory.FlightTrajectory,
	capability trajectoryeligibility.Capability,
) (scopeguard.FilterResult, error) {
	return executor.scopeGuard.Filter(
		items,
		capability,
	)
}
