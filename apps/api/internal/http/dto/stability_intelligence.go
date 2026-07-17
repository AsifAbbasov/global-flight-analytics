package dto

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/stabilityintelligence/stabilityproduction"
)

type StabilityIntelligenceVersion struct {
	VersionID             string    `json:"version_id"`
	Ordinal               int       `json:"ordinal"`
	ParentVersionID       string    `json:"parent_version_id,omitempty"`
	MethodName            string    `json:"method_name"`
	MethodVersion         string    `json:"method_version"`
	PolicyVersion         string    `json:"policy_version"`
	ImplementationVersion string    `json:"implementation_version"`
	InputFingerprint      string    `json:"input_fingerprint"`
	OutputFingerprint     string    `json:"output_fingerprint"`
	DecisionFingerprint   string    `json:"decision_fingerprint"`
	CreatedAt             time.Time `json:"created_at"`
}

type StabilityIntelligenceTransitionMetrics struct {
	AlignedPointCount                       int     `json:"aligned_point_count"`
	AlignedPointShare                       float64 `json:"aligned_point_share"`
	MeanHorizontalShiftKilometers           float64 `json:"mean_horizontal_shift_kilometers"`
	MaximumHorizontalShiftKilometers        float64 `json:"maximum_horizontal_shift_kilometers"`
	AggregateConfidenceDelta                float64 `json:"aggregate_confidence_delta"`
	MeanRelativeHorizontalUncertaintyChange float64 `json:"mean_relative_horizontal_uncertainty_change"`
	ArrivalComparable                       bool    `json:"arrival_comparable"`
	ArrivalShiftSeconds                     float64 `json:"arrival_shift_seconds"`
	MethodChanged                           bool    `json:"method_changed"`
	PolicyChanged                           bool    `json:"policy_changed"`
	ImplementationChanged                   bool    `json:"implementation_changed"`
	InputChanged                            bool    `json:"input_changed"`
	OutputChanged                           bool    `json:"output_changed"`
}

type StabilityIntelligenceTransition struct {
	BaselineVersionID  string                                 `json:"baseline_version_id"`
	CandidateVersionID string                                 `json:"candidate_version_id"`
	Level              string                                 `json:"level"`
	Score              float64                                `json:"score"`
	Metrics            StabilityIntelligenceTransitionMetrics `json:"metrics"`
	InputFingerprint   string                                 `json:"input_fingerprint"`
	EvaluatedAt        time.Time                              `json:"evaluated_at"`
}

type StabilityIntelligenceAnalysisMetrics struct {
	VersionCount                     int     `json:"version_count"`
	TransitionCount                  int     `json:"transition_count"`
	ComparableTransitionCount        int     `json:"comparable_transition_count"`
	StableTransitionShare            float64 `json:"stable_transition_share"`
	ComparableTransitionShare        float64 `json:"comparable_transition_share"`
	MaterialChangeShare              float64 `json:"material_change_share"`
	MeanStabilityScore               float64 `json:"mean_stability_score"`
	MinimumStabilityScore            float64 `json:"minimum_stability_score"`
	ScoreStandardDeviation           float64 `json:"score_standard_deviation"`
	LongestStableRun                 int     `json:"longest_stable_run"`
	MethodChangeCount                int     `json:"method_change_count"`
	PolicyChangeCount                int     `json:"policy_change_count"`
	ImplementationChangeCount        int     `json:"implementation_change_count"`
	InputChangeCount                 int     `json:"input_change_count"`
	OutputChangeCount                int     `json:"output_change_count"`
	MeanHorizontalShiftKilometers    float64 `json:"mean_horizontal_shift_kilometers"`
	MaximumHorizontalShiftKilometers float64 `json:"maximum_horizontal_shift_kilometers"`
	LatestLevel                      string  `json:"latest_level"`
}

type StabilityIntelligenceConfidenceSummary struct {
	Status               string  `json:"status"`
	Score                float64 `json:"score"`
	Level                string  `json:"level"`
	TargetNodeID         string  `json:"target_node_id"`
	LimitingDependencyID string  `json:"limiting_dependency_id,omitempty"`
	InputFingerprint     string  `json:"input_fingerprint"`
}

type StabilityIntelligenceAnalysis struct {
	Status           string                               `json:"status"`
	Trend            string                               `json:"trend"`
	Health           string                               `json:"health"`
	Metrics          StabilityIntelligenceAnalysisMetrics `json:"metrics"`
	ConfidenceScore  float64                              `json:"confidence_score"`
	ConfidenceLevel  string                               `json:"confidence_level"`
	InputFingerprint string                               `json:"input_fingerprint"`
}

