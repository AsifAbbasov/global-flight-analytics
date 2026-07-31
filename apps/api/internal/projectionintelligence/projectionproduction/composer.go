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
	ErrComposerUnavailable            = errors.New("production projection composer is unavailable")
	ErrHorizonPlanInvalid             = errors.New("production projection horizon plan is invalid")
	ErrTrajectoryIDRequired           = errors.New("production projection trajectory identifier is required")
	ErrGeneratedAtInvalid             = errors.New("production projection generated-at time must not precede the as-of time")
	ErrRouteContractInvalid           = errors.New("production route intelligence contract is invalid")
	ErrNeighborSelectionFailed        = errors.New("production historical neighbor selection failed")
	ErrPatternConfidenceFailed        = errors.New("production pattern confidence evaluation failed")
	ErrFreshnessEvaluationFailed      = errors.New("production pattern freshness evaluation failed")
	ErrRouteFrequencyEvaluationFailed = errors.New("production route-frequency evaluation failed")
	ErrHistoricalProjectionFailed     = errors.New("production historical projection failed")
	ErrKinematicProjectionFailed      = errors.New("production kinematic projection failed")
	ErrArrivalEstimationFailed        = errors.New("production estimated arrival failed")
	ErrProjectionContractInvalid      = errors.New("production projection contract is invalid")
	ErrCompositionResultInvalid       = errors.New("production composition result is invalid")
	ErrEvidenceBindingInvalid         = errors.New("production historical evidence binding is invalid")
)

type Composer struct {
	config Config
}

func New(config Config) (*Composer, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("validate production projection composition config: %w", err)
	}
	return &Composer{config: config}, nil
}

type Request struct {
	CurrentTrajectory    trajectory.FlightTrajectory
	HistoricalCandidates []trajectory.FlightTrajectory

	HistoricalCandidateRouteScope projectionneighbors.RouteScope
	Route                         routecontract.Result
	RouteHistory                  *projectionroutefrequency.HistorySummary

	AsOfTime          time.Time
	RequestedDuration time.Duration
	GeneratedAt       time.Time
}

type compositionState struct {
	authorized *AuthorizedHistoricalEvidence
	routeScope *projectionneighbors.RouteScope

	strategy       Strategy
	fallbackReason string
	arrivalStatus  ArrivalStatus
	notices        []Notice
}

func (composer *Composer) Compose(request Request) (Result, error) {
	if composer == nil {
		return Result{}, ErrComposerUnavailable
	}

	snapshot := request.Clone()
	if strings.TrimSpace(snapshot.CurrentTrajectory.ID) == "" {
		return Result{}, ErrTrajectoryIDRequired
	}

	plan, err := composer.config.HorizonPlanner.Build(projectionhorizon.Request{
		AsOfTime:          snapshot.AsOfTime,
		RequestedDuration: snapshot.RequestedDuration,
	})
	if err != nil {
		return Result{}, fmt.Errorf("build production projection horizon: %w", err)
	}
	if err := plan.Validate(); err != nil {
		return Result{}, fmt.Errorf("%w: %w", ErrHorizonPlanInvalid, err)
	}
	if !snapshot.AsOfTime.UTC().Equal(plan.AsOfTime) ||
		snapshot.RequestedDuration != plan.RequestedDuration {
		return Result{}, fmt.Errorf("%w: request does not match built plan", ErrHorizonPlanInvalid)
	}

	snapshot.AsOfTime = plan.AsOfTime
	snapshot.GeneratedAt = snapshot.GeneratedAt.UTC()
	if snapshot.GeneratedAt.IsZero() || snapshot.GeneratedAt.Before(plan.AsOfTime) {
		return Result{}, ErrGeneratedAtInvalid
	}

	state := compositionState{
		strategy:      StrategyKinematic,
		arrivalStatus: ArrivalStatusSkipped,
	}
	routeValid, err := composer.authorizeRoute(snapshot, plan, &state)
	if err != nil {
		return Result{}, err
	}
	if routeValid {
		if err := composer.evaluateHistoricalEvidence(snapshot, plan, &state); err != nil {
			if composer.config.DependencyFailurePolicy == DependencyFailureReturnError {
				return Result{}, err
			}
			if state.fallbackReason == "" {
				state.fallbackReason = "historical_evidence_evaluation_failed"
			}
			state.notices = append(state.notices, Notice{
				Code:    "historical_evidence_evaluation_failed",
				Message: "Historical evidence evaluation failed, so the conservative kinematic baseline was selected.",
			})
		}
	}

	projection, err := composer.project(snapshot, plan, &state)
	if err != nil {
		return Result{}, err
	}
	if routeValid && projection.Status != projectioncontract.ResultStatusUnavailable {
		projection, state.arrivalStatus, err = composer.attachArrival(snapshot, projection, &state)
		if err != nil {
			return Result{}, err
		}
	} else if routeValid {
		state.notices = append(state.notices, Notice{
			Code:    "estimated_arrival_skipped_projection_unavailable",
			Message: "Estimated Arrival was skipped because the position projection is unavailable.",
		})
	}

	result := Result{
		Version:        Version,
		Strategy:       state.strategy,
		FallbackReason: state.fallbackReason,
		ArrivalStatus:  state.arrivalStatus,
		HorizonPlan:    plan.Clone(),
		Projection:     projection.Clone(),
		Notices:        normalizeNotices(state.notices),
		InputFingerprint: requestInputFingerprint(
			snapshot,
			plan.Fingerprint,
			composer.config,
		),
		GeneratedAt: projection.GeneratedAt.UTC(),
	}
	if state.authorized != nil {
		evidence := state.authorized.Clone()
		result.NeighborSelection = pointerToSelection(evidence.Selection)
		result.PatternConfidence = pointerToPattern(evidence.Pattern)
		result.Freshness = pointerToFreshness(evidence.Freshness)
		result.RouteFrequency = pointerToFrequency(evidence.Frequency)
	}

	result, err = Finalize(result)
	if err != nil {
		return Result{}, fmt.Errorf("%w: %w", ErrCompositionResultInvalid, err)
	}
	return result.Clone(), nil
}

