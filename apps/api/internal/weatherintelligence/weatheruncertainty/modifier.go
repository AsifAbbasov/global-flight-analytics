package weatheruncertainty

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weatherencounter"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weathertrust"
)

var (
	ErrInvalidPolicy     = errors.New("weather uncertainty policy is invalid")
	ErrProjectionInvalid = errors.New("weather uncertainty projection is invalid")
	ErrTrustInvalid      = errors.New("weather uncertainty trust result is invalid")
	ErrEncounterInvalid  = errors.New("weather uncertainty encounter profile is invalid")
	ErrInputMismatch     = errors.New("weather uncertainty inputs do not describe the same evidence boundary")
	ErrAlreadyAdjusted   = errors.New("projection already contains weather-adjusted uncertainty")
	ErrResultInvalid     = errors.New("weather uncertainty result is invalid")
)

const (
	weatherInputName  = "weather_encounter_profile"
	weatherReasonCode = "weather_adjusted_uncertainty"
)

type Request struct {
	Projection  projectioncontract.Result
	Trust       weathertrust.Result
	Encounter   weatherencounter.Result
	Policy      Policy
	GeneratedAt time.Time
}

func Apply(request Request) (Result, error) {
	if err := request.Policy.Validate(); err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrInvalidPolicy, err)
	}

	projectionReport := projectioncontract.Validate(request.Projection)
	if projectionReport.Status != projectioncontract.ValidationStatusValid {
		return Result{}, fmt.Errorf("%w: issues=%v", ErrProjectionInvalid, projectionReport.Issues)
	}
	if err := request.Trust.Validate(); err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrTrustInvalid, err)
	}
	if err := request.Encounter.Validate(); err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrEncounterInvalid, err)
	}
	if err := validateInputBoundary(request); err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrInputMismatch, err)
	}
	if projectionAlreadyAdjusted(request.Projection) {
		return Result{}, ErrAlreadyAdjusted
	}

	generatedAt := request.GeneratedAt.UTC()
	if generatedAt.IsZero() ||
		generatedAt.Before(request.Projection.Horizon.AsOfTime) ||
		generatedAt.Before(request.Projection.GeneratedAt) ||
		generatedAt.Before(request.Encounter.GeneratedAt) {
		return Result{}, fmt.Errorf("%w: generated-at time is invalid", ErrInputMismatch)
	}

	fingerprint := inputFingerprint(
		request.Projection,
		request.Trust,
		request.Encounter,
		request.Policy,
	)

	result := Result{
		Version:            Version,
		TrajectoryID:       strings.TrimSpace(request.Projection.TrajectoryID),
		AsOfTime:           request.Projection.Horizon.AsOfTime.UTC(),
		WeatherMultiplier:  1,
		Components:         request.Policy.components(0, 0, 0, 0, 0),
		AdjustedProjection: request.Projection.Clone(),
		Explanations: []Notice{
			{
				Code:    "weather_context_only",
				Message: "Weather context may widen uncertainty but does not prove pilot intent, controller intent, rerouting reason, or maneuver cause.",
			},
			{
				Code:    "project_derived_research_heuristic",
				Message: "Weather adjustment thresholds and weights are project-derived research heuristics and are not operational aviation limits.",
			},
		},
		InputFingerprint: fingerprint,
		GeneratedAt:      generatedAt,
	}

	if request.Projection.Status == projectioncontract.ResultStatusUnavailable {
		result.Status = StatusUnavailable
		result.Limitations = []Notice{
			{
				Code:    "projection_unavailable",
				Message: "Weather uncertainty cannot be applied because Projection Intelligence is unavailable.",
			},
		}
		return validateAndClone(result)
	}

	withheldCode, withheldMessage := withheldReason(request.Trust, request.Encounter)
	if withheldCode != "" {
		result.Status = StatusWithheld
		result.Limitations = []Notice{
			{
				Code:    withheldCode,
				Message: withheldMessage,
			},
		}
		return validateAndClone(result)
	}

	components := calculateComponents(request.Trust, request.Encounter, request.Policy)
	severity := weightedScore(components)
	multiplier := 1 + severity*(request.Policy.MaximumUncertaintyMultiplier-1)

	adjusted := request.Projection.Clone()
	pointAdjustments := make([]PointAdjustment, 0, len(adjusted.Points))

	for index := range adjusted.Points {
		originalPoint := request.Projection.Points[index]
		adjustedPoint := &adjusted.Points[index]

		progress := horizonProgress(adjusted.Horizon, adjustedPoint.ForecastTime)
		effectFraction := request.Policy.NearTermEffectFraction +
			(1-request.Policy.NearTermEffectFraction)*progress
		pointMultiplier := 1 + (multiplier-1)*effectFraction

		adjustment := PointAdjustment{
			Sequence:                  adjustedPoint.Sequence,
			ForecastTime:              adjustedPoint.ForecastTime.UTC(),
			HorizonProgress:           progress,
			Multiplier:                pointMultiplier,
			OriginalHorizontalRadiusM: originalPoint.Uncertainty.HorizontalRadiusM,
			OriginalVerticalRadiusM:   cloneFloat64(originalPoint.Uncertainty.VerticalRadiusM),
			OriginalConfidenceScore:   originalPoint.Confidence.Score,
		}

		adjustedPoint.Uncertainty.HorizontalRadiusM = scalePositive(
			originalPoint.Uncertainty.HorizontalRadiusM,
			pointMultiplier,
		)
		adjustment.AdjustedHorizontalRadiusM = adjustedPoint.Uncertainty.HorizontalRadiusM

		if originalPoint.Uncertainty.VerticalRadiusM != nil {
			vertical := scalePositive(*originalPoint.Uncertainty.VerticalRadiusM, pointMultiplier)
			adjustedPoint.Uncertainty.VerticalRadiusM = &vertical
			adjustment.AdjustedVerticalRadiusM = cloneFloat64(&vertical)
		}

		adjustedPoint.Confidence = adjustConfidence(
			originalPoint.Confidence,
			request.Policy.MaximumConfidenceReduction*severity*effectFraction,
			"Weather context widened the point uncertainty envelope.",
		)
		adjustment.AdjustedConfidenceScore = adjustedPoint.Confidence.Score
		pointAdjustments = append(pointAdjustments, adjustment)
	}

	adjusted.Confidence = adjustConfidence(
		request.Projection.Confidence,
		request.Policy.MaximumConfidenceReduction*severity,
		"Weather context widened the projection uncertainty envelope.",
	)

	var arrivalAdjustment *ArrivalAdjustment
	if request.Projection.Arrival != nil {
		arrivalAdjustment = adjustArrival(
			request.Projection.Arrival,
			adjusted.Arrival,
			adjusted.Horizon.AsOfTime,
			multiplier,
			request.Policy.MaximumConfidenceReduction*severity,
		)
	}

	adjusted.GeneratedAt = generatedAt
	adjusted.Provenance.InputFingerprint = fingerprint
	adjusted.Provenance.Inputs = appendWeatherInput(adjusted.Provenance.Inputs, request.Encounter)
	adjusted.Provenance.LatestInputObservedAt = latestObservedAt(adjusted.Provenance.Inputs)
	adjusted.Limitations = appendProjectionLimitation(
		adjusted.Limitations,
		projectioncontract.Limitation{
			Code:    weatherReasonCode,
			Message: "Projection uncertainty was preserved or increased using trusted contextual weather evidence.",
			Scope:   "projection_uncertainty",
		},
	)
	adjusted.Explanations = appendProjectionExplanation(
		adjusted.Explanations,
		projectioncontract.Explanation{
			Code:    weatherReasonCode,
			Message: "Weather context modified uncertainty only; projected coordinates were not changed.",
		},
	)

	adjustedReport := projectioncontract.Validate(adjusted)
	if adjustedReport.Status != projectioncontract.ValidationStatusValid {
		return Result{}, fmt.Errorf("%w: adjusted projection issues=%v", ErrResultInvalid, adjustedReport.Issues)
	}

	result.Status = StatusApplied
	if request.Trust.Decision == weathertrust.DecisionLimited ||
		request.Encounter.Status == weatherencounter.StatusLimited {
		result.Status = StatusAppliedLimited
		result.Limitations = append(
			result.Limitations,
			Notice{
				Code:    "weather_evidence_limited",
				Message: "Weather uncertainty was applied with limited upstream trust or encounter coverage.",
			},
		)
	}

	result.SeverityScore = severity
	result.WeatherMultiplier = multiplier
	result.Components = components
	result.PointAdjustments = pointAdjustments
	result.ArrivalAdjustment = arrivalAdjustment
	result.AdjustedProjection = adjusted
	result.Explanations = append(
		result.Explanations,
		Notice{
			Code:    weatherReasonCode,
			Message: "Existing uncertainty radii were multiplied by a horizon-aware weather factor; no radius was reduced.",
		},
	)

	return validateAndClone(result)
}