type StabilityIntelligenceFailure struct {
	Rank                 int      `json:"rank"`
	Code                 string   `json:"code"`
	Category             string   `json:"category"`
	Severity             string   `json:"severity"`
	Classification       string   `json:"classification"`
	Summary              string   `json:"summary"`
	Detail               string   `json:"detail"`
	Source               string   `json:"source"`
	BlocksUse            bool     `json:"blocks_use"`
	PriorityScore        float64  `json:"priority_score"`
	EvidenceFingerprints []string `json:"evidence_fingerprints"`
}

type StabilityIntelligenceFailureExplanation struct {
	Status            string                         `json:"status"`
	PrimaryCode       string                         `json:"primary_code"`
	BlockingCount     int                            `json:"blocking_count"`
	WarningCount      int                            `json:"warning_count"`
	UnknownCauseCount int                            `json:"unknown_cause_count"`
	ConfidenceScore   float64                        `json:"confidence_score"`
	ConfidenceLevel   string                         `json:"confidence_level"`
	Failures          []StabilityIntelligenceFailure `json:"failures"`
	InputFingerprint  string                         `json:"input_fingerprint"`
}

type StabilityIntelligenceInterventionGuard struct {
	Status                 string  `json:"status"`
	ClaimKind              string  `json:"claim_kind"`
	Decision               string  `json:"decision"`
	ConfidenceScore        float64 `json:"confidence_score"`
	EvidenceCount          int     `json:"evidence_count"`
	UnknownEvidenceCount   int     `json:"unknown_evidence_count"`
	EstimatedEvidenceCount int     `json:"estimated_evidence_count"`
	EvidenceCompleteness   float64 `json:"evidence_completeness"`
	InputFingerprint       string  `json:"input_fingerprint"`
}

type StabilityIntelligenceScopeViolation struct {
	Code      string `json:"code"`
	ClaimCode string `json:"claim_code"`
	Message   string `json:"message"`
	Blocking  bool   `json:"blocking"`
}

type StabilityIntelligenceScopeEnforcement struct {
	Status           string                                `json:"status"`
	Decision         string                                `json:"decision"`
	ClaimCount       int                                   `json:"claim_count"`
	AllowedCount     int                                   `json:"allowed_count"`
	LimitedCount     int                                   `json:"limited_count"`
	BlockedCount     int                                   `json:"blocked_count"`
	Violations       []StabilityIntelligenceScopeViolation `json:"violations"`
	InputFingerprint string                                `json:"input_fingerprint"`
}

type StabilityIntelligenceResponse struct {
	Version      string      `json:"version"`
	TrajectoryID string      `json:"trajectory_id"`
	AsOfTimes    []time.Time `json:"as_of_times"`

	Projections []ProjectionIntelligenceResponse       `json:"projections"`
	Versions    []StabilityIntelligenceVersion         `json:"forecast_versions"`
	Transitions []StabilityIntelligenceTransition      `json:"transitions"`
	Analysis    StabilityIntelligenceAnalysis          `json:"forecast_analysis"`
	Confidence  StabilityIntelligenceConfidenceSummary `json:"propagated_confidence"`

	FailureExplanation  StabilityIntelligenceFailureExplanation `json:"failure_explanation"`
	UnknownIntervention StabilityIntelligenceInterventionGuard  `json:"unknown_intervention"`
	ScopeEnforcement    StabilityIntelligenceScopeEnforcement   `json:"scope_enforcement"`

	ScopeGuards      []string  `json:"scope_guards"`
	InputFingerprint string    `json:"input_fingerprint"`
	GeneratedAt      time.Time `json:"generated_at"`
}

