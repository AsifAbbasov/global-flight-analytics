package proximityscanner

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/interactiongraph"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/interactionradius"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/localtrafficscene"
)

const scannerSourceName = "multi_aircraft_proximity_scanner"

func Scan(request Request, policy Policy) (Result, error) {
	if err := policy.Validate(); err != nil {
		return Result{}, err
	}
	if err := validateRequest(request, policy); err != nil {
		return Result{}, err
	}

	scene := request.Scene.Clone()
	aircraft := append([]localtrafficscene.Aircraft(nil), scene.Aircraft...)
	sort.Slice(aircraft, func(left int, right int) bool {
		return aircraft[left].NodeID < aircraft[right].NodeID
	})

	metrics := Metrics{
		AircraftCount:     len(aircraft),
		PossiblePairCount: possiblePairCount(len(aircraft)),
	}
	candidates := make([]Candidate, 0)
	edges := make([]interactiongraph.EdgeInput, 0)

	for leftIndex := 0; leftIndex < len(aircraft); leftIndex++ {
		for rightIndex := leftIndex + 1; rightIndex < len(aircraft); rightIndex++ {
			metrics.EvaluatedPairCount++
			candidate, rejection := evaluatePair(
				aircraft[leftIndex],
				aircraft[rightIndex],
				request.GeneratedAt.UTC(),
				policy,
			)
			switch rejection {
			case pairRejectionTemporal:
				metrics.TemporalRejectedPairCount++
				continue
			case pairRejectionHorizontal:
				metrics.HorizontalRejectedPairCount++
				continue
			case pairRejectionVertical:
				metrics.VerticalRejectedPairCount++
				continue
			}
			if !candidate.VerticalFilteringApplied {
				metrics.VerticalFilteringWithheldPairCount++
			}
			if candidate.Status == CandidateStatusComplete {
				metrics.CompleteCandidateCount++
			} else {
				metrics.LimitedCandidateCount++
			}
			candidates = append(candidates, candidate)
			edges = append(edges, edgeInput(candidate))
		}
	}

	metrics.CandidatePairCount = len(candidates)
	if metrics.PossiblePairCount > 0 {
		metrics.CandidateShare = float64(metrics.CandidatePairCount) /
			float64(metrics.PossiblePairCount)
	}

	graph, err := interactiongraph.Build(interactiongraph.Request{
		RegionCode:  scene.RegionCode,
		AsOfTime:    scene.AsOfTime,
		GeneratedAt: request.GeneratedAt.UTC(),
		Nodes:       scene.GraphNodeInputs(),
		Edges:       edges,
	})
	if err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrGraphBuild, err)
	}

	result := Result{
		SchemaVersion: SchemaVersionV1,
		Status:        resultStatus(scene, candidates),
		RegionCode:    scene.RegionCode,
		SceneStatus:   scene.Status,
		AsOfTime:      scene.AsOfTime.UTC(),
		Candidates:    candidates,
		Graph:         graph,
		Metrics:       metrics,
		Confidence:    buildResultConfidence(scene, aircraft, metrics, policy),
		Limitations:   buildResultLimitations(scene, metrics),
		Explanations:  buildResultExplanations(metrics),
		ScopeGuard:    ScopeGuardResearchOnly,
		Provenance: Provenance{
			SceneFingerprint: scene.Provenance.InputFingerprint,
			SourceNames:      append([]string(nil), scene.Provenance.SourceNames...),
			LatestObservedAt: scene.Provenance.LatestObservedAt,
		},
		GeneratedAt: request.GeneratedAt.UTC(),
	}
	result.Provenance.InputFingerprint = inputFingerprint(result, policy)

	if report := Validate(result, policy); report.Status != ValidationStatusValid {
		return Result{}, fmt.Errorf("%w: issues=%v", ErrInvalidResult, report.Issues)
	}
	return result.Clone(), nil
}

type pairRejection string

const (
	pairRejectionNone       pairRejection = ""
	pairRejectionTemporal   pairRejection = "temporal"
	pairRejectionHorizontal pairRejection = "horizontal"
	pairRejectionVertical   pairRejection = "vertical"
)

