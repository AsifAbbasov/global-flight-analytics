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
	scopeGuard          *scopeguard.Guard
	confidenceEvaluator *confidencereport.Evaluator
}

func New(
	_ *calculator.Calculator,
) *Executor {
	return NewWithDependencies(
		nil,
		scopeguard.NewDefault(),
		confidencereport.NewDefault(),
	)
}

func NewWithScopeGuard(
	_ *calculator.Calculator,
	guard *scopeguard.Guard,
) *Executor {
	return NewWithDependencies(
		nil,
		guard,
		confidencereport.NewDefault(),
	)
}

func NewWithDependencies(
	_ *calculator.Calculator,
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
		scopeGuard:          guard,
		confidenceEvaluator: confidenceEvaluator,
	}
}

func (
	executor *Executor,
) EvaluateConfidence(
	request confidencereport.Request,
) (confidencereport.Report, error) {
	return executor.confidenceEvaluator.Evaluate(
		request,
	)
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
