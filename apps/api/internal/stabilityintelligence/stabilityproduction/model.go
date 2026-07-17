package stabilityproduction

import (
	"context"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/stabilityintelligence/confidencepropagation"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/stabilityintelligence/failureexplanation"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/stabilityintelligence/forecastanalysis"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/stabilityintelligence/forecaststability"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/stabilityintelligence/scopeenforcement"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/stabilityintelligence/unknownintervention"
)

const (
	Version                = "stability-intelligence-production-v1"
	ScopeGuardResearchOnly = "research_only_stability_and_explainability_not_operational_authorization"
	MinimumAsOfTimeCount   = 2
	MaximumAsOfTimeCount   = 8
)

type ProjectionRequest struct {
	TrajectoryID      string
	AsOfTime          time.Time
	RequestedDuration time.Duration
}

type ProjectionReader interface {
	ReadProjection(
		context.Context,
		ProjectionRequest,
	) (projectionproduction.Result, error)
}

type Request struct {
	TrajectoryID      string
	AsOfTimes         []time.Time
	RequestedDuration time.Duration
}

type Result struct {
	Version      string
	TrajectoryID string
	AsOfTimes    []time.Time

	Projections      []projectionproduction.Result
	ForecastVersions []forecaststability.ForecastVersionRecord
	Transitions      []forecaststability.StabilityResult
	ForecastAnalysis forecastanalysis.Result

	PropagatedConfidence confidencepropagation.Result
	FailureExplanation   failureexplanation.Result
	UnknownIntervention  unknownintervention.Result
	ScopeEnforcement     scopeenforcement.Result

	ScopeGuards      []string
	InputFingerprint string
	GeneratedAt      time.Time
}

func (result Result) Clone() Result {
	cloned := result
	cloned.AsOfTimes = append([]time.Time(nil), result.AsOfTimes...)

	cloned.Projections = make(
		[]projectionproduction.Result,
		0,
		len(result.Projections),
	)
	for _, item := range result.Projections {
		cloned.Projections = append(
			cloned.Projections,
			item.Clone(),
		)
	}

	cloned.ForecastVersions = make(
		[]forecaststability.ForecastVersionRecord,
		0,
		len(result.ForecastVersions),
	)
	for _, item := range result.ForecastVersions {
		cloned.ForecastVersions = append(
			cloned.ForecastVersions,
			item.Clone(),
		)
	}

	cloned.Transitions = make(
		[]forecaststability.StabilityResult,
		0,
		len(result.Transitions),
	)
	for _, item := range result.Transitions {
		cloned.Transitions = append(
			cloned.Transitions,
			item.Clone(),
		)
	}

	cloned.ForecastAnalysis =
		result.ForecastAnalysis.Clone()
	cloned.PropagatedConfidence =
		result.PropagatedConfidence.Clone()
	cloned.FailureExplanation =
		result.FailureExplanation.Clone()
	cloned.UnknownIntervention =
		result.UnknownIntervention.Clone()
	cloned.ScopeEnforcement =
		result.ScopeEnforcement.Clone()
	cloned.ScopeGuards = append(
		[]string(nil),
		result.ScopeGuards...,
	)

	return cloned
}