func evaluatePair(
	left localtrafficscene.Aircraft,
	right localtrafficscene.Aircraft,
	evaluatedAt time.Time,
	policy Policy,
) (Candidate, pairRejection) {
	timeDifference := absoluteDuration(left.ObservedAt.Sub(right.ObservedAt))
	maximumTimeDifference := minimumDuration(
		left.RadiusDecision.MaximumPairTimeDifference,
		right.RadiusDecision.MaximumPairTimeDifference,
	)
	if timeDifference > maximumTimeDifference {
		return Candidate{}, pairRejectionTemporal
	}

	horizontalDistance := horizontalDistanceKilometers(
		left.Latitude,
		left.Longitude,
		right.Latitude,
		right.Longitude,
	)
	effectiveHorizontalRadius := math.Max(
		left.RadiusDecision.HorizontalRadiusKilometers,
		right.RadiusDecision.HorizontalRadiusKilometers,
	)
	if horizontalDistance > effectiveHorizontalRadius {
		return Candidate{}, pairRejectionHorizontal
	}

	verticalFilteringApplied := comparableVerticalEvidence(left, right)
	var verticalSeparation *float64
	var effectiveVerticalRadius *float64
	if verticalFilteringApplied {
		separation := math.Abs(*left.AltitudeMeters - *right.AltitudeMeters)
		radius := math.Max(
			left.RadiusDecision.VerticalRadiusMeters,
			right.RadiusDecision.VerticalRadiusMeters,
		)
		verticalSeparation = &separation
		effectiveVerticalRadius = &radius
		if separation > radius {
			return Candidate{}, pairRejectionVertical
		}
	}

	closingRate := closingRateMetersPerSecond(
		aircraftVector{
			latitude:             left.Latitude,
			longitude:            left.Longitude,
			speedMetersPerSecond: left.VelocityMetersPerSecond,
			headingDegrees:       left.HeadingDegrees,
		},
		aircraftVector{
			latitude:             right.Latitude,
			longitude:            right.Longitude,
			speedMetersPerSecond: right.VelocityMetersPerSecond,
			headingDegrees:       right.HeadingDegrees,
		},
	)
	kind := interactionKind(left, right, closingRate, policy)
	confidence := buildCandidateConfidence(
		left,
		right,
		timeDifference,
		maximumTimeDifference,
		verticalFilteringApplied,
		policy,
	)
	status := candidateStatus(left, right, verticalFilteringApplied)
	candidate := Candidate{
		ID:                                  canonicalPairID(left.NodeID, right.NodeID),
		SourceNodeID:                        left.NodeID,
		TargetNodeID:                        right.NodeID,
		Status:                              status,
		Kind:                                kind,
		HorizontalDistanceKilometers:        horizontalDistance,
		VerticalSeparationMeters:            cloneFloat64(verticalSeparation),
		ObservationTimeDifference:           timeDifference,
		EffectiveHorizontalRadiusKilometers: effectiveHorizontalRadius,
		EffectiveVerticalRadiusMeters:       cloneFloat64(effectiveVerticalRadius),
		VerticalFilteringApplied:            verticalFilteringApplied,
		ClosingRateMetersPerSecond:          closingRate,
		Confidence:                          confidence,
		Limitations:                         candidateLimitations(left, right, verticalFilteringApplied, confidence, policy),
		Explanations:                        candidateExplanations(kind, closingRate),
		EvaluatedAt:                         evaluatedAt,
	}
	return candidate, pairRejectionNone
}

func comparableVerticalEvidence(
	left localtrafficscene.Aircraft,
	right localtrafficscene.Aircraft,
) bool {
	return left.AltitudeMeters != nil &&
		right.AltitudeMeters != nil &&
		left.RadiusDecision.VerticalFilteringPermitted &&
		right.RadiusDecision.VerticalFilteringPermitted &&
		left.AltitudeReference != interactiongraph.AltitudeReferenceUnknown &&
		left.AltitudeReference == right.AltitudeReference
}

func interactionKind(
	left localtrafficscene.Aircraft,
	right localtrafficscene.Aircraft,
	closingRate float64,
	policy Policy,
) interactiongraph.InteractionKind {
	switch {
	case closingRate >= policy.ConvergingClosingRateMetersPerSecond:
		return interactiongraph.InteractionKindConverging
	case closingRate <= -policy.DivergingOpeningRateMetersPerSecond:
		return interactiongraph.InteractionKindDiverging
	case headingDifferenceDegrees(left.HeadingDegrees, right.HeadingDegrees) <=
		policy.ParallelHeadingToleranceDegrees:
		return interactiongraph.InteractionKindParallel
	default:
		return interactiongraph.InteractionKindNearby
	}
}