func validateInputBoundary(request Request) error {
	trajectoryID := strings.TrimSpace(request.Projection.TrajectoryID)
	if trajectoryID != strings.TrimSpace(request.Encounter.TrajectoryID) {
		return fmt.Errorf("projection and encounter trajectory identifiers differ")
	}

	asOfTime := request.Projection.Horizon.AsOfTime.UTC()
	if !request.Trust.AsOfTime.UTC().Equal(asOfTime) ||
		!request.Encounter.AsOfTime.UTC().Equal(asOfTime) {
		return fmt.Errorf("projection, trust, and encounter as-of times differ")
	}
	return nil
}

func withheldReason(
	trust weathertrust.Result,
	encounter weatherencounter.Result,
) (string, string) {
	if trust.Decision == weathertrust.DecisionBlocked || !trust.Usable {
		return "weather_trust_blocked",
			"Weather Trust Gate blocked analytical use, so projection uncertainty was preserved unchanged."
	}
	if !hasScope(trust.AllowedScopes, weathertrust.UsageScopeProjectionUncertainty) {
		return "projection_uncertainty_scope_withheld",
			"Weather Trust Gate did not authorize projection-uncertainty use, so projection values were preserved unchanged."
	}
	if encounter.Status == weatherencounter.StatusUnavailable || encounter.EncounterPointCount == 0 {
		return "weather_encounter_unavailable",
			"Weather Encounter Profile is unavailable, so projection uncertainty was preserved unchanged."
	}
	return "", ""
}

