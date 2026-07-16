package proximityscanner

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/interactiongraph"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/localtrafficscene"
)

const Version = "multi-aircraft-proximity-scan-v1"

type SchemaVersion string

const SchemaVersionV1 SchemaVersion = "multi-aircraft-proximity-scan-v1"

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

type CandidateStatus string

const (
	CandidateStatusLimited  CandidateStatus = "limited"
	CandidateStatusComplete CandidateStatus = "complete"
)

func (status CandidateStatus) IsKnown() bool {
	return status == CandidateStatusLimited || status == CandidateStatusComplete
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

const ScopeGuardResearchOnly ScopeGuard = "research_only_not_for_operational_separation_use"

type Request struct {
	Scene       localtrafficscene.Result
	GeneratedAt time.Time
}

type ConfidenceComponent struct {
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
	Components []ConfidenceComponent
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

type Candidate struct {
	ID           string
	SourceNodeID string
	TargetNodeID string
	Status       CandidateStatus
	Kind         interactiongraph.InteractionKind

	HorizontalDistanceKilometers        float64
	VerticalSeparationMeters            *float64
	ObservationTimeDifference           time.Duration
	EffectiveHorizontalRadiusKilometers float64
	EffectiveVerticalRadiusMeters       *float64
	VerticalFilteringApplied            bool
	ClosingRateMetersPerSecond          float64

	Confidence   Confidence
	Limitations  []Limitation
	Explanations []Explanation
	EvaluatedAt  time.Time
}

type Metrics struct {
	AircraftCount                      int
	PossiblePairCount                  int
	EvaluatedPairCount                 int
	CandidatePairCount                 int
	CompleteCandidateCount             int
	LimitedCandidateCount              int
	TemporalRejectedPairCount          int
	HorizontalRejectedPairCount        int
	VerticalRejectedPairCount          int
	VerticalFilteringWithheldPairCount int
	CandidateShare                     float64
}

type Provenance struct {
	InputFingerprint string
	SceneFingerprint string
	SourceNames      []string
	LatestObservedAt time.Time
}

type Result struct {
	SchemaVersion SchemaVersion
	Status        ResultStatus
	RegionCode    string
	SceneStatus   localtrafficscene.ResultStatus
	AsOfTime      time.Time

	Candidates []Candidate
	Graph      interactiongraph.Result
	Metrics    Metrics

	Confidence   Confidence
	Limitations  []Limitation
	Explanations []Explanation
	ScopeGuard   ScopeGuard
	Provenance   Provenance
	GeneratedAt  time.Time
}

func (result Result) Clone() Result {
	cloned := result
	cloned.Candidates = make([]Candidate, 0, len(result.Candidates))
	for _, candidate := range result.Candidates {
		cloned.Candidates = append(cloned.Candidates, candidate.Clone())
	}
	cloned.Graph = result.Graph.Clone()
	cloned.Confidence.Components = append(
		[]ConfidenceComponent(nil),
		result.Confidence.Components...,
	)
	cloned.Confidence.Reasons = append(
		[]ConfidenceReason(nil),
		result.Confidence.Reasons...,
	)
	cloned.Limitations = append([]Limitation(nil), result.Limitations...)
	cloned.Explanations = append([]Explanation(nil), result.Explanations...)
	cloned.Provenance.SourceNames = append(
		[]string(nil),
		result.Provenance.SourceNames...,
	)
	return cloned
}

func (candidate Candidate) Clone() Candidate {
	cloned := candidate
	cloned.VerticalSeparationMeters = cloneFloat64(candidate.VerticalSeparationMeters)
	cloned.EffectiveVerticalRadiusMeters = cloneFloat64(candidate.EffectiveVerticalRadiusMeters)
	cloned.Confidence.Components = append(
		[]ConfidenceComponent(nil),
		candidate.Confidence.Components...,
	)
	cloned.Confidence.Reasons = append(
		[]ConfidenceReason(nil),
		candidate.Confidence.Reasons...,
	)
	cloned.Limitations = append([]Limitation(nil), candidate.Limitations...)
	cloned.Explanations = append([]Explanation(nil), candidate.Explanations...)
	return cloned
}

func cloneFloat64(value *float64) *float64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
