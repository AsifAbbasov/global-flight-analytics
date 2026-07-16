package forecastanalysis

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/stabilityintelligence/forecaststability"
)

func AnalyzeForecastHistory(
	request Request,
	policy Policy,
	stabilityPolicy forecaststability.StabilityPolicy,
) (Result, error) {
	if err := policy.Validate(); err != nil {
		return Result{}, err
	}
	if err := stabilityPolicy.Validate(); err != nil {
		return Result{}, err
	}

	versions, evaluatedAt, err := normalizeRequest(request, policy)
	if err != nil {
		return Result{}, err
	}

	transitions := make(
		[]forecaststability.StabilityResult,
		0,
		len(versions)-1,
	)
	for index := 1; index < len(versions); index++ {
		item, evaluationErr := forecaststability.EvaluateDecisionStability(
			forecaststability.StabilityRequest{
				Baseline:    versions[index-1],
				Candidate:   versions[index],
				EvaluatedAt: evaluatedAt,
			},
			stabilityPolicy,
		)
		if evaluationErr != nil {
			return Result{}, fmt.Errorf(
				"%w: transition %d: %v",
				ErrInvalidRequest,
				index,
				evaluationErr,
			)
		}
		transitions = append(transitions, item)
	}

	metrics := buildMetrics(versions, transitions)
	trend := classifyTrend(transitions, metrics, policy)
	health := classifyHealth(metrics, policy)
	confidence := analysisConfidence(metrics)
	status := ResultStatusComplete
	if metrics.ComparableTransitionShare < policy.MinimumComparableShare {
		status = ResultStatusLimited
	}

	result := Result{
		SchemaVersion: SchemaVersionV1,
		Status:        status,
		TrajectoryID:  versions[0].TrajectoryID,
		Trend:         trend,
		Health:        health,
		Metrics:       metrics,
		Transitions:   transitions,
		Confidence:    confidence,
		Limitations: []Limitation{
			{
				Code:    "stability_is_not_accuracy",
				Message: "Repeatedly stable forecasts may still be systematically inaccurate.",
				Scope:   "interpretation",
			},
			{
				Code:    "no_ground_truth_comparison",
				Message: "This analysis compares forecast versions and does not compare them with later observed reality.",
				Scope:   "accuracy",
			},
			{
				Code:    "experimental_thresholds",
				Message: "Trend and health thresholds are project-derived experimental policy values requiring replay calibration.",
				Scope:   policy.Version,
			},
		},
		Explanations: []Explanation{
			{
				Code:    "history_trend",
				Message: "Forecast-version stability trend: " + string(trend) + ".",
			},
			{
				Code:    "stability_health",
				Message: "Forecast-version stability health: " + string(health) + ".",
			},
		},
		ScopeGuard: ScopeGuardResearchOnly,
		Provenance: Provenance{
			PolicyVersion: policy.Version,
		},
		EvaluatedAt: evaluatedAt,
	}

	for _, version := range versions {
		result.Provenance.VersionIDs = append(
			result.Provenance.VersionIDs,
			version.VersionID,
		)
		result.Provenance.OutputFingerprints = append(
			result.Provenance.OutputFingerprints,
			version.OutputFingerprint,
		)
	}
	result.Provenance.InputFingerprint = resultFingerprint(result)

	if err := ValidateResult(result, policy); err != nil {
		return Result{}, err
	}
	return result.Clone(), nil
}

