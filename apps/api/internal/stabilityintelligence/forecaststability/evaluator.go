package forecaststability

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

func EvaluateDecisionStability(
	request StabilityRequest,
	policy StabilityPolicy,
) (StabilityResult, error) {
	if err := policy.Validate(); err != nil {
		return StabilityResult{}, err
	}
	normalized, err := normalizeStabilityRequest(request)
	if err != nil {
		return StabilityResult{}, err
	}

	metrics, components := compareVersions(normalized.Baseline, normalized.Candidate, policy)
	level, score, reasons := classifyStability(normalized.Baseline, normalized.Candidate, metrics, components, policy)
	status := ResultStatusComplete
	if level == StabilityLevelIndeterminate {
		status = ResultStatusLimited
	}
	result := StabilityResult{
		SchemaVersion:      SchemaVersionV1,
		Status:             status,
		TrajectoryID:       normalized.Baseline.TrajectoryID,
		BaselineVersionID:  normalized.Baseline.VersionID,
		CandidateVersionID: normalized.Candidate.VersionID,
		Level:              level,
		Score:              score,
		Metrics:            metrics,
		Components:         components,
		Reasons:            reasons,
		Limitations:        stabilityLimitations(level, policy),
		Explanations:       stabilityExplanations(level),
		ScopeGuard:         ScopeGuardResearchOnly,
		Provenance: StabilityProvenance{
			BaselineVersionID:          normalized.Baseline.VersionID,
			CandidateVersionID:         normalized.Candidate.VersionID,
			BaselineOutputFingerprint:  normalized.Baseline.OutputFingerprint,
			CandidateOutputFingerprint: normalized.Candidate.OutputFingerprint,
		},
		EvaluatedAt: normalized.EvaluatedAt,
	}
	result.Provenance.InputFingerprint = stabilityInputFingerprint(result, policy)
	if err := ValidateStabilityResult(result, policy); err != nil {
		return StabilityResult{}, err
	}
	return result.Clone(), nil
}

func normalizeStabilityRequest(request StabilityRequest) (StabilityRequest, error) {
	normalized := StabilityRequest{
		Baseline:    request.Baseline.Clone(),
		Candidate:   request.Candidate.Clone(),
		EvaluatedAt: request.EvaluatedAt.UTC(),
	}
	versionPolicy := DefaultVersionPolicy()
	if err := ValidateVersionRecord(normalized.Baseline, versionPolicy); err != nil {
		return StabilityRequest{}, fmt.Errorf("%w: baseline: %v", ErrInvalidStabilityRequest, err)
	}
	if err := ValidateVersionRecord(normalized.Candidate, versionPolicy); err != nil {
		return StabilityRequest{}, fmt.Errorf("%w: candidate: %v", ErrInvalidStabilityRequest, err)
	}
	if normalized.Baseline.TrajectoryID != normalized.Candidate.TrajectoryID {
		return StabilityRequest{}, fmt.Errorf("%w: trajectory mismatch", ErrInvalidStabilityRequest)
	}
	if normalized.Candidate.Ordinal < normalized.Baseline.Ordinal {
		return StabilityRequest{}, fmt.Errorf("%w: candidate precedes baseline ordinal", ErrInvalidStabilityRequest)
	}
	if normalized.EvaluatedAt.IsZero() || normalized.EvaluatedAt.Before(normalized.Candidate.CreatedAt) {
		return StabilityRequest{}, fmt.Errorf("%w: evaluated-at chronology", ErrInvalidStabilityRequest)
	}
	return normalized, nil
}

