package projectionevaluation

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

var (
	ErrProjectionInvalid            = errors.New("projection contract is invalid")
	ErrTrajectoryIdentifierMismatch = errors.New("projection and actual trajectory identifiers must match")
	ErrEvaluatedAtInvalid           = errors.New("evaluation time must not precede projection generation")
	ErrActualArrivalInvalid         = errors.New("actual arrival evidence is invalid")
	ErrEvaluationResultInvalid      = errors.New("projection evaluation result is invalid")
)

type Evaluator struct{ config Config }

func New(config Config) (*Evaluator, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("validate projection evaluation config: %w", err)
	}
	return &Evaluator{config: config}, nil
}

type Request struct {
	Projection        projectioncontract.Result
	ActualTrajectory  trajectory.FlightTrajectory
	TruthAvailability []TruthAvailability
	ActualArrival     *ActualArrival
	EvaluatedAt       time.Time
}

func (evaluator *Evaluator) Evaluate(request Request) (Result, error) {
	if evaluator == nil {
		return Result{}, ErrEvaluationResultInvalid
	}
	projectionReport := projectioncontract.Validate(request.Projection)
	if projectionReport.Status != projectioncontract.ValidationStatusValid {
		return Result{}, fmt.Errorf("%w: %#v", ErrProjectionInvalid, projectionReport.Issues)
	}
	if strings.TrimSpace(request.ActualTrajectory.ID) == "" || request.ActualTrajectory.ID != request.Projection.TrajectoryID {
		return Result{}, ErrTrajectoryIdentifierMismatch
	}
	evaluatedAt := request.EvaluatedAt.UTC()
	if evaluatedAt.IsZero() || evaluatedAt.Before(request.Projection.GeneratedAt.UTC()) {
		return Result{}, ErrEvaluatedAtInvalid
	}
	actualArrival := normalizeActualArrival(request.ActualArrival)
	if err := validateActualArrival(actualArrival, request.Projection, evaluatedAt); err != nil {
		return Result{}, err
	}
	asOfTime := request.Projection.Horizon.AsOfTime.UTC()
	truth, err := normalizeTruthPoints(
		request.ActualTrajectory,
		request.TruthAvailability,
		asOfTime,
		evaluatedAt,
	)
	if err != nil {
		return Result{}, err
	}

	points, rejectedImplausible := evaluateForecastPoints(request.Projection, truth, evaluator.config)
	positionMetrics := buildPositionMetrics(len(request.Projection.Points), points)
	endpointMetrics := buildEndpointMetrics(request.Projection.Horizon.EndTime, points)
	confidenceMetrics := buildConfidenceMetrics(points)
	leadTimeMetrics := buildLeadTimeMetrics(points, evaluator.config.LeadTimeBucketSize)
	arrivalMetrics, arrivalLimitation := evaluateArrival(request.Projection.Arrival, actualArrival)

	limitations := make([]Notice, 0, 8)
	if truth.excludedAfterObservationCutoff > 0 {
		limitations = append(limitations, Notice{
			Code:    "truth_after_observation_cutoff_excluded",
			Message: fmt.Sprintf("%d actual trajectory points after the evaluation event-time cutoff were excluded.", truth.excludedAfterObservationCutoff),
		})
	}
	if truth.excludedAfterAvailabilityCutoff > 0 {
		limitations = append(limitations, Notice{
			Code:    "truth_after_availability_cutoff_excluded",
			Message: fmt.Sprintf("%d actual trajectory points not yet available to the system at evaluation time were excluded.", truth.excludedAfterAvailabilityCutoff),
		})
	}
	if rejectedImplausible > 0 {
		limitations = append(limitations, Notice{
			Code:    "implausible_truth_interpolation_rejected",
			Message: fmt.Sprintf("%d forecast truth matches were rejected because the surrounding observed segment exceeded configured physical movement limits.", rejectedImplausible),
		})
	}
	if arrivalLimitation != nil {
		limitations = append(limitations, *arrivalLimitation)
	}

	status := StatusUnavailable
	switch {
	case len(points) < evaluator.config.MinimumEvaluatedPointCount:
		limitations = append(limitations, Notice{
			Code:    "insufficient_evaluated_projection_points",
			Message: fmt.Sprintf("Evaluation requires at least %d forecast points with actual truth, but %d were available.", evaluator.config.MinimumEvaluatedPointCount, len(points)),
		})
	case positionMetrics.EvaluatedPointCount == positionMetrics.ForecastPointCount &&
		arrivalEvaluationComplete(request.Projection.Arrival, actualArrival, arrivalMetrics):
		status = StatusComplete
	default:
		status = StatusPartial
		if positionMetrics.MissingActualPointCount > 0 {
			limitations = append(limitations, Notice{
				Code:    "actual_trajectory_coverage_partial",
				Message: fmt.Sprintf("%d forecast points could not be matched to actual trajectory evidence.", positionMetrics.MissingActualPointCount),
			})
		}
	}

	policy := evaluator.config.snapshot()
	projectionFingerprint := projectionSnapshotFingerprint(request.Projection)
	truthFingerprint := truthSnapshotFingerprint(request.ActualTrajectory.ID, truth)
	truthEvidence := buildTruthEvidenceSummary(truth, rejectedImplausible, truthFingerprint)
	result := Result{
		Version:                       Version,
		Status:                        status,
		TrajectoryID:                  request.Projection.TrajectoryID,
		ProjectionMethod:              request.Projection.Method,
		ProjectionAsOfTime:            asOfTime,
		ProjectionHorizonEndTime:      request.Projection.Horizon.EndTime.UTC(),
		ForecastStep:                  request.Projection.Horizon.Step,
		ProjectionGeneratedAt:         request.Projection.GeneratedAt.UTC(),
		EvaluatedAt:                   evaluatedAt,
		Policy:                        policy,
		TruthEvidence:                 truthEvidence,
		ProjectionInputFingerprint:    request.Projection.Provenance.InputFingerprint,
		ProjectionSnapshotFingerprint: projectionFingerprint,
		TruthSnapshotFingerprint:      truthFingerprint,
		EvaluationInputFingerprint:    evaluationFingerprint(projectionFingerprint, truthFingerprint, actualArrival, evaluatedAt, policy),
		Points:                        append([]PointEvaluation(nil), points...),
		Position:                      positionMetrics,
		Endpoint:                      endpointMetrics,
		Confidence:                    confidenceMetrics,
		LeadTimes:                     append([]LeadTimeMetrics(nil), leadTimeMetrics...),
		Arrival:                       arrivalMetrics,
		Limitations:                   normalizeNotices(limitations),
	}
	if err := result.Validate(); err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrEvaluationResultInvalid, err)
	}
	return result.Clone(), nil
}

