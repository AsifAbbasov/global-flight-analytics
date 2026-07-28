package historicalsimilarity

import (
	"math"
	"regexp"
)

const (
	Version            = "historical-trajectory-similarity-v2"
	FingerprintVersion = "historical-trajectory-similarity-fingerprint-v2"

	DefaultMinimumPointCount = 4
	DefaultSampleCount       = 16
	MaximumSampleCount       = 4096
	MaximumInputPointCount   = 100000

	// MaximumRankLimit is retained only as the shared projection-neighbor
	// selection limit. Historical Similarity no longer owns a public Rank API.
	MaximumRankLimit = 100
)

const (
	componentCount    = 4
	numericTolerance  = 1e-12
	weightTolerance   = 1e-9
	earthRadiusKM     = 6371.0088
	coordinateEpsilon = 1e-12
)

type Level string

const (
	LevelNone   Level = "none"
	LevelLow    Level = "low"
	LevelMedium Level = "medium"
	LevelHigh   Level = "high"
)

type ConfidenceLevel string

const (
	ConfidenceLevelNone   ConfidenceLevel = "none"
	ConfidenceLevelLow    ConfidenceLevel = "low"
	ConfidenceLevelMedium ConfidenceLevel = "medium"
	ConfidenceLevelHigh   ConfidenceLevel = "high"
)

type ComponentName string

const (
	ComponentGeometry   ComponentName = "geometry"
	ComponentEndpoints  ComponentName = "endpoints"
	ComponentPathLength ComponentName = "path_length"
	ComponentDuration   ComponentName = "duration"
)

type Component struct {
	Name          ComponentName
	Score         float64
	Weight        float64
	ObservedValue float64
	Unit          string
}

type Notice struct {
	Code    string
	Message string
}

type ScoringPolicy struct {
	GeometryScoreScaleKM float64
	EndpointScoreScaleKM float64

	GeometryWeight   float64
	EndpointsWeight  float64
	PathLengthWeight float64
	DurationWeight   float64
}

type EvidenceQuality struct {
	Score float64

	DeclaredQualityScore     float64
	SegmentQualityScore      float64
	CoverageContinuityScore  float64
	ObservationCadenceScore  float64
	PointRetentionScore      float64
	InputPointCount          int
	UsablePointCount         int
	ExcludedPointCount       int
	EqualTimestampPointCount int
	CoverageGapCount         int
	RelevantSegmentCount     int
	NonObservedSegmentCount  int
	InvalidSegmentCount      int
	SourceName               string
}

type EvidenceConfidence struct {
	Score     float64
	Level     ConfidenceLevel
	Reference EvidenceQuality
	Candidate EvidenceQuality
	Reasons   []Notice
}

type Result struct {
	Version string

	ReferenceTrajectoryID string
	CandidateTrajectoryID string

	// Score and Level represent route-shape similarity only. They are not
	// confidence values. Confidence is published separately below.
	Score float64
	Level Level

	Confidence EvidenceConfidence
	Policy     ScoringPolicy

	ReferencePointCount int
	CandidatePointCount int
	SampleCount         int

	MeanDistanceKM           float64
	MaximumDistanceKM        float64
	StartEndpointDistanceKM  float64
	EndEndpointDistanceKM    float64
	ReferencePathLengthKM    float64
	CandidatePathLengthKM    float64
	ReferenceDurationSeconds float64
	CandidateDurationSeconds float64

	Components  []Component
	Reasons     []string
	Limitations []Notice

	InputFingerprint string
}

func (result Result) Clone() Result {
	cloned := result
	cloned.Components = append(
		[]Component(nil),
		result.Components...,
	)
	cloned.Reasons = append(
		[]string(nil),
		result.Reasons...,
	)
	cloned.Limitations = append(
		[]Notice(nil),
		result.Limitations...,
	)
	cloned.Confidence.Reasons = append(
		[]Notice(nil),
		result.Confidence.Reasons...,
	)
	return cloned
}

var fingerprintPattern = regexp.MustCompile(
	`^sha256:[0-9a-f]{64}$`,
)

func LevelForScore(score float64) Level {
	switch {
	case score >= 0.8:
		return LevelHigh
	case score >= 0.6:
		return LevelMedium
	case score > 0:
		return LevelLow
	default:
		return LevelNone
	}
}

func ConfidenceLevelForScore(
	score float64,
) ConfidenceLevel {
	switch {
	case score >= 0.8:
		return ConfidenceLevelHigh
	case score >= 0.6:
		return ConfidenceLevelMedium
	case score > 0:
		return ConfidenceLevelLow
	default:
		return ConfidenceLevelNone
	}
}

func ratio(value float64) bool {
	return finite(value) &&
		value >= 0 &&
		value <= 1
}

func finite(value float64) bool {
	return !math.IsNaN(value) &&
		!math.IsInf(value, 0)
}