func buildCandidateConfidence(
	left localtrafficscene.Aircraft,
	right localtrafficscene.Aircraft,
	timeDifference time.Duration,
	maximumTimeDifference time.Duration,
	verticalFilteringApplied bool,
	policy Policy,
) Confidence {
	qualityScore := mean(left.QualityScore, right.QualityScore)
	radiusScore := mean(
		left.RadiusDecision.Confidence.Score,
		right.RadiusDecision.Confidence.Score,
	)
	temporalScore := 1.0
	if maximumTimeDifference > 0 {
		temporalScore = clampUnit(
			1 - float64(timeDifference)/float64(maximumTimeDifference),
		)
	}
	verticalScore := 0.0
	if verticalFilteringApplied {
		verticalScore = 1
	}
	components := []ConfidenceComponent{
		{Name: "prepared_evidence_quality", Score: qualityScore, Weight: policy.CandidateConfidenceWeights.PreparedEvidenceQuality},
		{Name: "radius_decision_confidence", Score: radiusScore, Weight: policy.CandidateConfidenceWeights.RadiusDecisionConfidence},
		{Name: "temporal_proximity", Score: temporalScore, Weight: policy.CandidateConfidenceWeights.TemporalProximity},
		{Name: "vertical_evidence", Score: verticalScore, Weight: policy.CandidateConfidenceWeights.VerticalEvidence},
	}
	score := weightedScore(components)
	return Confidence{
		Score:      score,
		Level:      confidenceLevel(score, policy),
		Components: components,
		Reasons: []ConfidenceReason{
			{Code: "prepared_evidence_quality", Message: "Candidate confidence includes the mean quality of both prepared aircraft observations.", Contribution: qualityScore},
			{Code: "radius_decision_confidence", Message: "Candidate confidence includes the mean confidence of both Interaction Radius decisions.", Contribution: radiusScore},
			{Code: "temporal_proximity", Message: "Candidate confidence includes how closely the two observation times align.", Contribution: temporalScore},
			{Code: "vertical_evidence", Message: "Candidate confidence records whether comparable vertical evidence was available.", Contribution: verticalScore},
		},
	}
}

func candidateStatus(
	left localtrafficscene.Aircraft,
	right localtrafficscene.Aircraft,
	verticalFilteringApplied bool,
) CandidateStatus {
	if !verticalFilteringApplied ||
		left.RadiusDecision.Status == interactionradius.DecisionStatusLimited ||
		right.RadiusDecision.Status == interactionradius.DecisionStatusLimited {
		return CandidateStatusLimited
	}
	return CandidateStatusComplete
}

func candidateLimitations(
	left localtrafficscene.Aircraft,
	right localtrafficscene.Aircraft,
	verticalFilteringApplied bool,
	confidence Confidence,
	policy Policy,
) []Limitation {
	limitations := []Limitation{
		{
			Code:    "research_only_candidate",
			Message: "The pair is a research proximity candidate and is not an operational separation alert.",
			Scope:   "operational_use",
		},
	}
	if !verticalFilteringApplied {
		limitations = append(limitations, Limitation{
			Code:    "vertical_filtering_withheld",
			Message: "Comparable vertical evidence was unavailable, so candidate inclusion used horizontal and temporal evidence only.",
			Scope:   "vertical_evidence",
		})
	}
	if left.RadiusDecision.Status == interactionradius.DecisionStatusLimited ||
		right.RadiusDecision.Status == interactionradius.DecisionStatusLimited {
		limitations = append(limitations, Limitation{
			Code:    "limited_radius_decision_present",
			Message: "At least one aircraft has a limited Interaction Radius decision.",
			Scope:   "candidate_search",
		})
	}
	if confidence.Score < policy.MediumConfidenceMinimumScore {
		limitations = append(limitations, Limitation{
			Code:    "low_candidate_confidence",
			Message: "The candidate confidence is below the medium-confidence threshold.",
			Scope:   "confidence",
		})
	}
	return limitations
}

func candidateExplanations(
	kind interactiongraph.InteractionKind,
	closingRate float64,
) []Explanation {
	return []Explanation{
		{
			Code:    "bounded_pairwise_candidate",
			Message: "The pair passed the applicable temporal, horizontal, and vertical candidate filters.",
		},
		{
			Code:    "relative_motion_classification",
			Message: fmt.Sprintf("Relative horizontal motion classified the pair as %s with a signed closing rate of %.3f meters per second.", kind, closingRate),
		},
		{
			Code:    "candidate_is_not_risk",
			Message: "Candidate membership does not classify regulated separation risk or collision probability.",
		},
	}
}

