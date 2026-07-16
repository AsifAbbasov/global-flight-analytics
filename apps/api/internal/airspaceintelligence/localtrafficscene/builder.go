package localtrafficscene

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/interactiongraph"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/interactionradius"
)

func Build(
	request Request,
	policy Policy,
	radiusPolicy interactionradius.Policy,
) (Result, error) {
	if err := policy.Validate(); err != nil {
		return Result{}, err
	}
	if err := radiusPolicy.Validate(); err != nil {
		return Result{}, err
	}
	if err := validateRequest(request, policy); err != nil {
		return Result{}, err
	}

	normalizedRequest := normalizeRequest(request)
	candidates := make([]ObservationInput, 0, len(normalizedRequest.Observations))
	excluded := make([]ExcludedObservation, 0)

	for _, observation := range normalizedRequest.Observations {
		switch {
		case observation.OnGround:
			excluded = append(excluded, exclusion(
				observation,
				ExclusionReasonOnGround,
				"Ground observations are outside the airborne local traffic scene.",
			))
		case observation.ObservedAt.After(normalizedRequest.AsOfTime):
			excluded = append(excluded, exclusion(
				observation,
				ExclusionReasonFutureEvidence,
				"The observation occurs after the requested as-of time.",
			))
		case !contains(normalizedRequest.RegionBounds, observation.Latitude, observation.Longitude):
			excluded = append(excluded, exclusion(
				observation,
				ExclusionReasonOutsideRegion,
				"The observation is outside the requested region bounds.",
			))
		default:
			candidates = append(candidates, observation)
		}
	}

	candidates, duplicateExclusions := deduplicateCandidates(candidates)
	excluded = append(excluded, duplicateExclusions...)

	aircraft := make([]Aircraft, 0, len(candidates))
	for _, observation := range candidates {
		decision, err := interactionradius.Evaluate(
			radiusRequest(normalizedRequest, observation),
			radiusPolicy,
		)
		if err != nil {
			return Result{}, fmt.Errorf(
				"%w: radius decision for node %q: %v",
				ErrInvalidRequest,
				observation.ID,
				err,
			)
		}
		if decision.Status == interactionradius.DecisionStatusBlocked {
			excluded = append(excluded, exclusion(
				observation,
				ExclusionReasonRadiusPolicyBlocked,
				"The Interaction Radius Policy blocked this observation from pairwise search.",
			))
			continue
		}
		aircraft = append(aircraft, aircraftFrom(
			observation,
			decision,
			normalizedRequest.AsOfTime,
		))
	}

	sort.Slice(aircraft, func(left int, right int) bool {
		return aircraft[left].NodeID < aircraft[right].NodeID
	})
	sortExcluded(excluded)

	metrics := buildMetrics(
		len(normalizedRequest.Observations),
		len(candidates),
		aircraft,
		excluded,
	)
	result := Result{
		SchemaVersion:        SchemaVersionV1,
		Status:               statusFor(metrics, policy),
		RegionCode:           normalizedRequest.RegionCode,
		RegionBounds:         normalizedRequest.RegionBounds,
		AsOfTime:             normalizedRequest.AsOfTime,
		Aircraft:             aircraft,
		ExcludedObservations: excluded,
		Metrics:              metrics,
		Confidence:           buildConfidence(aircraft, metrics, policy),
		Limitations:          buildLimitations(aircraft, metrics),
		Explanations:         buildExplanations(metrics),
		ScopeGuard:           ScopeGuardResearchOnly,
		Provenance:           buildProvenance(aircraft),
		GeneratedAt:          normalizedRequest.GeneratedAt,
	}
	result.Provenance.InputFingerprint = inputFingerprint(
		normalizedRequest,
		policy,
		radiusPolicy,
		result,
	)

	report := Validate(result, policy)
	if report.Status != ValidationStatusValid {
		return Result{}, fmt.Errorf(
			"%w: issues=%v",
			ErrInvalidScene,
			report.Issues,
		)
	}
	return result.Clone(), nil
}

func normalizeRequest(request Request) Request {
	normalized := request
	normalized.RegionCode = strings.ToUpper(strings.TrimSpace(request.RegionCode))
	normalized.AsOfTime = request.AsOfTime.UTC()
	normalized.GeneratedAt = request.GeneratedAt.UTC()
	normalized.Observations = make([]ObservationInput, 0, len(request.Observations))
	for _, observation := range request.Observations {
		normalized.Observations = append(
			normalized.Observations,
			normalizeObservation(observation),
		)
	}
	return normalized
}

