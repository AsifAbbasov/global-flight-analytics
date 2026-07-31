package projectioncontinuation

import (
	"errors"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionhorizon"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionpatternconfidence"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

const (
	Version    = "local-historical-neighbor-continuation-v3"
	MethodName = "local_historical_neighbor_continuation"

	FingerprintVersion         = "local-historical-neighbor-continuation-fingerprint-v3"
	FallbackFingerprintVersion = "local-historical-neighbor-fallback-fingerprint-v3"
)

var (
	ErrTrajectoryIDRequired = errors.New(
		"projection trajectory identifier is required",
	)
	ErrGeneratedAtInvalid = errors.New(
		"projection generated-at time must not be before the as-of time",
	)
	ErrHorizonPlanInvalid = errors.New(
		"historical continuation horizon planner returned an invalid plan",
	)
	ErrContinuationContractInvalid = errors.New(
		"generated historical continuation contract is invalid",
	)
	ErrFallbackProjectionFailed = errors.New(
		"kinematic fallback projection failed",
	)
)

type Baseline struct {
	config Config
}

func New(
	config Config,
) (*Baseline, error) {
	config = config.normalized()
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf(
			"validate local historical continuation config: %w",
			err,
		)
	}

	return &Baseline{
		config: config,
	}, nil
}

type ApprovedEvidence struct {
	Selection projectionneighbors.Result
	Pattern   projectionpatternconfidence.Result
}

func (evidence ApprovedEvidence) Clone() ApprovedEvidence {
	return ApprovedEvidence{
		Selection: evidence.Selection.Clone(),
		Pattern:   evidence.Pattern.Clone(),
	}
}

type Request struct {
	CurrentTrajectory trajectory.FlightTrajectory
	Candidates        []trajectory.FlightTrajectory

	RouteScope        projectionneighbors.RouteScope
	AsOfTime          time.Time
	RequestedDuration time.Duration
	GeneratedAt       time.Time
}

func (
	baseline *Baseline,
) Project(
	request Request,
) (projectioncontract.Result, error) {
	if baseline == nil {
		return projectioncontract.Result{}, ErrHorizonPlannerRequired
	}
	plan, err := baseline.config.HorizonPlanner.Build(
		projectionhorizon.Request{
			AsOfTime:          request.AsOfTime,
			RequestedDuration: request.RequestedDuration,
		},
	)
	if err != nil {
		return projectioncontract.Result{}, fmt.Errorf(
			"build historical continuation horizon: %w",
			err,
		)
	}
	return baseline.projectWithPlan(request, plan, nil)
}

func (
	baseline *Baseline,
) ProjectApproved(
	request Request,
	evidence ApprovedEvidence,
) (projectioncontract.Result, error) {
	if baseline == nil {
		return projectioncontract.Result{}, ErrHorizonPlannerRequired
	}
	plan, err := baseline.config.HorizonPlanner.Build(
		projectionhorizon.Request{
			AsOfTime:          request.AsOfTime,
			RequestedDuration: request.RequestedDuration,
		},
	)
	if err != nil {
		return projectioncontract.Result{}, fmt.Errorf(
			"build historical continuation horizon: %w",
			err,
		)
	}
	cloned := evidence.Clone()
	return baseline.projectWithPlan(request, plan, &cloned)
}

// ProjectApprovedWithPlan consumes one immutable production-authorized plan
// and approved evidence without rebuilding either orchestration decision.
func (
	baseline *Baseline,
) ProjectApprovedWithPlan(
	request Request,
	plan projectionhorizon.Plan,
	evidence ApprovedEvidence,
) (projectioncontract.Result, error) {
	cloned := evidence.Clone()
	return baseline.projectWithPlan(request, plan, &cloned)
}

func (
	baseline *Baseline,
) projectWithPlan(
	request Request,
	plan projectionhorizon.Plan,
	approvedEvidence *ApprovedEvidence,
) (projectioncontract.Result, error) {
	if baseline == nil {
		return projectioncontract.Result{}, ErrHorizonPlannerRequired
	}
	if strings.TrimSpace(request.CurrentTrajectory.ID) == "" {
		return projectioncontract.Result{}, ErrTrajectoryIDRequired
	}
	if err := plan.Validate(); err != nil {
		return projectioncontract.Result{}, fmt.Errorf(
			"%w: %w",
			ErrHorizonPlanInvalid,
			err,
		)
	}
	if !request.AsOfTime.UTC().Equal(plan.AsOfTime) ||
		request.RequestedDuration != plan.RequestedDuration {
		return projectioncontract.Result{}, fmt.Errorf(
			"%w: request does not match authorized plan",
			ErrHorizonPlanInvalid,
		)
	}

	generatedAt := request.GeneratedAt.UTC()
	if generatedAt.IsZero() || generatedAt.Before(plan.AsOfTime) {
		return projectioncontract.Result{}, ErrGeneratedAtInvalid
	}

	preparation := baseline.prepareContinuation(
		request,
		plan,
		approvedEvidence,
	)
	if preparation.requiresFallback() {
		return baseline.fallback(
			request,
			plan,
			preparation.fallbackReason,
			preparation.selectionFingerprint,
			preparation.patternFingerprint,
		)
	}

	pointResult := baseline.projectForecastPoints(
		preparation,
		plan,
	)
	if pointResult.fallbackReason != "" {
		return baseline.fallback(
			request,
			plan,
			pointResult.fallbackReason,
			preparation.selectionFingerprint,
			preparation.patternFingerprint,
		)
	}

	return validateProjectionResult(
		baseline.buildContinuationResult(
			preparation,
			plan,
			pointResult,
			generatedAt,
		),
	)
}
