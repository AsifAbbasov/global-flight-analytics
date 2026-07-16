package airspaceregionanalytics

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/localtrafficscene"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/proximityscanner"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/separationrisk"
)

const Version = "airspace-region-analytics-v1"

type SchemaVersion string

const SchemaVersionV1 SchemaVersion = "airspace-region-analytics-v1"

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

type ComplexityLevel string

const (
	ComplexityLevelNone     ComplexityLevel = "none"
	ComplexityLevelLow      ComplexityLevel = "low"
	ComplexityLevelModerate ComplexityLevel = "moderate"
	ComplexityLevelHigh     ComplexityLevel = "high"
	ComplexityLevelSevere   ComplexityLevel = "severe"
)

func (level ComplexityLevel) IsKnown() bool {
	switch level {
	case ComplexityLevelNone,
		ComplexityLevelLow,
		ComplexityLevelModerate,
		ComplexityLevelHigh,
		ComplexityLevelSevere:
		return true
	default:
		return false
	}
}

type OccupancyTrend string

const (
	OccupancyTrendUnavailable OccupancyTrend = "unavailable"
	OccupancyTrendFalling     OccupancyTrend = "falling"
	OccupancyTrendStable      OccupancyTrend = "stable"
	OccupancyTrendRising      OccupancyTrend = "rising"
)