func edgeInput(candidate Candidate) interactiongraph.EdgeInput {
	limitations := make([]interactiongraph.Limitation, 0, len(candidate.Limitations))
	for _, limitation := range candidate.Limitations {
		limitations = append(limitations, interactiongraph.Limitation{
			Code:    limitation.Code,
			Message: limitation.Message,
			Scope:   limitation.Scope,
		})
	}
	return interactiongraph.EdgeInput{
		SourceNodeID:                 candidate.SourceNodeID,
		TargetNodeID:                 candidate.TargetNodeID,
		Kind:                         candidate.Kind,
		HorizontalDistanceKilometers: candidate.HorizontalDistanceKilometers,
		VerticalSeparationMeters:     cloneFloat64(candidate.VerticalSeparationMeters),
		EvaluatedAt:                  candidate.EvaluatedAt,
		SourceName:                   scannerSourceName,
		ConfidenceScore:              candidate.Confidence.Score,
		Limitations:                  limitations,
	}
}

func resultStatus(
	scene localtrafficscene.Result,
	candidates []Candidate,
) ResultStatus {
	if len(scene.Aircraft) == 0 {
		return ResultStatusUnavailable
	}
	if len(scene.Aircraft) < 2 || scene.Status != localtrafficscene.ResultStatusComplete {
		return ResultStatusLimited
	}
	for _, candidate := range candidates {
		if candidate.Status == CandidateStatusLimited {
			return ResultStatusLimited
		}
	}
	return ResultStatusComplete
}

func buildResultConfidence(
	scene localtrafficscene.Result,
	aircraft []localtrafficscene.Aircraft,
	metrics Metrics,
	policy Policy,
) Confidence {
	meanRadiusConfidence := 0.0
	if len(aircraft) > 0 {
		for _, item := range aircraft {
			meanRadiusConfidence += item.RadiusDecision.Confidence.Score
		}
		meanRadiusConfidence /= float64(len(aircraft))
	}
	pairCompleteness := 1.0
	if metrics.PossiblePairCount > 0 {
		pairCompleteness = float64(metrics.EvaluatedPairCount) /
			float64(metrics.PossiblePairCount)
	}
	components := []ConfidenceComponent{
		{Name: "scene_confidence", Score: scene.Confidence.Score, Weight: policy.ResultConfidenceWeights.SceneConfidence},
		{Name: "mean_radius_confidence", Score: meanRadiusConfidence, Weight: policy.ResultConfidenceWeights.MeanRadiusConfidence},
		{Name: "pair_evaluation_completeness", Score: pairCompleteness, Weight: policy.ResultConfidenceWeights.PairEvaluationCompleteness},
	}
	score := weightedScore(components)
	return Confidence{
		Score:      score,
		Level:      confidenceLevel(score, policy),
		Components: components,
		Reasons: []ConfidenceReason{
			{Code: "scene_confidence", Message: "Scan confidence includes the upstream Local Traffic Scene confidence.", Contribution: scene.Confidence.Score},
			{Code: "mean_radius_confidence", Message: "Scan confidence includes the mean confidence of all included Interaction Radius decisions.", Contribution: meanRadiusConfidence},
			{Code: "pair_evaluation_completeness", Message: "Scan confidence includes the share of possible aircraft pairs evaluated by the scanner.", Contribution: pairCompleteness},
		},
	}
}

func buildResultLimitations(
	scene localtrafficscene.Result,
	metrics Metrics,
) []Limitation {
	limitations := []Limitation{
		{
			Code:    "research_only_not_operational_separation",
			Message: "The scan and graph are research context and must not be used for operational separation or collision avoidance.",
			Scope:   "operational_use",
		},
	}
	if metrics.AircraftCount < 2 {
		limitations = append(limitations, Limitation{
			Code:    "insufficient_aircraft_for_pairwise_scan",
			Message: "Fewer than two prepared aircraft were available for pairwise scanning.",
			Scope:   "pairwise_context",
		})
	}
	if metrics.CandidatePairCount == 0 && metrics.AircraftCount >= 2 {
		limitations = append(limitations, Limitation{
			Code:    "no_proximity_candidates",
			Message: "No pair passed the candidate filters; this is not proof of safe operational separation.",
			Scope:   "candidate_coverage",
		})
	}
	if metrics.VerticalFilteringWithheldPairCount > 0 {
		limitations = append(limitations, Limitation{
			Code:    "vertical_filtering_withheld_for_candidates",
			Message: "One or more candidates were included without comparable vertical evidence.",
			Scope:   "vertical_evidence",
		})
	}
	if scene.Status != localtrafficscene.ResultStatusComplete {
		limitations = append(limitations, Limitation{
			Code:    "upstream_scene_limited",
			Message: "The upstream Local Traffic Scene is limited.",
			Scope:   "scene_coverage",
		})
	}
	return limitations
}