func buildTruthEvidenceSummary(
	truth normalizedTruth,
	rejectedImplausible int,
	fingerprint string,
) TruthEvidenceSummary {
	sourceSet := make(map[string]struct{})
	summary := TruthEvidenceSummary{
		KnowledgeCutoffMode:                   TruthKnowledgeCutoffMode,
		CanonicalPointCount:                   len(truth.points),
		ExcludedAfterObservationCutoffCount:   truth.excludedAfterObservationCutoff,
		ExcludedAfterAvailabilityCutoffCount:  truth.excludedAfterAvailabilityCutoff,
		RejectedImplausibleInterpolationCount: rejectedImplausible,
		InputFingerprint:                      fingerprint,
	}
	for _, item := range truth.points {
		if item.point.ObservedAt.After(summary.LatestObservedAt) {
			summary.LatestObservedAt = item.point.ObservedAt
		}
		if item.availableAt.After(summary.LatestAvailableAt) {
			summary.LatestAvailableAt = item.availableAt
		}
		sourceSet[item.point.SourceName] = struct{}{}
		sourceSet[item.evidenceSource] = struct{}{}
	}
	for source := range sourceSet {
		if strings.TrimSpace(source) != "" {
			summary.SourceNames = append(summary.SourceNames, source)
		}
	}
	sort.Strings(summary.SourceNames)
	return summary
}