func compareVersions(
	baseline ForecastVersionRecord,
	candidate ForecastVersionRecord,
	policy StabilityPolicy,
) (StabilityMetrics, []StabilityComponent) {
	metrics := StabilityMetrics{
		BaselinePointCount:      len(baseline.Projection.Points),
		CandidatePointCount:     len(candidate.Projection.Points),
		ProjectionStatusChanged: baseline.Projection.Status != candidate.Projection.Status,
		MethodChanged:           methodIdentity(baseline.Method) != methodIdentity(candidate.Method),
		PolicyChanged:           baseline.PolicyVersion != candidate.PolicyVersion,
		ImplementationChanged:   baseline.ImplementationVersion != candidate.ImplementationVersion,
		InputChanged:            baseline.InputFingerprint != candidate.InputFingerprint,
		OutputChanged:           baseline.OutputFingerprint != candidate.OutputFingerprint,
	}

	baselinePoints := pointsByTime(baseline.Projection.Points)
	candidatePoints := pointsByTime(candidate.Projection.Points)
	times := make([]time.Time, 0)
	for forecastTime := range baselinePoints {
		if _, exists := candidatePoints[forecastTime]; exists {
			times = append(times, forecastTime)
		}
	}
	sort.Slice(times, func(left, right int) bool { return times[left].Before(times[right]) })
	metrics.AlignedPointCount = len(times)
	maximumPointCount := maxInt(metrics.BaselinePointCount, metrics.CandidatePointCount)
	if maximumPointCount > 0 {
		metrics.AlignedPointShare = float64(metrics.AlignedPointCount) / float64(maximumPointCount)
	}

	horizontalShifts := make([]float64, 0, len(times))
	confidenceDeltas := make([]float64, 0, len(times))
	uncertaintyChanges := make([]float64, 0, len(times))
	for _, forecastTime := range times {
		left := baselinePoints[forecastTime]
		right := candidatePoints[forecastTime]
		shift := haversineKilometers(
			left.Position.Latitude,
			left.Position.Longitude,
			right.Position.Latitude,
			right.Position.Longitude,
		)
		horizontalShifts = append(horizontalShifts, shift)
		metrics.MaximumHorizontalShiftKilometers = math.Max(metrics.MaximumHorizontalShiftKilometers, shift)
		confidenceDeltas = append(confidenceDeltas, math.Abs(right.Confidence.Score-left.Confidence.Score))
		uncertaintyChanges = append(uncertaintyChanges, safeRelativeChange(
			left.Uncertainty.HorizontalRadiusM,
			right.Uncertainty.HorizontalRadiusM,
		))
	}
	metrics.MeanHorizontalShiftKilometers = mean(horizontalShifts)
	metrics.MeanAbsolutePointConfidenceDelta = mean(confidenceDeltas)
	metrics.MeanRelativeHorizontalUncertaintyChange = mean(uncertaintyChanges)
	metrics.AggregateConfidenceDelta = math.Abs(candidate.Projection.Confidence.Score - baseline.Projection.Confidence.Score)

	if baseline.Projection.Arrival != nil && candidate.Projection.Arrival != nil &&
		strings.EqualFold(baseline.Projection.Arrival.AirportICAOCode, candidate.Projection.Arrival.AirportICAOCode) {
		metrics.ArrivalComparable = true
		metrics.ArrivalShiftSeconds = math.Abs(candidate.Projection.Arrival.EstimatedTime.Sub(
			baseline.Projection.Arrival.EstimatedTime,
		).Seconds())
	}

	positionComparable := metrics.AlignedPointShare >= policy.MinimumAlignedPointShare && metrics.AlignedPointCount > 0
	positionStability := 0.0
	if positionComparable {
		meanInstability := normalizedThresholdScore(
			metrics.MeanHorizontalShiftKilometers,
			policy.StableMeanHorizontalShiftKilometers,
			policy.MaterialMeanHorizontalShiftKilometers,
		)
		maxInstability := normalizedThresholdScore(
			metrics.MaximumHorizontalShiftKilometers,
			policy.StableMaximumHorizontalShiftKilometers,
			policy.MaterialMaximumHorizontalShiftKilometers,
		)
		positionStability = 1 - math.Max(meanInstability, maxInstability)
	}

	uncertaintyComparable := positionComparable
	uncertaintyStability := 0.0
	if uncertaintyComparable {
		uncertaintyStability = 1 - normalizedThresholdScore(
			metrics.MeanRelativeHorizontalUncertaintyChange,
			policy.StableRelativeUncertaintyChange,
			policy.MaterialRelativeUncertaintyChange,
		)
	}

	confidenceComparable := true
	confidenceInstability := math.Max(
		normalizedThresholdScore(metrics.AggregateConfidenceDelta, policy.StableConfidenceDelta, policy.MaterialConfidenceDelta),
		normalizedThresholdScore(metrics.MeanAbsolutePointConfidenceDelta, policy.StableConfidenceDelta, policy.MaterialConfidenceDelta),
	)
	confidenceStability := 1 - confidenceInstability

	arrivalComparable := metrics.ArrivalComparable || (baseline.Projection.Arrival == nil && candidate.Projection.Arrival == nil)
	arrivalStability := 1.0
	if metrics.ArrivalComparable {
		arrivalStability = 1 - normalizedThresholdScore(
			metrics.ArrivalShiftSeconds,
			policy.StableArrivalShiftSeconds,
			policy.MaterialArrivalShiftSeconds,
		)
	} else if (baseline.Projection.Arrival == nil) != (candidate.Projection.Arrival == nil) {
		arrivalStability = 0
		arrivalComparable = true
	}

	decisionPenalty := 0.0
	for _, changed := range []bool{
		metrics.ProjectionStatusChanged,
		metrics.MethodChanged,
		metrics.PolicyChanged,
		metrics.ImplementationChanged,
	} {
		if changed {
			decisionPenalty += 0.25
		}
	}
	decisionStability := 1 - clampUnit(decisionPenalty)

	components := []StabilityComponent{
		{Name: "forecast_position", Stability: clampUnit(positionStability), Weight: policy.Weights.Position, Comparable: positionComparable, Explanation: "Compares aligned forecast positions by mean and maximum horizontal displacement."},
		{Name: "forecast_uncertainty", Stability: clampUnit(uncertaintyStability), Weight: policy.Weights.Uncertainty, Comparable: uncertaintyComparable, Explanation: "Compares relative changes in horizontal uncertainty for aligned forecast points."},
		{Name: "forecast_confidence", Stability: clampUnit(confidenceStability), Weight: policy.Weights.Confidence, Comparable: confidenceComparable, Explanation: "Compares aggregate and point-level confidence changes."},
		{Name: "arrival_decision", Stability: clampUnit(arrivalStability), Weight: policy.Weights.Arrival, Comparable: arrivalComparable, Explanation: "Compares Estimated Arrival when the same destination is available and records arrival appearance or disappearance."},
		{Name: "decision_metadata", Stability: clampUnit(decisionStability), Weight: policy.Weights.Decision, Comparable: true, Explanation: "Compares projection status, method, policy, and implementation identity."},
	}
	return metrics, components
}

