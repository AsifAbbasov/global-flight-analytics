package separationrisk

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/interactiongraph"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/proximityscanner"
)

func Evaluate(request Request, policy Policy) (Result, error) {
	if err := policy.Validate(); err != nil {
		return Result{}, err
	}
	if err := validateRequest(request, policy); err != nil {
		return Result{}, err
	}

	scan := request.Scan.Clone()
	assessments := make([]Assessment, 0, len(scan.Candidates))
	for _, candidate := range scan.Candidates {
		assessments = append(assessments, evaluateCandidate(candidate, request.GeneratedAt.UTC(), policy))
	}
	sort.Slice(assessments, func(left, right int) bool {
		return assessments[left].CandidateID < assessments[right].CandidateID
	})

	metrics := buildMetrics(assessments)
	result := Result{
		SchemaVersion: SchemaVersionV1,
		Status:        resultStatus(scan.Status, metrics),
		RegionCode:    strings.ToUpper(strings.TrimSpace(scan.RegionCode)),
		AsOfTime:      scan.AsOfTime.UTC(),
		Assessments:   assessments,
		Metrics:       metrics,
		Confidence:    resultConfidence(scan, assessments, metrics, policy),
		Limitations:   resultLimitations(scan, metrics),
		Explanations:  resultExplanations(metrics),
		ScopeGuard:    ScopeGuardResearchOnly,
		Provenance: Provenance{
			ScanFingerprint:  scan.Provenance.InputFingerprint,
			SourceNames:      append([]string(nil), scan.Provenance.SourceNames...),
			LatestObservedAt: scan.Provenance.LatestObservedAt.UTC(),
		},
		GeneratedAt: request.GeneratedAt.UTC(),
	}
	result.Provenance.InputFingerprint = inputFingerprint(result, policy)

	report := Validate(result, policy)
	if report.Status != ValidationStatusValid {
		return Result{}, fmt.Errorf("%w: issues=%v", ErrInvalidResult, report.Issues)
	}
	return result.Clone(), nil
}

func validateRequest(request Request, policy Policy) error {
	if request.GeneratedAt.IsZero() || request.Scan.AsOfTime.IsZero() || request.Scan.GeneratedAt.IsZero() ||
		strings.TrimSpace(request.Scan.RegionCode) == "" || strings.TrimSpace(request.Scan.Provenance.InputFingerprint) == "" {
		return fmt.Errorf("%w: scan identity and times are required", ErrInvalidRequest)
	}
	if request.GeneratedAt.Before(request.Scan.GeneratedAt) || request.Scan.GeneratedAt.Before(request.Scan.AsOfTime) {
		return fmt.Errorf("%w: invalid generation chronology", ErrInvalidRequest)
	}
	if request.Scan.ScopeGuard != proximityscanner.ScopeGuardResearchOnly {
		return fmt.Errorf("%w: proximity scan scope guard", ErrInvalidRequest)
	}
	if len(request.Scan.Candidates) > policy.MaximumCandidateCount {
		return fmt.Errorf("%w: candidate capacity exceeded", ErrInvalidRequest)
	}
	return nil
}

func evaluateCandidate(candidate proximityscanner.Candidate, evaluatedAt time.Time, policy Policy) Assessment {
	horizontalRatio := safeRatio(candidate.HorizontalDistanceKilometers, candidate.EffectiveHorizontalRadiusKilometers)
	assessment := Assessment{
		CandidateID:                  candidate.ID,
		SourceNodeID:                 candidate.SourceNodeID,
		TargetNodeID:                 candidate.TargetNodeID,
		Status:                       AssessmentStatusComplete,
		Kind:                         candidate.Kind,
		HorizontalDistanceKilometers: candidate.HorizontalDistanceKilometers,
		VerticalSeparationMeters:     cloneFloat64(candidate.VerticalSeparationMeters),
		ObservationTimeDifference:    candidate.ObservationTimeDifference,
		ClosingRateMetersPerSecond:   candidate.ClosingRateMetersPerSecond,
		HorizontalRadiusRatio:        &horizontalRatio,
		EvaluatedAt:                  evaluatedAt,
	}

	if !candidate.VerticalFilteringApplied || candidate.VerticalSeparationMeters == nil || candidate.EffectiveVerticalRadiusMeters == nil {
		assessment.Status = AssessmentStatusLimited
		assessment.Level = RiskLevelIndeterminate
		assessment.Confidence = assessmentConfidence(candidate, 0, false, policy)
		assessment.Limitations = append(baseAssessmentLimitations(), Limitation{
			Code:    "vertical_evidence_unavailable",
			Message: "Comparable vertical evidence is unavailable, so a determinate separation-risk level is withheld.",
			Scope:   "vertical_evidence",
		})
		assessment.Explanations = []Explanation{
			{Code: "indeterminate_without_vertical_evidence", Message: "Horizontal proximity alone is insufficient for a determinate research separation-risk classification."},
		}
		return assessment
	}

	verticalRatio := safeRatio(*candidate.VerticalSeparationMeters, *candidate.EffectiveVerticalRadiusMeters)
	assessment.VerticalRadiusRatio = &verticalRatio
	components := riskComponents(candidate, horizontalRatio, verticalRatio, policy)
	riskScore := weightedScore(components)
	assessment.RiskScore = &riskScore
	assessment.Level = classifyRisk(candidate, horizontalRatio, verticalRatio, riskScore, policy)
	assessment.Confidence = assessmentConfidence(candidate, verticalRatio, true, policy)
	assessment.Limitations = baseAssessmentLimitations()
	assessment.Explanations = assessmentExplanations(assessment.Level, candidate.Kind)
	return assessment
}