func normalizeRequest(
	request Request,
	policy Policy,
) ([]forecaststability.ForecastVersionRecord, time.Time, error) {
	if len(request.Versions) < policy.MinimumVersionCount ||
		len(request.Versions) > policy.MaximumVersionCount {
		return nil, time.Time{}, fmt.Errorf(
			"%w: version count",
			ErrInvalidRequest,
		)
	}

	versions := make(
		[]forecaststability.ForecastVersionRecord,
		0,
		len(request.Versions),
	)
	for _, item := range request.Versions {
		if err := forecaststability.ValidateVersionRecord(
			item,
			forecaststability.DefaultVersionPolicy(),
		); err != nil {
			return nil, time.Time{}, fmt.Errorf(
				"%w: invalid version: %v",
				ErrInvalidRequest,
				err,
			)
		}
		versions = append(versions, item.Clone())
	}

	sort.Slice(
		versions,
		func(left int, right int) bool {
			return versions[left].Ordinal < versions[right].Ordinal
		},
	)

	trajectoryID := versions[0].TrajectoryID
	for index, item := range versions {
		if item.TrajectoryID != trajectoryID {
			return nil, time.Time{}, fmt.Errorf(
				"%w: trajectory mismatch",
				ErrInvalidRequest,
			)
		}
		if index == 0 {
			continue
		}
		previous := versions[index-1]
		if item.Ordinal != previous.Ordinal+1 ||
			item.ParentVersionID != previous.VersionID {
			return nil, time.Time{}, fmt.Errorf(
				"%w: non-contiguous version chain",
				ErrInvalidRequest,
			)
		}
	}

	evaluatedAt := request.EvaluatedAt.UTC()
	if evaluatedAt.IsZero() ||
		evaluatedAt.Before(versions[len(versions)-1].CreatedAt) {
		return nil, time.Time{}, fmt.Errorf(
			"%w: evaluated-at chronology",
			ErrInvalidRequest,
		)
	}
	return versions, evaluatedAt, nil
}

func buildMetrics(
	versions []forecaststability.ForecastVersionRecord,
	transitions []forecaststability.StabilityResult,
) Metrics {
	metrics := Metrics{
		VersionCount:          len(versions),
		TransitionCount:       len(transitions),
		MinimumStabilityScore: 1,
	}

	scoreSum := 0.0
	squareSum := 0.0
	shiftSum := 0.0
	stableRun := 0

	for _, item := range transitions {
		metrics.LatestLevel = item.Level
		switch item.Level {
		case forecaststability.StabilityLevelUnchanged:
			metrics.UnchangedCount++
			stableRun++
		case forecaststability.StabilityLevelStable:
			metrics.StableCount++
			stableRun++
		case forecaststability.StabilityLevelChanged:
			metrics.ChangedCount++
			stableRun = 0
		case forecaststability.StabilityLevelMaterialChange:
			metrics.MaterialChangeCount++
			stableRun = 0
		case forecaststability.StabilityLevelIndeterminate:
			metrics.IndeterminateCount++
			stableRun = 0
		}
		if stableRun > metrics.LongestStableRun {
			metrics.LongestStableRun = stableRun
		}

		if item.Level != forecaststability.StabilityLevelIndeterminate {
			metrics.ComparableTransitionCount++
			scoreSum += item.Score
			squareSum += item.Score * item.Score
			if item.Score < metrics.MinimumStabilityScore {
				metrics.MinimumStabilityScore = item.Score
			}
			shiftSum += item.Metrics.MeanHorizontalShiftKilometers
			if item.Metrics.MaximumHorizontalShiftKilometers >
				metrics.MaximumHorizontalShiftKilometers {
				metrics.MaximumHorizontalShiftKilometers =
					item.Metrics.MaximumHorizontalShiftKilometers
			}
		}

		if item.Metrics.MethodChanged {
			metrics.MethodChangeCount++
		}
		if item.Metrics.PolicyChanged {
			metrics.PolicyChangeCount++
		}
		if item.Metrics.ImplementationChanged {
			metrics.ImplementationChangeCount++
		}
		if item.Metrics.InputChanged {
			metrics.InputChangeCount++
		}
		if item.Metrics.OutputChanged {
			metrics.OutputChangeCount++
		}
	}

	if metrics.TransitionCount > 0 {
		metrics.StableTransitionShare = float64(
			metrics.UnchangedCount+metrics.StableCount,
		) / float64(metrics.TransitionCount)
		metrics.ComparableTransitionShare = float64(
			metrics.ComparableTransitionCount,
		) / float64(metrics.TransitionCount)
		metrics.MaterialChangeShare = float64(
			metrics.MaterialChangeCount,
		) / float64(metrics.TransitionCount)
	}

	if metrics.ComparableTransitionCount > 0 {
		count := float64(metrics.ComparableTransitionCount)
		metrics.MeanStabilityScore = scoreSum / count
		variance := squareSum/count -
			metrics.MeanStabilityScore*metrics.MeanStabilityScore
		if variance < 0 {
			variance = 0
		}
		metrics.ScoreStandardDeviation = math.Sqrt(variance)
		metrics.MeanHorizontalShiftKilometers = shiftSum / count
	} else {
		metrics.MinimumStabilityScore = 0
	}

	return metrics
}