func calculateComponents(
	trust weathertrust.Result,
	encounter weatherencounter.Result,
	policy Policy,
) []Component {
	windSpeedScore := metricSeverity(
		encounter.WindSpeedMetersPerSecond,
		policy.WindSpeedReferenceMetersPerSecond,
		policy.WindSpeedHighMetersPerSecond,
	)
	windGustScore := metricSeverity(
		encounter.WindGustsMetersPerSecond,
		policy.WindGustReferenceMetersPerSecond,
		policy.WindGustHighMetersPerSecond,
	)
	precipitationScore := math.Max(
		metricSeverity(
			encounter.PrecipitationMillimeters,
			policy.PrecipitationReferenceMillimeters,
			policy.PrecipitationHighMillimeters,
		),
		metricSeverity(
			encounter.RainMillimeters,
			policy.PrecipitationReferenceMillimeters,
			policy.PrecipitationHighMillimeters,
		),
	)
	cloudCoverScore := metricSeverity(
		encounter.CloudCoverPercent,
		policy.CloudCoverReferencePercent,
		policy.CloudCoverHighPercent,
	)
	evidenceQualityScore := evidenceQualityPenalty(trust, encounter)

	return policy.components(
		windSpeedScore,
		windGustScore,
		precipitationScore,
		cloudCoverScore,
		evidenceQualityScore,
	)
}

func metricSeverity(
	summary weatherencounter.MetricSummary,
	reference float64,
	high float64,
) float64 {
	if summary.PresentCount == 0 {
		return 0
	}

	value := summary.Mean
	if summary.Maximum != nil {
		value = summary.Maximum
	}
	if value == nil {
		return 0
	}
	return normalizedThreshold(*value, reference, high)
}