func riskComponents(candidate proximityscanner.Candidate, horizontalRatio, verticalRatio float64, policy Policy) []ScoreComponent {
	return []ScoreComponent{
		{Name: "horizontal_proximity", Score: clampUnit(1 - horizontalRatio), Weight: policy.RiskWeights.HorizontalProximity},
		{Name: "vertical_proximity", Score: clampUnit(1 - verticalRatio), Weight: policy.RiskWeights.VerticalProximity},
		{Name: "closing_motion", Score: clampUnit(candidate.ClosingRateMetersPerSecond / policy.HighClosingRateMinimumMetersPerSecond), Weight: policy.RiskWeights.ClosingMotion},
		{Name: "temporal_alignment", Score: clampUnit(1 - float64(candidate.ObservationTimeDifference)/float64(policy.MaximumPairTimeDifference)), Weight: policy.RiskWeights.TemporalAlignment},
		{Name: "evidence_confidence", Score: candidate.Confidence.Score, Weight: policy.RiskWeights.EvidenceConfidence},
	}
}

func classifyRisk(candidate proximityscanner.Candidate, horizontalRatio, verticalRatio, score float64, policy Policy) RiskLevel {
	if candidate.Kind == interactiongraph.InteractionKindConverging &&
		score >= policy.HighRiskMinimumScore &&
		horizontalRatio <= policy.HighHorizontalRadiusRatioMaximum &&
		verticalRatio <= policy.HighVerticalRadiusRatioMaximum &&
		candidate.ClosingRateMetersPerSecond >= policy.HighClosingRateMinimumMetersPerSecond {
		return RiskLevelHigh
	}
	if candidate.Kind == interactiongraph.InteractionKindConverging &&
		score >= policy.ElevatedRiskMinimumScore &&
		horizontalRatio <= policy.ElevatedHorizontalRadiusRatioMaximum &&
		verticalRatio <= policy.ElevatedVerticalRadiusRatioMaximum &&
		candidate.ClosingRateMetersPerSecond >= policy.ElevatedClosingRateMinimumMetersPerSecond {
		return RiskLevelElevated
	}
	return RiskLevelContextual
}

func assessmentConfidence(candidate proximityscanner.Candidate, verticalRatio float64, verticalAvailable bool, policy Policy) Confidence {
	verticalScore := 0.0
	if verticalAvailable {
		verticalScore = 1
	}
	components := []ScoreComponent{
		{Name: "candidate_confidence", Score: candidate.Confidence.Score, Weight: 0.65},
		{Name: "vertical_evidence", Score: verticalScore, Weight: 0.25},
		{Name: "temporal_alignment", Score: clampUnit(1 - float64(candidate.ObservationTimeDifference)/float64(policy.MaximumPairTimeDifference)), Weight: 0.10},
	}
	score := weightedScore(components)
	return Confidence{
		Score:      score,
		Level:      confidenceLevel(score, policy),
		Components: components,
		Reasons: []ConfidenceReason{
			{Code: "candidate_confidence", Message: "Prepared proximity-candidate confidence contributes to the assessment confidence.", Contribution: candidate.Confidence.Score},
			{Code: "vertical_evidence", Message: "Comparable vertical evidence materially affects assessment confidence.", Contribution: verticalScore},
			{Code: "temporal_alignment", Message: "Observation-time alignment contributes to assessment confidence.", Contribution: components[2].Score},
		},
	}
}

func buildMetrics(assessments []Assessment) Metrics {
	metrics := Metrics{CandidateCount: len(assessments), HighestDeterminateRiskLevel: RiskLevelIndeterminate}
	for _, assessment := range assessments {
		if assessment.Kind == interactiongraph.InteractionKindConverging {
			metrics.ConvergingAssessmentCount++
		}
		switch assessment.Status {
		case AssessmentStatusComplete:
			metrics.CompleteAssessmentCount++
		case AssessmentStatusLimited:
			metrics.LimitedAssessmentCount++
		}
		switch assessment.Level {
		case RiskLevelIndeterminate:
			metrics.IndeterminateCount++
			metrics.VerticalEvidenceWithheld++
		case RiskLevelContextual:
			metrics.ContextualCount++
		case RiskLevelElevated:
			metrics.ElevatedCount++
		case RiskLevelHigh:
			metrics.HighCount++
		}
		if riskRank(assessment.Level) > riskRank(metrics.HighestDeterminateRiskLevel) {
			metrics.HighestDeterminateRiskLevel = assessment.Level
		}
	}
	return metrics
}