func classifyTrend(
	items []forecaststability.StabilityResult,
	metrics Metrics,
	policy Policy,
) Trend {
	if metrics.ComparableTransitionCount < policy.MinimumTrendTransitions {
		return TrendInsufficient
	}
	if metrics.MaterialChangeShare >=
		policy.MaterialChangeShareForUnstable ||
		metrics.ScoreStandardDeviation >=
			policy.VolatileScoreStandardDeviation {
		return TrendVolatile
	}

	scores := make([]float64, 0, len(items))
	for _, item := range items {
		if item.Level != forecaststability.StabilityLevelIndeterminate {
			scores = append(scores, item.Score)
		}
	}
	middle := len(scores) / 2
	if middle == 0 || middle == len(scores) {
		return TrendInsufficient
	}

	first := mean(scores[:middle])
	second := mean(scores[middle:])
	if second-first >= policy.TrendScoreDelta {
		return TrendImproving
	}
	if first-second >= policy.TrendScoreDelta {
		return TrendDegrading
	}
	return TrendSteady
}

func classifyHealth(metrics Metrics, policy Policy) Health {
	if metrics.ComparableTransitionShare < policy.MinimumComparableShare {
		return HealthInsufficient
	}
	if metrics.MaterialChangeShare >=
		policy.MaterialChangeShareForUnstable ||
		metrics.StableTransitionShare < policy.UnstableHealthShare {
		return HealthUnstable
	}
	if metrics.StableTransitionShare >= policy.StableHealthShare &&
		metrics.MeanStabilityScore >= policy.StableHealthShare {
		return HealthStable
	}
	return HealthWatch
}

func analysisConfidence(metrics Metrics) Confidence {
	score := metrics.ComparableTransitionShare
	level := "low"
	if score >= 0.80 {
		level = "high"
	} else if score >= 0.55 {
		level = "medium"
	}
	return Confidence{
		Score: score,
		Level: level,
		Reasons: []Reason{
			{
				Code:    "comparable_transition_share",
				Message: "Confidence reflects the share of version transitions that were structurally comparable.",
				Impact:  score,
			},
			{
				Code:    "history_length",
				Message: "Longer bounded histories provide more evidence about repeated decision behavior.",
				Impact: math.Min(
					1,
					float64(metrics.TransitionCount)/10,
				),
			},
		},
	}
}

func mean(values []float64) float64 {
	total := 0.0
	for _, value := range values {
		total += value
	}
	return total / float64(len(values))
}

func resultFingerprint(result Result) string {
	versionIDs := append([]string(nil), result.Provenance.VersionIDs...)
	outputs := append([]string(nil), result.Provenance.OutputFingerprints...)
	sort.Strings(versionIDs)
	sort.Strings(outputs)
	payload := struct {
		SchemaVersion string
		TrajectoryID  string
		Trend         Trend
		Health        Health
		Metrics       Metrics
		VersionIDs    []string
		Outputs       []string
		PolicyVersion string
		EvaluatedAt   time.Time
	}{
		SchemaVersion: result.SchemaVersion,
		TrajectoryID:  result.TrajectoryID,
		Trend:         result.Trend,
		Health:        result.Health,
		Metrics:       result.Metrics,
		VersionIDs:    versionIDs,
		Outputs:       outputs,
		PolicyVersion: result.Provenance.PolicyVersion,
		EvaluatedAt:   result.EvaluatedAt.UTC(),
	}
	encoded, _ := json.Marshal(payload)
	digest := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(digest[:])
}
