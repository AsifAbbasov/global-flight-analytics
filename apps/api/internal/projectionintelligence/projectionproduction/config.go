package projectionproduction

import (
	"errors"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionarrival"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionbaseline"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontinuation"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionfreshness"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionhorizon"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionpatternconfidence"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionroutefrequency"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

var (
	ErrHorizonPlannerRequired = errors.New(
		"production projection horizon planner is required",
	)
	ErrKinematicProjectorRequired = errors.New(
		"production kinematic projector is required",
	)
	ErrHistoricalProjectorRequired = errors.New(
		"production historical projector is required",
	)
	ErrNeighborSelectorRequired = errors.New(
		"production historical neighbor selector is required",
	)
	ErrPatternConfidenceEvaluatorRequired = errors.New(
		"production pattern confidence evaluator is required",
	)
	ErrFreshnessEvaluatorRequired = errors.New(
		"production freshness evaluator is required",
	)
	ErrRouteFrequencyEvaluatorRequired = errors.New(
		"production route-frequency evaluator is required",
	)
	ErrArrivalEstimatorRequired = errors.New(
		"production arrival estimator is required",
	)
	ErrLimitedEvidencePolicyInvalid = errors.New(
		"production limited-evidence policy is invalid",
	)
	ErrDependencyFailurePolicyInvalid = errors.New(
		"production dependency-failure policy is invalid",
	)
	ErrArrivalFailurePolicyInvalid = errors.New(
		"production arrival-failure policy is invalid",
	)
)

type HorizonPlanner interface {
	Build(
		projectionhorizon.Request,
	) (projectionhorizon.Plan, error)
}

type KinematicProjector interface {
	Project(
		projectionbaseline.Request,
	) (projectioncontract.Result, error)
}

type HistoricalProjector interface {
	Project(
		projectioncontinuation.Request,
	) (projectioncontract.Result, error)
}

type NeighborSelector interface {
	Select(
		projectionneighbors.Request,
	) (projectionneighbors.Result, error)
}

type PatternConfidenceEvaluator interface {
	Evaluate(
		projectionneighbors.Result,
	) (projectionpatternconfidence.Result, error)
}

type FreshnessEvaluator interface {
	Evaluate(
		projectionneighbors.Result,
		projectionpatternconfidence.Result,
	) (projectionfreshness.Result, error)
}

type RouteFrequencyEvaluator interface {
	Evaluate(
		routecontract.Result,
		projectionroutefrequency.HistorySummary,
	) (projectionroutefrequency.Result, error)
}

type ArrivalEstimator interface {
	Estimate(
		projectionarrival.Request,
	) (projectioncontract.Result, error)
}

type LimitedEvidencePolicy string

const (
	LimitedEvidenceReject LimitedEvidencePolicy = "reject_limited_evidence"
	LimitedEvidenceAllow  LimitedEvidencePolicy = "allow_limited_evidence"
)

func (policy LimitedEvidencePolicy) IsKnown() bool {
	switch policy {
	case LimitedEvidenceReject,
		LimitedEvidenceAllow:
		return true
	default:
		return false
	}
}

type DependencyFailurePolicy string

const (
	DependencyFailureReturnError         DependencyFailurePolicy = "return_error"
	DependencyFailureFallbackToKinematic DependencyFailurePolicy = "fallback_to_kinematic"
)

func (policy DependencyFailurePolicy) IsKnown() bool {
	switch policy {
	case DependencyFailureReturnError,
		DependencyFailureFallbackToKinematic:
		return true
	default:
		return false
	}
}

type ArrivalFailurePolicy string

const (
	ArrivalFailureReturnError        ArrivalFailurePolicy = "return_error"
	ArrivalFailurePreserveProjection ArrivalFailurePolicy = "preserve_position_projection"
)

func (policy ArrivalFailurePolicy) IsKnown() bool {
	switch policy {
	case ArrivalFailureReturnError,
		ArrivalFailurePreserveProjection:
		return true
	default:
		return false
	}
}

type Config struct {
	HorizonPlanner HorizonPlanner

	KinematicProjector  KinematicProjector
	HistoricalProjector HistoricalProjector

	NeighborSelector           NeighborSelector
	PatternConfidenceEvaluator PatternConfidenceEvaluator
	FreshnessEvaluator         FreshnessEvaluator
	RouteFrequencyEvaluator    RouteFrequencyEvaluator
	ArrivalEstimator           ArrivalEstimator

	FreshnessLimitedPolicy      LimitedEvidencePolicy
	RouteFrequencyLimitedPolicy LimitedEvidencePolicy
	DependencyFailurePolicy     DependencyFailurePolicy
	ArrivalFailurePolicy        ArrivalFailurePolicy
}

func (config Config) Validate() error {
	if config.HorizonPlanner == nil {
		return ErrHorizonPlannerRequired
	}
	if config.KinematicProjector == nil {
		return ErrKinematicProjectorRequired
	}
	if config.HistoricalProjector == nil {
		return ErrHistoricalProjectorRequired
	}
	if config.NeighborSelector == nil {
		return ErrNeighborSelectorRequired
	}
	if config.PatternConfidenceEvaluator == nil {
		return ErrPatternConfidenceEvaluatorRequired
	}
	if config.FreshnessEvaluator == nil {
		return ErrFreshnessEvaluatorRequired
	}
	if config.RouteFrequencyEvaluator == nil {
		return ErrRouteFrequencyEvaluatorRequired
	}
	if config.ArrivalEstimator == nil {
		return ErrArrivalEstimatorRequired
	}
	if !config.FreshnessLimitedPolicy.IsKnown() {
		return fmt.Errorf(
			"%w: %q",
			ErrLimitedEvidencePolicyInvalid,
			config.FreshnessLimitedPolicy,
		)
	}
	if !config.RouteFrequencyLimitedPolicy.IsKnown() {
		return fmt.Errorf(
			"%w: %q",
			ErrLimitedEvidencePolicyInvalid,
			config.RouteFrequencyLimitedPolicy,
		)
	}
	if !config.DependencyFailurePolicy.IsKnown() {
		return fmt.Errorf(
			"%w: %q",
			ErrDependencyFailurePolicyInvalid,
			config.DependencyFailurePolicy,
		)
	}
	if !config.ArrivalFailurePolicy.IsKnown() {
		return fmt.Errorf(
			"%w: %q",
			ErrArrivalFailurePolicyInvalid,
			config.ArrivalFailurePolicy,
		)
	}

	return nil
}