func (trend OccupancyTrend) IsKnown() bool {
	switch trend {
	case OccupancyTrendUnavailable,
		OccupancyTrendFalling,
		OccupancyTrendStable,
		OccupancyTrendRising:
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

const ScopeGuardResearchOnly ScopeGuard = "research_only_not_for_operational_separation_or_air_traffic_control_use"

type SnapshotInput struct {
	Scene localtrafficscene.Result
	Scan  proximityscanner.Result
	Risk  separationrisk.Result
}

type Request struct {
	RegionCode  string
	WindowStart time.Time
	WindowEnd   time.Time
	GeneratedAt time.Time
	Snapshots   []SnapshotInput
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

type OccupancyCell struct {
	ID                string
	BucketID          string
	BucketStart       time.Time
	BucketEnd         time.Time
	LatitudeIndex     int
	LongitudeIndex    int
	AltitudeBandIndex int
	AltitudeKnown     bool
	AircraftNodeIDs   []string
	AircraftCount     int
	MeanQualityScore  float64
}

type OccupancyBucketMetrics struct {
	AircraftCount        int
	OccupiedCellCount    int
	UnknownAltitudeCount int
	MeanQualityScore     float64
}

type OccupancyBucket struct {
	ID        string
	StartTime time.Time
	EndTime   time.Time
	Cells     []OccupancyCell
	Metrics   OccupancyBucketMetrics
}

type OccupancyIndexMetrics struct {
	BucketCount              int
	ExpectedBucketCount      int
	OccupiedCellCount        int
	AircraftObservationCount int
	UniqueAircraftCount      int
	UnknownAltitudeCount     int
	PeakAircraftPerBucket    int
	PeakOccupiedCells        int
	MeanAircraftPerBucket    float64
	TemporalCoverage         float64
}

type TemporalOccupancyIndex struct {
	BucketDuration       time.Duration
	LatitudeCellDegrees  float64
	LongitudeCellDegrees float64
	AltitudeBandMeters   float64
	Buckets              []OccupancyBucket
	Metrics              OccupancyIndexMetrics
}

type SectorComplexityReport struct {
	ID                     string
	BucketID               string
	BucketStart            time.Time
	BucketEnd              time.Time
	LatitudeIndex          int
	LongitudeIndex         int
	AircraftNodeIDs        []string
	AircraftCount          int
	AltitudeBandCount      int
	UnknownAltitudeCount   int
	CandidatePairCount     int
	ConvergingPairCount    int
	ContextualRiskCount    int
	ElevatedRiskCount      int
	HighRiskCount          int
	IndeterminateRiskCount int
	HeadingDispersion      float64
	SpeedVariability       float64
	Score                  float64
	Level                  ComplexityLevel
	Components             []ScoreComponent
	Confidence             Confidence
	Limitations            []Limitation
	Explanations           []Explanation
}

type RegionMetrics struct {
	SnapshotCount             int
	BucketCount               int
	UniqueAircraftCount       int
	AircraftObservationCount  int
	OccupiedCellCount         int
	SectorReportCount         int
	CurrentAircraftCount      int
	PeakAircraftPerBucket     int
	MeanAircraftPerBucket     float64
	MeanComplexityScore       float64
	PeakComplexityScore       float64
	AirspacePressureIndex     float64
	PeakAirspacePressureIndex float64
	ModerateSectorCount       int
	HighSectorCount           int
	SevereSectorCount         int
	ContextualRiskCount       int
	ElevatedRiskCount         int
	HighRiskCount             int
	IndeterminateRiskCount    int
	UnknownAltitudeCount      int
	TemporalCoverage          float64
	OccupancyTrend            OccupancyTrend
	HighestComplexityLevel    ComplexityLevel
}

type Provenance struct {
	InputFingerprint  string
	SceneFingerprints []string
	ScanFingerprints  []string
	RiskFingerprints  []string
	SourceNames       []string
	LatestObservedAt  time.Time
}

type Result struct {
	SchemaVersion    SchemaVersion
	Status           ResultStatus
	RegionCode       string
	WindowStart      time.Time
	WindowEnd        time.Time
	Occupancy        TemporalOccupancyIndex
	SectorComplexity []SectorComplexityReport
	Metrics          RegionMetrics
	Confidence       Confidence
	Limitations      []Limitation
	Explanations     []Explanation
	ScopeGuard       ScopeGuard
	Provenance       Provenance
	GeneratedAt      time.Time
}

func (result Result) Clone() Result {
	cloned := result
	cloned.Occupancy = result.Occupancy.Clone()
	cloned.SectorComplexity = make([]SectorComplexityReport, 0, len(result.SectorComplexity))
	for _, report := range result.SectorComplexity {
		cloned.SectorComplexity = append(cloned.SectorComplexity, report.Clone())
	}
	cloned.Confidence = result.Confidence.Clone()
	cloned.Limitations = append([]Limitation(nil), result.Limitations...)
	cloned.Explanations = append([]Explanation(nil), result.Explanations...)
	cloned.Provenance.SceneFingerprints = append([]string(nil), result.Provenance.SceneFingerprints...)
	cloned.Provenance.ScanFingerprints = append([]string(nil), result.Provenance.ScanFingerprints...)
	cloned.Provenance.RiskFingerprints = append([]string(nil), result.Provenance.RiskFingerprints...)
	cloned.Provenance.SourceNames = append([]string(nil), result.Provenance.SourceNames...)
	return cloned
}

func (index TemporalOccupancyIndex) Clone() TemporalOccupancyIndex {
	cloned := index
	cloned.Buckets = make([]OccupancyBucket, 0, len(index.Buckets))
	for _, bucket := range index.Buckets {
		cloned.Buckets = append(cloned.Buckets, bucket.Clone())
	}
	return cloned
}

func (bucket OccupancyBucket) Clone() OccupancyBucket {
	cloned := bucket
	cloned.Cells = make([]OccupancyCell, 0, len(bucket.Cells))
	for _, cell := range bucket.Cells {
		clonedCell := cell
		clonedCell.AircraftNodeIDs = append([]string(nil), cell.AircraftNodeIDs...)
		cloned.Cells = append(cloned.Cells, clonedCell)
	}
	return cloned
}

func (report SectorComplexityReport) Clone() SectorComplexityReport {
	cloned := report
	cloned.AircraftNodeIDs = append([]string(nil), report.AircraftNodeIDs...)
	cloned.Components = append([]ScoreComponent(nil), report.Components...)
	cloned.Confidence = report.Confidence.Clone()
	cloned.Limitations = append([]Limitation(nil), report.Limitations...)
	cloned.Explanations = append([]Explanation(nil), report.Explanations...)
	return cloned
}

func (confidence Confidence) Clone() Confidence {
	cloned := confidence
	cloned.Components = append([]ScoreComponent(nil), confidence.Components...)
	cloned.Reasons = append([]ConfidenceReason(nil), confidence.Reasons...)
	return cloned
}