func resultStatus(scanStatus proximityscanner.ResultStatus, metrics Metrics) ResultStatus {
	if metrics.CandidateCount == 0 {
		return ResultStatusUnavailable
	}
	if scanStatus != proximityscanner.ResultStatusComplete || metrics.LimitedAssessmentCount > 0 {
		return ResultStatusLimited
	}
	return ResultStatusComplete
}

func resultConfidence(scan proximityscanner.Result, assessments []Assessment, metrics Metrics, policy Policy) Confidence {
	assessmentMean := 0.0
	for _, assessment := range assessments {
		assessmentMean += assessment.Confidence.Score
	}
	if len(assessments) > 0 {
		assessmentMean /= float64(len(assessments))
	}
	completeness := 0.0
	if metrics.CandidateCount > 0 {
		completeness = float64(metrics.CompleteAssessmentCount) / float64(metrics.CandidateCount)
	}
	components := []ScoreComponent{
		{Name: "scan_confidence", Score: scan.Confidence.Score, Weight: policy.ResultConfidenceWeights.ScanConfidence},
		{Name: "assessment_confidence", Score: assessmentMean, Weight: policy.ResultConfidenceWeights.AssessmentConfidence},
		{Name: "evidence_completeness", Score: completeness, Weight: policy.ResultConfidenceWeights.EvidenceCompleteness},
	}
	score := weightedScore(components)
	return Confidence{
		Score:      score,
		Level:      confidenceLevel(score, policy),
		Components: components,
		Reasons: []ConfidenceReason{
			{Code: "scan_confidence", Message: "The proximity scan confidence contributes to the result confidence.", Contribution: scan.Confidence.Score},
			{Code: "assessment_confidence", Message: "Mean pair assessment confidence contributes to the result confidence.", Contribution: assessmentMean},
			{Code: "evidence_completeness", Message: "The share of assessments with comparable vertical evidence contributes to the result confidence.", Contribution: completeness},
		},
	}
}

func baseAssessmentLimitations() []Limitation {
	return []Limitation{{
		Code:    "research_only_not_operational_separation",
		Message: "This classification is research context and must not be used for operational separation, collision avoidance, or safety decisions.",
		Scope:   "operational_use",
	}}
}

func assessmentExplanations(level RiskLevel, kind interactiongraph.InteractionKind) []Explanation {
	return []Explanation{
		{Code: "multidimensional_context_score", Message: "The classification combines horizontal proximity, vertical proximity, relative motion, temporal alignment, and evidence confidence."},
		{Code: "risk_level", Message: fmt.Sprintf("The research classification is %s for interaction kind %s.", level, kind)},
	}
}

func resultLimitations(scan proximityscanner.Result, metrics Metrics) []Limitation {
	limitations := []Limitation{
		{Code: "research_only_not_operational_separation", Message: "The result is non-operational research analytics and is not certified separation or collision-avoidance logic.", Scope: "operational_use"},
		{Code: "candidate_only_analysis", Message: "Only pairs admitted by the upstream proximity scanner are assessed.", Scope: "candidate_coverage"},
	}
	if scan.Status != proximityscanner.ResultStatusComplete {
		limitations = append(limitations, Limitation{Code: "upstream_scan_limited", Message: "The upstream proximity scan is limited.", Scope: "upstream_evidence"})
	}
	if metrics.IndeterminateCount > 0 {
		limitations = append(limitations, Limitation{Code: "indeterminate_pairs_present", Message: "One or more pairs lack comparable vertical evidence and therefore have no determinate risk level.", Scope: "vertical_evidence"})
	}
	return limitations
}

func resultExplanations(metrics Metrics) []Explanation {
	return []Explanation{
		{Code: "highest_determinate_level", Message: fmt.Sprintf("The highest determinate research risk level is %s.", metrics.HighestDeterminateRiskLevel)},
		{Code: "level_is_not_operational_alert", Message: "A high or elevated level is an analytical visualization category, not an operational alert."},
	}
}

func weightedScore(components []ScoreComponent) float64 {
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

func safeRatio(numerator, denominator float64) float64 {
	if denominator <= 0 {
		return 1
	}
	return numerator / denominator
}

func riskRank(level RiskLevel) int {
	switch level {
	case RiskLevelContextual:
		return 1
	case RiskLevelElevated:
		return 2
	case RiskLevelHigh:
		return 3
	default:
		return 0
	}
}

func clampUnit(value float64) float64 {
	return math.Min(math.Max(value, 0), 1)
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
