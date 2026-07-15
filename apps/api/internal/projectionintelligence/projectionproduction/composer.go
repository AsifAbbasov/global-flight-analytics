package projectionproduction

import (
	"errors"
	"fmt"
	"strings"
	"time"

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
	ErrComposerUnavailable = errors.New(
		"production projection composer is unavailable",
	)
	ErrTrajectoryIDRequired = errors.New(
		"production projection trajectory identifier is required",
	)
	ErrGeneratedAtInvalid = errors.New(
		"production projection generated-at time must not precede the as-of time",
	)
	ErrRouteContractInvalid = errors.New(
		"production route intelligence contract is invalid",
	)
	ErrNeighborSelectionFailed = errors.New(
		"production historical neighbor selection failed",
	)
	ErrPatternConfidenceFailed = errors.New(
		"production pattern confidence evaluation failed",
	)
	ErrFreshnessEvaluationFailed = errors.New(
		"production pattern freshness evaluation failed",
	)
	ErrRouteFrequencyEvaluationFailed = errors.New(
		"production route-frequency evaluation failed",
	)
	ErrHistoricalProjectionFailed = errors.New(
		"production historical projection failed",
	)
	ErrKinematicProjectionFailed = errors.New(
		"production kinematic projection failed",
	)
	ErrArrivalEstimationFailed = errors.New(
		"production estimated arrival failed",
	)
	ErrProjectionContractInvalid = errors.New(
		"production projection contract is invalid",
	)
	ErrCompositionResultInvalid = errors.New(
		"production composition result is invalid",
	)
)

type Composer struct {
	config Config
}

func New(
	config Config,
) (*Composer, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf(
			"validate production projection composition config: %w",
			err,
		)
	}

	return &Composer{
		config: config,
	}, nil
}

type Request struct {
	CurrentTrajectory    trajectory.FlightTrajectory
	HistoricalCandidates []trajectory.FlightTrajectory

	Route        routecontract.Result
	RouteHistory *projectionroutefrequency.HistorySummary

	AsOfTime          time.Time
	RequestedDuration time.Duration
	GeneratedAt       time.Time
}

type compositionState struct {
	selection *projectionneighbors.Result
	pattern   *projectionpatternconfidence.Result
	freshness *projectionfreshness.Result
	frequency *projectionroutefrequency.Result

	strategy       Strategy
	fallbackReason string
	arrivalStatus  ArrivalStatus
	notices        []Notice
}

func (
	composer *Composer,
) Compose(
	request Request,
) (Result, error) {
	if composer == nil {
		return Result{},
			ErrComposerUnavailable
	}
	if strings.TrimSpace(
		request.CurrentTrajectory.ID,
	) == "" {
		return Result{},
			ErrTrajectoryIDRequired
	}

	plan, err := composer.config.
		HorizonPlanner.Build(
		projectionhorizon.Request{
			AsOfTime: request.AsOfTime,
			RequestedDuration: request.
				RequestedDuration,
		},
	)
	if err != nil {
		return Result{}, fmt.Errorf(
			"build production projection horizon: %w",
			err,
		)
	}

	generatedAt := request.GeneratedAt.UTC()
	if generatedAt.IsZero() ||
		generatedAt.Before(plan.AsOfTime) {
		return Result{},
			ErrGeneratedAtInvalid
	}

	state := compositionState{
		strategy:      StrategyKinematic,
		arrivalStatus: ArrivalStatusSkipped,
	}

	routeValid := false
	routeReport := routecontract.Validate(
		request.Route,
	)
	if routeReport.Status ==
		routecontract.ValidationStatusValid {
		routeValid = true
	} else {
		dependencyErr := fmt.Errorf(
			"%w: %#v",
			ErrRouteContractInvalid,
			routeReport.Issues,
		)
		if composer.config.DependencyFailurePolicy ==
			DependencyFailureReturnError {
			return Result{}, dependencyErr
		}
		state.fallbackReason =
			"route_contract_invalid"
		state.notices = append(
			state.notices,
			Notice{
				Code:    "route_contract_invalid",
				Message: "Route Intelligence contract was invalid, so historical projection and Estimated Arrival were skipped.",
			},
		)
	}

	if routeValid {
		if err := composer.evaluateHistoricalEvidence(
			request,
			plan,
			&state,
		); err != nil {
			if composer.config.DependencyFailurePolicy ==
				DependencyFailureReturnError {
				return Result{}, err
			}
			if state.fallbackReason == "" {
				state.fallbackReason =
					"historical_evidence_evaluation_failed"
			}
			state.notices = append(
				state.notices,
				Notice{
					Code:    "historical_evidence_evaluation_failed",
					Message: "Historical evidence evaluation failed, so the conservative kinematic baseline was selected.",
				},
			)
		}
	}

	projection, err := composer.project(
		request,
		&state,
	)
	if err != nil {
		return Result{}, err
	}

	projectionReport := projectioncontract.Validate(
		projection,
	)
	if projectionReport.Status !=
		projectioncontract.ValidationStatusValid {
		return Result{}, fmt.Errorf(
			"%w: %#v",
			ErrProjectionContractInvalid,
			projectionReport.Issues,
		)
	}

	if routeValid {
		projection, state.arrivalStatus, err =
			composer.attachArrival(
				request,
				projection,
				&state,
			)
		if err != nil {
			return Result{}, err
		}
	}

	result := Result{
		Version: Version,

		Strategy:       state.strategy,
		FallbackReason: state.fallbackReason,
		ArrivalStatus:  state.arrivalStatus,

		Projection: projection.Clone(),

		NeighborSelection: cloneSelection(
			state.selection,
		),
		PatternConfidence: clonePattern(
			state.pattern,
		),
		Freshness: cloneFreshness(
			state.freshness,
		),
		RouteFrequency: cloneFrequency(
			state.frequency,
		),

		Notices: normalizeNotices(
			state.notices,
		),
		GeneratedAt: projection.GeneratedAt.UTC(),
	}
	result.InputFingerprint = productionFingerprint(
		result,
		composer.config,
	)

	if err := result.Validate(); err != nil {
		return Result{}, fmt.Errorf(
			"%w: %v",
			ErrCompositionResultInvalid,
			err,
		)
	}

	return result.Clone(), nil
}