func buildResultExplanations(metrics Metrics) []Explanation {
	return []Explanation{
		{
			Code:    "all_unique_pairs_evaluated",
			Message: fmt.Sprintf("The scanner evaluated %d of %d possible unique aircraft pairs.", metrics.EvaluatedPairCount, metrics.PossiblePairCount),
		},
		{
			Code:    "radius_policy_candidate_filter",
			Message: "Each pair uses the broader published horizontal radius, the stricter published time boundary, and vertical filtering only when altitude evidence is comparable.",
		},
		{
			Code:    "graph_composed_from_candidates",
			Message: fmt.Sprintf("The scanner converted %d accepted pairwise candidates into deterministic Interaction Graph edges.", metrics.CandidatePairCount),
		},
		{
			Code:    "candidate_is_not_operational_alert",
			Message: "A graph edge is contextual research evidence and is not a certified loss-of-separation or collision alert.",
		},
	}
}

func validateRequest(request Request, policy Policy) error {
	scene := request.Scene
	if request.GeneratedAt.IsZero() ||
		request.GeneratedAt.Before(scene.AsOfTime) ||
		request.GeneratedAt.Before(scene.GeneratedAt) ||
		scene.SchemaVersion != localtrafficscene.SchemaVersionV1 ||
		!scene.Status.IsKnown() ||
		strings.TrimSpace(scene.RegionCode) == "" ||
		scene.AsOfTime.IsZero() ||
		scene.GeneratedAt.IsZero() ||
		scene.ScopeGuard != localtrafficscene.ScopeGuardResearchOnly ||
		strings.TrimSpace(scene.Provenance.InputFingerprint) == "" {
		return fmt.Errorf("%w: scene contract or generation time", ErrInvalidRequest)
	}
	if len(scene.Aircraft) > policy.MaximumAircraftCount {
		return fmt.Errorf("%w: aircraft count exceeds policy maximum", ErrInvalidRequest)
	}
	if possiblePairCount(len(scene.Aircraft)) > policy.MaximumPairCount {
		return fmt.Errorf("%w: pair count exceeds policy maximum", ErrInvalidRequest)
	}
	seen := make(map[string]struct{}, len(scene.Aircraft))
	for index, aircraft := range scene.Aircraft {
		if strings.TrimSpace(aircraft.NodeID) == "" ||
			strings.TrimSpace(aircraft.SourceName) == "" ||
			aircraft.ObservedAt.IsZero() ||
			aircraft.ObservedAt.After(scene.AsOfTime) ||
			!validLatitude(aircraft.Latitude) ||
			!validLongitude(aircraft.Longitude) ||
			!unitInterval(aircraft.QualityScore) {
			return fmt.Errorf("%w: aircraft[%d]", ErrInvalidRequest, index)
		}
		if _, exists := seen[aircraft.NodeID]; exists {
			return fmt.Errorf("%w: duplicate node %q", ErrInvalidRequest, aircraft.NodeID)
		}
		seen[aircraft.NodeID] = struct{}{}
		radiusReport := interactionradius.Validate(aircraft.RadiusDecision)
		if radiusReport.Status != interactionradius.ValidationStatusValid ||
			aircraft.RadiusDecision.Status == interactionradius.DecisionStatusBlocked {
			return fmt.Errorf("%w: aircraft[%d] radius decision", ErrInvalidRequest, index)
		}
	}
	return nil
}

func canonicalPairID(left string, right string) string {
	if left > right {
		left, right = right, left
	}
	return left + "--" + right
}

func possiblePairCount(aircraftCount int) int {
	return aircraftCount * (aircraftCount - 1) / 2
}

func absoluteDuration(value time.Duration) time.Duration {
	if value < 0 {
		return -value
	}
	return value
}

func minimumDuration(left time.Duration, right time.Duration) time.Duration {
	if left < right {
		return left
	}
	return right
}

func mean(left float64, right float64) float64 {
	return (left + right) / 2
}

func weightedScore(components []ConfidenceComponent) float64 {
	score := 0.0
	for _, component := range components {
		score += component.Score * component.Weight
	}
	return clampUnit(score)
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

func validLatitude(value float64) bool {
	return finite(value) && value >= -90 && value <= 90
}

func validLongitude(value float64) bool {
	return finite(value) && value >= -180 && value <= 180
}

func clampUnit(value float64) float64 {
	return math.Min(math.Max(value, 0), 1)
}
