package interactionradius

import (
	"fmt"
	"math"
	"strings"
	"time"
)

func Evaluate(request Request, policy Policy) (Decision, error) {
	if err := policy.Validate(); err != nil {
		return Decision{}, err
	}
	if err := validateRequest(request); err != nil {
		return Decision{}, err
	}

	asOfTime := request.AsOfTime.UTC()
	generatedAt := request.GeneratedAt.UTC()
	observedAt := request.ObservedAt.UTC()
	age := asOfTime.Sub(observedAt)
	motionClass := classifyMotion(request, policy)

	freshnessScore := clampUnit(
		1 - float64(age)/float64(policy.MaximumObservationAge),
	)
	motionScore := motionPlausibilityScore(request, policy)
	verticalScore := 1.0
	verticalFilteringPermitted := request.AltitudeMeters != nil &&
		request.AltitudeReference != AltitudeReferenceUnknown
	if !verticalFilteringPermitted {
		verticalScore = 0
	}

	components := []Component{
		{Name: "data_quality", Score: request.QualityScore, Weight: policy.Weights.DataQuality},
		{Name: "temporal_freshness", Score: freshnessScore, Weight: policy.Weights.TemporalFreshness},
		{Name: "motion_plausibility", Score: motionScore, Weight: policy.Weights.MotionPlausibility},
		{Name: "vertical_evidence", Score: verticalScore, Weight: policy.Weights.VerticalEvidence},
	}
	confidenceScore := weightedScore(components)
	status := decisionStatus(request.QualityScore, age, verticalFilteringPermitted, policy)

	horizontalRadius := horizontalRadiusKilometers(request, policy)
	verticalRadius := verticalRadiusMeters(request, verticalFilteringPermitted, policy)
	if status == DecisionStatusBlocked {
		horizontalRadius = 0
		verticalRadius = 0
	}

	decision := Decision{
		SchemaVersion: SchemaVersionV1,
		Status:        status,
		RegionCode: strings.ToUpper(
			strings.TrimSpace(request.RegionCode),
		),
		NodeID:                     strings.TrimSpace(request.NodeID),
		ICAO24:                     strings.ToUpper(strings.TrimSpace(request.ICAO24)),
		Callsign:                   strings.ToUpper(strings.TrimSpace(request.Callsign)),
		MotionClass:                motionClass,
		HorizontalRadiusKilometers: horizontalRadius,
		VerticalRadiusMeters:       verticalRadius,
		MaximumObservationAge:      policy.MaximumObservationAge,
		MaximumPairTimeDifference:  policy.MaximumPairTimeDifference,
		LookaheadDuration:          policy.HorizontalLookaheadDuration,
		VerticalFilteringPermitted: verticalFilteringPermitted,
		Components:                 components,
		Confidence: Confidence{
			Score: confidenceScore,
			Level: confidenceLevel(confidenceScore, policy),
			Reasons: confidenceReasons(
				request.QualityScore,
				freshnessScore,
				motionScore,
				verticalScore,
			),
		},
		Limitations: limitationsFor(
			status,
			age,
			verticalFilteringPermitted,
			request.QualityScore,
			policy,
		),
		Explanations: explanationsFor(status, motionClass),
		ScopeGuard:   ScopeGuardResearchOnly,
		Provenance: Provenance{
			SourceNames: []string{strings.TrimSpace(request.SourceName)},
			ObservedAt:  observedAt,
		},
		AsOfTime:    asOfTime,
		GeneratedAt: generatedAt,
	}
	decision.Provenance.InputFingerprint = inputFingerprint(request, policy)

	report := Validate(decision)
	if report.Status != ValidationStatusValid {
		return Decision{}, fmt.Errorf(
			"%w: issues=%v",
			ErrInvalidDecision,
			report.Issues,
		)
	}
	return decision.Clone(), nil
}