func (
	composer *Composer,
) evaluateHistoricalEvidence(
	request Request,
	plan projectionhorizon.Plan,
	state *compositionState,
) error {
	selection, err := composer.config.
		NeighborSelector.Select(
		projectionneighbors.Request{
			CurrentTrajectory: request.
				CurrentTrajectory,
			Candidates: request.
				HistoricalCandidates,
			AsOfTime:                     plan.AsOfTime,
			RequiredContinuationDuration: plan.EffectiveDuration,
		},
	)
	if err != nil {
		state.fallbackReason =
			"historical_neighbor_selection_failed"
		return fmt.Errorf(
			"%w: %v",
			ErrNeighborSelectionFailed,
			err,
		)
	}
	if err := selection.Validate(); err != nil {
		state.fallbackReason =
			"historical_neighbor_selection_invalid"
		return fmt.Errorf(
			"%w: %v",
			ErrNeighborSelectionFailed,
			err,
		)
	}
	state.selection = pointerToSelection(
		selection,
	)
	if selection.Status ==
		projectionneighbors.StatusUnavailable {
		state.fallbackReason =
			"historical_neighbors_unavailable"
		state.notices = append(
			state.notices,
			Notice{
				Code:    "historical_neighbors_unavailable",
				Message: "No usable historical neighbors were selected, so the kinematic baseline was selected.",
			},
		)
		return nil
	}

	pattern, err := composer.config.
		PatternConfidenceEvaluator.Evaluate(
		selection,
	)
	if err != nil {
		state.fallbackReason =
			"historical_pattern_confidence_failed"
		return fmt.Errorf(
			"%w: %v",
			ErrPatternConfidenceFailed,
			err,
		)
	}
	if err := pattern.Validate(); err != nil {
		state.fallbackReason =
			"historical_pattern_confidence_invalid"
		return fmt.Errorf(
			"%w: %v",
			ErrPatternConfidenceFailed,
			err,
		)
	}
	state.pattern = pointerToPattern(pattern)
	if !pattern.Usable {
		state.fallbackReason =
			"historical_pattern_not_usable"
		state.notices = append(
			state.notices,
			Notice{
				Code:    "historical_pattern_not_usable",
				Message: "Pattern Confidence rejected the selected historical neighbors, so the kinematic baseline was selected.",
			},
		)
		return nil
	}

	freshness, err := composer.config.
		FreshnessEvaluator.Evaluate(
		selection,
		pattern,
	)
	if err != nil {
		state.fallbackReason =
			"pattern_freshness_evaluation_failed"
		return fmt.Errorf(
			"%w: %v",
			ErrFreshnessEvaluationFailed,
			err,
		)
	}
	if err := freshness.Validate(); err != nil {
		state.fallbackReason =
			"pattern_freshness_result_invalid"
		return fmt.Errorf(
			"%w: %v",
			ErrFreshnessEvaluationFailed,
			err,
		)
	}
	state.freshness = pointerToFreshness(
		freshness,
	)
	if !freshness.Usable ||
		(freshness.Decision ==
			projectionfreshness.DecisionLimited &&
			composer.config.FreshnessLimitedPolicy ==
				LimitedEvidenceReject) {
		state.fallbackReason =
			"pattern_freshness_guard_blocked"
		state.notices = append(
			state.notices,
			Notice{
				Code:    "pattern_freshness_guard_blocked",
				Message: "Pattern Freshness policy did not permit historical continuation, so the kinematic baseline was selected.",
			},
		)
		return nil
	}

	if request.Route.Status !=
		routecontract.RouteStatusComplete {
		state.fallbackReason =
			"complete_route_unavailable"
		state.notices = append(
			state.notices,
			Notice{
				Code:    "complete_route_unavailable",
				Message: "Route Intelligence did not resolve a complete origin-destination route, so route-frequency support could not authorize historical continuation.",
			},
		)
		return nil
	}
	if request.RouteHistory == nil {
		state.fallbackReason =
			"route_history_unavailable"
		state.notices = append(
			state.notices,
			Notice{
				Code:    "route_history_unavailable",
				Message: "No route-history summary was supplied, so historical continuation was blocked.",
			},
		)
		return nil
	}

	frequency, err := composer.config.
		RouteFrequencyEvaluator.Evaluate(
		request.Route,
		request.RouteHistory.Clone(),
	)
	if err != nil {
		state.fallbackReason =
			"route_frequency_evaluation_failed"
		return fmt.Errorf(
			"%w: %v",
			ErrRouteFrequencyEvaluationFailed,
			err,
		)
	}
	if err := frequency.Validate(); err != nil {
		state.fallbackReason =
			"route_frequency_result_invalid"
		return fmt.Errorf(
			"%w: %v",
			ErrRouteFrequencyEvaluationFailed,
			err,
		)
	}
	state.frequency = pointerToFrequency(
		frequency,
	)
	if !frequency.Usable ||
		(frequency.Decision ==
			projectionroutefrequency.DecisionLimited &&
			composer.config.RouteFrequencyLimitedPolicy ==
				LimitedEvidenceReject) {
		state.fallbackReason =
			"route_frequency_guard_blocked"
		state.notices = append(
			state.notices,
			Notice{
				Code:    "route_frequency_guard_blocked",
				Message: "Low-Frequency Route policy did not permit historical continuation, so the kinematic baseline was selected.",
			},
		)
		return nil
	}

	state.strategy = StrategyHistoricalNeighbor
	state.fallbackReason = ""
	state.notices = append(
		state.notices,
		Notice{
			Code:    "historical_neighbor_continuation_authorized",
			Message: "Pattern confidence, freshness, and route-frequency evidence authorized historical-neighbor continuation.",
		},
	)

	return nil
}