func normalizeObservation(observation ObservationInput) ObservationInput {
	normalized := observation
	normalized.TrajectoryID = strings.TrimSpace(observation.TrajectoryID)
	normalized.FlightID = strings.TrimSpace(observation.FlightID)
	normalized.AircraftID = strings.TrimSpace(observation.AircraftID)
	normalized.ICAO24 = strings.ToUpper(strings.TrimSpace(observation.ICAO24))
	normalized.Callsign = strings.ToUpper(strings.TrimSpace(observation.Callsign))
	normalized.SourceName = strings.TrimSpace(observation.SourceName)
	normalized.ObservedAt = observation.ObservedAt.UTC()
	normalized.ID = canonicalNodeID(observation)
	if normalized.AltitudeReference == "" {
		normalized.AltitudeReference = interactiongraph.AltitudeReferenceUnknown
	}
	normalized.AltitudeMeters = cloneFloat64(observation.AltitudeMeters)
	return normalized
}

func canonicalNodeID(observation ObservationInput) string {
	if value := strings.TrimSpace(observation.ID); value != "" {
		return value
	}
	if value := strings.TrimSpace(observation.TrajectoryID); value != "" {
		return "trajectory:" + value
	}
	if value := strings.ToUpper(strings.TrimSpace(observation.ICAO24)); value != "" {
		return "icao24:" + value
	}
	return ""
}

func deduplicateCandidates(
	candidates []ObservationInput,
) ([]ObservationInput, []ExcludedObservation) {
	sort.Slice(candidates, func(left int, right int) bool {
		if candidates[left].ID != candidates[right].ID {
			return candidates[left].ID < candidates[right].ID
		}
		if !candidates[left].ObservedAt.Equal(candidates[right].ObservedAt) {
			return candidates[left].ObservedAt.After(candidates[right].ObservedAt)
		}
		if candidates[left].QualityScore != candidates[right].QualityScore {
			return candidates[left].QualityScore > candidates[right].QualityScore
		}
		return candidates[left].SourceName < candidates[right].SourceName
	})

	selected := make([]ObservationInput, 0, len(candidates))
	excluded := make([]ExcludedObservation, 0)
	seen := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		if _, exists := seen[candidate.ID]; exists {
			excluded = append(excluded, exclusion(
				candidate,
				ExclusionReasonSupersededDuplicate,
				"A newer or higher-quality observation was selected for the same scene node.",
			))
			continue
		}
		seen[candidate.ID] = struct{}{}
		selected = append(selected, candidate)
	}
	return selected, excluded
}

func radiusRequest(request Request, observation ObservationInput) interactionradius.Request {
	return interactionradius.Request{
		RegionCode:                  request.RegionCode,
		NodeID:                      observation.ID,
		ICAO24:                      observation.ICAO24,
		Callsign:                    observation.Callsign,
		VelocityMetersPerSecond:     observation.VelocityMetersPerSecond,
		VerticalRateMetersPerSecond: observation.VerticalRateMetersPerSecond,
		AltitudeMeters:              cloneFloat64(observation.AltitudeMeters),
		AltitudeReference:           radiusAltitudeReference(observation.AltitudeReference),
		OnGround:                    observation.OnGround,
		ObservedAt:                  observation.ObservedAt,
		AsOfTime:                    request.AsOfTime,
		GeneratedAt:                 request.GeneratedAt,
		SourceName:                  observation.SourceName,
		QualityScore:                observation.QualityScore,
	}
}

func radiusAltitudeReference(
	reference interactiongraph.AltitudeReference,
) interactionradius.AltitudeReference {
	switch reference {
	case interactiongraph.AltitudeReferenceBarometric:
		return interactionradius.AltitudeReferenceBarometric
	case interactiongraph.AltitudeReferenceGeometric:
		return interactionradius.AltitudeReferenceGeometric
	default:
		return interactionradius.AltitudeReferenceUnknown
	}
}

