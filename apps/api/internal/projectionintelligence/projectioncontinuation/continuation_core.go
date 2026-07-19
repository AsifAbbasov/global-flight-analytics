package projectioncontinuation

import (
	"errors"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionhorizon"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

const (
	Version    = "local-historical-neighbor-continuation-v1"
	MethodName = "local_historical_neighbor_continuation"

	FingerprintVersion         = "local-historical-neighbor-continuation-fingerprint-v1"
	FallbackFingerprintVersion = "local-historical-neighbor-fallback-fingerprint-v1"
)

var (
	ErrTrajectoryIDRequired = errors.New(
		"projection trajectory identifier is required",
	)
	ErrGeneratedAtInvalid = errors.New(
		"projection generated-at time must not be before the as-of time",
	)
	ErrCurrentTrajectoryUnavailable = errors.New(
		"current trajectory does not contain a usable as-of endpoint",
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

type Request struct {
	CurrentTrajectory trajectory.FlightTrajectory
	Candidates        []trajectory.FlightTrajectory

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
		return projectioncontract.Result{},
			ErrHorizonPlannerRequired
	}
	if strings.TrimSpace(
		request.CurrentTrajectory.ID,
	) == "" {
		return projectioncontract.Result{},
			ErrTrajectoryIDRequired
	}

	plan, err := baseline.config.
		HorizonPlanner.Build(
		projectionhorizon.Request{
			AsOfTime: request.AsOfTime,
			RequestedDuration: request.
				RequestedDuration,
		},
	)
	if err != nil {
		return projectioncontract.Result{},
			fmt.Errorf(
				"build historical continuation horizon: %w",
				err,
			)
	}

	generatedAt := request.GeneratedAt.UTC()
	if generatedAt.IsZero() ||
		generatedAt.Before(plan.AsOfTime) {
		return projectioncontract.Result{},
			ErrGeneratedAtInvalid
	}

	preparation :=
		baseline.prepareContinuation(
			request,
			plan,
		)
	if preparation.requiresFallback() {
		return baseline.fallback(
			request,
			preparation.fallbackReason,
			preparation.selectionFingerprint,
			preparation.patternFingerprint,
		)
	}

	pointResult :=
		baseline.projectForecastPoints(
			preparation,
			plan,
		)
	if pointResult.fallbackReason != "" {
		return baseline.fallback(
			request,
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
