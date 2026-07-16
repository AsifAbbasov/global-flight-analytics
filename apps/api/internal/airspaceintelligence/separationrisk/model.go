package separationrisk

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/interactiongraph"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/proximityscanner"
)

const Version = "separation-risk-intelligence-v1"

type SchemaVersion string

const SchemaVersionV1 SchemaVersion = "separation-risk-intelligence-v1"

type ResultStatus string

const (
	ResultStatusUnavailable ResultStatus = "unavailable"
	ResultStatusLimited     ResultStatus = "limited"
	ResultStatusComplete    ResultStatus = "complete"
)

func (status ResultStatus) IsKnown() bool {
	switch status {
	case ResultStatusUnavailable, ResultStatusLimited, ResultStatusComplete:
		return true
	default:
		return false
	}
}

type AssessmentStatus string

const (
	AssessmentStatusLimited  AssessmentStatus = "limited"
	AssessmentStatusComplete AssessmentStatus = "complete"
)

func (status AssessmentStatus) IsKnown() bool {
	return status == AssessmentStatusLimited || status == AssessmentStatusComplete
}

type RiskLevel string

const (
	RiskLevelIndeterminate RiskLevel = "indeterminate"
	RiskLevelContextual    RiskLevel = "contextual"
	RiskLevelElevated      RiskLevel = "elevated"
	RiskLevelHigh          RiskLevel = "high"
)

func (level RiskLevel) IsKnown() bool {
	switch level {
	case RiskLevelIndeterminate, RiskLevelContextual, RiskLevelElevated, RiskLevelHigh:
		return true
	default:
		return false
	}
}

type ConfidenceLevel string

const (
	ConfidenceLevelNone   ConfidenceLevel = "none"
	ConfidenceLevelLow    ConfidenceLevel = "low"
	ConfidenceLevelMedium ConfidenceLevel = "medium"
	ConfidenceLevelHigh   ConfidenceLevel = "high"
)

func (level ConfidenceLevel) IsKnown() bool {
	switch level {
	case ConfidenceLevelNone, ConfidenceLevelLow, ConfidenceLevelMedium, ConfidenceLevelHigh:
		return true
	default:
		return false
	}
}

type ScopeGuard string

const ScopeGuardResearchOnly ScopeGuard = "research_only_not_for_operational_separation_or_collision_avoidance_use"

type Request struct {
	Scan        proximityscanner.Result
	GeneratedAt time.Time
}

type ScoreComponent struct {
	Name   string
	Score  float64
	Weight float64
}

type ConfidenceReason struct {
	Code         string
	Message      string
	Contribution float64
}

type Confidence struct {
	Score      float64
	Level      ConfidenceLevel
	Components []ScoreComponent
	Reasons    []ConfidenceReason
}

type Limitation struct {
	Code    string
	Message string
	Scope   string
}

type Explanation struct {
	Code    string
	Message string
}

type Assessment struct {
	CandidateID  string
	SourceNodeID string
	TargetNodeID string
	Status       AssessmentStatus
	Level        RiskLevel
	Kind         interactiongraph.InteractionKind

	HorizontalDistanceKilometers float64
	VerticalSeparationMeters     *float64
	ObservationTimeDifference    time.Duration
	ClosingRateMetersPerSecond   float64

	HorizontalRadiusRatio *float64
	VerticalRadiusRatio   *float64
	RiskScore             *float64

	Confidence   Confidence
	Limitations  []Limitation
	Explanations []Explanation
	EvaluatedAt  time.Time
}

type Metrics struct {
	CandidateCount              int
	CompleteAssessmentCount     int
	LimitedAssessmentCount      int
	IndeterminateCount          int
	ContextualCount             int
	ElevatedCount               int
	HighCount                   int
	ConvergingAssessmentCount   int
	VerticalEvidenceWithheld    int
	HighestDeterminateRiskLevel RiskLevel
}

type Provenance struct {
	InputFingerprint string
	ScanFingerprint  string
	SourceNames      []string
	LatestObservedAt time.Time
}

type Result struct {
	SchemaVersion SchemaVersion
	Status        ResultStatus
	RegionCode    string
	AsOfTime      time.Time

	Assessments []Assessment
	Metrics     Metrics

	Confidence   Confidence
	Limitations  []Limitation
	Explanations []Explanation
	ScopeGuard   ScopeGuard
	Provenance   Provenance
	GeneratedAt  time.Time
}

func (result Result) Clone() Result {
	cloned := result
	cloned.Assessments = make([]Assessment, 0, len(result.Assessments))
	for _, assessment := range result.Assessments {
		cloned.Assessments = append(cloned.Assessments, assessment.Clone())
	}
	cloned.Confidence.Components = append([]ScoreComponent(nil), result.Confidence.Components...)
	cloned.Confidence.Reasons = append([]ConfidenceReason(nil), result.Confidence.Reasons...)
	cloned.Limitations = append([]Limitation(nil), result.Limitations...)
	cloned.Explanations = append([]Explanation(nil), result.Explanations...)
	cloned.Provenance.SourceNames = append([]string(nil), result.Provenance.SourceNames...)
	return cloned
}

func (assessment Assessment) Clone() Assessment {
	cloned := assessment
	cloned.VerticalSeparationMeters = cloneFloat64(assessment.VerticalSeparationMeters)
	cloned.HorizontalRadiusRatio = cloneFloat64(assessment.HorizontalRadiusRatio)
	cloned.VerticalRadiusRatio = cloneFloat64(assessment.VerticalRadiusRatio)
	cloned.RiskScore = cloneFloat64(assessment.RiskScore)
	cloned.Confidence.Components = append([]ScoreComponent(nil), assessment.Confidence.Components...)
	cloned.Confidence.Reasons = append([]ConfidenceReason(nil), assessment.Confidence.Reasons...)
	cloned.Limitations = append([]Limitation(nil), assessment.Limitations...)
	cloned.Explanations = append([]Explanation(nil), assessment.Explanations...)
	return cloned
}

func cloneFloat64(value *float64) *float64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