func (composer *Composer) authorizeRoute(
	request Request,
	plan projectionhorizon.Plan,
	state *compositionState,
) (bool, error) {
	err := validateRouteBinding(request, plan)
	if err == nil {
		return true, nil
	}
	if composer.config.DependencyFailurePolicy == DependencyFailureReturnError {
		return false, err
	}
	state.fallbackReason = "route_contract_invalid"
	state.notices = append(state.notices, Notice{
		Code:    "route_contract_invalid",
		Message: "Route Intelligence contract did not match the current production request, so historical projection and Estimated Arrival were skipped.",
	})
	return false, nil
}

func (composer *Composer) evaluateHistoricalEvidence(
	request Request,
	plan projectionhorizon.Plan,
	state *compositionState,
) error {
	routeScope := request.HistoricalCandidateRouteScope.Clone()
	if err := validateHistoricalCandidateRouteScope(request.Route, routeScope, request.HistoricalCandidates); err != nil {
		state.fallbackReason = "historical_candidate_route_scope_invalid"
		return wrapDependency(ErrNeighborSelectionFailed, err)
	}
	state.routeScope = &routeScope

	selection, err := composer.config.NeighborSelector.Select(projectionneighbors.Request{
		CurrentTrajectory:            request.CurrentTrajectory.Clone(),
		Candidates:                   cloneTrajectories(request.HistoricalCandidates),
		RouteScope:                   routeScope.Clone(),
		AsOfTime:                     plan.AsOfTime,
		RequiredContinuationDuration: plan.EffectiveDuration,
	})
	if err != nil {
		state.fallbackReason = "historical_neighbor_selection_failed"
		return wrapDependency(ErrNeighborSelectionFailed, err)
	}
	if err := selection.Validate(); err != nil {
		state.fallbackReason = "historical_neighbor_selection_invalid"
		return wrapDependency(ErrNeighborSelectionFailed, err)
	}
	if selection.CurrentTrajectoryID != request.CurrentTrajectory.ID ||
		!selection.AsOfTime.UTC().Equal(plan.AsOfTime) ||
		selection.RequiredContinuationDuration != plan.EffectiveDuration {
		state.fallbackReason = "historical_neighbor_selection_request_mismatch"
		return wrapDependency(ErrEvidenceBindingInvalid, fmt.Errorf("selection does not match request and plan"))
	}
	if selection.Status == projectionneighbors.StatusUnavailable {
		state.fallbackReason = "historical_neighbors_unavailable"
		state.notices = append(state.notices, Notice{Code: "historical_neighbors_unavailable", Message: "No usable historical neighbors were selected, so the kinematic baseline was selected."})
		return nil
	}

	pattern, err := evaluatePatternConfidence(
		composer.config.PatternConfidenceEvaluator,
		selection.Clone(),
		cloneTrajectories(request.HistoricalCandidates),
	)
	if err != nil {
		state.fallbackReason = "historical_pattern_confidence_failed"
		return wrapDependency(ErrPatternConfidenceFailed, err)
	}
	if err := pattern.Validate(); err != nil {
		state.fallbackReason = "historical_pattern_confidence_invalid"
		return wrapDependency(ErrPatternConfidenceFailed, err)
	}
	if pattern.SourceSelectionFingerprint != selection.InputFingerprint ||
		!sameStrings(selectedTrajectoryIDs(selection), pattern.SelectedTrajectoryIDs) {
		state.fallbackReason = "historical_pattern_selection_mismatch"
		return wrapDependency(ErrEvidenceBindingInvalid, fmt.Errorf("pattern does not match selection"))
	}
	if !pattern.Usable {
		state.fallbackReason = "historical_pattern_not_usable"
		state.notices = append(state.notices, Notice{Code: "historical_pattern_not_usable", Message: "Pattern Confidence rejected the selected historical neighbors, so the kinematic baseline was selected."})
		return nil
	}

	freshness, err := composer.config.FreshnessEvaluator.Evaluate(selection.Clone(), pattern.Clone())
	if err != nil {
		state.fallbackReason = "pattern_freshness_evaluation_failed"
		return wrapDependency(ErrFreshnessEvaluationFailed, err)
	}
	if err := freshness.Validate(); err != nil {
		state.fallbackReason = "pattern_freshness_result_invalid"
		return wrapDependency(ErrFreshnessEvaluationFailed, err)
	}
	if freshness.SourceSelectionFingerprint != selection.InputFingerprint ||
		freshness.SourcePatternFingerprint != pattern.InputFingerprint ||
		!freshness.AsOfTime.UTC().Equal(plan.AsOfTime) ||
		!sameStrings(selectedTrajectoryIDs(selection), freshness.SelectedTrajectoryIDs) {
		state.fallbackReason = "pattern_freshness_evidence_mismatch"
		return wrapDependency(ErrEvidenceBindingInvalid, fmt.Errorf("freshness does not match selection, pattern, or plan"))
	}
	if !freshness.Usable ||
		(freshness.Decision == projectionfreshness.DecisionLimited && composer.config.FreshnessLimitedPolicy == LimitedEvidenceReject) {
		state.fallbackReason = "pattern_freshness_guard_blocked"
		state.notices = append(state.notices, Notice{Code: "pattern_freshness_guard_blocked", Message: "Pattern Freshness policy did not permit historical continuation, so the kinematic baseline was selected."})
		return nil
	}

	if request.Route.Status != routecontract.RouteStatusComplete {
		state.fallbackReason = "complete_route_unavailable"
		state.notices = append(state.notices, Notice{Code: "complete_route_unavailable", Message: "Route Intelligence did not resolve a complete origin-destination route, so route-frequency support could not authorize historical continuation."})
		return nil
	}
	if request.RouteHistory == nil {
		state.fallbackReason = "route_history_unavailable"
		state.notices = append(state.notices, Notice{Code: "route_history_unavailable", Message: "No route-history summary was supplied, so historical continuation was blocked."})
		return nil
	}
	history := request.RouteHistory.Clone()
	if err := history.Validate(); err != nil {
		state.fallbackReason = "route_history_invalid"
		return wrapDependency(ErrRouteFrequencyEvaluationFailed, err)
	}
	expectedRouteKey, err := completeRouteKey(request.Route)
	if err != nil || history.RouteKey != expectedRouteKey || !history.AsOfTime.UTC().Equal(plan.AsOfTime) {
		state.fallbackReason = "route_history_request_mismatch"
		return wrapDependency(ErrEvidenceBindingInvalid, fmt.Errorf("route history does not match route and plan"))
	}

	frequency, err := composer.config.RouteFrequencyEvaluator.Evaluate(request.Route.Clone(), history.Clone())
	if err != nil {
		state.fallbackReason = "route_frequency_evaluation_failed"
		return wrapDependency(ErrRouteFrequencyEvaluationFailed, err)
	}
	if err := frequency.Validate(); err != nil {
		state.fallbackReason = "route_frequency_result_invalid"
		return wrapDependency(ErrRouteFrequencyEvaluationFailed, err)
	}
	if frequency.RouteKey != expectedRouteKey ||
		!frequency.AsOfTime.UTC().Equal(plan.AsOfTime) ||
		frequency.HistoryInputFingerprint != history.InputFingerprint {
		state.fallbackReason = "route_frequency_evidence_mismatch"
		return wrapDependency(ErrEvidenceBindingInvalid, fmt.Errorf("route-frequency result does not match history and plan"))
	}
	if !frequency.Usable ||
		(frequency.Decision == projectionroutefrequency.DecisionLimited && composer.config.RouteFrequencyLimitedPolicy == LimitedEvidenceReject) {
		state.fallbackReason = "route_frequency_guard_blocked"
		state.notices = append(state.notices, Notice{Code: "route_frequency_guard_blocked", Message: "Low-Frequency Route policy did not permit historical continuation, so the kinematic baseline was selected."})
		return nil
	}

	authorized := AuthorizedHistoricalEvidence{
		Plan:       plan.Clone(),
		Route:      request.Route.Clone(),
		History:    history.Clone(),
		RouteScope: routeScope.Clone(),
		Selection:  selection.Clone(),
		Pattern:    pattern.Clone(),
		Freshness:  freshness.Clone(),
		Frequency:  frequency.Clone(),
	}
	if err := authorized.Validate(request); err != nil {
		state.fallbackReason = "historical_evidence_binding_invalid"
		return wrapDependency(ErrEvidenceBindingInvalid, err)
	}
	state.authorized = &authorized
	state.strategy = StrategyHistoricalNeighbor
	state.fallbackReason = ""
	state.notices = append(state.notices, Notice{Code: "historical_neighbor_continuation_authorized", Message: "Pattern confidence, freshness, and route-frequency evidence authorized historical-neighbor continuation."})
	if freshness.Decision == projectionfreshness.DecisionLimited {
		state.notices = append(state.notices, Notice{Code: "historical_projection_authorized_with_limited_freshness", Message: "Historical continuation was authorized with limited Pattern Freshness evidence under the configured policy."})
	}
	if frequency.Decision == projectionroutefrequency.DecisionLimited {
		state.notices = append(state.notices, Notice{Code: "historical_projection_authorized_with_limited_route_frequency", Message: "Historical continuation was authorized with limited route-frequency evidence under the configured policy."})
	}
	return nil
}