func evidenceQualityPenalty(
	trust weathertrust.Result,
	encounter weatherencounter.Result,
) float64 {
	coreCoverage := (encounter.TemperatureCelsius.CoverageRatio +
		encounter.WindSpeedMetersPerSecond.CoverageRatio +
		encounter.WindDirectionDegrees.CoverageRatio) / 3

	penalty := ((1 - trust.Score) +
		(1 - encounter.ProfileCoverageRatio) +
		(1 - encounter.AlignmentCoverageRatio) +
		(1 - coreCoverage)) / 4

	if trust.Decision == weathertrust.DecisionLimited {
		penalty += 0.10
	}
	if encounter.Status == weatherencounter.StatusLimited {
		penalty += 0.10
	}
	return clampUnit(penalty)
}

func normalizedThreshold(value float64, reference float64, high float64) float64 {
	if !finite(value) || value <= reference {
		return 0
	}
	if value >= high {
		return 1
	}
	return clampUnit((value - reference) / (high - reference))
}

func horizonProgress(
	horizon projectioncontract.Horizon,
	forecastTime time.Time,
) float64 {
	duration := horizon.Duration()
	if duration <= 0 {
		return 1
	}
	elapsed := forecastTime.Sub(horizon.AsOfTime)
	return clampUnit(float64(elapsed) / float64(duration))
}

func scalePositive(value float64, multiplier float64) float64 {
	scaled := value * multiplier
	if !finite(scaled) || scaled < value {
		return value
	}
	return scaled
}

func adjustConfidence(
	confidence projectioncontract.Confidence,
	reduction float64,
	message string,
) projectioncontract.Confidence {
	adjusted := confidence
	reduction = clampUnit(reduction)
	adjusted.Score = clampUnit(confidence.Score * (1 - reduction))
	adjusted.Level = confidenceLevel(adjusted.Score)

	contribution := adjusted.Score - confidence.Score
	adjusted.Reasons = append(
		append([]projectioncontract.ConfidenceReason(nil), confidence.Reasons...),
		projectioncontract.ConfidenceReason{
			Code:         weatherReasonCode,
			Message:      message,
			Contribution: contribution,
		},
	)
	return adjusted
}

func confidenceLevel(score float64) projectioncontract.ConfidenceLevel {
	switch {
	case score <= 0:
		return projectioncontract.ConfidenceLevelNone
	case score < 0.55:
		return projectioncontract.ConfidenceLevelLow
	case score < 0.80:
		return projectioncontract.ConfidenceLevelMedium
	default:
		return projectioncontract.ConfidenceLevelHigh
	}
}

func adjustArrival(
	original *projectioncontract.ArrivalEstimate,
	adjusted *projectioncontract.ArrivalEstimate,
	asOfTime time.Time,
	multiplier float64,
	confidenceReduction float64,
) *ArrivalAdjustment {
	if original == nil || adjusted == nil {
		return nil
	}

	lowerDuration := original.EstimatedTime.Sub(original.EarliestTime)
	upperDuration := original.LatestTime.Sub(original.EstimatedTime)

	adjustedLower := scaleDuration(lowerDuration, multiplier)
	adjustedUpper := scaleDuration(upperDuration, multiplier)

	earliest := original.EstimatedTime.Add(-adjustedLower)
	if earliest.Before(asOfTime) {
		earliest = asOfTime
	}
	latest := original.EstimatedTime.Add(adjustedUpper)

	adjusted.EarliestTime = earliest
	adjusted.EstimatedTime = original.EstimatedTime
	adjusted.LatestTime = latest
	adjusted.Confidence = adjustConfidence(
		original.Confidence,
		confidenceReduction,
		"Weather context widened the arrival interval.",
	)
	adjusted.Limitations = appendProjectionLimitation(
		adjusted.Limitations,
		projectioncontract.Limitation{
			Code:    weatherReasonCode,
			Message: "Arrival interval was preserved or widened using contextual weather evidence.",
			Scope:   "arrival_interval",
		},
	)

	return &ArrivalAdjustment{
		Multiplier:              multiplier,
		OriginalEarliestTime:    original.EarliestTime.UTC(),
		OriginalEstimatedTime:   original.EstimatedTime.UTC(),
		OriginalLatestTime:      original.LatestTime.UTC(),
		AdjustedEarliestTime:    adjusted.EarliestTime.UTC(),
		AdjustedEstimatedTime:   adjusted.EstimatedTime.UTC(),
		AdjustedLatestTime:      adjusted.LatestTime.UTC(),
		OriginalConfidenceScore: original.Confidence.Score,
		AdjustedConfidenceScore: adjusted.Confidence.Score,
	}
}