func ToStabilityIntelligenceResponse(
	result stabilityproduction.Result,
) StabilityIntelligenceResponse {
	response := StabilityIntelligenceResponse{
		Version:          result.Version,
		TrajectoryID:     result.TrajectoryID,
		AsOfTimes:        append([]time.Time(nil), result.AsOfTimes...),
		ScopeGuards:      append([]string(nil), result.ScopeGuards...),
		InputFingerprint: result.InputFingerprint,
		GeneratedAt:      result.GeneratedAt,
	}

	response.Projections = make(
		[]ProjectionIntelligenceResponse,
		0,
		len(result.Projections),
	)
	for _, projection := range result.Projections {
		response.Projections = append(
			response.Projections,
			ToProjectionIntelligenceResponse(
				projection,
			),
		)
	}

	response.Versions = make(
		[]StabilityIntelligenceVersion,
		0,
		len(result.ForecastVersions),
	)
	for _, version := range result.ForecastVersions {
		response.Versions = append(
			response.Versions,
			StabilityIntelligenceVersion{
				VersionID:             version.VersionID,
				Ordinal:               version.Ordinal,
				ParentVersionID:       version.ParentVersionID,
				MethodName:            version.Method.Name,
				MethodVersion:         version.Method.Version,
				PolicyVersion:         version.PolicyVersion,
				ImplementationVersion: version.ImplementationVersion,
				InputFingerprint:      version.InputFingerprint,
				OutputFingerprint:     version.OutputFingerprint,
				DecisionFingerprint:   version.DecisionFingerprint,
				CreatedAt:             version.CreatedAt,
			},
		)
	}

	response.Transitions = make(
		[]StabilityIntelligenceTransition,
		0,
		len(result.Transitions),
	)
	for _, transition := range result.Transitions {
		response.Transitions = append(
			response.Transitions,
			StabilityIntelligenceTransition{
				BaselineVersionID:  transition.BaselineVersionID,
				CandidateVersionID: transition.CandidateVersionID,
				Level:              string(transition.Level),
				Score:              transition.Score,
				Metrics: StabilityIntelligenceTransitionMetrics{
					AlignedPointCount:                       transition.Metrics.AlignedPointCount,
					AlignedPointShare:                       transition.Metrics.AlignedPointShare,
					MeanHorizontalShiftKilometers:           transition.Metrics.MeanHorizontalShiftKilometers,
					MaximumHorizontalShiftKilometers:        transition.Metrics.MaximumHorizontalShiftKilometers,
					AggregateConfidenceDelta:                transition.Metrics.AggregateConfidenceDelta,
					MeanRelativeHorizontalUncertaintyChange: transition.Metrics.MeanRelativeHorizontalUncertaintyChange,
					ArrivalComparable:                       transition.Metrics.ArrivalComparable,
					ArrivalShiftSeconds:                     transition.Metrics.ArrivalShiftSeconds,
					MethodChanged:                           transition.Metrics.MethodChanged,
					PolicyChanged:                           transition.Metrics.PolicyChanged,
					ImplementationChanged:                   transition.Metrics.ImplementationChanged,
					InputChanged:                            transition.Metrics.InputChanged,
					OutputChanged:                           transition.Metrics.OutputChanged,
				},
				InputFingerprint: transition.Provenance.InputFingerprint,
				EvaluatedAt:      transition.EvaluatedAt,
			},
		)
	}

	analysisMetrics := result.ForecastAnalysis.Metrics
	response.Analysis = StabilityIntelligenceAnalysis{
		Status: string(result.ForecastAnalysis.Status),
		Trend:  string(result.ForecastAnalysis.Trend),
		Health: string(result.ForecastAnalysis.Health),
		Metrics: StabilityIntelligenceAnalysisMetrics{
			VersionCount:                     analysisMetrics.VersionCount,
			TransitionCount:                  analysisMetrics.TransitionCount,
			ComparableTransitionCount:        analysisMetrics.ComparableTransitionCount,
			StableTransitionShare:            analysisMetrics.StableTransitionShare,
			ComparableTransitionShare:        analysisMetrics.ComparableTransitionShare,
			MaterialChangeShare:              analysisMetrics.MaterialChangeShare,
			MeanStabilityScore:               analysisMetrics.MeanStabilityScore,
			MinimumStabilityScore:            analysisMetrics.MinimumStabilityScore,
			ScoreStandardDeviation:           analysisMetrics.ScoreStandardDeviation,
			LongestStableRun:                 analysisMetrics.LongestStableRun,
			MethodChangeCount:                analysisMetrics.MethodChangeCount,
			PolicyChangeCount:                analysisMetrics.PolicyChangeCount,
			ImplementationChangeCount:        analysisMetrics.ImplementationChangeCount,
			InputChangeCount:                 analysisMetrics.InputChangeCount,
			OutputChangeCount:                analysisMetrics.OutputChangeCount,
			MeanHorizontalShiftKilometers:    analysisMetrics.MeanHorizontalShiftKilometers,
			MaximumHorizontalShiftKilometers: analysisMetrics.MaximumHorizontalShiftKilometers,
			LatestLevel:                      string(analysisMetrics.LatestLevel),
		},
		ConfidenceScore:  result.ForecastAnalysis.Confidence.Score,
		ConfidenceLevel:  result.ForecastAnalysis.Confidence.Level,
		InputFingerprint: result.ForecastAnalysis.Provenance.InputFingerprint,
	}

	limitingDependencyID := ""
	for _, node := range result.PropagatedConfidence.Nodes {
		if node.NodeID == result.PropagatedConfidence.TargetNodeID {
			limitingDependencyID = node.LimitingDependencyID
			break
		}
	}
	response.Confidence = StabilityIntelligenceConfidenceSummary{
		Status:               string(result.PropagatedConfidence.Status),
		Score:                result.PropagatedConfidence.Score,
		Level:                result.PropagatedConfidence.Level,
		TargetNodeID:         result.PropagatedConfidence.TargetNodeID,
		LimitingDependencyID: limitingDependencyID,
		InputFingerprint:     result.PropagatedConfidence.Provenance.InputFingerprint,
	}

	failures := make(
		[]StabilityIntelligenceFailure,
		0,
		len(result.FailureExplanation.Failures),
	)
	for _, failure := range result.FailureExplanation.Failures {
		failures = append(
			failures,
			StabilityIntelligenceFailure{
				Rank:                 failure.Rank,
				Code:                 failure.Code,
				Category:             string(failure.Category),
				Severity:             string(failure.Severity),
				Classification:       string(failure.Classification),
				Summary:              failure.Summary,
				Detail:               failure.Detail,
				Source:               failure.Source,
				BlocksUse:            failure.BlocksUse,
				PriorityScore:        failure.PriorityScore,
				EvidenceFingerprints: append([]string(nil), failure.EvidenceFingerprints...),
			},
		)
	}
	response.FailureExplanation = StabilityIntelligenceFailureExplanation{
		Status:            string(result.FailureExplanation.Status),
		PrimaryCode:       result.FailureExplanation.PrimaryCode,
		BlockingCount:     result.FailureExplanation.Metrics.BlockingCount,
		WarningCount:      result.FailureExplanation.Metrics.WarningCount,
		UnknownCauseCount: result.FailureExplanation.Metrics.UnknownCauseCount,
		ConfidenceScore:   result.FailureExplanation.Confidence.Score,
		ConfidenceLevel:   result.FailureExplanation.Confidence.Level,
		Failures:          failures,
		InputFingerprint:  result.FailureExplanation.Provenance.InputFingerprint,
	}

	response.UnknownIntervention = StabilityIntelligenceInterventionGuard{
		Status:                 string(result.UnknownIntervention.Status),
		ClaimKind:              string(result.UnknownIntervention.ClaimKind),
		Decision:               string(result.UnknownIntervention.Decision),
		ConfidenceScore:        result.UnknownIntervention.ConfidenceScore,
		EvidenceCount:          result.UnknownIntervention.Metrics.EvidenceCount,
		UnknownEvidenceCount:   result.UnknownIntervention.Metrics.UnknownEvidenceCount,
		EstimatedEvidenceCount: result.UnknownIntervention.Metrics.EstimatedEvidenceCount,
		EvidenceCompleteness:   result.UnknownIntervention.Metrics.EvidenceCompleteness,
		InputFingerprint:       result.UnknownIntervention.Provenance.InputFingerprint,
	}

	violations := make(
		[]StabilityIntelligenceScopeViolation,
		0,
		len(result.ScopeEnforcement.Violations),
	)
	for _, violation := range result.ScopeEnforcement.Violations {
		violations = append(
			violations,
			StabilityIntelligenceScopeViolation{
				Code:      violation.Code,
				ClaimCode: violation.ClaimCode,
				Message:   violation.Message,
				Blocking:  violation.Blocking,
			},
		)
	}
	response.ScopeEnforcement = StabilityIntelligenceScopeEnforcement{
		Status:           string(result.ScopeEnforcement.Status),
		Decision:         string(result.ScopeEnforcement.Decision),
		ClaimCount:       result.ScopeEnforcement.Metrics.ClaimCount,
		AllowedCount:     result.ScopeEnforcement.Metrics.AllowedCount,
		LimitedCount:     result.ScopeEnforcement.Metrics.LimitedCount,
		BlockedCount:     result.ScopeEnforcement.Metrics.BlockedCount,
		Violations:       violations,
		InputFingerprint: result.ScopeEnforcement.Provenance.InputFingerprint,
	}

	return response
}