func (composer *Composer) project(
	request Request,
	plan projectionhorizon.Plan,
	state *compositionState,
) (projectioncontract.Result, error) {
	if state.strategy != StrategyHistoricalNeighbor || state.authorized == nil {
		return composer.kinematic(request, plan)
	}

	evidence := state.authorized.Clone()
	historicalRequest := projectioncontinuation.Request{
		CurrentTrajectory: request.CurrentTrajectory.Clone(),
		Candidates:        cloneTrajectories(request.HistoricalCandidates),
		RouteScope:        evidence.RouteScope.Clone(),
		AsOfTime:          plan.AsOfTime,
		RequestedDuration: plan.RequestedDuration,
		GeneratedAt:       request.GeneratedAt,
	}
	approvedEvidence := projectioncontinuation.ApprovedEvidence{
		Selection: evidence.Selection.Clone(),
		Pattern:   evidence.Pattern.Clone(),
	}
	outcome, err := composer.config.HistoricalProjector.ProjectApprovedWithPlan(
		historicalRequest,
		plan.Clone(),
		approvedEvidence.Clone(),
	)
	if err != nil {
		reason := "historical_projection_failed"
		if errors.Is(err, ErrHistoricalProjectionLineageInvalid) {
			reason = "historical_projection_lineage_mismatch"
		}
		return composer.historicalFailure(
			request,
			plan,
			state,
			reason,
			wrapDependency(ErrHistoricalProjectionFailed, err),
		)
	}
	if err := outcome.ValidateAgainst(plan, approvedEvidence); err != nil {
		return composer.historicalFailure(request, plan, state, "historical_projection_lineage_mismatch", err)
	}
	result := outcome.Projection.Clone()
	if result.Method.Name == projectionbaseline.MethodName {
		if err := validateProjectionPostconditions(result, request, plan, StrategyKinematic); err != nil {
			return composer.historicalFailure(request, plan, state, "historical_projector_internal_fallback_invalid", err)
		}
		state.strategy = StrategyKinematic
		state.fallbackReason = "historical_projector_internal_fallback"
		state.notices = append(state.notices, Notice{Code: "historical_projector_internal_fallback", Message: "The authorized historical projector selected its internal kinematic fallback."})
		return result.Clone(), nil
	}
	if err := validateProjectionPostconditions(result, request, plan, StrategyHistoricalNeighbor); err != nil {
		return composer.historicalFailure(request, plan, state, "historical_projection_postcondition_failed", err)
	}
	return result.Clone(), nil
}