func (
	composer *Composer,
) project(
	request Request,
	state *compositionState,
) (projectioncontract.Result, error) {
	if state.strategy ==
		StrategyHistoricalNeighbor {
		result, err := composer.config.
			HistoricalProjector.Project(
			projectioncontinuation.Request{
				CurrentTrajectory: request.
					CurrentTrajectory,
				Candidates: request.
					HistoricalCandidates,
				AsOfTime: request.AsOfTime,
				RequestedDuration: request.
					RequestedDuration,
				GeneratedAt: request.GeneratedAt,
			},
		)
		if err != nil {
			if composer.config.DependencyFailurePolicy ==
				DependencyFailureReturnError {
				return projectioncontract.Result{},
					fmt.Errorf(
						"%w: %v",
						ErrHistoricalProjectionFailed,
						err,
					)
			}
			state.strategy = StrategyKinematic
			state.fallbackReason =
				"historical_projection_failed"
			state.notices = append(
				state.notices,
				Notice{
					Code:    "historical_projection_failed",
					Message: "Historical continuation failed after authorization, so the kinematic baseline was selected.",
				},
			)
			return composer.kinematic(request)
		}

		report := projectioncontract.Validate(
			result,
		)
		if report.Status !=
			projectioncontract.ValidationStatusValid {
			if composer.config.DependencyFailurePolicy ==
				DependencyFailureReturnError {
				return projectioncontract.Result{},
					fmt.Errorf(
						"%w: %#v",
						ErrHistoricalProjectionFailed,
						report.Issues,
					)
			}
			state.strategy = StrategyKinematic
			state.fallbackReason =
				"historical_projection_invalid"
			state.notices = append(
				state.notices,
				Notice{
					Code:    "historical_projection_invalid",
					Message: "Historical continuation returned an invalid contract, so the kinematic baseline was selected.",
				},
			)
			return composer.kinematic(request)
		}

		if result.Method.Name ==
			projectionbaseline.MethodName {
			state.strategy = StrategyKinematic
			state.fallbackReason =
				"historical_projector_internal_fallback"
			state.notices = append(
				state.notices,
				Notice{
					Code:    "historical_projector_internal_fallback",
					Message: "The authorized historical projector selected its internal kinematic fallback.",
				},
			)
			return result.Clone(), nil
		}
		if result.Method.Name !=
			projectioncontinuation.MethodName {
			if composer.config.DependencyFailurePolicy ==
				DependencyFailureReturnError {
				return projectioncontract.Result{},
					fmt.Errorf(
						"%w: unexpected method %q",
						ErrHistoricalProjectionFailed,
						result.Method.Name,
					)
			}
			state.strategy = StrategyKinematic
			state.fallbackReason =
				"historical_projection_method_unexpected"
			state.notices = append(
				state.notices,
				Notice{
					Code:    "historical_projection_method_unexpected",
					Message: "Historical continuation returned an unexpected method, so the kinematic baseline was selected.",
				},
			)
			return composer.kinematic(request)
		}

		return result.Clone(), nil
	}

	return composer.kinematic(request)
}