func classifyStability(
	baseline ForecastVersionRecord,
	candidate ForecastVersionRecord,
	metrics StabilityMetrics,
	components []StabilityComponent,
	policy StabilityPolicy,
) (StabilityLevel, float64, []StabilityReason) {
	if baseline.VersionID == candidate.VersionID ||
		(baseline.DecisionFingerprint == candidate.DecisionFingerprint && baseline.OutputFingerprint == candidate.OutputFingerprint) {
		return StabilityLevelUnchanged, 1, []StabilityReason{{
			Code:    "immutable_version_unchanged",
			Message: "The baseline and candidate resolve to the same immutable forecast decision and output.",
			Impact:  0,
		}}
	}

	comparableWeight := 0.0
	weightedStability := 0.0
	for _, component := range components {
		if !component.Comparable {
			continue
		}
		comparableWeight += component.Weight
		weightedStability += component.Stability * component.Weight
	}
	if comparableWeight < 0.65 || metrics.AlignedPointShare < policy.MinimumAlignedPointShare {
		return StabilityLevelIndeterminate, 0, []StabilityReason{{
			Code:    "insufficient_comparable_forecast_evidence",
			Message: "Too little aligned forecast evidence is available for a reliable decision-stability classification.",
			Impact:  1,
		}}
	}
	score := clampUnit(weightedStability / comparableWeight)
	reasons := stabilityReasons(metrics, policy)

	material := metrics.MethodChanged ||
		metrics.ProjectionStatusChanged ||
		metrics.MeanHorizontalShiftKilometers >= policy.MaterialMeanHorizontalShiftKilometers ||
		metrics.MaximumHorizontalShiftKilometers >= policy.MaterialMaximumHorizontalShiftKilometers ||
		metrics.AggregateConfidenceDelta >= policy.MaterialConfidenceDelta ||
		metrics.MeanAbsolutePointConfidenceDelta >= policy.MaterialConfidenceDelta ||
		metrics.MeanRelativeHorizontalUncertaintyChange >= policy.MaterialRelativeUncertaintyChange ||
		(metrics.ArrivalComparable && metrics.ArrivalShiftSeconds >= policy.MaterialArrivalShiftSeconds) ||
		((baseline.Projection.Arrival == nil) != (candidate.Projection.Arrival == nil))
	if material {
		return StabilityLevelMaterialChange, score, reasons
	}

	changed := metrics.MeanHorizontalShiftKilometers > policy.StableMeanHorizontalShiftKilometers ||
		metrics.MaximumHorizontalShiftKilometers > policy.StableMaximumHorizontalShiftKilometers ||
		metrics.AggregateConfidenceDelta > policy.StableConfidenceDelta ||
		metrics.MeanAbsolutePointConfidenceDelta > policy.StableConfidenceDelta ||
		metrics.MeanRelativeHorizontalUncertaintyChange > policy.StableRelativeUncertaintyChange ||
		(metrics.ArrivalComparable && metrics.ArrivalShiftSeconds > policy.StableArrivalShiftSeconds) ||
		metrics.PolicyChanged || metrics.ImplementationChanged
	if changed {
		return StabilityLevelChanged, score, reasons
	}
	return StabilityLevelStable, score, reasons
}