func (composer *Composer) historicalFailure(
	request Request,
	plan projectionhorizon.Plan,
	state *compositionState,
	reason string,
	failure error,
) (projectioncontract.Result, error) {
	if composer.config.DependencyFailurePolicy == DependencyFailureReturnError {
		return projectioncontract.Result{}, failure
	}
	state.strategy = StrategyKinematic
	state.fallbackReason = reason
	state.notices = append(state.notices, Notice{Code: reason, Message: "Historical continuation failed its production postconditions, so the kinematic baseline was selected."})
	return composer.kinematic(request, plan)
}

func (composer *Composer) kinematic(
	request Request,
	plan projectionhorizon.Plan,
) (projectioncontract.Result, error) {
	result, err := composer.config.KinematicProjector.ProjectWithPlan(
		projectionbaseline.Request{
			Trajectory:        request.CurrentTrajectory.Clone(),
			AsOfTime:          plan.AsOfTime,
			RequestedDuration: plan.RequestedDuration,
			GeneratedAt:       request.GeneratedAt,
		},
		plan.Clone(),
	)
	if err != nil {
		return projectioncontract.Result{}, wrapDependency(ErrKinematicProjectionFailed, err)
	}
	if err := validateProjectionPostconditions(result, request, plan, StrategyKinematic); err != nil {
		return projectioncontract.Result{}, err
	}
	return result.Clone(), nil
}

