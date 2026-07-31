package projectionproduction

import (
	"errors"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
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
	ErrHorizonPlannerRequired             = errors.New("production projection horizon planner is required")
	ErrKinematicProjectorRequired         = errors.New("production kinematic projector is required")
	ErrHistoricalProjectorRequired        = errors.New("production historical projector is required")
	ErrNeighborSelectorRequired           = errors.New("production historical neighbor selector is required")
	ErrPatternConfidenceEvaluatorRequired = errors.New("production pattern confidence evaluator is required")
	ErrFreshnessEvaluatorRequired         = errors.New("production freshness evaluator is required")
	ErrRouteFrequencyEvaluatorRequired    = errors.New("production route-frequency evaluator is required")
	ErrArrivalEstimatorRequired           = errors.New("production arrival estimator is required")
	ErrLimitedEvidencePolicyInvalid       = errors.New("production limited-evidence policy is invalid")
	ErrDependencyFailurePolicyInvalid     = errors.New("production dependency-failure policy is invalid")
	ErrArrivalFailurePolicyInvalid        = errors.New("production arrival-failure policy is invalid")
)

type HorizonPlanner interface {
	Build(projectionhorizon.Request) (projectionhorizon.Plan, error)
}

type KinematicProjector interface {
	ProjectWithPlan(
		projectionbaseline.Request,
		projectionhorizon.Plan,
	) (projectioncontract.Result, error)
}

type HistoricalProjector interface {
	ProjectApprovedWithPlan(
		projectioncontinuation.Request,
		projectionhorizon.Plan,
		projectioncontinuation.ApprovedEvidence,
	) (projectioncontract.Result, error)
}

type NeighborSelector interface {
	Select(projectionneighbors.Request) (projectionneighbors.Result, error)
}

type PatternConfidenceEvaluator interface {
	EvaluateWithContinuations(
		projectionneighbors.Result,
		[]trajectory.FlightTrajectory,
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

// ArrivalEstimator is intentionally narrow: production composition accepts
// only an Estimated Arrival delta and never a replacement position projection.
type ArrivalEstimator interface {
	EstimateArrival(projectionarrival.Request) (ArrivalOutcome, error)
}

type LimitedEvidencePolicy string

const (
	LimitedEvidenceReject LimitedEvidencePolicy = "reject_limited_evidence"
	LimitedEvidenceAllow  LimitedEvidencePolicy = "allow_limited_evidence"
)

func (policy LimitedEvidencePolicy) IsKnown() bool {
	return policy == LimitedEvidenceReject || policy == LimitedEvidenceAllow
}

type DependencyFailurePolicy string

const (
	DependencyFailureReturnError         DependencyFailurePolicy = "return_error"
	DependencyFailureFallbackToKinematic DependencyFailurePolicy = "fallback_to_kinematic"
)

func (policy DependencyFailurePolicy) IsKnown() bool {
	return policy == DependencyFailureReturnError || policy == DependencyFailureFallbackToKinematic
}

type ArrivalFailurePolicy string

const (
	ArrivalFailureReturnError        ArrivalFailurePolicy = "return_error"
	ArrivalFailurePreserveProjection ArrivalFailurePolicy = "preserve_position_projection"
)

func (policy ArrivalFailurePolicy) IsKnown() bool {
	return policy == ArrivalFailureReturnError || policy == ArrivalFailurePreserveProjection
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
	switch {
	case config.HorizonPlanner == nil:
		return ErrHorizonPlannerRequired
	case config.KinematicProjector == nil:
		return ErrKinematicProjectorRequired
	case config.HistoricalProjector == nil:
		return ErrHistoricalProjectorRequired
	case config.NeighborSelector == nil:
		return ErrNeighborSelectorRequired
	case config.PatternConfidenceEvaluator == nil:
		return ErrPatternConfidenceEvaluatorRequired
	case config.FreshnessEvaluator == nil:
		return ErrFreshnessEvaluatorRequired
	case config.RouteFrequencyEvaluator == nil:
		return ErrRouteFrequencyEvaluatorRequired
	case config.ArrivalEstimator == nil:
		return ErrArrivalEstimatorRequired
	case !config.FreshnessLimitedPolicy.IsKnown():
		return fmt.Errorf("%w: %q", ErrLimitedEvidencePolicyInvalid, config.FreshnessLimitedPolicy)
	case !config.RouteFrequencyLimitedPolicy.IsKnown():
		return fmt.Errorf("%w: %q", ErrLimitedEvidencePolicyInvalid, config.RouteFrequencyLimitedPolicy)
	case !config.DependencyFailurePolicy.IsKnown():
		return fmt.Errorf("%w: %q", ErrDependencyFailurePolicyInvalid, config.DependencyFailurePolicy)
	case !config.ArrivalFailurePolicy.IsKnown():
		return fmt.Errorf("%w: %q", ErrArrivalFailurePolicyInvalid, config.ArrivalFailurePolicy)
	default:
		return nil
	}
}