func (
	composer *Composer,
) kinematic(
	request Request,
) (projectioncontract.Result, error) {
	result, err := composer.config.
		KinematicProjector.Project(
		projectionbaseline.Request{
			Trajectory: request.CurrentTrajectory,
			AsOfTime:   request.AsOfTime,
			RequestedDuration: request.
				RequestedDuration,
			GeneratedAt: request.GeneratedAt,
		},
	)
	if err != nil {
		return projectioncontract.Result{},
			fmt.Errorf(
				"%w: %v",
				ErrKinematicProjectionFailed,
				err,
			)
	}

	return result.Clone(), nil
}

func (
	composer *Composer,
) attachArrival(
	request Request,
	projection projectioncontract.Result,
	state *compositionState,
) (
	projectioncontract.Result,
	ArrivalStatus,
	error,
) {
	result, err := composer.config.
		ArrivalEstimator.Estimate(
		projectionarrival.Request{
			Projection: projection,
			Route:      request.Route,
			CurrentTrajectory: request.
				CurrentTrajectory,
			GeneratedAt: request.GeneratedAt,
		},
	)
	if err != nil {
		if composer.config.ArrivalFailurePolicy ==
			ArrivalFailureReturnError {
			return projectioncontract.Result{},
				ArrivalStatusFailed,
				fmt.Errorf(
					"%w: %v",
					ErrArrivalEstimationFailed,
					err,
				)
		}

		state.notices = append(
			state.notices,
			Notice{
				Code:    "estimated_arrival_failed_projection_preserved",
				Message: "Estimated Arrival failed, but the configured policy preserved the valid position projection.",
			},
		)
		return projection.Clone(),
			ArrivalStatusFailed,
			nil
	}

	report := projectioncontract.Validate(result)
	if report.Status !=
		projectioncontract.ValidationStatusValid {
		if composer.config.ArrivalFailurePolicy ==
			ArrivalFailureReturnError {
			return projectioncontract.Result{},
				ArrivalStatusFailed,
				fmt.Errorf(
					"%w: %#v",
					ErrArrivalEstimationFailed,
					report.Issues,
				)
		}
		state.notices = append(
			state.notices,
			Notice{
				Code:    "estimated_arrival_invalid_projection_preserved",
				Message: "Estimated Arrival returned an invalid result, but the configured policy preserved the valid position projection.",
			},
		)
		return projection.Clone(),
			ArrivalStatusFailed,
			nil
	}

	if result.Arrival == nil {
		return result.Clone(),
			ArrivalStatusWithheld,
			nil
	}

	return result.Clone(),
		ArrivalStatusAttached,
		nil
}

func pointerToSelection(
	value projectionneighbors.Result,
) *projectionneighbors.Result {
	cloned := value.Clone()
	return &cloned
}

func pointerToPattern(
	value projectionpatternconfidence.Result,
) *projectionpatternconfidence.Result {
	cloned := value.Clone()
	return &cloned
}

func pointerToFreshness(
	value projectionfreshness.Result,
) *projectionfreshness.Result {
	cloned := value.Clone()
	return &cloned
}

func pointerToFrequency(
	value projectionroutefrequency.Result,
) *projectionroutefrequency.Result {
	cloned := value.Clone()
	return &cloned
}

func cloneSelection(
	value *projectionneighbors.Result,
) *projectionneighbors.Result {
	if value == nil {
		return nil
	}
	return pointerToSelection(*value)
}

func clonePattern(
	value *projectionpatternconfidence.Result,
) *projectionpatternconfidence.Result {
	if value == nil {
		return nil
	}
	return pointerToPattern(*value)
}

func cloneFreshness(
	value *projectionfreshness.Result,
) *projectionfreshness.Result {
	if value == nil {
		return nil
	}
	return pointerToFreshness(*value)
}

func cloneFrequency(
	value *projectionroutefrequency.Result,
) *projectionroutefrequency.Result {
	if value == nil {
		return nil
	}
	return pointerToFrequency(*value)
}