func stabilityReasons(metrics StabilityMetrics, policy StabilityPolicy) []StabilityReason {
	reasons := []StabilityReason{
		{Code: "aligned_point_share", Message: fmt.Sprintf("Aligned forecast point share is %.3f.", metrics.AlignedPointShare), Impact: 1 - metrics.AlignedPointShare},
		{Code: "mean_horizontal_shift", Message: fmt.Sprintf("Mean horizontal forecast shift is %.3f kilometers.", metrics.MeanHorizontalShiftKilometers), Impact: normalizedThresholdScore(metrics.MeanHorizontalShiftKilometers, policy.StableMeanHorizontalShiftKilometers, policy.MaterialMeanHorizontalShiftKilometers)},
		{Code: "maximum_horizontal_shift", Message: fmt.Sprintf("Maximum horizontal forecast shift is %.3f kilometers.", metrics.MaximumHorizontalShiftKilometers), Impact: normalizedThresholdScore(metrics.MaximumHorizontalShiftKilometers, policy.StableMaximumHorizontalShiftKilometers, policy.MaterialMaximumHorizontalShiftKilometers)},
		{Code: "aggregate_confidence_delta", Message: fmt.Sprintf("Aggregate confidence changed by %.3f.", metrics.AggregateConfidenceDelta), Impact: normalizedThresholdScore(metrics.AggregateConfidenceDelta, policy.StableConfidenceDelta, policy.MaterialConfidenceDelta)},
		{Code: "uncertainty_change", Message: fmt.Sprintf("Mean relative horizontal uncertainty change is %.3f.", metrics.MeanRelativeHorizontalUncertaintyChange), Impact: normalizedThresholdScore(metrics.MeanRelativeHorizontalUncertaintyChange, policy.StableRelativeUncertaintyChange, policy.MaterialRelativeUncertaintyChange)},
	}
	if metrics.ArrivalComparable {
		reasons = append(reasons, StabilityReason{
			Code:    "arrival_shift",
			Message: fmt.Sprintf("Estimated Arrival shifted by %.0f seconds.", metrics.ArrivalShiftSeconds),
			Impact:  normalizedThresholdScore(metrics.ArrivalShiftSeconds, policy.StableArrivalShiftSeconds, policy.MaterialArrivalShiftSeconds),
		})
	}
	if metrics.MethodChanged {
		reasons = append(reasons, StabilityReason{Code: "method_changed", Message: "The projection method identity changed.", Impact: 1})
	}
	if metrics.PolicyChanged {
		reasons = append(reasons, StabilityReason{Code: "policy_changed", Message: "The forecast policy version changed.", Impact: 0.5})
	}
	if metrics.ImplementationChanged {
		reasons = append(reasons, StabilityReason{Code: "implementation_changed", Message: "The implementation version changed.", Impact: 0.5})
	}
	sort.Slice(reasons, func(left, right int) bool { return reasons[left].Code < reasons[right].Code })
	return reasons
}

func stabilityLimitations(level StabilityLevel, policy StabilityPolicy) []Limitation {
	limitations := []Limitation{
		{
			Code:    "stability_is_not_accuracy",
			Message: "Decision stability describes change between forecast versions and does not prove that either forecast is accurate.",
			Scope:   "forecast_accuracy",
		},
		{
			Code:    "experimental_project_policy_thresholds",
			Message: fmt.Sprintf("Thresholds are explicit experimental project policy %s and require future historical replay calibration.", policy.Version),
			Scope:   "threshold_calibration",
		},
	}
	if level == StabilityLevelIndeterminate {
		limitations = append(limitations, Limitation{
			Code:    "insufficient_alignment_for_stability",
			Message: "The compared forecasts do not contain enough aligned points for a reliable stability decision.",
			Scope:   "comparability",
		})
	}
	return limitations
}

func stabilityExplanations(level StabilityLevel) []Explanation {
	return []Explanation{
		{
			Code:    "decision_stability_classification",
			Message: "Decision stability combines aligned forecast displacement, uncertainty change, confidence change, Estimated Arrival change, and decision metadata continuity.",
		},
		{
			Code:    "stability_level_" + string(level),
			Message: "The published stability level is " + string(level) + ".",
		},
	}
}

func pointsByTime(points []projectioncontract.ProjectionPoint) map[time.Time]projectioncontract.ProjectionPoint {
	result := make(map[time.Time]projectioncontract.ProjectionPoint, len(points))
	for _, point := range points {
		result[point.ForecastTime.UTC()] = point
	}
	return result
}

func normalizedThresholdScore(value float64, stableBoundary float64, materialBoundary float64) float64 {
	if value <= stableBoundary {
		return 0
	}
	if value >= materialBoundary {
		return 1
	}
	return (value - stableBoundary) / (materialBoundary - stableBoundary)
}

func mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	total := 0.0
	for _, value := range values {
		total += value
	}
	return total / float64(len(values))
}

func maxInt(left int, right int) int {
	if left > right {
		return left
	}
	return right
}