func scaleDuration(value time.Duration, multiplier float64) time.Duration {
	if value <= 0 || multiplier <= 1 {
		return value
	}
	scaled := float64(value) * multiplier
	if math.IsNaN(scaled) || math.IsInf(scaled, 0) || scaled > float64(math.MaxInt64) {
		return value
	}
	return time.Duration(scaled)
}

func appendWeatherInput(
	inputs []projectioncontract.InputReference,
	encounter weatherencounter.Result,
) []projectioncontract.InputReference {
	result := make([]projectioncontract.InputReference, 0, len(inputs)+1)
	for _, input := range inputs {
		if strings.TrimSpace(input.Name) == weatherInputName {
			continue
		}
		result = append(result, input)
	}

	observedAt := encounter.AsOfTime.UTC()
	if encounter.EncounterEndedAt != nil {
		observedAt = encounter.EncounterEndedAt.UTC()
	}

	result = append(
		result,
		projectioncontract.InputReference{
			Name:           weatherInputName,
			Classification: projectioncontract.InputClassificationDerived,
			SourceName:     "weather_intelligence",
			ObservedAt:     observedAt,
			RetrievedAt:    encounter.GeneratedAt.UTC(),
			Limitation:     "Contextual weather evidence only; not proof of intent or maneuver cause.",
		},
	)
	return result
}

func latestObservedAt(inputs []projectioncontract.InputReference) time.Time {
	var latest time.Time
	for _, input := range inputs {
		if input.ObservedAt.IsZero() {
			continue
		}
		observedAt := input.ObservedAt.UTC()
		if latest.IsZero() || observedAt.After(latest) {
			latest = observedAt
		}
	}
	return latest
}

func projectionAlreadyAdjusted(projection projectioncontract.Result) bool {
	for _, explanation := range projection.Explanations {
		if strings.TrimSpace(explanation.Code) == weatherReasonCode {
			return true
		}
	}
	for _, input := range projection.Provenance.Inputs {
		if strings.TrimSpace(input.Name) == weatherInputName {
			return true
		}
	}
	return false
}

func appendProjectionLimitation(
	items []projectioncontract.Limitation,
	item projectioncontract.Limitation,
) []projectioncontract.Limitation {
	for _, existing := range items {
		if existing.Code == item.Code && existing.Scope == item.Scope {
			return append([]projectioncontract.Limitation(nil), items...)
		}
	}
	result := append([]projectioncontract.Limitation(nil), items...)
	return append(result, item)
}

func appendProjectionExplanation(
	items []projectioncontract.Explanation,
	item projectioncontract.Explanation,
) []projectioncontract.Explanation {
	for _, existing := range items {
		if existing.Code == item.Code {
			return append([]projectioncontract.Explanation(nil), items...)
		}
	}
	result := append([]projectioncontract.Explanation(nil), items...)
	return append(result, item)
}

func hasScope(scopes []weathertrust.UsageScope, target weathertrust.UsageScope) bool {
	for _, scope := range scopes {
		if scope == target {
			return true
		}
	}
	return false
}

func validateAndClone(result Result) (Result, error) {
	if err := result.Validate(); err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrResultInvalid, err)
	}
	return result.Clone(), nil
}