func aircraftFrom(
	observation ObservationInput,
	decision interactionradius.Decision,
	asOfTime time.Time,
) Aircraft {
	return Aircraft{
		NodeID:                      observation.ID,
		TrajectoryID:                observation.TrajectoryID,
		FlightID:                    observation.FlightID,
		AircraftID:                  observation.AircraftID,
		ICAO24:                      observation.ICAO24,
		Callsign:                    observation.Callsign,
		Latitude:                    observation.Latitude,
		Longitude:                   observation.Longitude,
		AltitudeMeters:              cloneFloat64(observation.AltitudeMeters),
		AltitudeReference:           observation.AltitudeReference,
		VelocityMetersPerSecond:     observation.VelocityMetersPerSecond,
		HeadingDegrees:              observation.HeadingDegrees,
		VerticalRateMetersPerSecond: observation.VerticalRateMetersPerSecond,
		ObservedAt:                  observation.ObservedAt,
		ObservationAge:              asOfTime.Sub(observation.ObservedAt),
		SourceName:                  observation.SourceName,
		QualityScore:                observation.QualityScore,
		RadiusDecision:              decision.Clone(),
	}
}

func exclusion(
	observation ObservationInput,
	reason ExclusionReason,
	message string,
) ExcludedObservation {
	return ExcludedObservation{
		NodeID:     observation.ID,
		ICAO24:     observation.ICAO24,
		Callsign:   observation.Callsign,
		ObservedAt: observation.ObservedAt,
		Reason:     reason,
		Message:    message,
		SourceName: observation.SourceName,
	}
}

func sortExcluded(excluded []ExcludedObservation) {
	sort.Slice(excluded, func(left int, right int) bool {
		if excluded[left].NodeID != excluded[right].NodeID {
			return excluded[left].NodeID < excluded[right].NodeID
		}
		if excluded[left].Reason != excluded[right].Reason {
			return excluded[left].Reason < excluded[right].Reason
		}
		if !excluded[left].ObservedAt.Equal(excluded[right].ObservedAt) {
			return excluded[left].ObservedAt.Before(excluded[right].ObservedAt)
		}
		return excluded[left].SourceName < excluded[right].SourceName
	})
}

func buildMetrics(
	inputCount int,
	candidateCount int,
	aircraft []Aircraft,
	excluded []ExcludedObservation,
) SceneMetrics {
	metrics := SceneMetrics{
		InputObservationCount:     inputCount,
		CandidateObservationCount: candidateCount,
		IncludedAircraftCount:     len(aircraft),
		ExcludedObservationCount:  len(excluded),
	}
	for _, item := range aircraft {
		switch item.RadiusDecision.Status {
		case interactionradius.DecisionStatusAllowed:
			metrics.AllowedAircraftCount++
		case interactionradius.DecisionStatusLimited:
			metrics.LimitedAircraftCount++
		}
	}
	for _, item := range excluded {
		switch item.Reason {
		case ExclusionReasonOnGround:
			metrics.GroundExcludedCount++
		case ExclusionReasonOutsideRegion:
			metrics.OutsideRegionExcludedCount++
		case ExclusionReasonFutureEvidence:
			metrics.FutureEvidenceExcludedCount++
			metrics.MaterialEvidenceRejectedCount++
		case ExclusionReasonSupersededDuplicate:
			metrics.DuplicateExcludedCount++
		case ExclusionReasonRadiusPolicyBlocked:
			metrics.RadiusPolicyBlockedCount++
			metrics.MaterialEvidenceRejectedCount++
		}
	}
	eligibleCount := inputCount - metrics.GroundExcludedCount - metrics.OutsideRegionExcludedCount
	if eligibleCount > 0 {
		metrics.SceneCoverage = float64(metrics.IncludedAircraftCount) / float64(eligibleCount)
		metrics.SceneCoverage = math.Min(math.Max(metrics.SceneCoverage, 0), 1)
	}
	return metrics
}

func statusFor(metrics SceneMetrics, policy Policy) ResultStatus {
	switch {
	case metrics.IncludedAircraftCount == 0:
		return ResultStatusUnavailable
	case metrics.IncludedAircraftCount < policy.MinimumCompleteAircraftCount,
		metrics.LimitedAircraftCount > 0,
		metrics.MaterialEvidenceRejectedCount > 0:
		return ResultStatusLimited
	default:
		return ResultStatusComplete
	}
}