func validateRequest(request Request) error {
	if strings.TrimSpace(request.RegionCode) == "" ||
		strings.TrimSpace(request.NodeID) == "" ||
		strings.TrimSpace(request.SourceName) == "" ||
		request.ObservedAt.IsZero() ||
		request.AsOfTime.IsZero() ||
		request.GeneratedAt.IsZero() {
		return fmt.Errorf("%w: identity, source, and times are required", ErrInvalidRequest)
	}
	if request.OnGround {
		return fmt.Errorf("%w: airborne evidence is required", ErrInvalidRequest)
	}
	if request.ObservedAt.After(request.AsOfTime) ||
		request.GeneratedAt.Before(request.AsOfTime) {
		return fmt.Errorf("%w: future evidence or invalid generation time", ErrInvalidRequest)
	}
	if !unitInterval(request.QualityScore) ||
		!nonNegativeFinite(request.VelocityMetersPerSecond) ||
		!finite(request.VerticalRateMetersPerSecond) ||
		!request.AltitudeReference.IsKnown() {
		return fmt.Errorf("%w: quality, motion, or altitude reference", ErrInvalidRequest)
	}
	if request.AltitudeMeters != nil && !finite(*request.AltitudeMeters) {
		return fmt.Errorf("%w: altitude", ErrInvalidRequest)
	}
	return nil
}

func classifyMotion(request Request, policy Policy) MotionClass {
	if math.Abs(request.VerticalRateMetersPerSecond) >=
		policy.VerticalChangeMinimumMetersPerSecond {
		return MotionClassVerticalChange
	}
	switch {
	case request.VelocityMetersPerSecond <= policy.LowSpeedMaximumMetersPerSecond:
		return MotionClassLowSpeed
	case request.VelocityMetersPerSecond >= policy.HighSpeedMinimumMetersPerSecond:
		return MotionClassHighSpeed
	default:
		return MotionClassTransit
	}
}

func horizontalRadiusKilometers(request Request, policy Policy) float64 {
	motionDistance := request.VelocityMetersPerSecond *
		policy.HorizontalLookaheadDuration.Seconds() / 1000
	qualityMultiplier := 1 +
		(1-request.QualityScore)*policy.QualityUncertaintyFraction
	return clamp(
		(policy.BaseHorizontalRadiusKilometers+motionDistance)*qualityMultiplier,
		policy.MinimumHorizontalRadiusKilometers,
		policy.MaximumHorizontalRadiusKilometers,
	)
}

func verticalRadiusMeters(
	request Request,
	verticalFilteringPermitted bool,
	policy Policy,
) float64 {
	if !verticalFilteringPermitted {
		return policy.MaximumVerticalRadiusMeters
	}
	motionDistance := math.Abs(request.VerticalRateMetersPerSecond) *
		policy.VerticalLookaheadDuration.Seconds()
	qualityMultiplier := 1 +
		(1-request.QualityScore)*policy.QualityUncertaintyFraction
	return clamp(
		(policy.BaseVerticalRadiusMeters+motionDistance)*qualityMultiplier,
		policy.MinimumVerticalRadiusMeters,
		policy.MaximumVerticalRadiusMeters,
	)
}

func decisionStatus(
	qualityScore float64,
	age time.Duration,
	verticalFilteringPermitted bool,
	policy Policy,
) DecisionStatus {
	switch {
	case age > policy.MaximumObservationAge,
		qualityScore < policy.MinimumUsableQuality:
		return DecisionStatusBlocked
	case qualityScore < policy.MinimumAllowedQuality,
		!verticalFilteringPermitted:
		return DecisionStatusLimited
	default:
		return DecisionStatusAllowed
	}
}

func motionPlausibilityScore(request Request, policy Policy) float64 {
	if request.VelocityMetersPerSecond > policy.HighSpeedMinimumMetersPerSecond*2 {
		return 0.50
	}
	if math.Abs(request.VerticalRateMetersPerSecond) >
		policy.VerticalChangeMinimumMetersPerSecond*5 {
		return 0.50
	}
	return 1
}