func (composer *Composer) attachArrival(
	request Request,
	projection projectioncontract.Result,
	state *compositionState,
) (projectioncontract.Result, ArrivalStatus, error) {
	outcome, err := composer.config.ArrivalEstimator.EstimateArrival(projectionarrival.Request{
		Projection:        projection.Clone(),
		Route:             request.Route.Clone(),
		CurrentTrajectory: request.CurrentTrajectory.Clone(),
		GeneratedAt:       request.GeneratedAt,
	})
	if err != nil {
		if composer.config.ArrivalFailurePolicy == ArrivalFailureReturnError {
			return projectioncontract.Result{}, ArrivalStatusFailed, wrapDependency(ErrArrivalEstimationFailed, err)
		}
		state.notices = append(state.notices, Notice{Code: "estimated_arrival_failed_projection_preserved", Message: "Estimated Arrival failed, but the configured policy preserved the authorized position projection."})
		return projection.Clone(), ArrivalStatusFailed, nil
	}
	if err := outcome.Validate(); err != nil {
		if composer.config.ArrivalFailurePolicy == ArrivalFailureReturnError {
			return projectioncontract.Result{}, ArrivalStatusFailed, wrapDependency(ErrArrivalEstimationFailed, err)
		}
		state.notices = append(state.notices, Notice{Code: "estimated_arrival_invalid_projection_preserved", Message: "Estimated Arrival returned an invalid delta, but the configured policy preserved the authorized position projection."})
		return projection.Clone(), ArrivalStatusFailed, nil
	}
	state.notices = append(state.notices, outcome.Notices...)
	result := projection.Clone()
	result.Arrival = cloneArrivalEstimate(outcome.Estimate)
	if outcome.Status == ArrivalOutcomeWithheld {
		return result, ArrivalStatusWithheld, nil
	}
	return result, ArrivalStatusAttached, nil
}

func wrapDependency(sentinel error, cause error) error {
	return fmt.Errorf("%w: %w", sentinel, cause)
}

func pointerToSelection(value projectionneighbors.Result) *projectionneighbors.Result {
	cloned := value.Clone()
	return &cloned
}

func pointerToPattern(value projectionpatternconfidence.Result) *projectionpatternconfidence.Result {
	cloned := value.Clone()
	return &cloned
}

func pointerToFreshness(value projectionfreshness.Result) *projectionfreshness.Result {
	cloned := value.Clone()
	return &cloned
}

func pointerToFrequency(value projectionroutefrequency.Result) *projectionroutefrequency.Result {
	cloned := value.Clone()
	return &cloned
}