func buildConfidence(
	aircraft []Aircraft,
	metrics SceneMetrics,
	policy Policy,
) Confidence {
	if len(aircraft) == 0 {
		return Confidence{
			Score: 0,
			Level: ConfidenceLevelNone,
			Reasons: []ConfidenceReason{
				{
					Code:         "scene_aircraft_unavailable",
					Message:      "No aircraft passed the local traffic scene evidence gates.",
					Contribution: 0,
				},
			},
		}
	}
	decisionTotal := 0.0
	for _, item := range aircraft {
		decisionTotal += item.RadiusDecision.Confidence.Score
	}
	decisionMean := decisionTotal / float64(len(aircraft))
	score := decisionMean*policy.ConfidenceWeights.RadiusDecisionConfidence +
		metrics.SceneCoverage*policy.ConfidenceWeights.SceneCoverage
	return Confidence{
		Score: score,
		Level: confidenceLevel(score, policy),
		Reasons: []ConfidenceReason{
			{
				Code:         "mean_radius_decision_confidence",
				Message:      "Scene confidence includes the mean confidence of included Interaction Radius decisions.",
				Contribution: decisionMean,
			},
			{
				Code:         "scene_coverage",
				Message:      "Scene confidence includes the share of eligible observations represented by included aircraft.",
				Contribution: metrics.SceneCoverage,
			},
		},
	}
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

func buildLimitations(aircraft []Aircraft, metrics SceneMetrics) []Limitation {
	limitations := []Limitation{
		{
			Code:    "research_only_not_operational_separation",
			Message: "The local traffic scene is research context and must not be used for operational separation or collision avoidance.",
			Scope:   "operational_use",
		},
	}
	if len(aircraft) == 0 {
		limitations = append(limitations, Limitation{
			Code:    "scene_aircraft_unavailable",
			Message: "No airborne observation passed the region, time, quality, and radius-policy gates.",
			Scope:   "scene_coverage",
		})
	}
	if len(aircraft) == 1 {
		limitations = append(limitations, Limitation{
			Code:    "single_aircraft_scene",
			Message: "A single-aircraft scene cannot produce pairwise interaction evidence.",
			Scope:   "pairwise_context",
		})
	}
	if metrics.LimitedAircraftCount > 0 {
		limitations = append(limitations, Limitation{
			Code:    "limited_radius_decisions_present",
			Message: "At least one aircraft is included with a limited Interaction Radius decision.",
			Scope:   "candidate_search",
		})
	}
	if metrics.FutureEvidenceExcludedCount > 0 {
		limitations = append(limitations, Limitation{
			Code:    "future_evidence_excluded",
			Message: "One or more observations after the as-of time were excluded.",
			Scope:   "temporal_evidence",
		})
	}
	if metrics.RadiusPolicyBlockedCount > 0 {
		limitations = append(limitations, Limitation{
			Code:    "radius_policy_blocked_observations",
			Message: "One or more observations were excluded because their freshness or quality did not support pairwise search.",
			Scope:   "candidate_search",
		})
	}
	return limitations
}

func buildExplanations(metrics SceneMetrics) []Explanation {
	return []Explanation{
		{
			Code:    "bounded_as_of_scene",
			Message: "The scene contains one deterministic prepared observation per aircraft identity at or before the requested as-of time.",
		},
		{
			Code:    "region_and_airborne_filtering",
			Message: "Ground observations and observations outside the requested region bounds are excluded from the airborne scene.",
		},
		{
			Code:    "radius_policy_preparation",
			Message: fmt.Sprintf("The scene includes %d aircraft prepared with allowed or limited Interaction Radius decisions.", metrics.IncludedAircraftCount),
		},
		{
			Code:    "scene_is_not_risk_classification",
			Message: "Scene membership prepares evidence for later pairwise scanning and does not classify separation risk.",
		},
	}
}

func buildProvenance(aircraft []Aircraft) Provenance {
	sourceSet := make(map[string]struct{})
	latestObservedAt := time.Time{}
	for _, item := range aircraft {
		sourceSet[item.SourceName] = struct{}{}
		if item.ObservedAt.After(latestObservedAt) {
			latestObservedAt = item.ObservedAt
		}
	}
	sources := make([]string, 0, len(sourceSet))
	for source := range sourceSet {
		sources = append(sources, source)
	}
	sort.Strings(sources)
	return Provenance{
		SourceNames:      sources,
		LatestObservedAt: latestObservedAt,
	}
}

func contains(bounds Bounds, latitude float64, longitude float64) bool {
	return latitude >= bounds.MinimumLatitude &&
		latitude <= bounds.MaximumLatitude &&
		longitude >= bounds.MinimumLongitude &&
		longitude <= bounds.MaximumLongitude
}