func confidenceReasons(
	quality float64,
	freshness float64,
	motion float64,
	vertical float64,
) []ConfidenceReason {
	return []ConfidenceReason{
		{Code: "data_quality", Message: "Prepared node quality contributes to the radius decision.", Contribution: quality},
		{Code: "temporal_freshness", Message: "Observation freshness contributes to the radius decision.", Contribution: freshness},
		{Code: "motion_plausibility", Message: "Motion plausibility contributes to the radius decision.", Contribution: motion},
		{Code: "vertical_evidence", Message: "Comparable vertical evidence contributes to the radius decision.", Contribution: vertical},
	}
}

func limitationsFor(
	status DecisionStatus,
	age time.Duration,
	verticalFilteringPermitted bool,
	qualityScore float64,
	policy Policy,
) []Limitation {
	limitations := []Limitation{
		{
			Code:    "research_only_not_operational_separation",
			Message: "The radius is a research prefilter and must not be used as operational separation or collision-avoidance logic.",
			Scope:   "operational_use",
		},
	}
	if age > policy.MaximumObservationAge {
		limitations = append(limitations, Limitation{
			Code:    "observation_too_old",
			Message: "The aircraft observation is older than the policy boundary.",
			Scope:   "temporal_evidence",
		})
	}
	if qualityScore < policy.MinimumUsableQuality {
		limitations = append(limitations, Limitation{
			Code:    "quality_below_usable_threshold",
			Message: "Prepared node quality is below the usable policy threshold.",
			Scope:   "data_quality",
		})
	} else if qualityScore < policy.MinimumAllowedQuality {
		limitations = append(limitations, Limitation{
			Code:    "quality_supports_limited_search_only",
			Message: "Prepared node quality permits only a limited contextual search.",
			Scope:   "data_quality",
		})
	}
	if !verticalFilteringPermitted {
		limitations = append(limitations, Limitation{
			Code:    "vertical_filtering_withheld",
			Message: "Comparable altitude evidence is unavailable, so vertical filtering is withheld and the broad maximum vertical radius is published.",
			Scope:   "vertical_evidence",
		})
	}
	if status == DecisionStatusBlocked {
		limitations = append(limitations, Limitation{
			Code:    "interaction_search_blocked",
			Message: "The policy blocks pairwise interaction search for this node.",
			Scope:   "interaction_search",
		})
	}
	return limitations
}

func explanationsFor(status DecisionStatus, motionClass MotionClass) []Explanation {
	return []Explanation{
		{
			Code:    "motion_scaled_horizontal_radius",
			Message: "The horizontal search radius combines a base radius, bounded motion lookahead, and a data-quality uncertainty margin.",
		},
		{
			Code:    "vertical_radius_is_context_filter",
			Message: "The vertical radius is a contextual candidate filter and is not a regulatory separation minimum.",
		},
		{
			Code:    "decision_status",
			Message: fmt.Sprintf("The interaction search decision is %s for motion class %s.", status, motionClass),
		},
	}
}

func weightedScore(components []Component) float64 {
	total := 0.0
	for _, component := range components {
		total += component.Score * component.Weight
	}
	return clampUnit(total)
}

func confidenceLevel(score float64, policy Policy) ConfidenceLevel {
	switch {
	case score <= 0:
		return ConfidenceLevelNone
	case score < policy.MediumConfidenceMinimumScore:
		return ConfidenceLevelLow
	case score < policy.HighConfidenceMinimumScore:
		return ConfidenceLevelMedium
	default:
		return ConfidenceLevelHigh
	}
}

func clamp(value float64, minimum float64, maximum float64) float64 {
	return math.Min(math.Max(value, minimum), maximum)
}

func clampUnit(value float64) float64 {
	return clamp(value, 0, 1)
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func positiveFinite(value float64) bool {
	return finite(value) && value > 0
}

func nonNegativeFinite(value float64) bool {
	return finite(value) && value >= 0
}

func unitInterval(value float64) bool {
	return finite(value) && value >= 0 && value <= 1
}
